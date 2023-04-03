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
	"strconv"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/go-co-op/gocron"
	"github.com/go-openapi/strfmt"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	chiprometheus "github.com/stablecog/chi-prometheus"
	"github.com/stablecog/sc-go/database"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/ent/generation"
	"github.com/stablecog/sc-go/database/ent/generationoutput"
	"github.com/stablecog/sc-go/database/repository"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/server/analytics"
	"github.com/stablecog/sc-go/server/api/rest"
	"github.com/stablecog/sc-go/server/api/sse"
	"github.com/stablecog/sc-go/server/discord"
	"github.com/stablecog/sc-go/server/middleware"
	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/server/responses"
	"github.com/stablecog/sc-go/shared"
	"github.com/stablecog/sc-go/utils"
	stripe "github.com/stripe/stripe-go/v74/client"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/data/replication"
	"github.com/weaviate/weaviate/entities/models"
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
	loadEmbeddings := flag.Bool("load-embeddings", false, "Load embeddings into database")
	cursorEmbeddings := flag.String("cursor", "", "Cursor to start from")
	loadWeaviate := flag.Bool("load-weaviate", false, "Load weaviate data into database")

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

	// Setup milvus client
	// milvusClient, err := database.NewMilvusClient(ctx)
	// if err != nil {
	// 	log.Error("Failed to create milvus client", "err", err)
	// 	os.Exit(1)
	// }
	// defer milvusClient.Close()

	// // Create milvus schema
	// err = milvusClient.CreateCollectionIfNotExists()
	// if err != nil {
	// 	log.Error("Failed to create milvus collection", "err", err)
	// 	os.Exit(1)
	// }

	// // Create milvus index
	// err = milvusClient.CreateIndexes()
	// if err != nil {
	// 	log.Error("Failed to create milvus indexes", "err", err)
	// 	os.Exit(1)
	// }

	// // Load milvus collection
	// err = milvusClient.LoadCollection()
	// if err != nil {
	// 	log.Error("Failed to load milvus collection", "err", err)
	// 	os.Exit(1)
	// }

	// Setup weaviate
	weaviate := database.NewWeaviateClient(ctx)
	if err := weaviate.CreateSchema(); err != nil {
		log.Warn("Failed to create weaviate schema", "err", err)
	}

	// Create repository (database access)
	repo := &repository.Repository{
		DB:       entClient,
		ConnInfo: dbconn,
		Redis:    redis,
		Ctx:      ctx,
		// Milvus:   milvusClient,
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

	if *loadWeaviate {
		log.Info("🏡 Loading weaviate data...")
		rawQ := "select id, image_path, embedding, generation_id, created_at from generation_outputs where embedding is not null order by created_at desc limit 50;"
		res, err := repo.DB.QueryContext(ctx, rawQ)
		if err != nil {
			log.Fatal("Failed to load generation outputs", "err", err)
		}
		// Load res into struct
		type GenerationOutput struct {
			ID              string `json:"id"`
			ImagePath       string `json:"image_path"`
			Embedding       string `json:"embedding"`
			EmbeddingVector []float32
			GenerationID    string    `json:"generation_id"`
			GenerationUUID  uuid.UUID `json:"generation_uuid"`
			PromptText      string    `json:"prompt_text"`
			CreatedAt       time.Time `json:"created_at"`
		}
		var generationOutputs []GenerationOutput
		for res.Next() {
			var generationOutput GenerationOutput
			err = res.Scan(&generationOutput.ID, &generationOutput.ImagePath, &generationOutput.Embedding, &generationOutput.GenerationID, &generationOutput.CreatedAt)
			if err != nil {
				log.Fatal("Failed to scan generation output", "err", err)
			}
			trimmed := strings.Trim(generationOutput.Embedding, "[]")
			split := strings.Split(trimmed, ",")
			floats := make([]float32, len(split))
			for i, s := range split {
				f, err := strconv.ParseFloat(s, 32)
				if err != nil {
					log.Fatal("Failed to parse float", "err", err)
				}
				floats[i] = float32(f)
			}
			generationOutput.EmbeddingVector = floats
			generationOutput.GenerationUUID = uuid.MustParse(generationOutput.GenerationID)
			g, err := repo.DB.Generation.Query().Where(generation.IDEQ(generationOutput.GenerationUUID)).WithPrompt().Only(ctx)
			if err != nil {
				log.Fatal("Failed to load generation", "err", err)
			}
			generationOutput.PromptText = g.Edges.Prompt.Text
			generationOutputs = append(generationOutputs, generationOutput)
		}

		// Load to weaviate
		var objects []*models.Object
		for _, generationOutput := range generationOutputs {
			objects = append(objects, &models.Object{
				Class: "Test",
				ID:    strfmt.UUID(generationOutput.ID),
				Properties: map[string]interface{}{
					"image_path": generationOutput.ImagePath,
					"prompt":     generationOutput.PromptText,
				},
				Vector: generationOutput.EmbeddingVector,
			})
		}
		b, err := weaviate.Client.Batch().ObjectsBatcher().WithConsistencyLevel(replication.ConsistencyLevel.ALL).WithObjects(objects...).Do(weaviate.Ctx)
		if err != nil {
			log.Fatal("Failed to batch objects", "err", err)
		}
		log.Info("Batched objects", "count", len(b))

		// Repeat where created_at lt last created_at
		cursor := generationOutputs[len(generationOutputs)-1].CreatedAt
		for len(generationOutputs) < 5000000 {
			log.Info("Loading more generation outputs", "cursor", cursor)
			rawQ := "select id, image_path, embedding, generation_id, created_at from generation_outputs where embedding is not null and created_at < $1 order by created_at desc limit 50;"
			res, err := repo.DB.QueryContext(ctx, rawQ, cursor)
			if err != nil {
				log.Fatal("Failed to load generation outputs", "err", err)
			}
			// Load res into struct
			var generationOutputs []GenerationOutput
			for res.Next() {
				var generationOutput GenerationOutput
				err = res.Scan(&generationOutput.ID, &generationOutput.ImagePath, &generationOutput.Embedding, &generationOutput.GenerationID, &generationOutput.CreatedAt)
				if err != nil {
					log.Fatal("Failed to scan generation output", "err", err)
				}
				trimmed := strings.Trim(generationOutput.Embedding, "[]")
				split := strings.Split(trimmed, ",")
				floats := make([]float32, len(split))
				for i, s := range split {
					f, err := strconv.ParseFloat(s, 32)
					if err != nil {
						log.Fatal("Failed to parse float", "err", err)
					}
					floats[i] = float32(f)
				}
				generationOutput.EmbeddingVector = floats
				generationOutput.GenerationUUID = uuid.MustParse(generationOutput.GenerationID)
				g, err := repo.DB.Generation.Query().Where(generation.IDEQ(generationOutput.GenerationUUID)).WithPrompt().Only(ctx)
				if err != nil {
					log.Fatal("Failed to load generation", "err", err)
				}
				generationOutput.PromptText = g.Edges.Prompt.Text
				generationOutputs = append(generationOutputs, generationOutput)
				cursor = generationOutputs[len(generationOutputs)-1].CreatedAt
			}
			// Load to weaviate
			var objects []*models.Object
			for _, generationOutput := range generationOutputs {
				objects = append(objects, &models.Object{
					Class: "Test",
					ID:    strfmt.UUID(generationOutput.ID),
					Properties: map[string]interface{}{
						"image_path": generationOutput.ImagePath,
						"prompt":     generationOutput.PromptText,
					},
					Vector: generationOutput.EmbeddingVector,
				})
			}
			b, err := weaviate.Client.Batch().ObjectsBatcher().WithObjects(objects...).WithConsistencyLevel(replication.ConsistencyLevel.ALL).Do(weaviate.Ctx)
			if err != nil {
				log.Fatal("Failed to batch objects", "err", err)
			}
			log.Info("Batched objects", "count", len(b))
		}

		log.Info("Loaded generation outputs", "count", len(generationOutputs))
		os.Exit(0)
	}

	if *loadEmbeddings {
		log.Info("🏡 Loading embeddings...")
		secret := os.Getenv("CLIPAPI_SECRET")
		endpoint := os.Getenv("CLIPAPI_ENDPOINT")
		max := 500000
		each := 100
		cur := 0
		var cursor *time.Time
		if *cursorEmbeddings != "" {
			t, err := time.Parse(time.RFC3339, *cursorEmbeddings)
			if err != nil {
				log.Fatal("Failed to parse cursor", "err", err)
			}
			cursor = &t
		}
		for cur < max {
			log.Info("Loading batch of embeddings", "cur", cur, "max", max, "each", each)
			cur += each
			// Get N G output IDs
			var gOutputIDs []requests.ClipAPIImageRequest
			start := time.Now()
			bSelect := repo.DB.GenerationOutput.Query().Select(generationoutput.FieldID, generationoutput.FieldCreatedAt, generationoutput.FieldImagePath, generationoutput.FieldGenerationID).Where(func(s *sql.Selector) {
				s.Where(sql.NotNull("embedding"))
			})
			if cursor != nil {
				bSelect = bSelect.Where(generationoutput.CreatedAtLT(*cursor))
			}
			bSelect.WithGenerations(func(b *ent.GenerationQuery) {
				b.WithPrompt()
			})
			gOutputs, err := bSelect.Order(ent.Desc(generationoutput.FieldCreatedAt)).Limit(each).All(ctx)

			if err != nil {
				log.Fatal("Failed to get generation outputs", "err", err)
			}
			if len(gOutputs) == 0 {
				log.Info("No more generation outputs to load")
				os.Exit(0)
			}
			cursor = &gOutputs[len(gOutputs)-1].CreatedAt
			for _, gOutput := range gOutputs {
				gOutputIDs = append(gOutputIDs, requests.ClipAPIImageRequest{
					ID: gOutput.ID,
					// Image:   utils.GetURLFromImagePath(gOutput.ImagePath),
					ImageID: gOutput.ImagePath,
				})
			}
			end := time.Now()
			log.Infof("Got generation outputs in %fs", end.Sub(start).Seconds())

			// Http POST to endpoint with secret
			// Marshal req
			b, err := json.Marshal(gOutputIDs)
			if err != nil {
				log.Infof("Last cursor: %v", cursor.Format(time.RFC3339Nano))
				log.Fatalf("Error marshalling req %v", err)
			}
			request, _ := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(b))
			request.Header.Set("Authorization", secret)
			request.Header.Set("Content-Type", "application/json")
			// Do
			resp, err := http.DefaultClient.Do(request)
			if err != nil {
				log.Infof("Last cursor: %v", cursor.Format(time.RFC3339Nano))
				log.Fatalf("Error making request %v", err)
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

			// Update weaviate
			start = time.Now()
			idEmbeddingMap := make(map[string][]float32)
			for _, embedding := range clipAPIResponse.Embeddings {
				if embedding.Error != "" {
					log.Info("Skipping", "id", embedding.ID, "error", embedding.Error)
					continue
				}
				idEmbeddingMap[embedding.ID.String()] = embedding.Embedding
			}
			// Load to weaviate
			var objects []*models.Object
			for _, generationOutput := range gOutputs {
				embedding, ok := idEmbeddingMap[generationOutput.ID.String()]
				if !ok {
					log.Info("Skipping", "id", generationOutput.ID.String())
					continue
				}
				objects = append(objects, &models.Object{
					Class: "Test",
					ID:    strfmt.UUID(generationOutput.ID.String()),
					Properties: map[string]interface{}{
						"image_path": generationOutput.ImagePath,
						"prompt":     generationOutput.Edges.Generations.Edges.Prompt.Text,
					},
					Vector: embedding,
				})
			}
			batch, err := weaviate.Client.Batch().ObjectsBatcher().WithObjects(objects...).WithConsistencyLevel(replication.ConsistencyLevel.ALL).Do(weaviate.Ctx)
			if err != nil {
				log.Fatal("Failed to batch objects", "err", err)
			}
			log.Info("Batched objects", "count", len(batch))
			end = time.Now()
			log.Infof("Loaded batch in %fs", end.Sub(start).Seconds())
			log.Infof("Last cursor: %v", cursor.Format(time.RFC3339Nano))

			// // Update generation outputs
			// start = time.Now()
			// for _, embedding := range clipAPIResponse.Embeddings {
			// 	if embedding.Error != "" {
			// 		log.Info("Skipping", "id", embedding.ID, "error", embedding.Error)
			// 		continue
			// 	}
			// 	_, err = repo.DB.ExecContext(ctx, "UPDATE generation_outputs SET embedding = $1 WHERE id = $2", pgvector.NewVector(embedding.Embedding), embedding.ID)
			// 	// err = repo.DB.Debug().GenerationOutput.UpdateOneID(embedding.ID).SetEmbedding(pgvector.NewVector(embedding.Embedding)).Exec(ctx)
			// 	if err != nil {
			// 		log.Infof("Last cursor: %v", cursor.Format(time.RFC3339Nano))
			// 		log.Fatal("Failed to update generation output", "err", err)
			// 	}
			// }
			// end = time.Now()
			// log.Infof("Loaded batch in %fs", end.Sub(start).Seconds())
			// log.Infof("Last cursor: %v", cursor.Format(time.RFC3339Nano))
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
		Weaviate:       weaviate,
		// Milvus:         milvusClient,
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
	app.Get("/clip", hc.HandleClipSearch)
	app.Get("/clippg", hc.HandleClipSearchPGVector)
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
