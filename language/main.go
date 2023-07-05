package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/stablecog/sc-go/language/api"
	"github.com/stablecog/sc-go/language/shared"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/server/middleware"
	"github.com/stablecog/sc-go/utils"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

var Version = "dev"

func main() {
	log.Infof("SC Language API: %s", Version)

	// Load .env
	err := godotenv.Load("../.env")
	if err != nil {
		log.Warn("Error loading .env file (this is fine)", "err", err)
	}

	// Setup language detector
	detector := shared.NewLanguageDetector()

	app := chi.NewRouter()

	// Create controller
	hc := api.Controller{
		LanguageDetector: detector,
	}

	// Routes
	app.Route("/lingua", func(r chi.Router) {
		// Lingua API
		r.Route("/", func(r chi.Router) {
			r.Get("/health", hc.HandleHealth)
			r.Route("/", func(r chi.Router) {
				r.Use(middleware.Logger)
				r.Post("/", hc.HandleGetTargetFloresCode)
			})
		})
	})

	// Start server
	port := utils.GetEnv("PORT", "13339")
	log.Info("Starting language server", "port", port)

	h2s := &http2.Server{}
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: h2c.NewHandler(app, h2s),
	}

	log.Info(srv.ListenAndServe())
}
