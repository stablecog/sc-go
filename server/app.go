package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	"ariga.io/atlas/sql/migrate"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql/schema"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	"github.com/stablecog/go-apps/database"
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
	klog.Infoln("🦋 Running migrations...")
	if err := entClient.Schema.Create(
		ctx,
		schema.WithApplyHook(func(next schema.Applier) schema.Applier {
			return schema.ApplyFunc(func(ctx context.Context, conn dialect.ExecQuerier, plan *migrate.Plan) error {
				// Omit changes to users table
				filteredChanges := make([]*migrate.Change, 0)
				for _, c := range plan.Changes {
					if !strings.Contains(strings.ToLower(c.Cmd), "table \"users\"") {
						filteredChanges = append(filteredChanges, c)
						continue
					}
					klog.Infof("Skipping migration: %s", c.Cmd)
				}
				return next.Apply(ctx, conn, plan)
			})
		}),
	); err != nil {
		klog.Fatalf("Failed to run migrations: %v", err)
		os.Exit(1)
	}

	// Create controller
	hc := controller.HttpController{}

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
