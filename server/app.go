package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-co-op/gocron"
	"github.com/joho/godotenv"
	"github.com/stablecog/sc-go/database"
	"github.com/stablecog/sc-go/database/repository"
	"github.com/stablecog/sc-go/server/api/rest"
	"github.com/stablecog/sc-go/server/api/sse"
	"github.com/stablecog/sc-go/server/middleware"
	"github.com/stablecog/sc-go/server/responses"
	"github.com/stablecog/sc-go/shared"
	"github.com/stablecog/sc-go/utils"
	stripe "github.com/stripe/stripe-go/client"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"k8s.io/klog/v2"
)

func main() {
	// Load .env
	err := godotenv.Load("../.env")
	if err != nil {
		klog.Warningf("Error loading .env file (this is fine): %v", err)
	}

	// Setup logger
	klog.InitFlags(nil)
	flag.Set("logtostderr", "true")
	flag.Set("stderrthreshold", "INFO")
	flag.Set("v", "3")

	flag.Parse()

	// Setup database
	ctx := context.Background()

	// Setup sql
	klog.Infoln("🏡 Connecting to database...")
	dbconn, err := database.GetSqlDbConn(false)
	if err != nil {
		klog.Fatalf("Failed to connect to database: %v", err)
		os.Exit(1)
	}
	entClient, err := database.NewEntClient(dbconn)
	if err != nil {
		klog.Fatalf("Failed to create ent client: %v", err)
		os.Exit(1)
	}
	defer entClient.Close()
	// Run migrations
	// We can't run on supabase, :(
	if utils.GetEnv("RUN_MIGRATIONS", "") == "true" {
		klog.Infoln("🦋 Running migrations...")
		if err := entClient.Schema.Create(ctx); err != nil {
			klog.Fatalf("Failed to run migrations: %v", err)
			os.Exit(1)
		}
	}

	// Setup redis
	redis, err := database.NewRedis(ctx)
	if err != nil {
		klog.Fatalf("Error connecting to redis: %v", err)
		os.Exit(1)
	}

	// Create repository (database access)
	repo := &repository.Repository{
		DB:    entClient,
		Redis: redis,
		Ctx:   ctx,
	}

	app := chi.NewRouter()

	// Cors middleware
	app.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"http://localhost:3000", "http://localhost:5173", "http://localhost:8000"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	// Get models, schedulers and put in cache
	klog.Infof("📦 Updating cache...")
	err = repo.UpdateCache()
	if err != nil {
		// ! Not getting these is fatal and will result in crash
		panic(err)
	}
	// Update periodically
	s := gocron.NewScheduler(time.UTC)
	s.Every(5).Minutes().StartAt(time.Now().Add(5 * time.Minute)).Do(func() {
		klog.Infof("📦 Updating cache...")
		err = repo.UpdateCache()
		if err != nil {
			klog.Errorf("Error updating cache: %v", err)
		}
	})
	// Start cron scheduler
	s.StartAsync()

	// Create SSE hub
	sseHub := sse.NewHub(redis, repo)
	go sseHub.Run()

	// Create stripe client
	stripeClient := stripe.New(utils.GetEnv("STRIPE_SECRET_KEY", ""), nil)

	// Create controller
	hc := rest.RestAPI{
		Repo:         repo,
		Redis:        redis,
		Hub:          sseHub,
		StripeClient: stripeClient,
	}

	// Create middleware
	mw := middleware.Middleware{
		SupabaseAuth: database.NewSupabaseAuth(),
		Repo:         repo,
		Redis:        redis,
	}

	// Routes
	app.Route("/v1", func(r chi.Router) {
		r.Get("/health", hc.HandleHealth)

		// r.Get("/userget", hc.HandleGetUser)

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

		// Routes that require authentication
		r.Route("/user", func(r chi.Router) {
			r.Use(mw.AuthMiddleware)
			r.Use(chimiddleware.Logger)
			r.Use(mw.RateLimit())

			// Create Generation
			r.Post("/generation", hc.HandleCreateGeneration)
			// Mark generation for deletion
			r.Delete("/generation", hc.HandleDeleteGenerationOutputForUser)

			// Query Generation (outputs + generations)
			r.Get("/outputs", hc.HandleQueryGenerations)

			// Create upscale
			r.Post("/upscale", hc.HandleUpscale)

			// Query credits
			r.Get("/credits", hc.HandleQueryCredits)
		})

		// Admin only routes
		r.Route("/admin", func(r chi.Router) {
			r.Route("/gallery", func(r chi.Router) {
				r.Use(mw.AuthMiddleware)
				r.Use(mw.AdminMiddleware)
				r.Use(chimiddleware.Logger)
				r.Put("/", hc.HandleReviewGallerySubmission)
			})
			r.Route("/generation", func(r chi.Router) {
				r.Use(mw.AuthMiddleware)
				r.Use(mw.SuperAdminMiddleware)
				r.Use(chimiddleware.Logger)
				r.Delete("/", hc.HandleDeleteGenerationOutput)
			})
		})
	})

	// This redis subscription has the following purpose:
	// - We receive events from the cog (e.g. started/succeeded/failed)
	// - We update the database with the new status
	// - Our repository will broadcast to another redis channel, the SSE one
	pubsub := redis.Client.Subscribe(ctx, shared.COG_REDIS_EVENT_CHANNEL)
	defer pubsub.Close()

	// Process messages from cog
	go func() {
		klog.Infof("Listening for cog messages on channel: %s", shared.COG_REDIS_EVENT_CHANNEL)
		for msg := range pubsub.Channel() {
			klog.Infof("Received %s message: %s", shared.COG_REDIS_EVENT_CHANNEL, msg.Payload)
			var cogMessage responses.CogStatusUpdate
			err := json.Unmarshal([]byte(msg.Payload), &cogMessage)
			if err != nil {
				klog.Errorf("--- Error unmarshalling webhook message: %v", err)
				continue
			}

			// Process live page message
			go func() {
				// Live page update
				livePageMsg := cogMessage.Input.LivePageData
				var status shared.LivePageStatus
				switch cogMessage.Status {
				case responses.CogProcessing:
					status = shared.LivePageProcessing
				case responses.CogSucceeded:
					status = shared.LivePageSucceeded
				case responses.CogFailed:
					status = shared.LivePageFailed
				default:
					klog.Errorf("Unknown cog status in live page message parsing: %s", cogMessage.Status)
					return
				}

				livePageMsg.Status = status
				now := time.Now()
				if cogMessage.Status == responses.CogProcessing {
					livePageMsg.StartedAt = &now
				}
				if cogMessage.Status == responses.CogSucceeded || cogMessage.Status == responses.CogFailed {
					livePageMsg.CompletedAt = &now
				}
				sseHub.BroadcastLivePageMessage(livePageMsg)
			}()

			// Process in database
			repo.ProcessCogMessage(cogMessage)
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
		klog.Infof("Listening for cog messages on channel: %s", shared.REDIS_SSE_BROADCAST_CHANNEL)
		for msg := range pubsubSSEMessages.Channel() {
			klog.Infof("Received %s message: %s", shared.REDIS_SSE_BROADCAST_CHANNEL, msg.Payload)
			var sseMessage responses.SSEStatusUpdateResponse
			err := json.Unmarshal([]byte(msg.Payload), &sseMessage)
			if err != nil {
				klog.Errorf("--- Error unmarshalling sse message: %v", err)
				continue
			}

			// The hub will broadcast this to our clients if it's supposed to
			sseHub.BroadcastStatusUpdate(sseMessage)
		}
	}()

	// Start server
	port := utils.GetEnv("PORT", "13337")
	klog.Infof("Starting server on port %s", port)

	h2s := &http2.Server{}
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: h2c.NewHandler(app, h2s),
	}
	klog.Infoln(srv.ListenAndServe())
}
