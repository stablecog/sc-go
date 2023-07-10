package rest

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/ent/upscale"
	"github.com/stablecog/sc-go/database/enttypes"
	"github.com/stablecog/sc-go/database/repository"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/server/responses"
	"github.com/stablecog/sc-go/server/stripe"
	"github.com/stablecog/sc-go/shared"
	"github.com/stablecog/sc-go/utils"
	"golang.org/x/exp/slices"
)

// POST generate endpoint
// Handles creating a generation with API token
func (c *RestAPI) HandleCreateGenerationToken(w http.ResponseWriter, r *http.Request) {
	var user *ent.User
	if user = c.GetUserIfAuthenticated(w, r); user == nil {
		return
	}
	var apiToken *ent.ApiToken
	if apiToken = c.GetApiToken(w, r); apiToken == nil {
		return
	}

	free := user.ActiveProductID == nil
	if free {
		// Re-evaluate if they have paid credits
		count, err := c.Repo.HasPaidCredits(user.ID)
		if err != nil {
			log.Error("Error getting paid credit sum for users", "err", err)
			responses.ErrInternalServerError(w, r, "An unknown error has occurred")
			return
		}
		free = count <= 0
	}

	var qMax int
	roles, err := c.Repo.GetRoles(user.ID)
	if err != nil {
		log.Error("Error getting roles for user", "err", err)
		responses.ErrInternalServerError(w, r, "An unknown error has occurred")
		return
	}
	isSuperAdmin := slices.Contains(roles, "SUPER_ADMIN")
	if isSuperAdmin {
		qMax = math.MaxInt64
	} else {
		qMax = shared.MAX_QUEUED_ITEMS_FREE
	}
	if !isSuperAdmin && user.ActiveProductID != nil {
		switch *user.ActiveProductID {
		// Starter
		case stripe.GetProductIDs()[1]:
			qMax = shared.MAX_QUEUED_ITEMS_STARTER
			// Pro
		case stripe.GetProductIDs()[2]:
			qMax = shared.MAX_QUEUED_ITEMS_PRO
		// Ultimate
		case stripe.GetProductIDs()[3]:
			qMax = shared.MAX_QUEUED_ITEMS_ULTIMATE
		default:
			log.Warn("Unknown product ID", "product_id", *user.ActiveProductID)
		}
		// // Get product level
		// for level, product := range GetProductIDs() {
		// 	if product == *user.ActiveProductID {
		// 		prodLevel = level
		// 		break
		// 	}
		// }
	}
	for _, role := range roles {
		switch role {
		case "ULTIMATE":
			free = false
			if qMax < shared.MAX_QUEUED_ITEMS_ULTIMATE {
				qMax = shared.MAX_QUEUED_ITEMS_ULTIMATE
			}
		case "PRO":
			free = false
			if qMax < shared.MAX_QUEUED_ITEMS_PRO {
				qMax = shared.MAX_QUEUED_ITEMS_PRO
			}
		case "STARTER":
			free = false
			if qMax < shared.MAX_QUEUED_ITEMS_STARTER {
				qMax = shared.MAX_QUEUED_ITEMS_STARTER
			}
		}
	}

	// Parse request body
	reqBody, _ := io.ReadAll(r.Body)
	var generateReq requests.CreateGenerationRequest
	err = json.Unmarshal(reqBody, &generateReq)
	if err != nil {
		responses.ErrUnableToParseJson(w, r)
		return
	}

	if user.BannedAt != nil {
		remainingCredits, _ := c.Repo.GetNonExpiredCreditTotalForUser(user.ID, nil)
		render.Status(r, http.StatusOK)
		render.JSON(w, r, &responses.TaskQueuedResponse{
			ID:               uuid.NewString(),
			RemainingCredits: remainingCredits,
		})
		return
	}

	// Validation
	err = generateReq.Validate(true)
	if err != nil {
		responses.ErrBadRequest(w, r, err.Error(), "")
		return
	}

	// Set settings resp
	initSettings := responses.ImageGenerationSettingsResponse{
		ModelId:        *generateReq.ModelId,
		SchedulerId:    *generateReq.SchedulerId,
		Width:          *generateReq.Width,
		Height:         *generateReq.Height,
		NumOutputs:     *generateReq.NumOutputs,
		GuidanceScale:  *generateReq.GuidanceScale,
		InferenceSteps: *generateReq.InferenceSteps,
		Seed:           generateReq.Seed,
		InitImageURL:   generateReq.InitImageUrl,
		PromptStrength: generateReq.PromptStrength,
	}

	// The URL we send worker
	var signedInitImageUrl string
	// See if init image specified, validate it belongs to user, validate it exists in bucket
	if generateReq.InitImageUrl != "" {
		if utils.IsValidHTTPURL(generateReq.InitImageUrl) {
			// Custom image, do some validation on size, format, etc.
			_, _, err = utils.GetImageWidthHeightFromUrl(generateReq.InitImageUrl, shared.MAX_GENERATE_IMAGE_SIZE)
			if err != nil {
				responses.ErrBadRequest(w, r, "image_url_width_height_error", "")
				return
			}
			signedInitImageUrl = generateReq.InitImageUrl
		} else {
			// Remove s3 prefix
			signedInitImageUrl = strings.TrimPrefix(generateReq.InitImageUrl, "s3://")
			// Hash user ID to see if it belongs to this user
			uidHash := utils.Sha256(user.ID.String())
			if !strings.HasPrefix(signedInitImageUrl, fmt.Sprintf("%s/", uidHash)) {
				responses.ErrUnauthorized(w, r)
				return
			}
			// Verify exists in bucket
			_, err := c.S3.HeadObject(&s3.HeadObjectInput{
				Bucket: aws.String(os.Getenv("S3_IMG2IMG_BUCKET_NAME")),
				Key:    aws.String(signedInitImageUrl),
			})
			if err != nil {
				if aerr, ok := err.(awserr.Error); ok {
					switch aerr.Code() {
					case "NotFound": // s3.ErrCodeNoSuchKey does not work, aws is missing this error code so we hardwire a string
						responses.ErrBadRequest(w, r, "init_image_not_found", "")
						return
					default:
						log.Error("Error checking if init image exists in bucket", "err", err)
						responses.ErrInternalServerError(w, r, "An unknown error has occured")
						return
					}
				}
				responses.ErrBadRequest(w, r, "init_image_not_found", "")
				return
			}
			// Sign object URL to pass to worker
			req, _ := c.S3.GetObjectRequest(&s3.GetObjectInput{
				Bucket: aws.String(os.Getenv("S3_IMG2IMG_BUCKET_NAME")),
				Key:    aws.String(signedInitImageUrl),
			})
			urlStr, err := req.Presign(5 * time.Minute)
			if err != nil {
				log.Error("Error signing init image URL", "err", err)
				responses.ErrInternalServerError(w, r, "An unknown error has occured")
				return
			}
			signedInitImageUrl = urlStr
		}
	}

	// Get queue count
	nq, err := c.QueueThrottler.NumQueued(fmt.Sprintf("g:%s", user.ID.String()))
	if err != nil {
		log.Warn("Error getting queue count", "err", err, "user_id", user.ID.String())
	}
	if err == nil && nq > qMax {
		// Get queue overflow size
		overflowSize, err := c.QueueThrottler.NumQueued(fmt.Sprintf("of:%s", user.ID.String()))
		if err != nil {
			log.Warn("Error getting queue overflow count", "err", err, "user_id", user.ID.String())
		}
		// If overflow size is greater than max, return error
		if overflowSize > shared.QUEUE_OVERFLOW_MAX {
			responses.ErrBadRequest(w, r, "queue_limit_reached", "")
			return
		}
		// Overflow size can be 0 so we need to add 1
		overflowSize++
		c.QueueThrottler.IncrementBy(1, fmt.Sprintf("of:%s", user.ID.String()))
		for {
			time.Sleep(time.Duration(shared.QUEUE_OVERFLOW_PENALTY_MS*overflowSize) * time.Millisecond)
			nq, err = c.QueueThrottler.NumQueued(fmt.Sprintf("g:%s", user.ID.String()))
			if err != nil {
				log.Warn("Error getting queue count", "err", err, "user_id", user.ID.String())
			}
			if err == nil && nq <= qMax {
				c.QueueThrottler.DecrementBy(1, fmt.Sprintf("of:%s", user.ID.String()))
				break
			}
			// Update overflow size
			overflowSize, err = c.QueueThrottler.NumQueued(fmt.Sprintf("of:%s", user.ID.String()))
			if err != nil {
				log.Warn("Error getting queue overflow count", "err", err, "user_id", user.ID.String())
			}
			overflowSize++
		}
	}

	// Enforce submit to gallery
	if free {
		generateReq.SubmitToGallery = true
	}

	// Parse request headers
	countryCode := utils.GetCountryCode(r)
	deviceInfo := utils.GetClientDeviceInfo(r)

	// Get model and scheduler name for cog
	modelName := shared.GetCache().GetGenerationModelNameFromID(*generateReq.ModelId)
	schedulerName := shared.GetCache().GetSchedulerNameFromID(*generateReq.SchedulerId)
	if modelName == "" || schedulerName == "" {
		log.Error("Error getting model or scheduler name: %s - %s", modelName, schedulerName)
		responses.ErrInternalServerError(w, r, "An unknown error has occured")
		return
	}

	// Format prompts
	generateReq.Prompt = utils.FormatPrompt(generateReq.Prompt)
	generateReq.NegativePrompt = utils.FormatPrompt(generateReq.NegativePrompt)

	// For live page update
	var livePageMsg shared.LivePageMessage
	// For keeping track of this request as it gets sent to the worker
	var requestId uuid.UUID
	// Cog request
	var cogReqBody requests.CogQueueRequest

	// Credits left after this operation
	var remainingCredits int

	// Create channel to track request
	// Create channel
	activeChl := make(chan requests.CogWebhookMessage)
	// Cleanup
	defer close(activeChl)

	// Wrap everything in a DB transaction
	// We do this since we want our credit deduction to be atomic with the whole process
	if err := c.Repo.WithTx(func(tx *ent.Tx) error {
		// Bind a client to the transaction
		DB := tx.Client()
		// Deduct credits from user
		deducted, err := c.Repo.DeductCreditsFromUser(user.ID, *generateReq.NumOutputs, DB)
		if err != nil {
			log.Error("Error deducting credits", "err", err)
			responses.ErrInternalServerError(w, r, "Error deducting credits from user")
			return err
		} else if !deducted {
			responses.ErrInsufficientCredits(w, r)
			return responses.InsufficientCreditsErr
		}

		// Translate prompts
		translatedPrompt, translatedNegativePrompt, err := c.SafetyChecker.TranslatePrompt(generateReq.Prompt, generateReq.NegativePrompt)
		if err != nil {
			log.Error("Error translating prompt", "err", err)
			responses.ErrInternalServerError(w, r, "An unknown error has occured")
			return err
		}
		generateReq.Prompt = translatedPrompt
		generateReq.NegativePrompt = translatedNegativePrompt
		// Check NSFW
		isNSFW, reason, err := c.SafetyChecker.IsPromptNSFW(generateReq.Prompt)
		if err != nil {
			log.Error("Error checking prompt NSFW", "err", err)
			responses.ErrInternalServerError(w, r, "An unknown error has occured")
			return err
		}
		if isNSFW {
			responses.ErrBadRequest(w, r, "nsfw_prompt", reason)
			return fmt.Errorf("nsfw: %s", reason)
		}

		remainingCredits, err = c.Repo.GetNonExpiredCreditTotalForUser(user.ID, DB)
		if err != nil {
			log.Error("Error getting remaining credits", "err", err)
			responses.ErrInternalServerError(w, r, "An unknown error has occured")
			return err
		}

		// Create generation
		g, err := c.Repo.CreateGeneration(
			user.ID,
			string(deviceInfo.DeviceType),
			deviceInfo.DeviceOs,
			deviceInfo.DeviceBrowser,
			countryCode,
			generateReq,
			user.ActiveProductID,
			&apiToken.ID,
			enttypes.SourceTypeAPI,
			DB)
		if err != nil {
			log.Error("Error creating generation", "err", err)
			responses.ErrInternalServerError(w, r, "Error creating generation")
			return err
		}

		// Request Id matches generation ID
		requestId = g.ID

		// For live page update
		livePageMsg = shared.LivePageMessage{
			ProcessType:      shared.GENERATE,
			ID:               utils.Sha256(requestId.String()),
			CountryCode:      countryCode,
			Status:           shared.LivePageQueued,
			TargetNumOutputs: *generateReq.NumOutputs,
			Width:            generateReq.Width,
			Height:           generateReq.Height,
			CreatedAt:        g.CreatedAt,
			ProductID:        user.ActiveProductID,
			Source:           enttypes.SourceTypeAPI,
		}

		cogReqBody = requests.CogQueueRequest{
			WebhookEventsFilter: []requests.CogEventFilter{requests.CogEventFilterStart, requests.CogEventFilterStart},
			WebhookUrl:          fmt.Sprintf("%s/v1/worker/webhook", utils.GetEnv("PUBLIC_API_URL", "")),
			Input: requests.BaseCogRequest{
				SkipSafetyChecker:    true,
				SkipTranslation:      true,
				APIRequest:           true,
				ID:                   requestId,
				IP:                   utils.GetIPAddress(r),
				UserID:               &user.ID,
				DeviceInfo:           deviceInfo,
				LivePageData:         &livePageMsg,
				Prompt:               generateReq.Prompt,
				NegativePrompt:       generateReq.NegativePrompt,
				Width:                generateReq.Width,
				Height:               generateReq.Height,
				NumInferenceSteps:    generateReq.InferenceSteps,
				GuidanceScale:        generateReq.GuidanceScale,
				Model:                modelName,
				ModelId:              *generateReq.ModelId,
				Scheduler:            schedulerName,
				SchedulerId:          *generateReq.SchedulerId,
				Seed:                 generateReq.Seed,
				NumOutputs:           generateReq.NumOutputs,
				OutputImageExtension: string(shared.DEFAULT_GENERATE_OUTPUT_EXTENSION),
				OutputImageQuality:   utils.ToPtr(shared.DEFAULT_GENERATE_OUTPUT_QUALITY),
				ProcessType:          shared.GENERATE,
				SubmitToGallery:      generateReq.SubmitToGallery,
				InitImageUrl:         signedInitImageUrl,
				PromptStrength:       generateReq.PromptStrength,
			},
		}

		if cogReqBody.Input.InitImageUrl != "" {
			cogReqBody.Input.InitImageUrlS3 = generateReq.InitImageUrl
		}

		err = c.Redis.EnqueueCogRequest(r.Context(), shared.COG_REDIS_QUEUE, cogReqBody)
		if err != nil {
			log.Error("Failed to write request %s to queue: %v", requestId, err)
			responses.ErrInternalServerError(w, r, "Failed to queue generate request")
			return err
		}

		c.QueueThrottler.IncrementBy(1, fmt.Sprintf("g:%s", user.ID.String()))
		return nil
	}); err != nil {
		log.Error("Error in transaction", "err", err)
		return
	}

	// Add channel to sync array (basically a thread-safe map)
	c.SMap.Put(requestId.String(), activeChl)
	defer c.SMap.Delete(requestId.String())
	defer c.QueueThrottler.DecrementBy(1, fmt.Sprintf("g:%s", user.ID.String()))

	// Send live page update
	go func() {
		liveResp := repository.TaskStatusUpdateResponse{
			ForLivePage:     true,
			LivePageMessage: &livePageMsg,
		}
		respBytes, err := json.Marshal(liveResp)
		if err != nil {
			log.Error("Error marshalling sse live response", "err", err)
			return
		}
		err = c.Redis.Client.Publish(c.Redis.Ctx, shared.REDIS_SSE_BROADCAST_CHANNEL, respBytes).Err()
		if err != nil {
			log.Error("Failed to publish live page update", "err", err)
		}
	}()

	// Analytics
	go c.Track.GenerationStarted(user, cogReqBody.Input, utils.GetIPAddress(r))

	// Wait for result
	for {
		select {
		case cogMsg := <-activeChl:
			switch cogMsg.Status {
			case requests.CogProcessing:
				err := c.Repo.SetGenerationStarted(requestId.String())
				if err != nil {
					log.Error("Failed to set generate started", "id", requestId, "err", err)
					responses.ErrInternalServerError(w, r, "An unknown error occurred")
					return
				}
				// Send live page update
				go func() {
					cogMsg.Input.LivePageData.Status = shared.LivePageProcessing
					now := time.Now()
					cogMsg.Input.LivePageData.StartedAt = &now
					liveResp := repository.TaskStatusUpdateResponse{
						ForLivePage:     true,
						LivePageMessage: cogMsg.Input.LivePageData,
					}
					respBytes, err := json.Marshal(liveResp)
					if err != nil {
						log.Error("Error marshalling sse live response", "err", err)
						return
					}
					err = c.Redis.Client.Publish(c.Redis.Ctx, shared.REDIS_SSE_BROADCAST_CHANNEL, respBytes).Err()
					if err != nil {
						log.Error("Failed to publish live page update", "err", err)
					}
				}()
			case requests.CogSucceeded:
				outputs, err := c.Repo.SetGenerationSucceeded(requestId.String(), generateReq.Prompt, generateReq.NegativePrompt, cogMsg.Output, cogMsg.NSFWCount)
				if err != nil {
					log.Error("Failed to set generation succeeded", "id", upscale.ID, "err", err)
					responses.ErrInternalServerError(w, r, "An unknown error occurred")
					return
				}
				// Send live page update
				go func() {
					cogMsg.Input.LivePageData.Status = shared.LivePageSucceeded
					now := time.Now()
					cogMsg.Input.LivePageData.CompletedAt = &now
					cogMsg.Input.LivePageData.ActualNumOutputs = len(outputs)
					liveResp := repository.TaskStatusUpdateResponse{
						ForLivePage:     true,
						LivePageMessage: cogMsg.Input.LivePageData,
					}
					respBytes, err := json.Marshal(liveResp)
					if err != nil {
						log.Error("Error marshalling sse live response", "err", err)
						return
					}
					err = c.Redis.Client.Publish(c.Redis.Ctx, shared.REDIS_SSE_BROADCAST_CHANNEL, respBytes).Err()
					if err != nil {
						log.Error("Failed to publish live page update", "err", err)
					}
				}()
				// Analytics
				generation, err := c.Repo.GetGeneration(requestId)
				if err != nil {
					log.Error("Error getting generation for analytics", "err", err)
				}
				// Get durations in seconds
				if generation.StartedAt == nil {
					log.Error("Generation started at is nil", "id", cogMsg.Input.ID)
				}
				duration := time.Now().Sub(*generation.StartedAt).Seconds()
				qDuration := (*generation.StartedAt).Sub(generation.CreatedAt).Seconds()
				go c.Track.GenerationSucceeded(user, cogMsg.Input, duration, qDuration, utils.GetIPAddress(r))

				// Format response
				resOutputs := make([]responses.ApiOutput, len(outputs))
				for i, output := range outputs {
					resOutputs[i] = responses.ApiOutput{
						URL:      utils.GetURLFromImagePath(output.ImagePath),
						ImageURL: utils.ToPtr(utils.GetURLFromImagePath(output.ImagePath)),
						ID:       output.ID,
					}
				}

				// Set token used
				err = c.Repo.SetTokenUsedAndIncrementCreditsSpent(int(*generateReq.NumOutputs), *generation.APITokenID)
				if err != nil {
					log.Error("Failed to set token used", "err", err)
				}

				render.Status(r, http.StatusOK)
				render.JSON(w, r, responses.ApiSucceededResponse{
					Outputs:          resOutputs,
					RemainingCredits: remainingCredits,
					Settings:         initSettings,
				})
				return
			case requests.CogFailed:
				if err := c.Repo.WithTx(func(tx *ent.Tx) error {
					DB := tx.Client()
					err := c.Repo.SetGenerationFailed(requestId.String(), cogMsg.Error, cogMsg.NSFWCount, DB)
					if err != nil {
						log.Error("Failed to set generation failed", "id", upscale.ID, "err", err)
						responses.ErrInternalServerError(w, r, "An unknown error occurred")
						return err
					}
					// Send live page update
					go func() {
						cogMsg.Input.LivePageData.Status = shared.LivePageFailed
						now := time.Now()
						cogMsg.Input.LivePageData.CompletedAt = &now
						liveResp := repository.TaskStatusUpdateResponse{
							ForLivePage:     true,
							LivePageMessage: cogMsg.Input.LivePageData,
						}
						respBytes, err := json.Marshal(liveResp)
						if err != nil {
							log.Error("Error marshalling sse live response", "err", err)
							return
						}
						err = c.Redis.Client.Publish(c.Redis.Ctx, shared.REDIS_SSE_BROADCAST_CHANNEL, respBytes).Err()
						if err != nil {
							log.Error("Failed to publish live page update", "err", err)
						}
					}()
					// Analytics
					duration := time.Now().Sub(cogMsg.Input.LivePageData.CreatedAt).Seconds()
					go c.Track.GenerationFailed(user, cogMsg.Input, duration, cogMsg.Error, utils.GetIPAddress(r))
					// Refund credits
					_, err = c.Repo.RefundCreditsToUser(user.ID, *generateReq.NumOutputs, DB)
					if err != nil {
						log.Error("Failed to refund credits", "err", err)
						return err
					}
					return nil
				}); err != nil {
					log.Error("Failed to set generation failed", "id", requestId, "err", err)
					responses.ErrInternalServerError(w, r, "An unknown error occurred")
					return
				}

				render.Status(r, http.StatusInternalServerError)
				render.JSON(w, r, responses.ApiFailedResponse{
					Error:    cogMsg.Error,
					Settings: initSettings,
				})
				return
			}
		case <-time.After(shared.REQUEST_COG_TIMEOUT):
			if err := c.Repo.WithTx(func(tx *ent.Tx) error {
				DB := tx.Client()
				err := c.Repo.SetGenerationFailed(requestId.String(), shared.TIMEOUT_ERROR, 0, DB)
				if err != nil {
					log.Error("Failed to set generation failed", "id", upscale.ID, "err", err)
				}
				// Refund credits
				_, err = c.Repo.RefundCreditsToUser(user.ID, *generateReq.NumOutputs, DB)
				if err != nil {
					log.Error("Failed to refund credits", "err", err)
					return err
				}
				return nil
			}); err != nil {
				log.Error("Failed to set generation failed", "id", requestId, "err", err)
				responses.ErrInternalServerError(w, r, "An unknown error occurred")
				return
			}

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, responses.ApiFailedResponse{
				Error:    shared.TIMEOUT_ERROR,
				Settings: initSettings,
			})
			return
		}
	}
}
