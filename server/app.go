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
	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/shared"
	"github.com/stablecog/sc-go/utils"
	stripeBase "github.com/stripe/stripe-go/v74"
	stripe "github.com/stripe/stripe-go/v74/client"
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

	// Custom flags
	createMockData := flag.Bool("load-mock-data", false, "Create test data in database")
	// ! TODO - remove after production
	processCredits := flag.Bool("process-credits", false, "Process stripe subscription credits")

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

	if *createMockData {
		klog.Infoln("🏡 Creating mock data...")
		err = repo.CreateMockData(ctx)
		if err != nil {
			klog.Fatalf("Failed to create mock data: %v", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// Create stripe client
	stripeClient := stripe.New(utils.GetEnv("STRIPE_SECRET_KEY", ""), nil)

	if *processCredits {
		subList := stripeClient.Subscriptions.List(&stripeBase.SubscriptionListParams{
			Status: stripeBase.String("active"),
		})
		for subList.Next() {
			sub := subList.Subscription()
			if sub == nil {
				klog.Warningf("Subscription is nil")
				continue
			}
			if sub.LatestInvoice == nil || sub.LatestInvoice.Lines == nil {
				klog.Warningf("Latest invoice is nil")
				continue
			}
			// Get product ID
			var productId string
			var lineItemId string
			for _, line := range sub.LatestInvoice.Lines.Data {
				if line.Plan != nil {
					productId = line.Plan.Product.ID
					lineItemId = line.ID
					break
				}
			}
			// Find credit type with this id
			creditType, err := repo.GetCreditTypeByStripeProductID(productId)
			if err != nil {
				klog.Warningf("Error getting credit type: %v", err)
				continue
			} else if creditType == nil {
				klog.Warningf("Credit type doesn't exist with product ID %s", productId)
				continue
			}

			// Get user
			user, err := repo.GetUserByStripeCustomerId(sub.Customer.ID)
			if err != nil {
				klog.Warningf("Error getting user with ID %s: %v", sub.Customer.ID, err)
				continue
			}

			expiresAt := utils.SecondsSinceEpochToTime(sub.CurrentPeriodEnd)

			_, err = repo.AddCreditsIfEligible(creditType, user.ID, expiresAt, lineItemId, nil)
			if err != nil {
				klog.Warningf("Error adding credits to user %s: %v", user.ID, err)
				continue
			}
		}
		klog.Infof("Bye bye!")
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

	// Create SSE hub
	sseHub := sse.NewHub(redis, repo)
	go sseHub.Run()

	// Create controller
	hc := rest.RestAPI{
		Repo:         repo,
		Redis:        redis,
		Hub:          sseHub,
		StripeClient: stripeClient,
		Meili:        database.NewMeiliSearchClient(),
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

		// Stats
		r.Route("/stats", func(r chi.Router) {
			r.Use(chimiddleware.Logger)
			// 10 requests per second
			r.Use(mw.RateLimit(10, 1*time.Second))
			r.Get("/", hc.HandleGetStats)
		})

		// Gallery search
		r.Route("/gallery", func(r chi.Router) {
			r.Use(chimiddleware.Logger)
			// 20 requests per second
			r.Use(mw.RateLimit(20, 1*time.Second))
			r.Get("/", hc.HandleQueryGallery)
		})

		// Routes that require authentication
		r.Route("/user", func(r chi.Router) {
			r.Use(mw.AuthMiddleware(middleware.AuthLevelAny))
			r.Use(chimiddleware.Logger)
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

			// Create upscale
			r.Post("/upscale", hc.HandleUpscale)

			// Query credits
			r.Get("/credits", hc.HandleQueryCredits)

			// Subscriptions
			r.Post("/subscription/downgrade", hc.HandleSubscriptionDowngrade)
			r.Post("/subscription/checkout", hc.HandleCreateCheckoutSession)
			r.Post("/subscription/portal", hc.HandleCreatePortalSession)
		})

		// Admin only routes
		r.Route("/admin", func(r chi.Router) {
			r.Route("/gallery", func(r chi.Router) {
				r.Use(mw.AuthMiddleware(middleware.AuthLevelGalleryAdmin))
				r.Use(chimiddleware.Logger)
				r.Put("/", hc.HandleReviewGallerySubmission)
			})
			r.Route("/outputs", func(r chi.Router) {
				r.Use(mw.AuthMiddleware(middleware.AuthLevelSuperAdmin))
				r.Use(chimiddleware.Logger)
				r.Delete("/", hc.HandleDeleteGenerationOutput)
				r.Get("/", hc.HandleQueryGenerationsForAdmin)
			})
			r.Route("/users", func(r chi.Router) {
				r.Use(mw.AuthMiddleware(middleware.AuthLevelSuperAdmin))
				r.Use(chimiddleware.Logger)
				r.Get("/", hc.HandleQueryUsers)
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
			var cogMessage requests.CogRedisMessage
			err := json.Unmarshal([]byte(msg.Payload), &cogMessage)
			if err != nil {
				klog.Errorf("--- Error unmarshalling webhook message: %v", err)
				continue
			}

			// Process live page message
			go func() {
				// Live page update
				livePageMsg := cogMessage.Input.LivePageData
				if cogMessage.Status == requests.CogProcessing {
					livePageMsg.Status = shared.LivePageProcessing
				} else if cogMessage.Status == requests.CogSucceeded && len(cogMessage.Outputs) > 0 {
					livePageMsg.Status = shared.LivePageSucceeded
				} else if cogMessage.Status == requests.CogSucceeded && cogMessage.NSFWCount > 0 {
					livePageMsg.Status = shared.LivePageFailed
					livePageMsg.FailureReason = shared.NSFW_ERROR
				} else {
					livePageMsg.Status = shared.LivePageFailed
				}

				now := time.Now()
				if cogMessage.Status == requests.CogProcessing {
					livePageMsg.StartedAt = &now
				}
				if cogMessage.Status == requests.CogSucceeded || cogMessage.Status == requests.CogFailed {
					livePageMsg.CompletedAt = &now
					livePageMsg.ActualNumOutputs = len(cogMessage.Outputs)
					livePageMsg.NSFWCount = cogMessage.NSFWCount
				}
				// Send live page update
				liveResp := repository.TaskStatusUpdateResponse{
					ForLivePage:     true,
					LivePageMessage: livePageMsg,
				}
				respBytes, err := json.Marshal(liveResp)
				if err != nil {
					klog.Errorf("--- Error marshalling sse live response: %v", err)
					return
				}
				err = redis.Client.Publish(redis.Client.Context(), shared.REDIS_SSE_BROADCAST_CHANNEL, respBytes).Err()
				if err != nil {
					klog.Errorf("Failed to publish live page update: %v", err)
				}
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
			var sseMessage repository.TaskStatusUpdateResponse
			err := json.Unmarshal([]byte(msg.Payload), &sseMessage)
			if err != nil {
				klog.Errorf("--- Error unmarshalling sse message: %v", err)
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
	klog.Infof("Starting server on port %s", port)

	h2s := &http2.Server{}
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: h2c.NewHandler(app, h2s),
	}
	klog.Infoln(srv.ListenAndServe())
}
