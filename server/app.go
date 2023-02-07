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
	"github.com/stablecog/go-apps/database/ent"
	"github.com/stablecog/go-apps/database/repository"
	"github.com/stablecog/go-apps/server/api/rest"
	"github.com/stablecog/go-apps/server/api/websocket"
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
		AllowedOrigins: []string{"http://localhost:3000", "http://localhost:5173"},
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
	s.StartAsync()

	// Create a thread-safe HashMap to store in-flight cog requests
	// We use this to track requests that are in queue, and coordinate responses
	// To appropriate users over websocket
	cogRequestWebsocketConnMap := shared.NewSyncMap[string]()

	// Create and start websocket hub
	hub := websocket.NewHub()
	go hub.Run()

	// Create controller
	hc := rest.RestAPI{
		Repo:                       repo,
		Redis:                      redis,
		CogRequestWebsocketConnMap: cogRequestWebsocketConnMap,
		Hub:                        hub,
		LanguageDetector:           utils.NewLanguageDetector(),
	}

	// Create middleware
	mw := middleware.Middleware{
		SupabaseAuth: database.NewSupabaseAuth(),
		Repo:         repo,
	}

	// Routes
	app.Route("/v1", func(r chi.Router) {
		r.Get("/health", hc.HandleHealth)

		// Websocket
		r.Route("/ws", func(r chi.Router) {
			r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				websocket.ServeWS(hub, w, r)
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
	// ! TODO - move the bulk of this logic somewhere else
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

			// See if this request belongs to our instance
			websocketIdStr := cogRequestWebsocketConnMap.Get(cogMessage.Input.Id)
			if websocketIdStr == "" {
				continue
			}
			// Ensure is valid
			if !utils.IsSha256Hash(websocketIdStr) {
				// Not sure how we get here, we never should
				klog.Errorf("--- Invalid websocket id: %s", websocketIdStr)
				continue
			}

			// Processing, update this generation as started in database
			var outputs []*ent.GenerationOutput
			var cogErr string
			if cogMessage.Status == responses.CogProcessing {
				// In goroutine since we want them to know it started asap
				go repo.SetGenerationStarted(cogMessage.Input.Id)
			} else if cogMessage.Status == responses.CogFailed {
				repo.SetGenerationFailed(cogMessage.Input.Id, cogMessage.Error)
				cogErr = cogMessage.Error
			} else if cogMessage.Status == responses.CogSucceeded {
				// NSFW counts as failure
				if len(cogMessage.Outputs) == 0 && cogMessage.NSFWCount > 0 {
					cogErr = "NSFW"
					repo.SetGenerationFailed(cogMessage.Input.Id, cogErr)
					cogMessage.Status = responses.CogFailed
				} else if len(cogMessage.Outputs) == 0 && cogMessage.NSFWCount == 0 {
					cogErr = "No outputs"
					repo.SetGenerationFailed(cogMessage.Input.Id, cogErr)
					cogMessage.Status = responses.CogFailed
				} else {
					outputs, err = repo.SetGenerationSucceeded(cogMessage.Input.Id, cogMessage.Outputs)
					if err != nil {
						klog.Errorf("--- Error setting generation succeeded for ID %s: %v", cogMessage.Input.Id, err)
						continue
					}
				}
			} else {
				klog.Warningf("--- Unknown webhook status, not sure how to proceed so not proceeding at all: %s", cogMessage.Status)
				continue
			}

			// Regardless of the status, we always send over websocket so user knows what's up
			// Send message to user
			resp := responses.WebsocketStatusUpdateResponse{
				Status:    cogMessage.Status,
				Id:        cogMessage.Input.Id,
				NSFWCount: cogMessage.NSFWCount,
				Error:     cogErr,
			}

			if cogMessage.Status == responses.CogSucceeded {
				generateOutputs := make([]responses.WebhookStatusUpdateOutputs, len(outputs))
				for i, output := range outputs {
					generateOutputs[i] = responses.WebhookStatusUpdateOutputs{
						ID:               output.ID,
						ImageUrl:         output.ImageURL,
						UpscaledImageUrl: output.UpscaledImageURL,
						GalleryStatus:    output.GalleryStatus,
					}
				}
				resp.Outputs = generateOutputs
			}
			// Marshal
			respBytes, err := json.Marshal(resp)
			if err != nil {
				klog.Errorf("--- Error marshalling websocket started response: %v", err)
				continue
			}
			hub.BroadcastToClientsWithUid(websocketIdStr, respBytes)
		}
	}()

	// Upscale
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

			// See if this request belongs to our instance
			websocketIdStr := cogRequestWebsocketConnMap.Get(cogMessage.Input.Id)
			if websocketIdStr == "" {
				continue
			}
			// Ensure is valid
			if !utils.IsSha256Hash(websocketIdStr) {
				// Not sure how we get here, we never should
				klog.Errorf("--- Invalid websocket id: %s", websocketIdStr)
				continue
			}

			// Processing, update this upscale as started in database
			var output *ent.UpscaleOutput
			var cogErr string
			if cogMessage.Status == responses.CogProcessing {
				// In goroutine since we want them to know it started asap
				go repo.SetUpscaleStarted(cogMessage.Input.Id)
			} else if cogMessage.Status == responses.CogFailed {
				repo.SetUpscaleFailed(cogMessage.Input.Id, cogMessage.Error)
				cogErr = cogMessage.Error
			} else if cogMessage.Status == responses.CogSucceeded {
				if len(cogMessage.Outputs) == 0 {
					cogErr = "No outputs"
					repo.SetUpscaleFailed(cogMessage.Input.Id, cogErr)
					cogMessage.Status = responses.CogFailed
				} else {
					output, err = repo.SetUpscaleSucceeded(cogMessage.Input.Id, cogMessage.Input.GenerationOutputID, cogMessage.Outputs[0])
					if err != nil {
						klog.Errorf("--- Error setting upscale succeeded for ID %s: %v", cogMessage.Input.Id, err)
						continue
					}
				}

			} else {
				klog.Warningf("--- Unknown webhook status, not sure how to proceed so not proceeding at all: %s", cogMessage.Status)
				continue
			}

			// Regardless of the status, we always send over websocket so user knows what's up
			// Send message to user
			resp := responses.WebsocketStatusUpdateResponse{
				Status:    cogMessage.Status,
				Id:        cogMessage.Input.Id,
				NSFWCount: cogMessage.NSFWCount,
				Error:     cogErr,
			}
			if cogMessage.Status == responses.CogSucceeded {
				resp.Outputs = []responses.WebhookStatusUpdateOutputs{
					{
						ID:       output.ID,
						ImageUrl: output.ImageURL,
					},
				}
			}
			// Marshal
			respBytes, err := json.Marshal(resp)
			if err != nil {
				klog.Errorf("--- Error marshalling websocket started response: %v", err)
				continue
			}
			hub.BroadcastToClientsWithUid(websocketIdStr, respBytes)
		}
	}()

	// Start server
	port := utils.GetEnv("PORT", "13337")
	klog.Infof("Starting server on port %s", port)
	klog.Infoln(http.ListenAndServe(fmt.Sprintf(":%s", port), app))
}
