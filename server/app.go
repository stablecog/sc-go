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
	"github.com/stablecog/go-apps/database"
	"github.com/stablecog/go-apps/database/repository"
	"github.com/stablecog/go-apps/server/api/rest"
	"github.com/stablecog/go-apps/server/api/sse"
	"github.com/stablecog/go-apps/server/middleware"
	"github.com/stablecog/go-apps/server/responses"
	"github.com/stablecog/go-apps/shared"
	"github.com/stablecog/go-apps/utils"
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
	loadTiers := flag.Bool("loadTiers", false, "Load subscription tiers into database")
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
		DB:  entClient,
		Ctx: ctx,
	}

	if *loadTiers {
		klog.Infoln("📦 Loading subscription tiers...")
		err = repo.CreateSubscriptionTiers()
		if err != nil {
			klog.Fatalf("Error loading subscription tiers: %v", err)
			os.Exit(1)
		}
		os.Exit(0)
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

	// Create a thread-safe HashMap to store in-flight cog requests
	// We use this to track requests that are in queue, and coordinate responses
	// To appropriate users over SSE
	cogRequestSSEConnMap := shared.NewSyncMap[string]()

	// Create SSE hub
	sseHub := sse.NewHub(cogRequestSSEConnMap, repo)
	go sseHub.Run()

	// Create controller
	hc := rest.RestAPI{
		Repo:                 repo,
		Redis:                redis,
		CogRequestSSEConnMap: cogRequestSSEConnMap,
		Hub:                  sseHub,
		LanguageDetector:     utils.NewLanguageDetector(),
	}

	// Create middleware
	mw := middleware.Middleware{
		SupabaseAuth: database.NewSupabaseAuth(),
		Repo:         repo,
	}

	// Routes
	app.Route("/v1", func(r chi.Router) {
		r.Get("/health", hc.HandleHealth)

		// SSE
		r.Route("/sse", func(r chi.Router) {
			r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				sseHub.ServeSSE(w, r)
			})
		})

		// Routes that require authentication
		r.Route("/user", func(r chi.Router) {
			r.Use(mw.AuthMiddleware)
			r.Use(chimiddleware.Logger)
			r.Route("/generation", func(r chi.Router) {
				r.Post("/create", hc.HandleCreateGeneration)
				r.Get("/query", hc.HandleQueryGenerations)
			})
			r.Route("/upscale", func(r chi.Router) {
				r.Post("/create", hc.HandleUpscale)
			})
			r.Route("/gallery", func(r chi.Router) {
				r.Post("/submit", hc.HandleSubmitGenerationToGallery)
			})
		})

		// Admin only routes
		r.Route("/admin", func(r chi.Router) {
			r.Route("/gallery", func(r chi.Router) {
				r.Use(mw.AuthMiddleware)
				r.Use(mw.AdminMiddleware)
				r.Use(chimiddleware.Logger)
				r.Post("/review", hc.HandleReviewGallerySubmission)
			})
			r.Route("/generation", func(r chi.Router) {
				r.Use(mw.AuthMiddleware)
				r.Use(mw.SuperAdminMiddleware)
				r.Use(chimiddleware.Logger)
				r.Delete("/delete", hc.HandleDeleteGeneration)
			})
		})
	})

	// Subscribe to cog events for generations
	pubsubGenerate := redis.Client.Subscribe(ctx, shared.COG_REDIS_GENERATE_EVENT_CHANNEL)
	defer pubsubGenerate.Close()
	pubsubUpscale := redis.Client.Subscribe(ctx, shared.COG_REDIS_UPSCALE_EVENT_CHANNEL)
	defer pubsubUpscale.Close()

	// Listen for messages
	go func() {
		klog.Infof("Listening for webhook messages on channel: %s", shared.COG_REDIS_GENERATE_EVENT_CHANNEL)
		for msg := range pubsubGenerate.Channel() {
			klog.Infof("Received webhook message: %s", msg.Payload)
			var cogMessage responses.CogStatusUpdate
			err := json.Unmarshal([]byte(msg.Payload), &cogMessage)
			if err != nil {
				klog.Errorf("--- Error unmarshalling webhook message: %v", err)
				continue
			}

			sseHub.BroadcastGenerateMessage(cogMessage)
		}
	}()

	// Upscale message from worker
	go func() {
		klog.Infof("Listening for webhook messages on channel: %s", shared.COG_REDIS_UPSCALE_EVENT_CHANNEL)
		for msg := range pubsubUpscale.Channel() {
			klog.Infof("Received webhook message: %s", msg.Payload)
			var cogMessage responses.CogStatusUpdate
			err := json.Unmarshal([]byte(msg.Payload), &cogMessage)
			if err != nil {
				klog.Errorf("--- Error unmarshalling webhook message: %v", err)
				continue
			}

			sseHub.BroadcastUpscaleMessage(cogMessage)
		}
	}()

	// Start server
	port := utils.GetEnv("PORT", "13337")
	klog.Infof("Starting server on port %s", port)
	klog.Infoln(http.ListenAndServe(fmt.Sprintf(":%s", port), app))
}
