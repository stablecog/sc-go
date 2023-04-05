package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
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
	"github.com/stablecog/sc-go/database/ent/generation"
	"github.com/stablecog/sc-go/database/ent/generationoutput"
	"github.com/stablecog/sc-go/database/qdrant"
	"github.com/stablecog/sc-go/database/repository"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/server/analytics"
	"github.com/stablecog/sc-go/server/api/rest"
	"github.com/stablecog/sc-go/server/api/sse"
	"github.com/stablecog/sc-go/server/discord"
	"github.com/stablecog/sc-go/server/middleware"
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
		loaded := 0
		for loaded < 50000 {
			rawQ := fmt.Sprintf("select id, image_path, upscaled_image_path, gallery_status, embedding, is_favorited, generation_id, created_at, updated_at from generation_outputs where embedding is not null and has_embedding = false order by created_at desc limit 100 offset %d;", loaded)
			res, err := repo.DB.QueryContext(ctx, rawQ)
			if err != nil {
				log.Fatal("Failed to load generation outputs", "err", err)
			}
			// Load res into struct
			type GenerationOutput struct {
				ID                 string  `json:"id"`
				ImagePath          string  `json:"image_path"`
				UpscaledImagePath  *string `json:"upscaled_image_path,omitempty"`
				GalleryStatus      string  `json:"gallery_status"`
				Embedding          string  `json:"embedding"`
				IsFavorited        bool    `json:"is_favorited"`
				EmbeddingVector    []float32
				GenerationID       string    `json:"generation_id"`
				GenerationUUID     uuid.UUID `json:"generation_uuid"`
				PromptText         string    `json:"prompt_text"`
				NegativePromptText *string   `json:"negative_prompt_text"`
				Generation         *ent.Generation
				CreatedAt          time.Time `json:"created_at"`
				UpdatedAt          time.Time `json:"updated_at"`
			}
			var generationOutputs []GenerationOutput
			for res.Next() {
				var generationOutput GenerationOutput
				err = res.Scan(&generationOutput.ID, &generationOutput.ImagePath, &generationOutput.UpscaledImagePath, &generationOutput.GalleryStatus, &generationOutput.Embedding, &generationOutput.IsFavorited, &generationOutput.GenerationID, &generationOutput.CreatedAt, &generationOutput.UpdatedAt)
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
				g, err := repo.DB.Generation.Query().Where(generation.IDEQ(generationOutput.GenerationUUID)).WithPrompt().WithNegativePrompt().Only(ctx)
				if err != nil {
					log.Fatal("Failed to load generation", "err", err)
				}
				generationOutput.PromptText = g.Edges.Prompt.Text
				if g.Edges.NegativePrompt != nil {
					generationOutput.NegativePromptText = &g.Edges.NegativePrompt.Text
				}
				generationOutput.Generation = g
				generationOutputs = append(generationOutputs, generationOutput)
			}

			// Load to qdrant
			var payloads []map[string]interface{}
			var ids []uuid.UUID
			for _, gOutput := range generationOutputs {
				ids = append(ids, gOutput.GenerationUUID)
				payload := map[string]interface{}{
					"image_path":      gOutput.ImagePath,
					"gallery_status":  gOutput.GalleryStatus,
					"is_favorited":    gOutput.IsFavorited,
					"created_at":      gOutput.CreatedAt.Unix(),
					"updated_at":      gOutput.UpdatedAt.Unix(),
					"guidance_scale":  gOutput.Generation.GuidanceScale,
					"inference_steps": gOutput.Generation.InferenceSteps,
					"prompt_strength": gOutput.Generation.PromptStrength,
					"height":          gOutput.Generation.Height,
					"width":           gOutput.Generation.Width,
					"model":           gOutput.Generation.ModelID.String(),
					"scheduler":       gOutput.Generation.SchedulerID.String(),
					"user_id":         gOutput.Generation.UserID.String(),
					"prompt":          gOutput.PromptText,
				}
				payload["embedding"] = gOutput.EmbeddingVector
				payload["id"] = gOutput.ID
				if gOutput.UpscaledImagePath != nil {
					payload["upscaled_image_path"] = *gOutput.UpscaledImagePath
				}
				if gOutput.Generation.InitImageURL != nil {
					payload["init_image_url"] = *gOutput.Generation.InitImageURL
				}
				if gOutput.NegativePromptText != nil {
					payload["negative_prompt"] = *gOutput.NegativePromptText
				}
				payloads = append(payloads, payload)
			}
			err = qdrantClient.BatchUpsert(payloads, false)
			if err != nil {
				log.Fatal("Failed to batch objects", "err", err)
			}
			err = repo.DB.GenerationOutput.Update().Where(generationoutput.IDIn(ids...)).SetHasEmbeddings(true).Exec(ctx)
			if err != nil {
				log.Fatal("Failed to update generation outputs", "err", err)
			}
			log.Info("Batched objects", "count", len(payloads))
			loaded += len(payloads)
		}

		log.Info("Loaded generation outputs", "count", loaded)
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
