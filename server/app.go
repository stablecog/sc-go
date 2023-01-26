package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/go-co-op/gocron"
	"github.com/joho/godotenv"
	"github.com/stablecog/go-apps/database"
	"github.com/stablecog/go-apps/database/repository"
	"github.com/stablecog/go-apps/server/controller"
	"github.com/stablecog/go-apps/server/controller/websocket"
	"github.com/stablecog/go-apps/server/middleware"
	"github.com/stablecog/go-apps/server/requests"
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
	klog.Infof("📦 Updating cache...")
	err = shared.GetCache().UpdateCache(repo)
	if err != nil {
		// ! Not getting these is fatal and will result in crash
		panic(err)
	}
	// Update periodically
	s := gocron.NewScheduler(time.UTC)
	s.Every(5).Minutes().StartAt(time.Now().Add(5 * time.Minute)).Do(func() {
		klog.Infof("📦 Updating cache...")
		err = shared.GetCache().UpdateCache(repo)
		if err != nil {
			klog.Errorf("Error updating cache: %v", err)
		}
	})
	s.StartAsync()

	// Create a thread-safe HashMap to store in-flight cog requests
	// We use this to track requests that are in queue, and coordinate responses
	// To appropriate users over websocket
	cogRequestWebsocketConnMap := shared.NewSyncMap[string]()

	// Setup s3
	s3Data := utils.GetS3Data()
	s3resolver := aws.EndpointResolverWithOptionsFunc(
		func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL: fmt.Sprintf("https://%s.r2.cloudflarestorage.com", s3Data.AccountId),
			}, nil
		},
	)
	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithEndpointResolverWithOptions(s3resolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(s3Data.AccessKeyId, s3Data.SecretKey, "")),
	)
	s3Client := s3.NewFromConfig(cfg)
	s3PresignClient := s3.NewPresignClient(s3Client)

	// Create and start websocket hub
	hub := websocket.NewHub()
	go hub.Run()

	// Create controller
	hc := controller.HttpController{
		Repo:                       repo,
		Redis:                      redis,
		S3Client:                   s3Client,
		S3PresignClient:            s3PresignClient,
		CogRequestWebsocketConnMap: cogRequestWebsocketConnMap,
		Hub:                        hub,
	}

	// Create middleware
	mw := middleware.Middleware{
		SupabaseAuth: database.NewSupabaseAuth(),
	}

	// Routes
	app.Route("/v1", func(r chi.Router) {
		r.Get("/health", hc.GetHealth)

		// Auth
		r.Route("/auth", func(r chi.Router) {
			r.Use(mw.AuthMiddleware)
			r.Post("/generate", hc.PostGenerate)
		})

		// Webhook
		r.Route("/queue", func(r chi.Router) {
			// ! TODO - remove QUEUE_SECRET after STA-22 is implemented
			r.Put(fmt.Sprintf("/upload/%s/*", utils.GetEnv("QUEUE_SECRET", "")), hc.PutUploadFile)
			r.Post(fmt.Sprintf("/webhook/%s", utils.GetEnv("QUEUE_SECRET", "")), hc.PostWebhook)
		})

		// Websocket, optional auth
		r.Route("/ws", func(r chi.Router) {
			r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				websocket.ServeWS(hub, w, r)
			})
		})
	})

	// Subscribe to webhook events
	pubsub := redis.Client.Subscribe(ctx, shared.COG_REDIS_WEBHOOK_QUEUE_CHANNEL)
	defer pubsub.Close()

	// Listen for messages
	go func() {
		klog.Infof("Listening for webhook messages on channel: %s", shared.COG_REDIS_WEBHOOK_QUEUE_CHANNEL)
		for msg := range pubsub.Channel() {
			klog.Infof("Received webhook message: %s", msg.Payload)
			var webhookMessage requests.WebhookRequest
			err := json.Unmarshal([]byte(msg.Payload), &webhookMessage)
			if err != nil {
				klog.Errorf("--- Error unmarshalling webhook message: %v", err)
				continue
			}

			// See if this request belongs to our instance
			websocketIdStr := cogRequestWebsocketConnMap.Get(webhookMessage.Input.Id)
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
			if webhookMessage.Status == requests.WebhookProcessing {
				// In goroutine since we don't care, like whatever man
				go repo.SetGenerationStarted(webhookMessage.Input.Id)
			} else if webhookMessage.Status == requests.WebhookFailed {
				go repo.SetGenerationFailed(webhookMessage.Input.Id, webhookMessage.Error)
			} else if webhookMessage.Status == requests.WebhookSucceeded {
				go repo.SetGenerationSucceeded(webhookMessage.Input.Id, webhookMessage.Output)
			} else {
				klog.Warningf("--- Unknown webhook status, not sure how to proceed so not proceeding at all: %s", webhookMessage.Status)
				continue
			}

			// Regardless of the status, we always send over websocket so user knows what's up
			// Send message to user
			resp := responses.WebsocketStatusUpdateResponse{
				Status: webhookMessage.Status,
				Id:     webhookMessage.Input.Id,
			}
			if webhookMessage.Status == requests.WebhookSucceeded {
				resp.Outputs = webhookMessage.Output
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
