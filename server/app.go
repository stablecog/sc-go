package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/go-co-op/gocron"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	chiprometheus "github.com/stablecog/chi-prometheus"
	"github.com/stablecog/sc-go/database"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/ent/generationoutput"
	"github.com/stablecog/sc-go/database/qdrant"
	"github.com/stablecog/sc-go/database/repository"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/server/analytics"
	"github.com/stablecog/sc-go/server/api/rest"
	"github.com/stablecog/sc-go/server/api/sse"
	"github.com/stablecog/sc-go/server/clip"
	"github.com/stablecog/sc-go/server/discord"
	"github.com/stablecog/sc-go/server/middleware"
	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/server/responses"
	"github.com/stablecog/sc-go/shared"
	"github.com/stablecog/sc-go/utils"
	stripe "github.com/stripe/stripe-go/v74/client"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

var Version = "dev"
var CommitMsg = "dev"

// Used to track the build time from our CI
var BuildStart = ""

func main() {
	log.Infof("SC Server: %s", Version)

	// Load .env
	err := godotenv.Load("../.env")
	if err != nil {
		log.Warn("Error loading .env file (this is fine)", "err", err)
	}

	// Custom flags
	createMockData := flag.Bool("load-mock-data", false, "Create test data in database")
	loadQdrant := flag.Bool("load-qdrant", false, "Load qdrant data")
	cursorEmbeddings := flag.String("cursor-embeddings", "", "Cursor for loading embeddings")

	flag.Parse()

	// Setup database
	ctx := context.Background()

	// Setup sql
	log.Info("🏡 Connecting to database...")
	dbconn, err := database.GetSqlDbConn(false)
	if err != nil {
		log.Fatal("Failed to connect to database", "err", err)
		os.Exit(1)
	}
	entClient, err := database.NewEntClient(dbconn)
	if err != nil {
		log.Fatal("Failed to create ent client", "err", err)
		os.Exit(1)
	}
	defer entClient.Close()
	// Run migrations
	// We can't run on supabase, :(
	if utils.GetEnv("RUN_MIGRATIONS", "") == "true" {
		log.Info("🦋 Running migrations...")
		if err := entClient.Schema.Create(ctx); err != nil {
			log.Fatal("Failed to run migrations", "err", err)
			os.Exit(1)
		}
	}

	// Setup redis
	redis, err := database.NewRedis(ctx)
	if err != nil {
		log.Fatal("Error connecting to redis", "err", err)
		os.Exit(1)
	}

	// Setup qdrant
	qdrantClient, err := qdrant.NewQdrantClient(ctx)
	if err != nil {
		log.Fatal("Error connecting to qdrant", "err", err)
		os.Exit(1)
	}
	err = qdrantClient.CreateCollectionIfNotExists(false)
	if err != nil {
		log.Fatal("Error creating qdrant collection", "err", err)
		os.Exit(1)
	}

	// Create indexes in Qdrant
	err = qdrantClient.CreateAllIndexes()
	if err != nil {
		log.Warn("Error creating qdrant indexes", "err", err)
	}

	// Create repository (database access)
	repo := &repository.Repository{
		DB:       entClient,
		ConnInfo: dbconn,
		Redis:    redis,
		Ctx:      ctx,
		QDrant:   qdrantClient,
	}

	if *loadQdrant {
		log.Info("🏡 Loading qdrant data...")
		secret := os.Getenv("CLIPAPI_SECRET")
		urlStr := os.Getenv("CLIPAPI_URLS")
		urls := strings.Split(urlStr, ",")
		each := 100
		cur := 0
		urlIdx := 0
		var cursor *time.Time
		if *cursorEmbeddings != "" {
			t, err := time.Parse(time.RFC3339, *cursorEmbeddings)
			if err != nil {
				log.Fatal("Failed to parse cursor", "err", err)
			}
			cursor = &t
		}

		promptEmbeddings := make(map[string][]float32)

		for {
			if urlIdx >= len(urls) {
				urlIdx = 0
			}
			log.Info("Loading batch of embeddings", "cur", cur, "each", each)
			start := time.Now()
			q := repo.DB.GenerationOutput.Query().Where(generationoutput.HasEmbeddings(false))
			if cursor != nil {
				q = q.Where(generationoutput.CreatedAtLT(*cursor))
			}
			gens, err := q.Order(ent.Desc(generationoutput.FieldCreatedAt)).WithGenerations(func(gq *ent.GenerationQuery) {
				gq.WithPrompt()
				gq.WithNegativePrompt()
			}).Limit(each).All(ctx)
			if err != nil {
				if cursor != nil {
					log.Info("Last cursor", "cursor", cursor.Format(time.RFC3339Nano))
				}
				log.Fatal("Failed to load generation outputs", "err", err)
			}
			log.Infof("Retreived generations in %s", time.Since(start))

			// Update cursor
			cursor = &gens[len(gens)-1].CreatedAt

			if len(gens) == 0 {
				break
			}

			ids := make([]uuid.UUID, len(gens))
			var clipReq []requests.ClipAPIImageRequest
			promptMap := make(map[uuid.UUID]string)
			for i, gen := range gens {
				ids[i] = gen.ID
				clipReq = append(clipReq, requests.ClipAPIImageRequest{
					ID:      gen.ID,
					ImageID: gen.ImagePath,
				})
				if _, ok := promptEmbeddings[gen.Edges.Generations.Edges.Prompt.Text]; !ok {
					promptMap[gen.GenerationID] = gen.Edges.Generations.Edges.Prompt.Text
				}
			}
			for k, gen := range promptMap {
				clipReq = append(clipReq, requests.ClipAPIImageRequest{
					ID:   k,
					Text: gen,
				})
			}

			// Make API request to clip
			start = time.Now()
			b, err := json.Marshal(clipReq)
			if err != nil {
				log.Infof("Last cursor: %v", cursor.Format(time.RFC3339Nano))
				log.Fatalf("Error marshalling req %v", err)
			}
			request, _ := http.NewRequest(http.MethodPost, urls[urlIdx], bytes.NewReader(b))
			urlIdx++
			request.Header.Set("Authorization", secret)
			request.Header.Set("Content-Type", "application/json")
			// Do
			resp, err := http.DefaultClient.Do(request)
			if err != nil {
				log.Infof("Last cursor: %v", cursor.Format(time.RFC3339Nano))
				log.Warnf("Error making request %v", err)
				time.Sleep(30 * time.Second)
				continue
			}
			defer resp.Body.Close()

			readAll, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Infof("Last cursor: %v", cursor.Format(time.RFC3339Nano))
				log.Fatal(err)
			}
			var clipAPIResponse responses.EmbeddingsResponse
			err = json.Unmarshal(readAll, &clipAPIResponse)
			if err != nil {
				log.Infof("Last cursor: %v", cursor.Format(time.RFC3339Nano))
				log.Fatalf("Error unmarshalling resp %v", err)
				return
			}

			// Builds maps of embeddings
			embeddings := make(map[uuid.UUID][]float32)
			for _, embedding := range clipAPIResponse.Embeddings {
				if embedding.Error != "" {
					log.Warn("Error from clip api", "err", embedding.Error)
					continue
				}
				embeddings[embedding.ID] = embedding.Embedding
			}

			log.Infof("Retreived embeddings in %s", time.Since(start))

			// Build payloads for qdrant
			var payloads []map[string]interface{}

			start = time.Now()
			for _, gOutput := range gens {
				payload := map[string]interface{}{
					"image_path":      gOutput.ImagePath,
					"gallery_status":  gOutput.GalleryStatus,
					"is_favorited":    gOutput.IsFavorited,
					"created_at":      gOutput.CreatedAt.Unix(),
					"updated_at":      gOutput.UpdatedAt.Unix(),
					"guidance_scale":  gOutput.Edges.Generations.GuidanceScale,
					"inference_steps": gOutput.Edges.Generations.InferenceSteps,
					"prompt_strength": gOutput.Edges.Generations.PromptStrength,
					"height":          gOutput.Edges.Generations.Height,
					"width":           gOutput.Edges.Generations.Width,
					"model":           gOutput.Edges.Generations.ModelID.String(),
					"scheduler":       gOutput.Edges.Generations.SchedulerID.String(),
					"user_id":         gOutput.Edges.Generations.UserID.String(),
					"prompt":          gOutput.Edges.Generations.Edges.Prompt.Text,
				}
				var ok bool
				payload["embedding"], ok = embeddings[gOutput.ID]
				if !ok {
					log.Warn("Missing embedding", "id", gOutput.ID)
					continue
				}
				payload["text_embedding"], ok = embeddings[gOutput.Edges.Generations.ID]
				if !ok {
					payload["text_embedding"], ok = promptEmbeddings[gOutput.Edges.Generations.Edges.Prompt.Text]
					if !ok {
						log.Warn("Missing text embedding", "id", gOutput.Edges.Generations.ID)
						continue
					}
				} else {
					promptEmbeddings[gOutput.Edges.Generations.Edges.Prompt.Text] = embeddings[gOutput.Edges.Generations.ID]
				}
				payload["id"] = gOutput.ID.String()
				if gOutput.UpscaledImagePath != nil {
					payload["upscaled_image_path"] = *gOutput.UpscaledImagePath
				}
				if gOutput.Edges.Generations.InitImageURL != nil {
					payload["init_image_url"] = *gOutput.Edges.Generations.InitImageURL
				}
				if gOutput.Edges.Generations.Edges.NegativePrompt != nil {
					payload["negative_prompt"] = gOutput.Edges.Generations.Edges.NegativePrompt.Text
				}
				payloads = append(payloads, payload)
			}

			// QD Upsert
			err = qdrantClient.BatchUpsert(payloads, false)
			if err != nil {
				log.Info("Last cursor", "cursor", cursor.Format(time.RFC3339Nano))
				log.Warn("Failed to batch objects", "err", err)
				continue
			}

			log.Infof("Batched objects for qdrant in %s", time.Since(start))

			err = repo.DB.GenerationOutput.Update().Where(generationoutput.IDIn(ids...)).SetHasEmbeddings(true).Exec(ctx)
			if err != nil {
				log.Info("Last cursor", "cursor", cursor.Format(time.RFC3339Nano))
				log.Fatal("Failed to update generation outputs", "err", err)
			}
			log.Info("Batched objects", "count", len(payloads))
			cur += len(payloads)

			// Log cursor
			log.Info("Last cursor", "cursor", cursor.Format(time.RFC3339Nano))
		}

		log.Info("Loaded generation outputs", "count", cur)
		os.Exit(0)
	}

	if *createMockData {
		log.Info("🏡 Creating mock data...")
		err = repo.CreateMockData(ctx)
		if err != nil {
			log.Fatal("Failed to create mock data", "err", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// Create stripe client
	stripeClient := stripe.New(utils.GetEnv("STRIPE_SECRET_KEY", ""), nil)

	app := chi.NewRouter()

	// Prometheus middleware
	promMiddleware := chiprometheus.NewMiddleware("sc-server")
	app.Use(promMiddleware)

	// Cors middleware
	app.Use(cors.Handler(cors.Options{
		AllowedOrigins: utils.GetCorsOrigins(),
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	// Get models, schedulers and put in cache
	log.Info("📦 Populating cache...")
	err = repo.UpdateCache()
	if err != nil {
		// ! Not getting these is fatal and will result in crash
		panic(err)
	}
	// Update periodically
	s := gocron.NewScheduler(time.UTC)
	s.Every(5).Minutes().StartAt(time.Now().Add(5 * time.Minute)).Do(func() {
		log.Info("📦 Updating cache...")
		err = repo.UpdateCache()
		if err != nil {
			log.Error("Error updating cache", "err", err)
		}
	})

	// Create SSE hub
	sseHub := sse.NewHub(redis, repo)
	go sseHub.Run()

	// Need to send keepalive every 30 seconds
	s.Every(30).Seconds().StartAt(time.Now().Add(30 * time.Second)).Do(func() {
		sseHub.BraodcastKeepalive()
	})

	// Start cron scheduler
	s.StartAsync()

	// Create analytics service
	analyticsService := analytics.NewAnalyticsService()
	defer analyticsService.Close()

	// Q Throttler
	qThrottler := shared.NewQueueThrottler(shared.REQUEST_COG_TIMEOUT)

	// Setup S3 Client
	region := os.Getenv("S3_IMG2IMG_REGION")
	accessKey := os.Getenv("S3_IMG2IMG_ACCESS_KEY")
	secretKey := os.Getenv("S3_IMG2IMG_SECRET_KEY")

	s3Config := &aws.Config{
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
		Endpoint:    aws.String(os.Getenv("S3_IMG2IMG_ENDPOINT")),
		Region:      aws.String(region),
	}

	newSession := session.New(s3Config)
	s3Client := s3.New(newSession)

	// Create controller
	hc := rest.RestAPI{
		Repo:           repo,
		Redis:          redis,
		Hub:            sseHub,
		StripeClient:   stripeClient,
		Meili:          database.NewMeiliSearchClient(),
		Track:          analyticsService,
		QueueThrottler: qThrottler,
		S3:             s3Client,
		QDrant:         qdrantClient,
		Clip:           clip.NewClipService(),
	}

	// Create middleware
	mw := middleware.Middleware{
		SupabaseAuth: database.NewSupabaseAuth(),
		Repo:         repo,
		Redis:        redis,
	}

	// Routes
	app.Get("/", hc.HandleHealth)
	app.Handle("/metrics", middleware.BasicAuth(promhttp.Handler(), "user", "password", "Authentication required"))
	app.Get("/clipq", hc.HandleClipQSearch)
	app.Route("/v1", func(r chi.Router) {
		r.Get("/health", hc.HandleHealth)

		// SSE
		r.Route("/sse", func(r chi.Router) {
			r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				sseHub.ServeSSE(w, r)
			})
		})

		// Stripe
		r.Route("/stripe", func(r chi.Router) {
			r.Post("/webhook", hc.HandleStripeWebhook)
		})

		// SCWorker
		r.Route("/worker", func(r chi.Router) {
			r.Post("/webhook", hc.HandleSCWorkerWebhook)
		})

		// Stats
		r.Route("/stats", func(r chi.Router) {
			r.Use(middleware.Logger)
			// 10 requests per second
			r.Use(mw.RateLimit(10, 1*time.Second))
			r.Get("/", hc.HandleGetStats)
		})

		// Gallery search
		r.Route("/gallery", func(r chi.Router) {
			r.Use(middleware.Logger)
			// 20 requests per second
			r.Use(mw.RateLimit(20, 1*time.Second))
			r.Get("/", hc.HandleQueryGallery)
			r.Get("/semantic", hc.HandleSemanticSearchGallery)
		})

		// Routes that require authentication
		r.Route("/user", func(r chi.Router) {
			r.Use(mw.AuthMiddleware(middleware.AuthLevelAny))
			r.Use(middleware.Logger)
			// 10 requests per second
			r.Use(mw.RateLimit(10, 1*time.Second))

			// Get user summary
			r.Get("/", hc.HandleGetUser)

			// Create Generation
			r.Post("/generation", hc.HandleCreateGeneration)
			// Mark generation for deletion
			r.Delete("/generation", hc.HandleDeleteGenerationOutputForUser)

			// Query Generation (outputs + generations)
			r.Get("/outputs", hc.HandleQueryGenerations)

			// Favorite
			r.Post("/outputs/favorite", hc.HandleFavoriteGenerationOutputsForUser)

			// Create upscale
			r.Post("/upscale", hc.HandleUpscale)

			// Query credits
			r.Get("/credits", hc.HandleQueryCredits)

			// Submit to gallery
			r.Put("/gallery", hc.HandleSubmitGenerationToGallery)

			// Subscriptions
			r.Post("/subscription/downgrade", hc.HandleSubscriptionDowngrade)
			r.Post("/subscription/checkout", hc.HandleCreateCheckoutSession)
			r.Post("/subscription/portal", hc.HandleCreatePortalSession)
		})

		// Admin only routes
		r.Route("/admin", func(r chi.Router) {
			r.Route("/gallery", func(r chi.Router) {
				r.Use(mw.AuthMiddleware(middleware.AuthLevelGalleryAdmin))
				r.Use(middleware.Logger)
				r.Put("/", hc.HandleReviewGallerySubmission)
			})
			r.Route("/outputs", func(r chi.Router) {
				// TODO - this is auth level gallery admin, but delete route manually enforces super admin
				r.Use(mw.AuthMiddleware(middleware.AuthLevelGalleryAdmin))
				r.Use(middleware.Logger)
				r.Delete("/", hc.HandleDeleteGenerationOutput)
				r.Get("/", hc.HandleQueryGenerationsForAdmin)
			})
			r.Route("/users", func(r chi.Router) {
				r.Use(mw.AuthMiddleware(middleware.AuthLevelSuperAdmin))
				r.Use(middleware.Logger)
				r.Get("/", hc.HandleQueryUsers)
			})
			r.Route("/credit", func(r chi.Router) {
				r.Use(mw.AuthMiddleware(middleware.AuthLevelSuperAdmin))
				r.Use(middleware.Logger)
				r.Get("/types", hc.HandleQueryCreditTypes)
				r.Post("/add", hc.HandleAddCreditsToUser)
			})
		})
	})

	// This redis subscription has the following purpose:
	// When a generation/upscale is finished - we want to un-throttle them from the queue across all instances
	pubsubThrottleMessages := redis.Client.Subscribe(ctx, shared.REDIS_QUEUE_THROTTLE_CHANNEL)
	defer pubsubThrottleMessages.Close()

	// Start SSE redis subscription
	go func() {
		log.Info("Listening for throttle messages", "channel", shared.REDIS_QUEUE_THROTTLE_CHANNEL)
		for msg := range pubsubThrottleMessages.Channel() {
			var unthrottleMsg repository.UnthrottleUserResponse
			err := json.Unmarshal([]byte(msg.Payload), &unthrottleMsg)
			if err != nil {
				log.Error("Error unmarshalling unthrottle message", "err", err)
				continue
			}
			qThrottler.Decrement(unthrottleMsg.RequestID, unthrottleMsg.UserID)
		}
	}()

	// This redis subscription has the following purpose:
	// After we are done processing a cog message, we want to broadcast it to
	// our subscribed SSE clients matching that stream ID
	// the purpose of this instead of just directly sending the message to the SSE is that
	// our service can scale, and we may have many instances running and we care about SSE connections
	// on all of them.
	pubsubSSEMessages := redis.Client.Subscribe(ctx, shared.REDIS_SSE_BROADCAST_CHANNEL)
	defer pubsubSSEMessages.Close()

	// Start SSE redis subscription
	go func() {
		log.Info("Listening for cog messages", "channel", shared.REDIS_SSE_BROADCAST_CHANNEL)
		for msg := range pubsubSSEMessages.Channel() {
			var sseMessage repository.TaskStatusUpdateResponse
			err := json.Unmarshal([]byte(msg.Payload), &sseMessage)
			if err != nil {
				log.Error("Error unmarshalling sse message", "err", err)
				continue
			}

			// Live page separate broadcast stream
			if sseMessage.ForLivePage {
				sseHub.BroadcastLivePageMessage(*sseMessage.LivePageMessage)
				continue
			}

			// Sanitize
			sseMessage.LivePageMessage = nil
			// The hub will broadcast this to our clients if it's supposed to
			sseHub.BroadcastStatusUpdate(sseMessage)
		}
	}()

	// Start server
	port := utils.GetEnv("PORT", "13337")
	log.Info("Starting server", "port", port)

	h2s := &http2.Server{}
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: h2c.NewHandler(app, h2s),
	}

	// Send discord notification
	go func() {
		err = discord.FireServerReadyWebhook(Version, CommitMsg, BuildStart)
		if err != nil {
			log.Error("Error firing discord ready webhook", "err", err)
		}
	}()
	log.Info(srv.ListenAndServe())
}
