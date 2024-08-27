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
	"github.com/stablecog/sc-go/server/translator"
	"github.com/stablecog/sc-go/shared"
	"github.com/stablecog/sc-go/shared/queue"
	uapi "github.com/stablecog/sc-go/uploadapi/api"
	"github.com/stablecog/sc-go/utils"
	stripe "github.com/stripe/stripe-go/v74/client"
	"golang.org/x/exp/slices"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

var Version = "dev"
var CommitMsg = "dev"

// Used to track the build time from our CI
var BuildStart = ""

func main() {
	log.Infof("SC Server: %s", Version)

	// Close loki if exists
	defer log.CloseLoki()

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
	syncDeletedAt := flag.Bool("sync-deleted-at", false, "Sync deleted_at to qdrant")
	syncIsPublic := flag.Bool("sync-is-public", false, "Sync is_public to qdrant")
	syncGalleryStatus := flag.Bool("sync-gallery-status", false, "Sync gallery_status to qdrant")
	loadQdrant := flag.Bool("load-qdrant", false, "Load qdrant with all data")
	loadQdrantStop := flag.Int("load-qdrant-stop", 0, "Stop loading qdrant at this many")
	migrateQdrant := flag.Bool("migrate-qdrant", false, "Migrate qdrant data")
	testEmbeddings := flag.Bool("test-embeddings", false, "Test embeddings API")
	testClipService := flag.Bool("test-clip-service", false, "Test clip service")
	batchSize := flag.Int("batch-size", 100, "Batch size for loading qdrant")
	reverse := flag.Bool("reverse", false, "Reverse the order of the embeddings")
	clipUrlOverride := flag.String("clip-url", "", "Clip url to process")
	clipUrl2 := flag.String("clip-url-2", "", "Clip url to process")
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
	if utils.GetEnv().RunMigrations {
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

	if *syncDeletedAt {
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
			q := repo.DB.GenerationOutput.Query().Where(generationoutput.HasEmbeddings(true), generationoutput.DeletedAtNotNil(), generationoutput.ImagePathNEQ("placeholder.webp")).WithGenerations()
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
				properties["deleted_at"] = gen.DeletedAt.Unix()
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
		clipAuthToken := utils.GetEnv().ClipApiAuthToken
		var clipUrl string
		if *clipUrlOverride != "" {
			clipUrl = *clipUrlOverride
		}
		if clipUrl == "" {
			clipUrl = utils.GetEnv().ClipApiUrl + "/embed"
		}
		each := *batchSize
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

		for {
			if *loadQdrantStop > 0 && cur >= *loadQdrantStop {
				log.Info("Reached stop limit", "stop", *loadQdrantStop)
				break
			}
			log.Info("Loading batch of embeddings", "cur", cur, "each", each)
			start := time.Now()
			q := repo.DB.GenerationOutput.Query().Where(generationoutput.HasEmbeddingsNew(true), generationoutput.ImagePathNEQ("placeholder.webp"))
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
			var clipReq []requests.ClipApiEmbeddingRequest
			for i, gen := range gens {
				ids[i] = gen.ID
				clipReq = append(clipReq, requests.ClipApiEmbeddingRequest{
					ID:    gen.ID.String(),
					Image: gen.ImagePath,
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
			request.Header.Set("Authorization", clipAuthToken)
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
			var clipAPIResponse responses.ClipEmbeddingResponse
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
				embUuid, embErr := uuid.Parse(embedding.ID)
				if embErr != nil {
					log.Warn("Failed to parse uuid", "err", embErr)
					continue
				}
				embeddings[embUuid] = embedding.Embedding
			}

			log.Infof("Retreived embeddings in %s", time.Since(start))

			// Build payloads for qdrant
			var payloads []map[string]interface{}

			start = time.Now()
			for _, gOutput := range gens {
				payload := map[string]interface{}{
					"image_path":               gOutput.ImagePath,
					"gallery_status":           gOutput.GalleryStatus,
					"is_favorited":             gOutput.IsFavorited,
					"was_auto_submitted":       gOutput.Edges.Generations.WasAutoSubmitted,
					"created_at":               gOutput.CreatedAt.Unix(),
					"updated_at":               gOutput.UpdatedAt.Unix(),
					"generation_id":            gOutput.Edges.Generations.ID.String(),
					"guidance_scale":           gOutput.Edges.Generations.GuidanceScale,
					"inference_steps":          gOutput.Edges.Generations.InferenceSteps,
					"prompt_strength":          gOutput.Edges.Generations.PromptStrength,
					"height":                   gOutput.Edges.Generations.Height,
					"width":                    gOutput.Edges.Generations.Width,
					"model":                    gOutput.Edges.Generations.ModelID.String(),
					"scheduler":                gOutput.Edges.Generations.SchedulerID.String(),
					"user_id":                  gOutput.Edges.Generations.UserID.String(),
					"prompt":                   gOutput.Edges.Generations.Edges.Prompt.Text,
					"is_public":                gOutput.IsPublic,
					"prompt_id":                gOutput.Edges.Generations.PromptID.String(),
					"aesthetic_rating_score":   gOutput.AestheticRatingScore,
					"aesthetic_artifact_score": gOutput.AestheticArtifactScore,
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

			err = repo.DB.GenerationOutput.Update().Where(generationoutput.IDIn(ids...)).SetHasEmbeddingsNew(true).Exec(ctx)
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

	if *migrateQdrant {
		log.Info("🏡 Loading qdrant data...")
		each := *batchSize
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
			if *loadQdrantStop > 0 && cur >= *loadQdrantStop {
				log.Info("Reached stop limit", "stop", *loadQdrantStop)
				break
			}
			log.Info("Loading batch of embeddings", "cur", cur, "each", each)
			start := time.Now()
			q := repo.DB.GenerationOutput.Query().Select(generationoutput.FieldID, generationoutput.FieldCreatedAt).Where(generationoutput.IsMigratedEQ(false), generationoutput.ImagePathNEQ("placeholder.webp"))
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
			gens, err := q.Limit(each).All(ctx)
			if err != nil {
				if cursor != nil {
					log.Info("Last cursor", "cursor", cursor.Format(time.RFC3339Nano))
				}
				log.Fatal("Failed to load generation outputs", "err", err)
			}
			log.Infof("Retrieved generations in %s", time.Since(start))

			if len(gens) == 0 {
				break
			}

			// Update cursor
			cursor = &gens[len(gens)-1].CreatedAt

			ids := make([]uuid.UUID, len(gens))
			for i, gen := range gens {
				ids[i] = gen.ID
			}

			// Get points from qdrant
			res, err := qdrantClient.GetPoints(ids, false)
			if err != nil || res == nil {
				log.Info("Last cursor", "cursor", cursor.Format(time.RFC3339Nano))
				log.Warn("Failed to get points from qdrant", "err", err)
				continue
			}

			// New QD Upsert
			url := fmt.Sprintf("%s/collections/%s/points", os.Getenv("NEW_QDRANT_URL"), qdrantClient.CollectionName)

			requestBody, err := json.Marshal(qdrant.UpsertPointRequest{Points: res.Result})
			if err != nil {
				log.Info("Last cursor", "cursor", cursor.Format(time.RFC3339Nano))
				log.Warn("Failed to marshal upsert request", "err", err)
				continue
			}

			req, err := http.NewRequest("PUT", url, bytes.NewBuffer(requestBody))
			if err != nil {
				log.Info("Last cursor", "cursor", cursor.Format(time.RFC3339Nano))
				log.Warn("Failed to create request to upsert points to new qdrant", "err", err)
				continue
			}

			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				log.Info("Last cursor", "cursor", cursor.Format(time.RFC3339Nano))
				log.Warn("Failed to execute upsert in qdrant", "err", err)
				continue
			}

			// Close the response body immediately after reading
			func() {
				defer resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					log.Info("Last cursor", "cursor", cursor.Format(time.RFC3339Nano))
					log.Warn("Received non-200 status code from new qdrant", "code", resp.StatusCode)
					return
				}
			}()

			log.Infof("Batched objects for qdrant in %s", time.Since(start))

			err = repo.DB.GenerationOutput.Update().Where(generationoutput.IDIn(ids...)).SetIsMigrated(true).Exec(ctx)
			if err != nil {
				log.Info("Last cursor", "cursor", cursor.Format(time.RFC3339Nano))
				log.Fatal("Failed to update generation outputs", "err", err)
			}
			log.Info("Batched objects", "count", len(ids))
			cur += len(ids)

			// Log cursor
			log.Info("Last cursor", "cursor", cursor.Format(time.RFC3339Nano))
		}

		log.Info("Loaded generation outputs", "count", cur)
		os.Exit(0)
	}

	if *testEmbeddings {
		log.Info("🏡 Test test...")
		clipAuthToken := utils.GetEnv().ClipApiAuthToken
		var clipUrl string
		if *clipUrlOverride != "" {
			clipUrl = *clipUrlOverride
		}
		if clipUrl == "" {
			clipUrl = utils.GetEnv().ClipApiUrl + "/embed"
		}
		each := *batchSize
		cur := 0
		urlIdx := 0
		var cursor *time.Time

		for {
			log.Info("Fetching batch of embeddings", "cur", cur, "each", each)
			start := time.Now()
			q := repo.DB.GenerationOutput.Query().Where(generationoutput.ImagePathNEQ("placeholder.webp"))
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
			var clipReq []requests.ClipApiEmbeddingRequest
			for i, gen := range gens {
				ids[i] = gen.ID
				clipReq = append(clipReq, requests.ClipApiEmbeddingRequest{
					ID:    gen.ID.String(),
					Image: gen.ImagePath,
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
			request.Header.Set("Authorization", clipAuthToken)
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
			var clipAPIResponse responses.ClipEmbeddingResponse
			err = json.Unmarshal(readAll, &clipAPIResponse)
			if err != nil {
				log.Infof("Last cursor: %v", cursor.Format(time.RFC3339Nano))
				log.Warnf("Error unmarshalling resp clip %v", err)
				continue
			}

			log.Infof("Retreived embeddings in %s", time.Since(start))

			log.Info("Batched objects", "count", len(clipAPIResponse.Embeddings))
			cur += len(clipAPIResponse.Embeddings)

			// Log cursor
			log.Info("Last cursor", "cursor", cursor.Format(time.RFC3339Nano))
		}

		log.Info("Hit API this many times", "count", cur)
		os.Exit(0)
	}

	if *testClipService {
		log.Info("🏡 Comparing clip...")
		clipAuthToken := utils.GetEnv().ClipApiAuthToken
		clipUrl := *clipUrlOverride
		clipUrl2 := *clipUrl2
		if *clipUrlOverride != "" {
			clipUrl = *clipUrlOverride
		}
		if clipUrl2 == "" && *clipUrlOverride == "" {
			log.Fatal("Both clip URLs is required")
		}
		each := *batchSize
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
			var clipReq []requests.ClipApiEmbeddingRequest
			for i, gen := range gens {
				ids[i] = gen.ID
				clipReq = append(clipReq, requests.ClipApiEmbeddingRequest{
					ID:    gen.ID.String(),
					Image: gen.ImagePath,
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
			request.Header.Set("Authorization", clipAuthToken)
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
			var clipAPIResponse responses.ClipEmbeddingResponse
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
				embUuid, embErr := uuid.Parse(embedding.ID)
				if embErr != nil {
					log.Warn("Failed to parse uuid", "err", embErr)
					continue
				}
				embeddings[embUuid] = embedding.Embedding
			}

			// Clip 2 request
			request2, _ := http.NewRequest(http.MethodPost, clipUrl2, bytes.NewReader(b))
			urlIdx++
			request2.Header.Set("Authorization", clipAuthToken)
			request2.Header.Set("Content-Type", "application/json")
			// Do
			resp2, err := http.DefaultClient.Do(request2)
			if err != nil {
				log.Infof("Last cursor: %v", cursor.Format(time.RFC3339Nano))
				log.Warnf("Error making request %v", err)
				time.Sleep(30 * time.Second)
				continue
			}
			defer resp2.Body.Close()

			readAll2, err := io.ReadAll(resp2.Body)
			if err != nil {
				log.Infof("Last cursor: %v", cursor.Format(time.RFC3339Nano))
				log.Fatal(err)
			}
			var clipAPIResponse2 responses.ClipEmbeddingResponse
			err = json.Unmarshal(readAll2, &clipAPIResponse2)
			if err != nil {
				log.Infof("Last cursor: %v", cursor.Format(time.RFC3339Nano))
				log.Warnf("Error unmarshalling resp clip %v", err)
				continue
			}

			// Builds maps of embeddings
			embeddings2 := make(map[uuid.UUID][]float32)
			for _, embedding := range clipAPIResponse2.Embeddings {
				if embedding.Error != "" {
					log.Warn("Error from clip api", "err", embedding.Error)
					continue
				}
				embUuid, embErr := uuid.Parse(embedding.ID)
				if embErr != nil {
					log.Warn("Failed to parse uuid", "err", embErr)
					continue
				}
				embeddings2[embUuid] = embedding.Embedding
			}

			start = time.Now()
			for i, embedding := range embeddings {
				if _, ok := embeddings2[i]; !ok {
					log.Warn("Missing embedding 2", "id", i)
					continue
				}
				if len(embedding) != len(embeddings2[i]) {
					log.Warn("Embeddings are different lengths", "id", i)
					continue
				}
				if slices.Compare(embedding, embeddings2[i]) != 0 {
					log.Warn("Embeddings are different", "id", i)
				}
			}

			// Log cursor
			log.Info("Last cursor", "cursor", cursor.Format(time.RFC3339Nano))
		}

		log.Info("Done", "count", cur)
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

	if *syncGalleryStatus {
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
			q := repo.DB.GenerationOutput.Query().Where(generationoutput.HasEmbeddings(true), generationoutput.DeletedAtIsNil(), generationoutput.ImagePathNEQ("placeholder.webp"), generationoutput.GalleryStatusEQ(generationoutput.GalleryStatusApproved))
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
				properties["gallery_status"] = gen.GalleryStatus
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
	stripeClient := stripe.New(utils.GetEnv().StripeSecretKey, nil)

	app := chi.NewRouter()

	// Prometheus middleware
	promMiddleware := chiprometheus.NewMiddleware("sc-server")
	app.Use(promMiddleware)

	// Cors middleware
	app.Use(cors.Handler(cors.Options{
		AllowedOrigins: utils.GetEnv().GetCorsOrigins(),
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
	region := utils.GetEnv().S3Img2ImgRegion
	accessKey := utils.GetEnv().S3Img2ImgAccessKey
	secretKey := utils.GetEnv().S3Img2ImgSecretKey

	s3ConfigImg := &aws.Config{
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
		Endpoint:    aws.String(utils.GetEnv().S3Img2ImgEndpoint),
		Region:      aws.String(region),
	}

	newSessionImg := session.New(s3ConfigImg)
	s3ClientImg := s3.New(newSessionImg)

	// Setup S3 Client regular
	region = utils.GetEnv().S3Region
	accessKey = utils.GetEnv().S3AccessKey
	secretKey = utils.GetEnv().S3SecretKey

	s3Config := &aws.Config{
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
		Endpoint:    aws.String(utils.GetEnv().S3Endpoint),
		Region:      aws.String(region),
	}

	newSession := session.New(s3Config)
	s3Client := s3.New(newSession)

	// Setup rabbitmq client
	rabbitmqClient, err := queue.NewRabbitMQClient(ctx, utils.GetEnv().RabbitMQAMQPUrl)
	if err != nil {
		log.Fatalf("Error connecting to rabbitmq: %v", err)
	}
	defer rabbitmqClient.Close()

	// Create controller
	apiTokenSmap := shared.NewSyncMap[chan requests.CogWebhookMessage]()
	safetyChecker := translator.NewTranslatorSafetyChecker(ctx, utils.GetEnv().OpenAIApiKey, false, redis)
	hc := rest.RestAPI{
		Repo:           repo,
		Redis:          redis,
		Hub:            sseHub,
		StripeClient:   stripeClient,
		Track:          analyticsService,
		QueueThrottler: qThrottler,
		S3:             s3ClientImg,
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
			S3Img:          s3ClientImg,
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
		region := utils.GetEnv().S3Img2ImgRegion
		accessKey := utils.GetEnv().S3Img2ImgAccessKey
		secretKey := utils.GetEnv().S3Img2ImgSecretKey

		s3Config := &aws.Config{
			Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
			Endpoint:    aws.String(utils.GetEnv().S3Img2ImgEndpoint),
			Region:      aws.String(region),
		}

		newSession := session.New(s3Config)
		s3Client := s3.New(newSession)

		sourceHash := utils.Sha256(sourceID.String())
		targetHash := utils.Sha256(targetID.String())

		out, err := s3Client.ListObjects(&s3.ListObjectsInput{
			Bucket: aws.String(utils.GetEnv().S3Img2ImgBucketName),
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
				Bucket:     aws.String(utils.GetEnv().S3Img2ImgBucketName),
				CopySource: aws.String(url.QueryEscape(fmt.Sprintf("%s/%s", utils.GetEnv().S3Img2ImgBucketName, src))),
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
	// app.Get("/clipq", hc.HandleClipQSearch)
	app.Route("/v1", func(r chi.Router) {
		r.Get("/health", hc.HandleHealth)

		r.Get("/test-html.html", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`<html><body><h1>Test HTML</h1><p>Lorem ipsum dolor sit amet, consectetur adipiscing elit. Phasellus vehicula nunc dictum, consequat felis vulputate, faucibus turpis. Proin viverra eros quis ultricies consectetur. Pellentesque condimentum id ipsum non sollicitudin. Cras congue eleifend tellus non varius. In quis est a nunc tincidunt dignissim sed eget sapien. Integer eu arcu eu nisi varius porta. Ut consectetur sapien varius felis gravida, id fringilla leo gravida. Etiam varius vehicula imperdiet. Nulla laoreet sem mattis, gravida nibh id, feugiat ligula. Cras vel urna fringilla, sodales ipsum sit amet, convallis libero. Phasellus consequat nulla eget rhoncus porttitor. Donec tempus neque at varius tristique. Etiam ultricies augue ac magna iaculis mattis. Cras ornare aliquet libero, eu ornare enim egestas eu.

Etiam ullamcorper gravida facilisis. Maecenas pretium cursus turpis ut pharetra. Vivamus non dignissim velit. Suspendisse vehicula arcu vel lorem eleifend, vitae rhoncus turpis tincidunt. Orci varius natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Nullam tincidunt ornare accumsan. Aliquam scelerisque neque elit, ac lobortis urna feugiat nec. Duis vel magna maximus, tristique neque ac, tempus purus. Nunc non est posuere, venenatis dui et, lobortis magna. Cras euismod tincidunt enim ut bibendum. In commodo tempor leo, nec blandit eros vestibulum a.

Integer vel sollicitudin erat. Praesent semper dolor vitae molestie vulputate. Cras bibendum mattis purus, quis facilisis lectus. Donec rutrum euismod sem vel auctor. Sed hendrerit, magna in mattis vestibulum, eros augue tincidunt ipsum, in commodo erat lectus ut orci. Mauris posuere urna et nulla pulvinar malesuada. Nulla vel mi dui. Nullam lacinia feugiat commodo. Nam viverra arcu ex, in ullamcorper lacus condimentum vitae. Fusce eget augue iaculis, placerat nibh a, ultrices ex.

Maecenas et elit dignissim elit condimentum ultricies. Etiam imperdiet consequat tempor. Phasellus at tellus nisl. Vivamus interdum quam quis felis laoreet efficitur. Aliquam non felis cursus massa blandit condimentum. Duis lorem mauris, aliquam et nibh vitae, feugiat scelerisque tellus. Sed at mi quis sem pulvinar hendrerit. Fusce facilisis lorem in purus eleifend semper. Quisque at varius libero. Vestibulum efficitur leo quam, eget commodo mi semper at. Proin iaculis elementum libero ut sagittis. Morbi sapien est, dignissim id iaculis sollicitudin, consequat congue nunc. Quisque sed maximus risus, eget auctor est. Nulla et tortor purus. Maecenas sed blandit orci. Suspendisse finibus tellus at magna placerat, finibus ornare erat sodales.

Phasellus a est ut nulla commodo blandit id sed nibh. Ut venenatis turpis ac mi convallis, at laoreet turpis convallis. Vestibulum ac purus elit. Vestibulum magna metus, molestie quis ipsum eu, imperdiet placerat enim. Donec convallis diam ultricies tempus ullamcorper. Aliquam odio lectus, tristique nec convallis eu, malesuada at mauris. Sed mauris arcu, aliquam sit amet metus vitae, laoreet porttitor nisl.

Nunc bibendum ullamcorper facilisis. Aliquam tincidunt nec justo sodales consectetur. Quisque quis sem massa. Suspendisse at nulla consequat, dapibus diam ut, suscipit elit. Integer sed blandit dolor, congue lobortis nisi. Donec porta sapien ut libero tempus ornare. Vestibulum mollis sed sapien sit amet facilisis. Donec sed felis tortor. Ut maximus nulla et rutrum tristique. Pellentesque sollicitudin pretium lacus, a volutpat purus aliquam non. Quisque porttitor laoreet sodales. Class aptent taciti sociosqu ad litora torquent per conubia nostra, per inceptos himenaeos.

Nam arcu magna, sodales sed dolor ut, viverra malesuada orci. Proin in lacinia diam. Aenean ac semper ipsum, dictum dapibus justo. Cras ultricies id eros at venenatis. Nam semper, purus ut elementum scelerisque, dolor velit ultricies lacus, at feugiat mauris erat ac neque. Nulla nec posuere nisi, vel fermentum justo. Aenean eu urna arcu. Aenean lorem eros, bibendum nec placerat a, mollis nec tortor. Cras ultricies pretium augue sit amet ultrices. Fusce ut varius turpis, in convallis lorem. Etiam vitae ultricies nisl. Aliquam erat volutpat.

Suspendisse finibus mauris eros, nec lobortis nisl consectetur in. Suspendisse a velit vulputate, posuere urna luctus, porta sem. Phasellus condimentum, ante non mattis tempus, ex eros ullamcorper urna, eget rhoncus nisi neque sed tortor. Nam vitae ante at dolor porttitor porttitor. Cras non semper ipsum. Sed venenatis vel leo sed ornare. Donec a felis sed lectus aliquam mollis. Etiam laoreet malesuada urna sed pulvinar. Pellentesque quis quam nisi. Praesent odio ante, tincidunt ut interdum ut, rutrum vel lorem. Donec ut odio eget dolor ultrices mattis. Ut maximus lectus at felis pretium, id ornare tellus malesuada. Nulla eget consectetur massa. Nam a feugiat leo, quis luctus mauris. Aenean in tincidunt eros.

Fusce pretium, quam a molestie facilisis, massa mauris hendrerit lacus, et auctor ligula eros non libero. Ut quis gravida orci, a blandit erat. Donec pellentesque tempor nibh ut bibendum. Vivamus tellus felis, auctor nec rutrum vitae, eleifend et mi. Quisque lorem neque, finibus in malesuada sed, mollis eget metus. Sed lectus risus, venenatis ac dolor non, condimentum vehicula urna. Aliquam nisl sapien, tempor et mauris sed, imperdiet congue metus. Nunc fermentum condimentum lectus, a condimentum elit consectetur ac.

Pellentesque ullamcorper interdum lectus, a eleifend magna hendrerit eu. Quisque iaculis, lectus aliquam consequat mollis, urna dui bibendum dui, eu placerat sem ipsum et purus. Morbi mauris eros, dapibus eget purus vitae, blandit maximus quam. Quisque sed nibh et risus suscipit dignissim vel id dolor. Maecenas faucibus tellus vel massa volutpat sollicitudin. Ut ornare, enim sed ullamcorper posuere, dolor mi rhoncus ligula, sit amet suscipit nulla felis id ante. Suspendisse potenti. Fusce ac lectus elementum, sollicitudin ligula eget, faucibus lectus. Mauris placerat elit a turpis malesuada, quis pharetra eros vulputate.

Donec sed purus lacus. Sed maximus risus vitae ullamcorper tristique. Suspendisse potenti. Aliquam ac eros eget felis facilisis porta. Fusce hendrerit leo quis rutrum cursus. Nam sed sapien et massa cursus ultrices eget nec nisl. Nunc vehicula aliquet felis, vitae hendrerit arcu. In hendrerit enim at dui auctor condimentum. Maecenas at libero malesuada, fermentum orci in, vehicula arcu. Duis at massa eget risus pretium dignissim a id leo. Praesent laoreet tempor leo at pharetra. Suspendisse eros nunc, mattis sit amet leo nec, lacinia accumsan ex. Phasellus lobortis sapien sed volutpat mollis.

Mauris sollicitudin neque id metus vulputate mollis. Fusce ut volutpat augue. Donec eget nulla nec massa congue consequat. Ut vel posuere sem. Phasellus imperdiet elit vel dolor vestibulum ultrices. Quisque at mi pellentesque, tincidunt purus ac, laoreet metus. Suspendisse dignissim lorem ac ex suscipit imperdiet. Duis ut arcu sit amet eros porta ullamcorper. Sed felis leo, luctus sed lacinia ut, pretium tincidunt ex. Donec eu gravida libero. Fusce sagittis ullamcorper turpis, vel commodo orci elementum interdum. Donec tristique justo non ante ultrices, at egestas nisl sodales. Sed vitae bibendum risus.

Sed auctor non augue vitae accumsan. Interdum et malesuada fames ac ante ipsum primis in faucibus. Nullam ultrices libero id dapibus ultrices. Ut sit amet elit lectus. Vestibulum eu vestibulum lacus. Donec vel risus id diam varius bibendum. Duis et urna sed nisi porttitor fringilla in hendrerit lectus. Etiam fermentum sapien sit amet magna gravida posuere. Curabitur luctus ante lacinia laoreet mollis. Vestibulum dapibus efficitur tempor. Donec ullamcorper, diam sit amet varius sagittis, erat purus congue sem, vitae sollicitudin orci dui non nisl. Etiam volutpat felis id ligula elementum molestie. Ut ut interdum augue, et laoreet libero.

Aliquam non orci nisl. Nulla vel sem non est ornare tristique. Pellentesque aliquet turpis mauris, non convallis ex gravida nec. Nullam vitae ex eleifend, dignissim magna vitae, pretium felis. Interdum et malesuada fames ac ante ipsum primis in faucibus. Nunc consequat cursus pretium. Aliquam sit amet scelerisque risus, et varius quam.

Sed rutrum vestibulum suscipit. Aenean a imperdiet dolor. Quisque interdum erat eu velit finibus, ac iaculis ipsum bibendum. Nullam dui urna, ultrices blandit fermentum et, faucibus a urna. Integer consectetur neque id tempus semper. Ut sit amet tellus mauris. Sed consequat sapien ac mauris ornare posuere. Duis ultrices viverra urna sed ultrices. Phasellus eu velit iaculis, vulputate orci bibendum, tincidunt ante. Sed sollicitudin metus sed massa euismod pellentesque. Donec dictum dolor quis quam condimentum posuere. Vivamus sed tincidunt risus. Praesent nec nibh vitae risus lacinia auctor eget sed lacus.

Donec mi arcu, rutrum eu scelerisque eget, finibus sed neque. Donec et varius est. Morbi at velit sed ante congue pretium eu molestie ligula. Aliquam pharetra ullamcorper lobortis. Ut tempor massa vel orci eleifend, vel gravida leo faucibus. Sed bibendum massa lacinia velit varius, et scelerisque ipsum ornare. Ut mattis ullamcorper orci nec ornare.

Morbi ut velit mollis, consectetur mi ac, elementum erat. Suspendisse maximus, leo non ornare eleifend, quam lectus dapibus nulla, imperdiet sollicitudin elit ante vel leo. Sed eu molestie ligula, ut pretium velit. In eleifend suscipit egestas. Curabitur non imperdiet massa. In et dolor porta, dictum turpis non, rutrum nunc. Pellentesque quis nisi tellus. Integer placerat porttitor pulvinar. Maecenas viverra maximus lorem tempor venenatis. Quisque eget sem felis.

Cras bibendum porta hendrerit. Donec feugiat nunc mauris, non accumsan tortor imperdiet ut. In gravida sapien ut hendrerit commodo. Cras tempor vestibulum porta. Ut blandit dignissim arcu nec consequat. Maecenas vitae eleifend sapien. Aliquam commodo malesuada ligula non maximus. Proin venenatis gravida mi, et pulvinar massa blandit vel. Vestibulum sed nisi ante. Nam id pharetra ligula, at convallis ipsum.

Ut sed vehicula augue. Vivamus nec maximus est. Sed a euismod justo, ut bibendum arcu. Etiam porta ex id odio consequat gravida. Nulla facilisi. Praesent eu lacinia eros. Praesent eget ante lacus. Sed iaculis blandit tempor. In ut congue lorem. Pellentesque habitant morbi tristique senectus et netus et malesuada fames ac turpis egestas. In porttitor ex at vehicula pulvinar. Morbi imperdiet blandit ullamcorper. Vivamus ante sem, vulputate et nunc a, luctus elementum massa.

Quisque vitae lacus quis sem elementum aliquam et nec enim. Fusce porttitor viverra ipsum nec ornare. Sed est dolor, sollicitudin vitae nisi eget, porttitor efficitur purus. Maecenas tincidunt faucibus mattis. Cras a dui id eros posuere malesuada in ut felis. Proin in mauris condimentum, elementum nunc vel, auctor augue. Praesent arcu eros, sagittis nec nibh sed, placerat interdum mauris. Phasellus vel leo justo. Curabitur vitae est in velit pretium placerat. Pellentesque quis lacus tellus. Mauris interdum pulvinar massa vitae ultricies. Etiam a tempor neque, eu facilisis diam.

Vivamus gravida vel nibh ut viverra. Phasellus quis ornare sem. Phasellus venenatis dui quis mattis ultricies. Vestibulum dignissim odio eleifend, vulputate nisl eget, ullamcorper leo. Ut fringilla, justo nec sodales efficitur, est ligula varius tellus, id fermentum nibh turpis eu ex. Aliquam eu tellus ac tellus consectetur vehicula ac a dolor. Cras ut lectus sit amet orci dapibus tristique non sit amet odio. Vivamus quis erat sodales, lobortis metus eget, condimentum nisi. Donec efficitur eros quis risus dictum, ac luctus augue suscipit. Pellentesque iaculis, risus sed feugiat pretium, lorem justo venenatis mi, et consequat libero quam venenatis dolor.

Nulla fringilla egestas lorem, id euismod sem ultricies at. Nunc suscipit convallis odio eu vestibulum. Donec suscipit nisl vitae ligula porttitor placerat. Sed commodo rutrum eros, sit amet elementum odio cursus sed. Nulla vehicula volutpat justo, tincidunt ullamcorper enim pulvinar ac. Etiam at imperdiet tellus, id feugiat lorem. Nullam quis mauris in mauris consectetur hendrerit sit amet ac arcu. Duis a cursus augue. Duis egestas accumsan erat, nec varius risus accumsan eget. Nulla in vestibulum quam, in consectetur velit.

Vestibulum ut vehicula odio. Nullam non ante sapien. Curabitur urna odio, viverra ut lobortis consectetur, tristique sit amet est. Sed dapibus augue lorem, quis blandit arcu accumsan sed. Nam justo lacus, hendrerit id sollicitudin eu, ornare sed risus. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed viverra efficitur convallis. Proin turpis massa, aliquam et turpis at, posuere lacinia libero. Duis pulvinar dignissim ante, lacinia molestie ex commodo eget. Pellentesque pulvinar, tortor vitae faucibus tempus, tortor massa fermentum justo, et vehicula mi purus id est. Etiam tempus posuere elit nec sagittis.

Vestibulum sed sodales odio. Quisque dapibus accumsan ante in eleifend. Mauris diam nibh, cursus eu placerat sit amet, congue et mauris. Sed vestibulum varius orci ac sodales. Cras venenatis nisi vitae dignissim bibendum. Sed lacinia commodo ipsum, rhoncus condimentum dui. Aliquam volutpat neque sit amet dolor rutrum mollis. Sed ipsum odio, egestas eget nunc id, posuere interdum felis. Suspendisse finibus dui nisl, eget laoreet arcu molestie ut. Nulla facilisi. Nulla a turpis ut nunc scelerisque vestibulum eget a lacus. Sed molestie nisl non quam luctus, et varius libero dignissim.

Etiam mattis sit amet velit sed vestibulum. Nunc at dolor in ipsum convallis dignissim. Vestibulum quis erat condimentum, malesuada velit id, condimentum felis. Nam ac nulla dictum, convallis sem et, suscipit velit. Vestibulum vitae est pellentesque, pulvinar libero eget, dignissim velit. Nullam laoreet mi eget elit mollis, at bibendum orci aliquam. Nulla in diam efficitur, luctus arcu non, accumsan enim. Pellentesque ut commodo erat. Vivamus egestas tincidunt lacus, a dictum purus pretium non. Quisque auctor eros in fermentum convallis. Proin lacinia risus nisi, sed sagittis velit convallis ac. Morbi porttitor euismod libero, non laoreet velit ultricies vel. Maecenas eu pretium nulla. Suspendisse consectetur purus eu leo facilisis, in gravida orci ultricies. Curabitur laoreet nisl a eros eleifend elementum.

Pellentesque ullamcorper eu nunc nec varius. Nulla a rhoncus velit. Nunc rutrum eu ex vitae molestie. Vestibulum euismod nisl tincidunt purus bibendum egestas. Aliquam lectus velit, maximus vitae iaculis et, rutrum non justo. Vestibulum nisi nibh, pellentesque eget varius sit amet, tristique et mi. Maecenas a semper ligula. Fusce porttitor placerat volutpat. Proin suscipit nisl et augue finibus, sit amet convallis dolor congue. Vestibulum ante ipsum primis in faucibus orci luctus et ultrices posuere cubilia curae; Duis quis scelerisque diam, vitae ornare urna.

Quisque in sodales turpis. Integer malesuada accumsan risus id lobortis. Pellentesque congue vulputate elit, vel accumsan velit feugiat sed. Nullam luctus tempor odio, quis lacinia ex eleifend ut. In leo odio, eleifend vel odio eu, ultrices pellentesque ipsum. Integer maximus nibh in velit sodales, a dignissim arcu tincidunt. Nunc egestas lectus ac diam facilisis gravida. Cras tristique malesuada condimentum. In accumsan, sem vel pellentesque aliquam, ante dolor tempus ipsum, congue dictum est sapien sit amet libero. Mauris porttitor sollicitudin interdum. Nam nec ante egestas, ornare urna nec, lacinia arcu.

Vestibulum elementum ultricies semper. Fusce finibus ex dignissim ipsum congue, non dictum enim mattis. Pellentesque id placerat magna, vel volutpat turpis. Donec sit amet tellus in urna porta aliquet auctor at est. Morbi ut dui ac ex vulputate bibendum. Aliquam ut fringilla diam. Sed mattis pellentesque elit. Morbi ac semper massa. Quisque eget vehicula odio. Quisque diam eros, gravida eget ante eget, porta tincidunt quam. Nulla justo nulla, aliquet vel tincidunt sed, mollis quis erat. Sed ligula elit, tempor in mi id, accumsan posuere tellus. Curabitur malesuada placerat condimentum. Suspendisse ac augue quam. Nullam sit amet accumsan mi. Maecenas non molestie lacus.

Sed efficitur nisl eros, id congue ipsum suscipit in. Maecenas at convallis risus. In suscipit suscipit euismod. Cras semper sem vel facilisis accumsan. Vestibulum rutrum finibus porttitor. Maecenas quis ex vel magna semper sagittis at a justo. Maecenas hendrerit enim nec elit pellentesque, eget dapibus lacus consectetur. Mauris elit justo, aliquam a odio ac, viverra interdum urna. Nunc maximus nisi odio, id accumsan risus sollicitudin dapibus. Aliquam id euismod tellus, a volutpat arcu. In sit amet odio diam. In hac habitasse platea dictumst. Aenean vehicula felis lectus, vel vestibulum nisl eleifend non.

Ut tincidunt elementum lacus nec pulvinar. Nam tincidunt consectetur neque a euismod. Aliquam erat volutpat. Fusce quis magna sed nulla ultrices ultricies vitae vel ex. Mauris nec commodo dolor, id vulputate nulla. Integer quam mauris, lobortis eget hendrerit a, bibendum eget mi. Vestibulum ante ipsum primis in faucibus orci luctus et ultrices posuere cubilia curae; Cras faucibus lorem sed mattis accumsan. Vestibulum nibh felis, pretium non metus sed, consequat fermentum nulla. Morbi a pellentesque enim. Praesent finibus augue eget malesuada iaculis. Ut nec ornare mauris.

Aliquam erat volutpat. Fusce tellus urna, gravida non pharetra in, sollicitudin ac turpis. Suspendisse potenti. Sed venenatis finibus ligula, sed tincidunt sapien. Etiam aliquam mi efficitur condimentum ullamcorper. Praesent non suscipit urna. Sed porttitor vitae magna ut tempus. Fusce at risus maximus, facilisis augue quis, rhoncus arcu. Nullam faucibus lacinia sapien, id facilisis justo tincidunt eget. Integer ac risus tristique, congue erat vitae, sollicitudin neque. Donec ut neque sapien.

Fusce placerat erat neque, sit amet aliquam nibh auctor sed. In nunc urna, vehicula sed mollis quis, luctus quis nisl. Cras et felis turpis. Duis pellentesque vestibulum odio eu eleifend. Morbi augue ex, lacinia non leo at, feugiat molestie augue. Ut vel convallis lectus, et maximus libero. Curabitur malesuada nibh ac fermentum bibendum. Maecenas vitae massa nibh. Cras eget mauris sit amet libero dictum rutrum. Praesent vel ullamcorper ligula. Aenean sollicitudin lacus sed mi varius euismod. Etiam ultricies, enim et ornare eleifend, eros risus mollis elit, nec venenatis est metus at quam.

Vivamus vitae nibh tincidunt, accumsan nunc in, sollicitudin tortor. Mauris at consequat orci. Duis auctor tincidunt eros at ornare. Fusce quis feugiat dolor, a ornare ante. Proin commodo urna urna, in interdum felis posuere eu. Ut ut mi ornare dui aliquam auctor. Nulla nisl sapien, varius ac maximus ut, viverra ac augue. Maecenas dignissim leo a finibus tincidunt. Sed facilisis imperdiet elit, ornare ultricies leo feugiat sed. Mauris urna orci, scelerisque eu justo eget, ultricies laoreet dui. Morbi porta turpis vel nunc malesuada, venenatis malesuada nisi ullamcorper. Quisque volutpat mauris et scelerisque tempor.

Proin quis tortor et erat tempor semper pharetra id mauris. Phasellus quis finibus sem, non porta arcu. Quisque mattis, lorem ut volutpat fermentum, ligula metus imperdiet massa, at consequat lacus nisi vitae augue. Maecenas rhoncus elementum ante, ac faucibus augue vulputate id. Vestibulum lacinia dolor ut quam tincidunt fringilla. Nunc ac leo quis quam porta accumsan. Suspendisse sodales eleifend fringilla. Praesent sagittis aliquam tortor et consectetur. Praesent nulla nisi, pretium a magna quis, venenatis volutpat mi. Integer erat erat, egestas tempus pretium sit amet, consequat non mi. Nulla vulputate vehicula tristique. Duis quis luctus orci, ac viverra nisl. Mauris et varius turpis.

Phasellus eu aliquam libero. Aenean scelerisque, tellus sit amet ornare interdum, metus nunc auctor tellus, eu rutrum lorem urna ac turpis. Nulla tristique convallis dui, vehicula pretium lacus cursus quis. Integer pretium ultrices dignissim. Integer ac dignissim tortor. Etiam vitae condimentum justo. Integer rutrum nulla lorem, in condimentum nibh pellentesque sit amet. Suspendisse eu iaculis odio, in auctor neque.

Sed a magna metus. Etiam eu consequat dolor. Etiam vel tellus eu lorem auctor commodo ut ac augue. Vestibulum aliquam sodales leo a tempor. Sed et facilisis augue. Sed congue maximus velit nec semper. Sed volutpat nec odio quis euismod.

Sed ornare feugiat nibh sit amet scelerisque. Nulla hendrerit mi eu turpis laoreet condimentum. Suspendisse suscipit consectetur arcu non maximus. Integer eget ultricies magna. Curabitur sollicitudin ante vitae urna posuere commodo quis nec odio. Donec commodo sed justo id pretium. Nam cursus mauris at quam gravida ornare. Curabitur nisl turpis, viverra nec facilisis ac, egestas vel velit. Morbi vitae semper dolor. Duis metus nibh, scelerisque a est vitae, volutpat lobortis purus. Curabitur rutrum felis ac tempor tincidunt. Phasellus blandit diam eget augue iaculis, vel condimentum tellus placerat.

Class aptent taciti sociosqu ad litora torquent per conubia nostra, per inceptos himenaeos. Sed pellentesque nunc in ipsum finibus hendrerit vel a orci. Nunc congue nunc sit amet nibh volutpat, non ultrices dui vestibulum. Curabitur eget fermentum elit. In hac habitasse platea dictumst. Cras laoreet nunc a convallis sodales. Aliquam imperdiet viverra libero id condimentum.

Pellentesque consectetur et ex in eleifend. Sed posuere lorem sed lacus condimentum varius. Maecenas a nibh quis lectus tempor gravida. Morbi imperdiet rhoncus velit in eleifend. Mauris neque ex, condimentum sit amet fermentum eget, lobortis eget risus. Nunc lacinia ex in imperdiet tempor. Aenean feugiat risus eget felis rutrum molestie. Sed non varius elit, id accumsan augue. Aliquam mollis pretium diam. Morbi nec turpis sollicitudin, imperdiet mauris ut, pharetra ex.

Quisque id fermentum dui. Vivamus porttitor tortor enim, malesuada tincidunt dui vehicula et. In sed rutrum dui. Fusce id tincidunt tortor. Donec vel convallis sem. Curabitur mattis lacinia purus nec pretium. Vestibulum lobortis aliquet ullamcorper. Vivamus condimentum orci dui, scelerisque pellentesque leo suscipit in. Etiam pulvinar arcu nunc.

Suspendisse condimentum lobortis sodales. Pellentesque vehicula, dui ut pellentesque varius, tortor purus aliquam ligula, ut efficitur est dolor rhoncus nibh. Proin sem dui, vehicula ac enim id, malesuada ultrices nisl. In id eros enim. Praesent sit amet risus lectus. Vestibulum sagittis non ante id dapibus. Aliquam tortor odio, vulputate sit amet sem et, luctus gravida elit. Proin in lectus nec elit sodales placerat quis sed sapien. Nulla placerat nisl eros, sed egestas enim vestibulum quis. Sed tempus vulputate facilisis. Vivamus lobortis lacus ut ligula tristique, vel pharetra lacus faucibus. Vivamus gravida enim ac magna venenatis auctor. Nam condimentum orci et tincidunt posuere.

Sed laoreet metus vel neque fermentum vehicula. Pellentesque leo mauris, vestibulum id pellentesque non, condimentum non metus. Vestibulum molestie, metus a molestie condimentum, elit quam eleifend leo, sed feugiat dolor augue quis erat. Cras tempor libero vitae mattis lacinia. Morbi tellus dolor, mattis a facilisis vitae, feugiat vel nisl. Vivamus mollis nisi metus, eu tristique justo lobortis nec. Curabitur egestas rutrum ante dictum dapibus. In eu quam vel sapien finibus vehicula. Fusce sit amet fermentum nisi. Vestibulum ornare a purus a ullamcorper. Suspendisse vel dolor tellus. Fusce molestie nec tortor sit amet varius. Vestibulum tincidunt felis eu elit mattis volutpat. Ut bibendum mauris quis arcu imperdiet, in commodo odio sodales.

Aliquam nec lacus et tortor dapibus pretium sit amet ac ex. Fusce ac nibh turpis. Fusce a auctor risus, id sollicitudin metus. Pellentesque iaculis dictum elit vel placerat. Vestibulum feugiat justo viverra, faucibus erat in, hendrerit elit. Vestibulum eget enim lacus. Duis malesuada et ligula et venenatis. Proin dictum ante non vehicula elementum. Phasellus at tristique velit. In nulla arcu, laoreet eu laoreet id, porta vel metus. Nam pretium tempus urna, et tempor metus.

Maecenas lacinia ullamcorper nunc sed malesuada. Curabitur non ullamcorper nulla. Duis bibendum odio et diam feugiat scelerisque. Nam efficitur sed mauris sit amet vehicula. Cras leo mauris, consequat in porttitor a, ultricies vel sapien. Integer bibendum in eros a commodo. Praesent pharetra sem in mi dignissim efficitur.

Sed dapibus ligula et nibh vulputate, in efficitur mauris pharetra. Donec venenatis tempus elit non gravida. Etiam eget semper magna. Aliquam quis tellus ut elit rhoncus tempor in et sapien. Nullam interdum bibendum pellentesque. Vivamus pharetra justo a nisi dignissim, nec cursus lorem condimentum. Mauris quis elit nec leo vulputate porttitor. Praesent a magna ut erat consectetur efficitur. Sed est turpis, interdum nec sodales vel, finibus quis mauris. Nunc fringilla vehicula nisl, vel fermentum lacus porttitor at. Sed metus felis, maximus sit amet convallis ultricies, bibendum at urna.

Nunc finibus, metus nec malesuada ornare, nisl enim ultricies justo, nec rutrum mauris augue quis nisi. Proin rhoncus sapien at felis vehicula porttitor. Ut iaculis luctus vestibulum. Donec volutpat ipsum vel massa tincidunt venenatis. Sed suscipit eu ex porta dignissim. Maecenas ac tincidunt nunc. Donec vel tortor lectus.

Nunc condimentum aliquet metus, eu gravida elit pulvinar id. Sed tincidunt scelerisque libero, et congue nunc. Aenean interdum nisi in ipsum dapibus, eget egestas ipsum cursus. Sed placerat sollicitudin enim non vehicula. Vivamus gravida, enim eu mollis fringilla, est purus ornare arcu, in elementum odio enim ullamcorper augue. Aliquam quis nunc aliquet, bibendum sem a, ornare lorem. Suspendisse sed metus non lacus cursus varius eu sit amet metus. Ut dolor turpis, luctus eget massa non, tempus ultrices purus. Donec a ipsum fringilla, semper enim a, viverra lorem. Nunc varius neque vel nunc porttitor lacinia. Ut blandit justo ac mauris tempor, eget eleifend ipsum interdum. Integer turpis nunc, porta sed neque quis, consequat ultrices massa. Integer sed tortor sed justo mollis semper. In nulla urna, faucibus hendrerit tortor vel, malesuada mattis augue. In faucibus tempus auctor. Sed tempus purus sed orci vestibulum cursus.

Donec tincidunt diam et odio rhoncus, a fermentum nulla suscipit. Fusce feugiat mi arcu, eget laoreet lorem ornare consectetur. Integer ac bibendum felis. Vivamus aliquam libero sit amet nunc ultrices egestas. Duis suscipit erat tristique, fringilla tellus ac, hendrerit velit. Donec feugiat viverra faucibus. In mattis elementum mi, at mattis ex convallis et. Nam vitae nisi eleifend, placerat nunc in, scelerisque ligula. Duis nec consequat dolor, at viverra orci. Phasellus vulputate eu risus quis ornare. Integer non libero consequat, aliquam mi non, ultricies ipsum. Nullam mattis ligula purus, posuere venenatis libero pulvinar et. Praesent urna nulla, condimentum eu erat non, elementum fermentum est. Maecenas fermentum rutrum volutpat. Proin id lobortis augue, vel imperdiet leo.

Sed id neque at ligula accumsan rutrum quis eu massa. Praesent eget justo iaculis, malesuada justo ut, molestie erat. Maecenas vulputate efficitur elit. Nam congue metus id ipsum suscipit laoreet quis vitae tellus. Suspendisse tincidunt dignissim orci ac vestibulum. Cras dignissim elit non enim pretium elementum mattis a purus. Nulla facilisi.

Aliquam quis ex arcu. Donec ac fermentum purus. Nunc aliquet elementum bibendum. Morbi ornare vehicula lacus. Integer aliquet diam sit amet magna sagittis, vitae congue felis fermentum. Vestibulum eu posuere nunc, sit amet luctus augue. Nulla eu tincidunt metus. Phasellus dapibus lacus at leo volutpat, sed convallis velit suscipit. Ut urna mi, tincidunt vitae eleifend ut, mollis non nibh. Etiam nec dui eu ante accumsan bibendum. Sed accumsan laoreet mi eu placerat. Sed rhoncus, nisi sit amet malesuada maximus, turpis magna aliquam tortor, id tincidunt lectus augue in ligula. Morbi id lorem interdum, ullamcorper eros a, volutpat tellus.

Orci varius natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Suspendisse sem erat, vehicula vel nibh eu, venenatis sagittis ante. Proin pretium massa sit amet consectetur elementum. Pellentesque volutpat volutpat tellus, mattis vestibulum neque dictum nec. Cras consequat justo eu auctor sollicitudin. Phasellus facilisis sed dolor sed feugiat. Nam interdum leo vel diam interdum eleifend. Praesent facilisis et nulla vitae ornare. Nunc at auctor est, a tempor ligula. Sed luctus pellentesque orci, nec tempor neque convallis non. Proin interdum ullamcorper sodales. Nulla congue lacus at felis fermentum, dictum accumsan orci tincidunt. Sed condimentum consectetur lectus sed luctus.

Donec consectetur, magna at lacinia facilisis, eros quam aliquam turpis, et molestie diam dolor id nibh. Vivamus ligula nunc, facilisis at elementum in, tempor id ante. Ut blandit luctus arcu mattis dictum. Sed a consectetur lacus, lacinia feugiat purus. Morbi maximus eros sit amet enim vestibulum placerat. Proin aliquam vestibulum ligula, at pulvinar diam ultricies id. Nunc eget nunc non nisl laoreet interdum ut nec mi. Donec sed hendrerit turpis.

Integer euismod rutrum mi, eget auctor justo. Nam eu ultrices metus, at lobortis urna. Curabitur vulputate sem ac quam iaculis hendrerit. Pellentesque pharetra lacus vitae purus eleifend hendrerit. Pellentesque ac porttitor ipsum, nec euismod leo. Interdum et malesuada fames ac ante ipsum primis in faucibus. Donec in iaculis quam, pharetra interdum mauris. Proin ante urna, posuere vitae ullamcorper in, dictum ut tortor. Integer ac metus nisl. Phasellus rutrum, nisl vel cursus convallis, ipsum ligula vulputate massa, nec facilisis nulla mauris a arcu.

Nullam tristique orci at ante blandit suscipit. Vivamus condimentum ultrices libero id gravida. Duis mollis convallis nibh, vel lobortis nisl lobortis eget. Nulla ultricies lectus porta risus sagittis, eget malesuada erat ultrices. Aenean vulputate, nulla at pulvinar mollis, nisi leo efficitur massa, at accumsan arcu diam at odio. Aenean ultrices accumsan elit non bibendum. Donec facilisis lectus sed libero facilisis, et tempus nisi ornare. In volutpat nisi in lectus mattis cursus. Proin blandit ex venenatis augue sagittis feugiat. Nunc in quam nec risus tincidunt ullamcorper id ut sem. Aliquam sagittis, augue in fringilla tempus, justo enim dictum sapien, eget pretium urna massa vitae nulla.

Duis rutrum neque a rhoncus commodo. Cras vitae libero et sem venenatis feugiat vitae nec elit. Duis non porttitor elit. Ut et aliquet dolor, nec rhoncus nisl. Suspendisse potenti. Quisque volutpat mollis magna eget mollis. Sed posuere purus ante, at tempus magna hendrerit sed. Morbi varius, nisl eu convallis fringilla, orci mauris ullamcorper ante, vitae ultricies eros nisl a arcu. Donec porta mi at dapibus blandit. Aliquam et accumsan nisi, et suscipit lorem. Curabitur sit amet turpis diam. Etiam a magna viverra, suscipit purus at, dignissim urna.

Phasellus sollicitudin ante in est ultrices, vel dictum velit ultricies. In hac habitasse platea dictumst. In nec neque euismod, varius est vel, accumsan tellus. Pellentesque porta, nisl vel consectetur gravida, ante magna finibus erat, et mattis lorem purus non elit. Donec laoreet vehicula dui, id rhoncus velit volutpat et. Sed id rutrum nulla. Nulla finibus pharetra dui, at tempor erat convallis sed. In elementum elit dolor, finibus facilisis diam ultricies non. Suspendisse egestas nisl turpis, id interdum nibh euismod nec. Nam vitae tristique est. Etiam rhoncus ante commodo sollicitudin dapibus. Nulla gravida nibh sed augue laoreet dignissim. Pellentesque sapien augue, tristique eget purus sit amet, maximus ornare sapien. Praesent varius arcu enim, sed hendrerit sapien iaculis at.

Aliquam erat volutpat. Nullam non elit vulputate, tempor nisi ut, egestas magna. Ut facilisis viverra dui non pellentesque. Morbi faucibus, nibh in lobortis aliquam, leo ipsum molestie ipsum, luctus elementum nisl nunc et nulla. Nunc lacinia libero mauris, sed elementum erat consectetur vitae. Duis ac efficitur massa. Maecenas rutrum, ligula ac iaculis placerat, ligula nisi egestas massa, sed vulputate nibh quam eget velit. Sed volutpat vestibulum scelerisque. Sed faucibus diam turpis, vel scelerisque sem egestas ac. Phasellus sed consectetur mi. Nullam at magna purus. In id nunc ut lorem commodo faucibus et eget turpis.

Donec eu laoreet neque, ut venenatis dolor. Integer nulla ex, finibus nec dictum sit amet, iaculis quis lectus. Duis condimentum a augue ut tincidunt. Vivamus viverra augue et aliquet rutrum. Ut euismod quam in ex tempus, et placerat diam rhoncus. Pellentesque habitant morbi tristique senectus et netus et malesuada fames ac turpis egestas. Nam ut magna metus.

Vestibulum sollicitudin metus eu justo feugiat pellentesque. Vestibulum commodo sapien mattis massa lacinia, non dapibus augue semper. Maecenas elit purus, imperdiet id purus ut, aliquam blandit risus. Sed maximus malesuada elit vel elementum. Cras ac ante fermentum, semper sapien nec, pellentesque nibh. Fusce vulputate, neque non accumsan faucibus, elit erat varius quam, et suscipit ipsum libero ut eros. Donec cursus consequat aliquet.

Maecenas imperdiet efficitur elit vel faucibus. Integer quis justo pellentesque, vehicula lorem vel, posuere velit. Mauris auctor, lectus vitae ultrices rhoncus, odio mi tincidunt arcu, et tristique ligula elit sed enim. Pellentesque libero nisl, fermentum sit amet interdum nec, sollicitudin non metus. Fusce elementum condimentum feugiat. Pellentesque elementum, nulla ut consectetur pharetra, tortor risus consectetur risus, sit amet interdum libero neque vel nisi. Praesent eget ante nec nulla posuere laoreet. In hac habitasse platea dictumst. Ut eget tellus id nulla accumsan tincidunt. Aliquam erat volutpat. Aliquam erat volutpat. Maecenas pulvinar justo rutrum enim suscipit, quis suscipit leo interdum.

In porta metus nibh, vitae facilisis quam sodales ut. Pellentesque molestie interdum massa quis consequat. Pellentesque posuere mollis felis, ut tincidunt ipsum mollis eget. Fusce lobortis justo ac arcu porta, non facilisis risus eleifend. Phasellus dui ipsum, tincidunt vel vestibulum vel, pharetra a est. Aenean id nunc sed turpis dictum tristique sed et ipsum. Nunc vulputate, leo consectetur placerat aliquet, leo nulla lobortis risus, quis ultricies velit mi ac arcu. Nunc sed metus vitae risus accumsan ornare sit amet vitae nunc. Mauris vulputate erat ipsum, eget pulvinar odio euismod non. Sed volutpat urna ut leo imperdiet, eu porttitor erat vulputate. Maecenas eu urna nec diam aliquam placerat. Interdum et malesuada fames ac ante ipsum primis in faucibus. Pellentesque dignissim vehicula erat dapibus luctus.

Suspendisse nunc neque, sollicitudin et nunc et, sodales tincidunt arcu. Praesent porta, risus quis vestibulum ultricies, ligula quam porta ipsum, eu elementum ligula nisi non tellus. Proin eu tempor nisl. Proin molestie eget neque ut aliquet. Vestibulum felis ligula, placerat eget lectus quis, sagittis sodales turpis. Fusce ac massa rhoncus, luctus magna at, tincidunt erat. Aenean eleifend vel erat id dictum. Maecenas id sapien hendrerit, condimentum ante accumsan, pretium eros. Curabitur quis mauris eget nibh ultrices placerat. Donec mattis lorem orci, eget pellentesque sem bibendum vel. Pellentesque habitant morbi tristique senectus et netus et malesuada fames ac turpis egestas. Vivamus venenatis elit quis accumsan fringilla. Maecenas pellentesque pulvinar suscipit. Sed mattis accumsan posuere.

Mauris id dignissim quam. Phasellus tempus dolor a consectetur pharetra. Ut at interdum dolor, eu posuere purus. Donec efficitur vehicula ligula et semper. Morbi fermentum, orci at tempor lacinia, risus dui lobortis mauris, faucibus rhoncus massa eros ac magna. Ut in nisi porta, cursus arcu at, fringilla massa. Phasellus imperdiet, sapien a mattis ornare, metus turpis malesuada enim, vitae ornare nibh lacus at mauris. Sed at ornare libero. Cras a magna sed neque porta sollicitudin eget ut felis. In sit amet odio magna. Nunc congue efficitur iaculis. Nunc augue diam, fermentum vel tincidunt pretium, eleifend non justo. Vivamus lectus tellus, bibendum non sapien et, viverra tempus tortor.

Maecenas malesuada pretium dapibus. Nullam pharetra gravida tortor et scelerisque. Vivamus blandit ante sed tortor laoreet lacinia. Aliquam sodales ut erat ullamcorper pretium. Donec eget enim felis. Praesent congue, libero sit amet semper blandit, diam elit faucibus augue, vitae molestie nulla dui ut leo. Phasellus facilisis nunc id felis molestie, vel luctus dui bibendum.

Phasellus hendrerit tellus id arcu luctus blandit. Aliquam tincidunt libero in odio faucibus, vitae pellentesque lectus elementum. Mauris ac justo sed enim luctus mollis. Fusce rhoncus a diam tempor fermentum. Maecenas et urna aliquet, tristique metus molestie, accumsan purus. Nulla id euismod velit, non convallis urna. Nullam pulvinar, nisi a aliquet dapibus, orci felis mattis lectus, in convallis magna diam at ante. Morbi ultrices a tortor ac sollicitudin. Integer quis suscipit sapien, quis eleifend lorem.

Vestibulum ante ipsum primis in faucibus orci luctus et ultrices posuere cubilia curae; Maecenas bibendum dolor at elementum pulvinar. Donec in sapien lacus. Vivamus sed enim ipsum. Nam facilisis lobortis bibendum. Phasellus at pretium massa, a placerat sapien. Ut aliquet turpis ac dolor gravida finibus. Pellentesque at volutpat diam, vitae aliquet nisl. Vivamus dapibus tellus pharetra, tempor odio eget, ultricies mauris. Ut sagittis blandit nunc, eget pulvinar leo tempor ut. Proin scelerisque erat vel enim auctor interdum. Praesent lacus erat, dapibus eget rhoncus ac, vehicula vel metus. Nam feugiat, erat id euismod bibendum, lacus est varius neque, eu posuere lectus lacus non nisi. Fusce gravida risus eu laoreet posuere. Morbi mattis mattis lacus nec condimentum. Ut id mauris vel erat iaculis blandit.

Nam cursus accumsan nisi, eget scelerisque purus sodales eget. Sed tincidunt massa sit amet volutpat ultrices. Cras fringilla purus nec nibh bibendum aliquam. Integer efficitur blandit velit id molestie. Duis id gravida eros. Nulla id rutrum leo. Nullam massa velit, venenatis eleifend accumsan non, eleifend eget dolor. Donec eu luctus felis. Aliquam vel nisl nibh. Suspendisse varius dolor lectus, sit amet lobortis nisi mollis vel.

Nam lacinia interdum odio, et vehicula erat pretium sit amet. Praesent vitae sapien pulvinar orci commodo viverra. Nunc finibus dui vitae mauris congue, sed tempor nisl feugiat. Aenean ac lectus eget sapien malesuada imperdiet nec ut erat. Etiam iaculis viverra nisl at vestibulum. Morbi risus leo, scelerisque feugiat magna ut, pellentesque gravida libero. Curabitur finibus ante eu nisi condimentum, sit amet porttitor arcu elementum. Cras dictum justo vitae urna hendrerit fringilla. Vestibulum mollis rhoncus lacinia. Duis lobortis libero eu facilisis dignissim. Cras eu nibh sem. Phasellus pulvinar purus eu ornare laoreet. Proin a diam ac erat tincidunt pellentesque a sit amet est. Curabitur imperdiet massa vitae dolor congue luctus. Sed purus orci, venenatis a congue in, vulputate in purus. Mauris ultricies enim et purus volutpat, at tincidunt dolor tincidunt.

Cras mollis accumsan nulla. Vivamus fringilla sodales tortor. Donec justo diam, tempor at blandit viverra, feugiat quis velit. Nunc dapibus dictum massa, in tincidunt ante laoreet id. Vivamus eget nunc eu eros ultrices lobortis. Proin pharetra diam dictum mattis convallis. Quisque vitae eleifend turpis. Maecenas metus tellus, faucibus et ligula vel, auctor rhoncus eros. In efficitur aliquet sapien.

Nam id tempus arcu. Integer sed mi eu ex pellentesque iaculis ac in ligula. Suspendisse ullamcorper arcu ac massa commodo, eget mollis urna lobortis. Mauris placerat dolor tortor, ullamcorper sodales nulla finibus in. Duis eros lorem, vestibulum non gravida vel, commodo ac ante. Ut id mauris viverra, egestas nulla sit amet, eleifend dui. Integer sed dolor ut ligula placerat faucibus nec eget arcu. Praesent vestibulum venenatis augue, quis dignissim metus rhoncus at. Nam laoreet nisl sed leo volutpat faucibus. Morbi feugiat vitae nibh nec pretium. Nam accumsan facilisis lorem eget feugiat. Aliquam pretium lacus arcu, ut faucibus est hendrerit at. Donec lectus odio, consequat vitae nulla eleifend, tincidunt feugiat ex. Aenean lobortis velit a eros faucibus, luctus pretium neque sagittis. Mauris imperdiet iaculis ipsum ac elementum.

Sed nec purus ut urna viverra dignissim non id erat. Donec dictum interdum viverra. Curabitur sodales laoreet elit. Quisque at accumsan ante. Fusce et ultricies ligula. Duis varius ut sem quis hendrerit. Donec porta nisl sit amet sem consectetur, at scelerisque ante aliquet. Etiam eget lorem sollicitudin, pellentesque justo id, aliquet dolor. Praesent lacinia urna sit amet ipsum accumsan, ut porta lacus finibus. Cras condimentum accumsan ligula, a ullamcorper turpis lacinia quis. Etiam eu mauris a tortor auctor mollis. In mauris est, tempor varius neque vel, maximus condimentum arcu. Sed id enim vitae orci rutrum finibus sed id nunc.

Donec bibendum sodales quam quis placerat. Curabitur maximus elementum convallis. Fusce eleifend arcu non vulputate lobortis. Duis ut turpis nibh. Cras volutpat magna eget ipsum varius, eu tristique nulla rhoncus. Morbi sapien nisi, vehicula id risus vel, scelerisque venenatis lacus. Aenean accumsan sollicitudin metus, eu tempor dui viverra euismod.

Mauris sodales porta scelerisque. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Etiam pellentesque nulla ut mi mattis tincidunt. Praesent sit amet convallis eros, eu finibus libero. Aliquam id leo eu risus tincidunt vulputate. Nunc elementum in velit id vehicula. Fusce sit amet massa eu augue bibendum tincidunt eu in est. Nullam faucibus ultricies dui ac finibus. Nunc dapibus urna et arcu laoreet, et posuere dui euismod. Mauris vestibulum, orci quis aliquet pellentesque, erat mi varius mauris, eu mollis massa nisi et urna. Quisque sit amet odio leo. Quisque vitae cursus tellus, sed ultricies justo. Praesent sit amet accumsan nisl, ac aliquet dolor. Sed scelerisque ullamcorper tempor.

In laoreet pretium nibh, sed sollicitudin urna luctus id. Vestibulum tincidunt rhoncus nibh non viverra. Integer dignissim augue interdum, laoreet diam semper, sodales tellus. Vestibulum ante risus, vestibulum quis convallis ut, mattis ac risus. Nulla neque ante, luctus ut mattis quis, vehicula ac felis. Donec id enim quis erat fermentum commodo sed vitae sem. Nulla sed lobortis orci. Maecenas tincidunt, purus id pulvinar congue, nunc lacus ultricies leo, ut tempus augue leo at nulla. Nam lectus leo, consequat quis enim eu, rhoncus laoreet lorem. Nulla non enim vitae nunc faucibus consectetur. Integer iaculis commodo dolor, id ultricies tellus iaculis sit amet. Phasellus in sollicitudin tortor. Donec ac turpis porttitor, semper nisi vel, facilisis elit.

Mauris placerat, augue quis volutpat vestibulum, lorem erat molestie lectus, a tempus mi felis id lorem. Vivamus porttitor sapien eget erat cursus lacinia. Aliquam tincidunt congue libero. In hac habitasse platea dictumst. Praesent pellentesque sed erat nec mollis. Nam id est a metus sodales lobortis. Fusce nibh ligula, finibus id lacus at, maximus vehicula augue. Aenean sit amet tempus nunc. Duis eleifend et ligula sit amet pretium. Cras enim turpis, fermentum a eros nec, pharetra facilisis libero. Quisque eu vehicula libero. Donec feugiat mi at nunc finibus facilisis.

Proin dapibus nulla vitae justo mollis, sit amet venenatis diam dictum. Cras porttitor ac enim varius ultricies. Sed luctus vel purus in fermentum. Nam at ligula in orci accumsan sagittis non laoreet orci. Mauris quis ex non quam tempus pellentesque id sit amet lectus. Vivamus sodales, nunc et sagittis interdum, ligula urna malesuada purus, eget scelerisque justo neque gravida enim. Phasellus tincidunt malesuada leo ullamcorper finibus. Vestibulum faucibus felis dictum, tempus justo at, luctus neque. Proin ut orci at ex blandit ultricies sit amet in lectus. Praesent turpis ex, bibendum quis tortor vel, iaculis scelerisque orci. Lorem ipsum dolor sit amet, consectetur adipiscing elit.

Pellentesque sed urna sagittis, posuere ex non, consequat nisl. Quisque interdum eros ut nulla mattis consectetur. Ut nec neque consequat, efficitur quam quis, dictum odio. Donec enim urna, suscipit at sem ut, scelerisque viverra metus. Etiam lacus nulla, condimentum in purus eu, dapibus facilisis diam. In vel mauris non erat pellentesque porttitor vitae in nisi. Nullam sit amet blandit risus, ac aliquam massa. Nulla facilisi. Pellentesque et posuere est. Vestibulum vel egestas dolor. Cras rhoncus purus ut consectetur faucibus. Ut egestas, elit ut condimentum auctor, massa urna cursus orci, sit amet rhoncus nisl sapien a libero. Quisque commodo euismod leo et pulvinar. Donec molestie mauris ut felis placerat tempus.

Etiam varius metus enim. Donec eget est lobortis, scelerisque sapien at, pretium nulla. Fusce at massa cursus diam bibendum vulputate. Morbi ac quam sed ipsum faucibus finibus. Vestibulum egestas lacinia ante, et ornare augue feugiat sit amet. Maecenas finibus egestas diam, cursus posuere lacus luctus a. Aenean eleifend nulla ut hendrerit euismod. Quisque non ornare odio, nec dignissim lectus. Etiam vel sapien id nunc consequat placerat at vitae turpis.

Sed ultricies venenatis mi, a rhoncus tellus sollicitudin in. Praesent enim nisi, lacinia ac justo at, maximus ornare elit. Proin eget ultricies tellus. Fusce et ipsum et tortor placerat varius. Nunc malesuada nisl et sem tristique ullamcorper. Duis at sapien et purus sollicitudin feugiat quis quis tortor. Integer id sapien a libero posuere tempus. Vestibulum ut nisi eu neque convallis vehicula.

Pellentesque habitant morbi tristique senectus et netus et malesuada fames ac turpis egestas. Praesent semper lacus id ullamcorper tempor. Nullam nec neque diam. Fusce ac pretium dui. Donec eget condimentum massa, lacinia suscipit augue. Vestibulum ultricies est quam, at pulvinar nisi ullamcorper id. Duis et dapibus enim, sit amet sodales enim. Fusce volutpat metus a urna tempus porttitor.

Donec ullamcorper malesuada ultrices. Maecenas blandit augue in mi pretium, non molestie quam cursus. Pellentesque habitant morbi tristique senectus et netus et malesuada fames ac turpis egestas. Suspendisse mattis scelerisque ipsum a ullamcorper. Nam malesuada a eros vitae vehicula. Fusce lobortis ullamcorper nibh. Nulla luctus justo id erat rutrum, eu suscipit nibh fermentum. Phasellus hendrerit vulputate mollis. Nunc consectetur elementum dolor, et sollicitudin ex bibendum viverra. Mauris sodales elit non quam scelerisque, in efficitur dui maximus. Cras at maximus sapien.

Vivamus ac neque lectus. Aenean in felis orci. Nam blandit urna sit amet porta mollis. Nulla sed faucibus libero. Duis sollicitudin imperdiet congue. Pellentesque rutrum odio quam, sed faucibus nunc scelerisque id. Morbi eu laoreet odio, nec blandit metus. Mauris nec ultricies enim, vel egestas nibh. Donec tellus orci, pretium commodo elit venenatis, suscipit porta leo. Ut consequat id justo eu lobortis. Quisque quis ex sed turpis condimentum fringilla. Nam venenatis, dui sed sollicitudin ullamcorper, eros velit volutpat tortor, non tempor purus odio non orci. Morbi pharetra convallis sapien. Proin ut fringilla sapien.

Sed non risus ac ipsum hendrerit dapibus. Phasellus nulla eros, ultrices id quam at, commodo tristique magna. Suspendisse scelerisque ligula at finibus fermentum. Pellentesque sagittis, orci finibus ultrices efficitur, arcu mi placerat ipsum, in facilisis lorem felis non nisl. Quisque et accumsan nibh. Nunc in lorem et diam fringilla viverra eget nec ipsum. Curabitur posuere tincidunt diam. Duis sodales elit eget eros ullamcorper gravida. Nullam faucibus sapien a ipsum sollicitudin fringilla. Nunc non condimentum diam. Mauris volutpat turpis ut eleifend accumsan. Integer vel suscipit sem. Cras nec volutpat neque. Nulla sodales vitae orci sed facilisis. Quisque pharetra augue nunc, eget sodales est eleifend eu.

Quisque elit velit, ultrices ut lectus quis, mollis eleifend magna. Curabitur sodales semper lobortis. Fusce dapibus ante vitae sollicitudin tincidunt. Nullam interdum libero quam, at lacinia metus ornare nec. Fusce efficitur commodo lorem in vestibulum. Quisque tincidunt nibh vel ante blandit, quis eleifend urna luctus. Proin a fringilla leo. Pellentesque tempor elementum viverra. Vestibulum non leo finibus, commodo elit ut, dictum justo. Aenean ornare est quis mi pretium dignissim. Aliquam faucibus, metus at tempus cursus, erat mauris eleifend enim, id tincidunt magna nisi nec arcu. Ut porttitor, nisl tempor pretium pulvinar, odio est blandit ante, sit amet lacinia nisl dui in sapien. Donec tempus nisl ut pharetra facilisis. Aenean rutrum magna ut erat lobortis, non congue quam egestas.

Donec faucibus in massa a rutrum. Nulla sit amet nibh consequat, sodales felis eu, euismod purus. Etiam aliquet ligula vitae dolor fermentum pharetra. Integer tincidunt magna vel ante aliquam, eget luctus tellus dignissim. Aliquam vulputate dolor non tristique dictum. Phasellus eu mattis mi. Ut est mauris, porttitor non tellus sit amet, eleifend commodo purus. Duis ultrices nulla ut urna placerat, ornare cursus felis rhoncus. Aenean cursus sollicitudin dolor, eu dignissim ex sollicitudin sodales. Praesent eu nisl neque. Vivamus eleifend eu dui non sollicitudin. Quisque pharetra aliquet tincidunt. Morbi neque lacus, fringilla dapibus risus a, facilisis varius diam. In hendrerit enim eu tempor maximus. In ut ex eu nisi egestas rutrum pretium at orci. Sed maximus, nunc id placerat rhoncus, ante ligula molestie nulla, at lacinia velit ligula non justo.

Donec vitae quam ante. Vivamus tincidunt porttitor blandit. Suspendisse potenti. Phasellus aliquet hendrerit ex, vitae semper elit vulputate nec. Phasellus id velit quis nunc fermentum placerat et non libero. Etiam commodo elementum augue, id auctor magna egestas at. Quisque bibendum ipsum tortor, eget viverra leo venenatis a. Aliquam sem eros, scelerisque sit amet accumsan quis, scelerisque vel justo. Duis tempus lacus et purus porta, a faucibus lorem ultrices. Duis in enim ipsum.

Mauris sit amet bibendum purus. Suspendisse et consequat turpis, vitae porta risus. Nullam accumsan dolor id diam euismod, at lacinia velit vestibulum. Donec pretium sollicitudin sapien, ut rhoncus purus euismod suscipit. Donec consectetur ante sed sapien accumsan dignissim. Pellentesque tempus semper velit et commodo. Proin eget ligula nisl. Integer sit amet erat ut magna condimentum faucibus. Integer rhoncus tincidunt metus. Maecenas maximus nisl at dui tincidunt, eu finibus nisi aliquam. Praesent sit amet massa nec purus scelerisque congue. Suspendisse porta diam eget neque sagittis mollis. Phasellus non orci odio. Mauris quam est, congue vel fermentum nec, cursus a lectus.

Vivamus id libero eros. Sed finibus nec ligula ut venenatis. Aenean id rutrum turpis, quis eleifend mi. Mauris pharetra ligula sit amet purus finibus viverra. Etiam nec nisl tristique, lobortis nunc sit amet, convallis tortor. Nunc sit amet velit sed enim malesuada pharetra. Morbi dignissim porta vulputate. Fusce accumsan nisl ex, eget dictum purus sollicitudin a. Vestibulum in leo in nisi accumsan pellentesque. Vivamus tempus mi tempor, congue velit dapibus, pretium felis. Fusce eget lectus velit. Curabitur tempor mollis malesuada. Suspendisse cursus risus odio, cursus luctus urna ultricies eu. Aenean in aliquam lacus. Donec eget interdum neque. Praesent efficitur eros at massa tristique, eu maximus felis rhoncus.

Duis sollicitudin risus eu est ullamcorper, in feugiat odio accumsan. Nunc convallis facilisis congue. Curabitur sem nunc, ullamcorper id ligula dignissim, commodo dapibus eros. Cras tristique ullamcorper elit in pretium. Fusce ac aliquam ipsum. Proin sodales sodales hendrerit. Ut iaculis nulla in consequat tincidunt. In mollis maximus lorem, eu tempus ipsum scelerisque quis.

Aenean hendrerit interdum turpis, eget rhoncus diam varius vel. Aliquam dolor augue, vulputate in varius ut, rutrum lobortis lectus. Quisque rhoncus lectus eros, ac sagittis arcu venenatis quis. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Cras faucibus, sapien quis porta tristique, risus dui mollis nisi, in feugiat turpis velit quis nunc. Aliquam eget odio porttitor, suscipit lacus non, tempor mi. Lorem ipsum dolor sit amet, consectetur adipiscing elit.</p></body></html>`))
		})

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
			r.Post("/webhook-subscription", hc.HandleStripeWebhookSubscription)
		})

		// SCWorker
		r.Route("/worker", func(r chi.Router) {
			r.Use(middleware.Logger)
			r.Get("/health", hc.HandleSCWorkerHealth)
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
			r.Use(mw.AuthMiddleware(middleware.AuthLevelOptional))
			// 20 requests per second
			r.Use(mw.RateLimit(20, "srv", 1*time.Second))
			r.Get("/", hc.HandleSemanticSearchGallery)
		})

		// User profiles
		r.Route("/profile", func(r chi.Router) {
			r.Use(middleware.Logger)
			r.Use(mw.AuthMiddleware(middleware.AuthLevelOptional))
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
			r.Get("/", hc.HandleGetUserV2)

			// Link to discord
			r.Post("/connect/discord", hc.HandleAuthorizeDiscord)

			// Create Generation
			r.Route("/image/generation/create", func(r chi.Router) {
				r.Use(mw.AbuseProtectorMiddleware())
				r.Post("/", hc.HandleCreateGeneration)
			})
			// ! Deprecated
			r.Route("/generation", func(r chi.Router) {
				r.Use(mw.AbuseProtectorMiddleware())
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

			// Like
			r.Post("/like", hc.HandleLikeGenerationOutputsForUser)

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
					r.Use(mw.AbuseProtectorMiddleware())
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
	port := utils.GetEnv().Port
	log.Info("Starting server", "port", port)

	h2s := &http2.Server{}
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
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
