package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	"github.com/stablecog/go-apps/database"
	"github.com/stablecog/go-apps/database/repository"
	"github.com/stablecog/go-apps/models"
	"github.com/stablecog/go-apps/server/controller"
	"github.com/stablecog/go-apps/server/controller/websocket"
	"github.com/stablecog/go-apps/server/middleware"
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
	flag.Set("v", "3")

	flag.Parse()

	app := chi.NewRouter()
	ctx := context.Background()

	// Cors middleware
	app.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"http://localhost:3000", "http://localhost:5173"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	// Log middleware
	app.Use(middleware.LogMiddleware)

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

	// Get models, schedulers and put in cache
	// ! Not getting these is fatal and will result in crash
	generationModels, err := repo.GetAllGenerationModels()
	if err != nil {
		klog.Fatalf("Failed to get generation_models: %v", err)
		panic(err)
	}
	models.GetCache().UpdateGenerationModels(generationModels)
	schedulers, err := repo.GetAllSchedulers()
	if err != nil {
		klog.Fatalf("Failed to get schedulers: %v", err)
		panic(err)
	}
	models.GetCache().UpdateSchedulers(schedulers)

	// Create controller
	hc := controller.HttpController{
		Repo:  repo,
		Redis: redis,
	}

	// Create middleware
	mw := middleware.Middleware{
		SupabaseAuth: database.NewSupabaseAuth(),
	}

	// Create and start websocket hub
	hub := websocket.NewHub()
	go hub.Run()

	// Routes
	app.Route("/v1", func(r chi.Router) {
		r.Get("/health", hc.GetHealth)

		// Auth
		r.Route("/auth", func(r chi.Router) {
			r.Use(mw.AuthMiddleware)
			r.Post("/generate", hc.PostGenerate)
		})

		// Websocket, optional auth
		r.Route("/ws", func(r chi.Router) {
			r.Use(mw.OptionalAuthMiddleware)
			r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				websocket.ServeWS(hub, w, r)
			})
		})
	})

	// Start server
	port := utils.GetEnv("PORT", "13337")
	klog.Infof("Starting server on port %s", port)
	klog.Infoln(http.ListenAndServe(fmt.Sprintf(":%s", port), app))
}
