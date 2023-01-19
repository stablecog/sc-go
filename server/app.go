package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/stablecog/go-apps/server/controller"
	"github.com/stablecog/go-apps/server/controller/websocket"
	"github.com/stablecog/go-apps/server/middleware"
	"github.com/stablecog/go-apps/utils"
	"k8s.io/klog/v2"
)

func main() {
	// Setup logger
	klog.InitFlags(nil)
	flag.Set("logtostderr", "true")
	flag.Set("stderrthreshold", "INFO")
	flag.Set("v", "3")

	flag.Parse()

	app := chi.NewRouter()

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

	// Create controller
	hc := controller.HttpController{}

	// Create and start websocket hub
	hub := websocket.NewHub()
	go hub.Run()

	// HTTP Routes
	app.Route("/v1", func(r chi.Router) {
		r.Post("/health", hc.PostHealth)
		// Websocket
		r.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
			websocket.ServeWS(hub, w, r)
		})
	})

	// Start server
	port := utils.GetEnv("PORT", "13337")
	klog.Infof("Starting server on port %s", port)
	klog.Infoln(http.ListenAndServe(fmt.Sprintf(":%s", port), app))
}
