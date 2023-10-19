package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
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
	"github.com/stablecog/sc-go/database/ent/user"
	"github.com/stablecog/sc-go/database/qdrant"
	"github.com/stablecog/sc-go/database/repository"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/server/analytics"
	"github.com/stablecog/sc-go/server/api/rest"
	"github.com/stablecog/sc-go/server/api/sse"
	"github.com/stablecog/sc-go/server/clip"
	"github.com/stablecog/sc-go/server/discord"
	"github.com/stablecog/sc-go/server/middleware"
	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/server/responses"
	"github.com/stablecog/sc-go/server/scworker"
	"github.com/stablecog/sc-go/shared"
	"github.com/stablecog/sc-go/shared/queue"
	uapi "github.com/stablecog/sc-go/uploadapi/api"
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
	transferUserData := flag.Bool("transfer-user", false, "transfer account data from one user to another")
	sourceUser := flag.String("source-user", "", "source user id")
	destUser := flag.String("dest-user", "", "destination user id")
	cursorEmbeddings := flag.String("cursor-embeddings", "", "Cursor for loading embeddings")
	syncPromptId := flag.Bool("sync-prompt-id", false, "Sync prompt_id to qdrant")
	syncIsPublic := flag.Bool("sync-is-public", false, "Sync is_public to qdrant")
	loadQdrant := flag.Bool("load-qdrant", false, "Load qdrant with all data")
	reverse := flag.Bool("reverse", false, "Reverse the order of the embeddings")
	clipUrlOverride := flag.String("clip-url", "", "Clip url to process")
	migrateUsername := flag.Bool("migrate-username", false, "Generate usernames for existing users")

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

	// Create indexes in Qdrant
	err = qdrantClient.CreateAllIndexes()
	if err != nil {
		log.Warn("Error creating qdrant indexes", "err", err)
	}

	// Q Throttler
	qThrottler := shared.NewQueueThrottler(ctx, redis.Client, shared.REQUEST_COG_TIMEOUT)

	// Create repository (database access)
	repo := &repository.Repository{
		DB:             entClient,
		ConnInfo:       dbconn,
		Redis:          redis,
		Ctx:            ctx,
		Qdrant:         qdrantClient,
		QueueThrottler: qThrottler,
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

	if *syncPromptId {
		log.Info("🏡 Loading qdrant data...")
		each := 500
		cur := 0
		var cursor *time.Time
		if *cursorEmbeddings != "" {
			t, err := time.Parse(time.RFC3339, *cursorEmbeddings)
			if err != nil {
				log.Fatal("Failed to parse cursor", "err", err)
			}
			cursor = &t
		}

		for {
			log.Info("Loading batch of embeddings", "cur", cur, "each", each)
			start := time.Now()
			q := repo.DB.GenerationOutput.Query().Where(generationoutput.HasEmbeddings(true), generationoutput.DeletedAtIsNil(), generationoutput.ImagePathNEQ("placeholder.webp")).WithGenerations()
			if cursor != nil {
				q = q.Where(generationoutput.CreatedAtLT(*cursor))
			}
			gens, err := q.Order(ent.Desc(generationoutput.FieldCreatedAt)).Limit(each).All(ctx)
			if err != nil {
				if cursor != nil {
					log.Info("Last cursor", "cursor", cursor.Format(time.RFC3339Nano))
				}
				log.Fatal("Failed to load generation outputs", "err", err)
			}
			log.Infof("Retreived generations in %s", time.Since(start))

			if len(gens) == 0 {
				break
			}

			for _, gen := range gens {
				properties := make(map[string]interface{})
				properties["prompt_id"] = gen.Edges.Generations.PromptID.String()
				err = qdrantClient.SetPayload(properties, []uuid.UUID{gen.ID}, false)
				if err != nil {
					log.Fatal("Failed to set payload", "err", err)
				}
			}
			// Update cursor
			cursor = &gens[len(gens)-1].CreatedAt
			log.Info("Last cursor", "cursor", cursor.Format(time.RFC3339Nano))
			log.Infof("Loaded %d generations", len(gens))
			cur += len(gens)
		}
		log.Infof("Done, sync'd %d", cur)
		os.Exit(0)
	}

	if *loadQdrant {
		log.Info("🏡 Loading qdrant data...")
		secret := os.Getenv("CLIPAPI_SECRET")
		clipUrl := os.Getenv("CLIP_URL_TEMP")
		if *clipUrlOverride != "" {
			clipUrl = *clipUrlOverride
		}
		each := 100
		cur := 0
		urlIdx := 0
		var cursor *time.Time
		if *cursorEmbeddings != "" {
			t, err := time.Parse(time.RFC3339, *cursorEmbeddings)
			if err != nil {
				log.Fatal("Failed to parse cursor", "err", err)
			}
			cursor = &t
		}

		promptEmbeddings := make(map[string][]float32)

		for {
			log.Info("Loading batch of embeddings", "cur", cur, "each", each)
			start := time.Now()
			q := repo.DB.GenerationOutput.Query().Where(generationoutput.HasEmbeddings(false), generationoutput.ImagePathNEQ("placeholder.webp"))
			if cursor != nil {
				if *reverse {
					q = q.Where(generationoutput.CreatedAtGT(*cursor))
				} else {
					q = q.Where(generationoutput.CreatedAtLT(*cursor))
				}
			}
			if *reverse {
				q.Order(ent.Asc(generationoutput.FieldCreatedAt))
			} else {
				q.Order(ent.Desc(generationoutput.FieldCreatedAt))
			}
			gens, err := q.WithGenerations(func(gq *ent.GenerationQuery) {
				gq.WithPrompt()
				gq.WithNegativePrompt()
			}).Limit(each).All(ctx)
			if err != nil {
				if cursor != nil {
					log.Info("Last cursor", "cursor", cursor.Format(time.RFC3339Nano))
				}
				log.Fatal("Failed to load generation outputs", "err", err)
			}
			log.Infof("Retreived generations in %s", time.Since(start))

			// Update cursor
			cursor = &gens[len(gens)-1].CreatedAt

			if len(gens) == 0 {
				break
			}

			ids := make([]uuid.UUID, len(gens))
			var clipReq []requests.ClipAPIImageRequest
			promptMap := make(map[uuid.UUID]string)
			for i, gen := range gens {
				ids[i] = gen.ID
				clipReq = append(clipReq, requests.ClipAPIImageRequest{
					ID:      gen.ID,
					ImageID: gen.ImagePath,
				})
				if _, ok := promptEmbeddings[gen.Edges.Generations.Edges.Prompt.Text]; !ok {
					promptMap[gen.GenerationID] = gen.Edges.Generations.Edges.Prompt.Text
				}
			}
			for k, gen := range promptMap {
				clipReq = append(clipReq, requests.ClipAPIImageRequest{
					ID:   k,
					Text: gen,
				})
			}

			// Make API request to clip
			start = time.Now()
			b, err := json.Marshal(clipReq)
			if err != nil {
				log.Infof("Last cursor: %v", cursor.Format(time.RFC3339Nano))
				log.Fatalf("Error marshalling req %v", err)
			}
			request, _ := http.NewRequest(http.MethodPost, clipUrl, bytes.NewReader(b))
			urlIdx++
			request.Header.Set("Authorization", secret)
			request.Header.Set("Content-Type", "application/json")
			// Do
			resp, err := http.DefaultClient.Do(request)
			if err != nil {
				log.Infof("Last cursor: %v", cursor.Format(time.RFC3339Nano))
				log.Warnf("Error making request %v", err)
				time.Sleep(30 * time.Second)
				continue
			}
			defer resp.Body.Close()

			readAll, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Infof("Last cursor: %v", cursor.Format(time.RFC3339Nano))
				log.Fatal(err)
			}
			var clipAPIResponse responses.EmbeddingsResponse
			err = json.Unmarshal(readAll, &clipAPIResponse)
			if err != nil {
				log.Infof("Last cursor: %v", cursor.Format(time.RFC3339Nano))
				log.Warnf("Error unmarshalling resp clip %v", err)
				continue
			}

			// Builds maps of embeddings
			embeddings := make(map[uuid.UUID][]float32)
			for _, embedding := range clipAPIResponse.Embeddings {
				if embedding.Error != "" {
					log.Warn("Error from clip api", "err", embedding.Error)
					continue
				}
				embeddings[embedding.ID] = embedding.Embedding
			}

			log.Infof("Retreived embeddings in %s", time.Since(start))

			// Build payloads for qdrant
			var payloads []map[string]interface{}

			start = time.Now()
			for _, gOutput := range gens {
				payload := map[string]interface{}{
					"image_path":         gOutput.ImagePath,
					"gallery_status":     gOutput.GalleryStatus,
					"is_favorited":       gOutput.IsFavorited,
					"was_auto_submitted": gOutput.Edges.Generations.WasAutoSubmitted,
					"created_at":         gOutput.CreatedAt.Unix(),
					"updated_at":         gOutput.UpdatedAt.Unix(),
					"generation_id":      gOutput.Edges.Generations.ID.String(),
					"guidance_scale":     gOutput.Edges.Generations.GuidanceScale,
					"inference_steps":    gOutput.Edges.Generations.InferenceSteps,
					"prompt_strength":    gOutput.Edges.Generations.PromptStrength,
					"height":             gOutput.Edges.Generations.Height,
					"width":              gOutput.Edges.Generations.Width,
					"model":              gOutput.Edges.Generations.ModelID.String(),
					"scheduler":          gOutput.Edges.Generations.SchedulerID.String(),
					"user_id":            gOutput.Edges.Generations.UserID.String(),
					"prompt":             gOutput.Edges.Generations.Edges.Prompt.Text,
					"is_public":          gOutput.IsPublic,
				}
				if gOutput.DeletedAt != nil {
					payload["deleted_at"] = gOutput.DeletedAt.Unix()
				}
				var ok bool
				payload["embedding"], ok = embeddings[gOutput.ID]
				if !ok {
					log.Warn("Missing embedding", "id", gOutput.ID)
					continue
				}
				payload["text_embedding"], ok = embeddings[gOutput.Edges.Generations.ID]
				if !ok {
					payload["text_embedding"], ok = promptEmbeddings[gOutput.Edges.Generations.Edges.Prompt.Text]
					if !ok {
						log.Warn("Missing text embedding", "id", gOutput.Edges.Generations.ID)
						continue
					}
				} else {
					promptEmbeddings[gOutput.Edges.Generations.Edges.Prompt.Text] = embeddings[gOutput.Edges.Generations.ID]
				}
				payload["id"] = gOutput.ID.String()
				if gOutput.UpscaledImagePath != nil {
					payload["upscaled_image_path"] = *gOutput.UpscaledImagePath
				}
				if gOutput.Edges.Generations.InitImageURL != nil {
					payload["init_image_url"] = *gOutput.Edges.Generations.InitImageURL
				}
				if gOutput.Edges.Generations.Edges.NegativePrompt != nil {
					payload["negative_prompt"] = gOutput.Edges.Generations.Edges.NegativePrompt.Text
				}
				payloads = append(payloads, payload)
			}

			// QD Upsert
			err = qdrantClient.BatchUpsert(payloads, false)
			if err != nil {
				log.Info("Last cursor", "cursor", cursor.Format(time.RFC3339Nano))
				log.Warn("Failed to batch objects", "err", err)
				continue
			}

			log.Infof("Batched objects for qdrant in %s", time.Since(start))

			err = repo.DB.GenerationOutput.Update().Where(generationoutput.IDIn(ids...)).SetHasEmbeddings(true).Exec(ctx)
			if err != nil {
				log.Info("Last cursor", "cursor", cursor.Format(time.RFC3339Nano))
				log.Fatal("Failed to update generation outputs", "err", err)
			}
			log.Info("Batched objects", "count", len(payloads))
			cur += len(payloads)

			// Log cursor
			log.Info("Last cursor", "cursor", cursor.Format(time.RFC3339Nano))
		}

		log.Info("Loaded generation outputs", "count", cur)
		os.Exit(0)
	}

	if *syncIsPublic {
		log.Info("🏡 Loading qdrant data...")
		each := 500
		cur := 0
		var cursor *time.Time
		if *cursorEmbeddings != "" {
			t, err := time.Parse(time.RFC3339, *cursorEmbeddings)
			if err != nil {
				log.Fatal("Failed to parse cursor", "err", err)
			}
			cursor = &t
		}

		for {
			log.Info("Loading batch of embeddings", "cur", cur, "each", each)
			start := time.Now()
			q := repo.DB.GenerationOutput.Query().Where(generationoutput.HasEmbeddings(true), generationoutput.DeletedAtIsNil(), generationoutput.ImagePathNEQ("placeholder.webp"))
			if cursor != nil {
				q = q.Where(generationoutput.CreatedAtLT(*cursor))
			}
			gens, err := q.Order(ent.Desc(generationoutput.FieldCreatedAt)).Limit(each).All(ctx)
			if err != nil {
				if cursor != nil {
					log.Info("Last cursor", "cursor", cursor.Format(time.RFC3339Nano))
				}
				log.Fatal("Failed to load generation outputs", "err", err)
			}
			log.Infof("Retreived generations in %s", time.Since(start))

			if len(gens) == 0 {
				break
			}

			for _, gen := range gens {
				properties := make(map[string]interface{})
				properties["is_public"] = gen.IsPublic
				err = qdrantClient.SetPayload(properties, []uuid.UUID{gen.ID}, false)
				if err != nil {
					log.Fatal("Failed to set payload", "err", err)
				}
			}
			// Update cursor
			cursor = &gens[len(gens)-1].CreatedAt
			log.Info("Last cursor", "cursor", cursor.Format(time.RFC3339Nano))
			log.Infof("Loaded %d generations", len(gens))
			cur += len(gens)
		}
		log.Infof("Done, sync'd %d", cur)
		os.Exit(0)
	}

	if *migrateUsername {
		log.Info("🏡 Generating usernames...")
		users, err := repo.DB.User.Query().Where(user.UsernameChangedAtIsNil()).All(ctx)
		if err != nil {
			log.Fatal("Failed to migrate usernames", "err", err)
			os.Exit(1)
		}
		for _, user := range users {
			repo.DB.User.UpdateOne(user).SetUsername(utils.GenerateUsername(nil)).SaveX(ctx)
		}
		log.Info("Done")
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
	const cacheIntervalSec = 30
	s.Every(cacheIntervalSec).Seconds().StartAt(time.Now().Add(cacheIntervalSec * time.Second)).Do(func() {
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

	// Setup rabbitmq client
	rabbitmqClient, err := queue.NewRabbitMQClient(ctx, os.Getenv("RABBITMQ_AMQP_URL"))
	if err != nil {
		log.Fatalf("Error connecting to rabbitmq: %v", err)
	}
	defer rabbitmqClient.Close()

	// Create controller
	apiTokenSmap := shared.NewSyncMap[chan requests.CogWebhookMessage]()
	safetyChecker := utils.NewTranslatorSafetyChecker(ctx, os.Getenv("OPENAI_API_KEY"), false)
	hc := rest.RestAPI{
		Repo:           repo,
		Redis:          redis,
		Hub:            sseHub,
		StripeClient:   stripeClient,
		Track:          analyticsService,
		QueueThrottler: qThrottler,
		S3:             s3Client,
		Qdrant:         qdrantClient,
		Clip:           clip.NewClipService(redis, safetyChecker),
		SMap:           apiTokenSmap,
		SafetyChecker:  safetyChecker,
		SCWorker: &scworker.SCWorker{
			Repo:           repo,
			Redis:          redis,
			QueueThrottler: qThrottler,
			Track:          analyticsService,
			SMap:           apiTokenSmap,
			SafetyChecker:  safetyChecker,
			S3:             s3Client,
			MQClient:       rabbitmqClient,
		},
	}

	// Create upload controller
	uploadHc := uapi.Controller{
		Repo:  repo,
		Redis: redis,
		S3:    s3Client,
	}

	// Create middleware
	mw := middleware.Middleware{
		SupabaseAuth: database.NewSupabaseAuth(),
		Repo:         repo,
		Redis:        redis,
	}

	if *transferUserData {
		sourceID := uuid.MustParse(*sourceUser)
		targetID := uuid.MustParse(*destUser)

		log.Info("📦 Transferring user data...", "source", sourceID, "target", targetID)
		// Get all generation output ids
		outputs := repo.DB.Generation.Query().Where(generation.UserIDEQ(targetID)).QueryGenerationOutputs().AllX(ctx)
		gOutputIDs := make([]uuid.UUID, len(outputs))
		for i, o := range outputs {
			gOutputIDs[i] = o.ID
		}

		// Set qdrant payload
		qdrantPayload := map[string]interface{}{
			"user_id": targetID.String(),
		}

		err := qdrantClient.SetPayload(qdrantPayload, gOutputIDs, false)
		if err != nil {
			log.Fatalf("Error setting qdrant payload: %v", err)
		}

		log.Infof("Sync'd qdrant")

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

		sourceHash := utils.Sha256(sourceID.String())
		targetHash := utils.Sha256(targetID.String())

		out, err := s3Client.ListObjects(&s3.ListObjectsInput{
			Bucket: aws.String(os.Getenv("S3_IMG2IMG_BUCKET_NAME")),
			Prefix: aws.String(fmt.Sprintf("%s/", sourceHash)),
		})
		if err != nil {
			log.Fatalf("Error listing img2img objects for user %s: %v", sourceID.String(), err)
		}

		for _, o := range out.Contents {
			// Move object
			src := *o.Key
			dst := strings.Replace(src, sourceHash, targetHash, 1)
			_, err := s3Client.CopyObject(&s3.CopyObjectInput{
				Bucket:     aws.String(os.Getenv("S3_IMG2IMG_BUCKET_NAME")),
				CopySource: aws.String(url.QueryEscape(fmt.Sprintf("%s/%s", os.Getenv("S3_IMG2IMG_BUCKET_NAME"), src))),
				Key:        aws.String(dst),
			})
			if err != nil {
				log.Fatalf("Error copying img2img object for user %s: %v", sourceID.String(), err)
			}
		}

		log.Infof("Finished sync")
		os.Exit(0)
	}

	// Routes
	app.Get("/", hc.HandleHealth)
	app.Handle("/metrics", middleware.BasicAuth(promhttp.Handler(), "user", "password", "Authentication required"))
	app.Get("/clipq", hc.HandleClipQSearch)
	app.Route("/v1", func(r chi.Router) {
		r.Get("/health", hc.HandleHealth)

		r.Route("/email", func(r chi.Router) {
			r.Use(middleware.Logger)
			r.Use(mw.RateLimit(5, "srv", 1*time.Second))
			r.Post("/check", hc.HandleVerifyEmailDomain)
		})

		// SSE
		r.Route("/sse", func(r chi.Router) {
			r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				sseHub.ServeSSE(w, r)
			})
		})

		// Stripe
		r.Route("/stripe", func(r chi.Router) {
			r.Use(middleware.Logger)
			r.Post("/webhook", hc.HandleStripeWebhook)
		})

		// SCWorker
		r.Route("/worker", func(r chi.Router) {
			r.Use(middleware.Logger)
			r.Post("/webhook", hc.HandleSCWorkerWebhook)
		})

		// Stats
		r.Route("/stats", func(r chi.Router) {
			r.Use(middleware.Logger)
			// 10 requests per second
			r.Use(mw.RateLimit(10, "srv", 1*time.Second))
			r.Get("/", hc.HandleGetStats)
		})

		// Gallery search
		r.Route("/gallery", func(r chi.Router) {
			r.Use(middleware.Logger)
			// 20 requests per second
			r.Use(mw.RateLimit(20, "srv", 1*time.Second))
			r.Get("/", hc.HandleSemanticSearchGallery)
		})

		// User profiles
		r.Route("/profile", func(r chi.Router) {
			r.Use(middleware.Logger)
			// 20 requests per second
			r.Use(mw.RateLimit(20, "srv", 1*time.Second))
			r.Get("/{username}/outputs", hc.HandleUserProfileSemanticSearch)
			r.Get("/{username}/metadata", hc.HandleGetUserProfileMetadata)
		})

		// Routes that require authentication
		r.Route("/user", func(r chi.Router) {
			r.Use(mw.AuthMiddleware(middleware.AuthLevelAny))
			r.Use(middleware.Logger)
			// 10 requests per second
			r.Use(mw.RateLimit(10, "srv", 1*time.Second))

			// Get user summary
			r.Get("/", hc.HandleGetUser)

			// Link to discord
			r.Post("/connect/discord", hc.HandleAuthorizeDiscord)

			// Create Generation
			r.Route("/image/generation/create", func(r chi.Router) {
				r.Use(mw.GeoIPMiddleware())
				r.Post("/", hc.HandleCreateGeneration)
			})
			// ! Deprecated
			r.Route("/generation", func(r chi.Router) {
				r.Use(mw.GeoIPMiddleware())
				r.Post("/", hc.HandleCreateGeneration)
			})
			// Mark generation for deletion
			r.Delete("/image/generation", hc.HandleDeleteGenerationOutputForUser)
			// ! Deprecated
			r.Delete("/generation", hc.HandleDeleteGenerationOutputForUser)

			// Query Generation (outputs + generations)
			r.Get("/image/generation/outputs", hc.HandleQueryGenerations)
			// ! Deprecated
			r.Get("/outputs", hc.HandleQueryGenerations)

			// Favorite
			r.Post("/image/generation/outputs/favorite", hc.HandleFavoriteGenerationOutputsForUser)
			// ! Deprecated
			r.Post("/outputs/favorite", hc.HandleFavoriteGenerationOutputsForUser)

			// Create upscale
			r.Post("/image/upscale/create", hc.HandleUpscale)
			// ! Deprecated
			r.Post("/upscale", hc.HandleUpscale)

			// Create voiceover
			r.Post("/audio/voiceover/create", hc.HandleVoiceover)

			// Query voiceover outputs
			r.Get("/audio/voiceover/outputs", hc.HandleQueryVoiceovers)
			r.Delete("/audio/voiceover", hc.HandleDeleteVoiceoverOutputForUser)

			// Query credits
			r.Get("/credits", hc.HandleQueryCredits)

			// ! Deprecated Submit to gallery
			r.Put("/gallery", hc.HandleSubmitGenerationToGallery)

			// Make generations public
			r.Put("/image/generation/outputs/make_public", hc.HandleMakeGenerationOutputsPublic)
			// Make generations private
			r.Put("/image/generation/outputs/make_private", hc.HandleMakeGenerationOutputsPrivate)

			// Subscriptions
			r.Post("/subscription/downgrade", hc.HandleSubscriptionDowngrade)
			r.Post("/subscription/checkout", hc.HandleCreateCheckoutSession)
			r.Post("/subscription/portal", hc.HandleCreatePortalSession)

			// API Tokens
			r.Post("/tokens", hc.HandleNewAPIToken)
			r.Get("/tokens", hc.HandleGetAPITokens)
			r.Delete("/tokens", hc.HandleDeactivateAPIToken)

			// Operations
			r.Get("/operations", hc.HandleQueryOperations)

			// Email preferences
			r.Post("/email", hc.HandleUpdateEmailPreferences)

			// Username
			r.Post("/username/change", hc.HandleUpdateUsername)
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
				r.Post("/ban", hc.HandleBanUser)
			})
			r.Route("/credit", func(r chi.Router) {
				r.Use(mw.AuthMiddleware(middleware.AuthLevelSuperAdmin))
				r.Use(middleware.Logger)
				r.Get("/types", hc.HandleQueryCreditTypes)
				r.Post("/add", hc.HandleAddCreditsToUser)
			})
			r.Route("/domains", func(r chi.Router) {
				r.Use(mw.AuthMiddleware(middleware.AuthLevelSuperAdmin))
				r.Use(middleware.Logger)
				r.Get("/disposable", hc.HandleGetDisposableDomains)
				r.Post("/ban", hc.HandleBanDomains)
			})
			r.Route("/clip", func(r chi.Router) {
				r.Use(mw.AuthMiddleware(middleware.AuthLevelSuperAdmin, middleware.AuthLevelAPIToken))
				r.Use(middleware.Logger)
				r.Post("/text", hc.HandleEmbedText)
			})
		})

		// For API tokens
		r.Route("/image", func(r chi.Router) {
			// txt2img/img2img
			r.Route("/generation/create", func(r chi.Router) {
				r.Route("/", func(r chi.Router) {
					r.Use(mw.AuthMiddleware(middleware.AuthLevelAPIToken))
					r.Use(middleware.Logger)
					r.Use(mw.GeoIPMiddleware())
					r.Use(mw.RateLimit(5, "api", 1*time.Second))
					r.Post("/", hc.HandleCreateGenerationToken)
				})
			})
			// ! Deprecated
			r.Route("/generate", func(r chi.Router) {
				r.Route("/", func(r chi.Router) {
					r.Use(mw.AuthMiddleware(middleware.AuthLevelAPIToken))
					r.Use(middleware.Logger)
					r.Use(mw.RateLimit(5, "api", 1*time.Second))
					r.Post("/", hc.HandleCreateGenerationToken)
				})
			})

			r.Route("/upscale/create", func(r chi.Router) {
				r.Route("/", func(r chi.Router) {
					r.Use(mw.AuthMiddleware(middleware.AuthLevelAPIToken))
					r.Use(middleware.Logger)
					r.Use(mw.RateLimit(5, "api", 1*time.Second))
					r.Post("/", hc.HandleCreateUpscaleToken)
				})
			})
			// ! Deprecated
			r.Route("/upscale", func(r chi.Router) {
				r.Route("/", func(r chi.Router) {
					r.Use(mw.AuthMiddleware(middleware.AuthLevelAPIToken))
					r.Use(middleware.Logger)
					r.Use(mw.RateLimit(5, "api", 1*time.Second))
					r.Post("/", hc.HandleCreateUpscaleToken)
				})
			})

			// Model info
			r.Route("/upscale/models", func(r chi.Router) {
				r.Use(middleware.Logger)
				r.Use(mw.RateLimit(10, "api", 1*time.Second))
				r.Get("/", hc.HandleGetUpscaleModels)
			})
			r.Route("/generation/models", func(r chi.Router) {
				r.Use(middleware.Logger)
				r.Use(mw.RateLimit(10, "api", 1*time.Second))
				r.Get("/", hc.HandleGetGenerationModels)
			})
			// ! Deprecated
			r.Route("/models", func(r chi.Router) {
				r.Use(middleware.Logger)
				r.Use(mw.RateLimit(10, "api", 1*time.Second))
				r.Get("/generate", hc.HandleGetGenerationModels)
				r.Get("/upscale", hc.HandleGetUpscaleModels)
			})

			// Defaults
			r.Route("/upscale/defaults", func(r chi.Router) {
				r.Use(middleware.Logger)
				r.Use(mw.RateLimit(10, "api", 1*time.Second))
				r.Get("/", hc.HandleGetUpscaleDefaults)
			})
			r.Route("/generation/defaults", func(r chi.Router) {
				r.Use(middleware.Logger)
				r.Use(mw.RateLimit(10, "api", 1*time.Second))
				r.Get("/", hc.HandleGetGenerationDefaults)
			})

			// ! Deprecated
			r.Route("/defaults", func(r chi.Router) {
				r.Use(middleware.Logger)
				r.Use(mw.RateLimit(10, "api", 1*time.Second))
				r.Get("/generate", hc.HandleGetGenerationDefaults)
				r.Get("/upscale", hc.HandleGetUpscaleDefaults)
			})

			// upload
			r.Route("/upload", func(r chi.Router) {
				r.Use(middleware.Logger)
				r.Use(mw.RateLimit(2, "uapi", 1*time.Second))
				r.Use(mw.AuthMiddleware(middleware.AuthLevelAPIToken))
				r.Post("/", uploadHc.HandleUpload)
			})

			// Querying user outputs
			r.Route("/generation/outputs", func(r chi.Router) {
				r.Use(middleware.Logger)
				r.Use(mw.RateLimit(10, "api", 1*time.Second))
				r.Use(mw.AuthMiddleware(middleware.AuthLevelAPIToken))
				r.Get("/", hc.HandleQueryGenerations)
			})
			// ! Deprecated
			r.Route("/outputs", func(r chi.Router) {
				r.Use(middleware.Logger)
				r.Use(mw.RateLimit(10, "api", 1*time.Second))
				r.Use(mw.AuthMiddleware(middleware.AuthLevelAPIToken))
				r.Get("/", hc.HandleQueryGenerations)
			})
		})

		r.Route("/audio", func(r chi.Router) {
			r.Route("/voiceover/create", func(r chi.Router) {
				r.Route("/", func(r chi.Router) {
					r.Use(mw.AuthMiddleware(middleware.AuthLevelAPIToken))
					r.Use(middleware.Logger)
					r.Use(mw.RateLimit(5, "api", 1*time.Second))
					r.Post("/", hc.HandleCreateVoiceoverToken)
				})
			})

			// Querying user outputs
			r.Route("/voiceover/outputs", func(r chi.Router) {
				r.Use(middleware.Logger)
				r.Use(mw.RateLimit(10, "api", 1*time.Second))
				r.Use(mw.AuthMiddleware(middleware.AuthLevelAPIToken))
				r.Get("/", hc.HandleQueryVoiceovers)
			})

			r.Route("/voiceover/defaults", func(r chi.Router) {
				r.Use(middleware.Logger)
				r.Use(mw.RateLimit(10, "api", 1*time.Second))
				r.Get("/", hc.HandleGetVoiceoverDefaults)
			})

			r.Route("/voiceover/models", func(r chi.Router) {
				r.Use(middleware.Logger)
				r.Use(mw.RateLimit(10, "api", 1*time.Second))
				r.Get("/", hc.HandleGetVoiceoverModels)
			})
		})

		r.Route("/credits", func(r chi.Router) {
			r.Use(middleware.Logger)
			r.Use(mw.RateLimit(10, "api", 1*time.Second))
			r.Use(mw.AuthMiddleware(middleware.AuthLevelAPIToken))
			r.Get("/", hc.HandleQueryCredits)
		})
	})

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

			// Broadcast queue update
			queueLog, err := repo.GetQueuedItems(nil)
			if err != nil {
				log.Error("Error getting queue log", "err", err)
				continue
			}
			sseHub.BroadcastQueueUpdate(queueLog)
		}
	}()

	// This redis subscription has the following purpose:
	// For API token requests, they are synchronous with API requests
	// so we need to send the response back to the appropriate channel
	apiTokenChannel := redis.Client.Subscribe(ctx, shared.REDIS_APITOKEN_COG_CHANNEL)
	defer apiTokenChannel.Close()

	// Start SSE redis subscription
	go func() {
		log.Info("Listening for api messages", "channel", shared.REDIS_APITOKEN_COG_CHANNEL)
		for msg := range apiTokenChannel.Channel() {
			var cogMessage requests.CogWebhookMessage
			err := json.Unmarshal([]byte(msg.Payload), &cogMessage)
			if err != nil {
				log.Error("Error unmarshalling cog webhook message", "err", err)
				continue
			}

			if chl := apiTokenSmap.Get(cogMessage.Input.ID.String()); chl != nil {
				chl <- cogMessage
			}
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
