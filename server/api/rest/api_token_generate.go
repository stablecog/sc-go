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
	"github.com/stablecog/sc-go/database/repository"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/server/responses"
	"github.com/stablecog/sc-go/shared"
	"github.com/stablecog/sc-go/utils"
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
	isSuperAdmin, _ := c.Repo.IsSuperAdmin(user.ID)
	if isSuperAdmin {
		qMax = math.MaxInt64
	} else {
		qMax = shared.MAX_QUEUED_ITEMS_FREE
	}
	if !isSuperAdmin && user.ActiveProductID != nil {
		switch *user.ActiveProductID {
		// Starter
		case GetProductIDs()[1]:
			qMax = shared.MAX_QUEUED_ITEMS_STARTER
			// Pro
		case GetProductIDs()[2]:
			qMax = shared.MAX_QUEUED_ITEMS_PRO
		// Ultimate
		case GetProductIDs()[3]:
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

	// Parse request body
	reqBody, _ := io.ReadAll(r.Body)
	var generateReq requests.CreateGenerationRequest
	err := json.Unmarshal(reqBody, &generateReq)
	if err != nil {
		responses.ErrUnableToParseJson(w, r)
		return
	}

	// Validation
	err = generateReq.Validate(true)
	if err != nil {
		responses.ErrBadRequest(w, r, err.Error(), "")
		return
	}

	// The URL we send worker
	var signedInitImageUrl string
	// See if init image specified, validate it belongs to user, validate it exists in bucket
	if generateReq.InitImageUrl != "" {
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
	modelName := shared.GetCache().GetGenerationModelNameFromID(generateReq.ModelId)
	schedulerName := shared.GetCache().GetSchedulerNameFromID(generateReq.SchedulerId)
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
	var requestId string
	// Cog request
	var cogReqBody requests.CogQueueRequest

	// Credits left after this operation
	var remainingCredits int

	// Create channel to track request
	// Create channel
	activeChl := make(chan requests.CogWebhookMessage)
	// Cleanup
	defer close(activeChl)
	defer c.SMap.Delete(requestId)

	// Wrap everything in a DB transaction
	// We do this since we want our credit deduction to be atomic with the whole process
	if err := c.Repo.WithTx(func(tx *ent.Tx) error {
		// Bind a client to the transaction
		DB := tx.Client()
		// Deduct credits from user
		deducted, err := c.Repo.DeductCreditsFromUser(user.ID, int32(generateReq.NumOutputs), DB)
		if err != nil {
			log.Error("Error deducting credits", "err", err)
			responses.ErrInternalServerError(w, r, "Error deducting credits from user")
			return err
		} else if !deducted {
			responses.ErrInsufficientCredits(w, r)
			return responses.InsufficientCreditsErr
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
			DB)
		if err != nil {
			log.Error("Error creating generation", "err", err)
			responses.ErrInternalServerError(w, r, "Error creating generation")
			return err
		}

		// Request Id matches generation ID
		requestId = g.ID.String()

		// For live page update
		livePageMsg = shared.LivePageMessage{
			ProcessType:      shared.GENERATE,
			ID:               utils.Sha256(requestId),
			CountryCode:      countryCode,
			Status:           shared.LivePageQueued,
			TargetNumOutputs: generateReq.NumOutputs,
			Width:            generateReq.Width,
			Height:           generateReq.Height,
			CreatedAt:        g.CreatedAt,
			ProductID:        user.ActiveProductID,
			Source:           shared.OperationSourceTypeAPI,
		}

		var promtpStrengthStr string
		if generateReq.PromptStrength != nil {
			promtpStrengthStr = fmt.Sprint(*generateReq.PromptStrength)
		}

		cogReqBody = requests.CogQueueRequest{
			WebhookEventsFilter: []requests.CogEventFilter{requests.CogEventFilterStart, requests.CogEventFilterStart},
			WebhookUrl:          fmt.Sprintf("%s/v1/worker/webhook", utils.GetEnv("PUBLIC_API_URL", "")),
			Input: requests.BaseCogRequest{
				APIRequest:           true,
				ID:                   requestId,
				IP:                   utils.GetIPAddress(r),
				UserID:               &user.ID,
				DeviceInfo:           deviceInfo,
				LivePageData:         &livePageMsg,
				Prompt:               generateReq.Prompt,
				NegativePrompt:       generateReq.NegativePrompt,
				Width:                fmt.Sprint(generateReq.Width),
				Height:               fmt.Sprint(generateReq.Height),
				NumInferenceSteps:    fmt.Sprint(generateReq.InferenceSteps),
				GuidanceScale:        fmt.Sprint(generateReq.GuidanceScale),
				Model:                modelName,
				ModelId:              generateReq.ModelId,
				Scheduler:            schedulerName,
				SchedulerId:          generateReq.SchedulerId,
				Seed:                 fmt.Sprint(generateReq.Seed),
				NumOutputs:           fmt.Sprint(generateReq.NumOutputs),
				OutputImageExtension: string(shared.DEFAULT_GENERATE_OUTPUT_EXTENSION),
				OutputImageQuality:   fmt.Sprint(shared.DEFAULT_GENERATE_OUTPUT_QUALITY),
				ProcessType:          shared.GENERATE,
				SubmitToGallery:      generateReq.SubmitToGallery,
				InitImageUrl:         signedInitImageUrl,
				PromptStrength:       promtpStrengthStr,
			},
		}

		if cogReqBody.Input.InitImageUrl != "" {
			cogReqBody.Input.InitImageUrlS3 = generateReq.InitImageUrl
		}

		// Add channel to sync array (basically a thread-safe map)
		c.SMap.Put(requestId, activeChl)

		err = c.Redis.EnqueueCogRequest(r.Context(), cogReqBody)
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
				err := c.Repo.SetGenerationStarted(requestId)
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
				outputs, err := c.Repo.SetGenerationSucceeded(requestId, generateReq.Prompt, generateReq.NegativePrompt, cogMsg.Output, cogMsg.NSFWCount)
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
				generation, err := c.Repo.GetGeneration(uuid.MustParse(requestId))
				if err != nil {
					log.Error("Error getting generation for analytics", "err", err)
				}
				// Get durations in seconds
				if generation.StartedAt == nil {
					log.Error("Generation started at is nil", "id", cogMsg.Input.ID)
				}
				duration := time.Now().Sub(*generation.StartedAt).Seconds()
				qDuration := (*generation.StartedAt).Sub(generation.CreatedAt).Seconds()
				go c.Track.GenerationSucceeded(user, cogMsg.Input, duration, qDuration, "system")

				// Format response
				resOutputs := make([]responses.ApiOutput, len(outputs))
				for i, output := range outputs {
					resOutputs[i] = responses.ApiOutput{
						URL: utils.GetURLFromImagePath(output.ImagePath),
						ID:  output.ID,
					}
				}

				// Set token used
				err = c.Repo.SetTokenUsedAndIncrementCreditsSpent(int(generateReq.NumOutputs), *generation.APITokenID)
				if err != nil {
					log.Error("Failed to set token used", "err", err)
				}

				render.Status(r, http.StatusOK)
				render.JSON(w, r, responses.ApiSucceededResponse{
					Outputs:          resOutputs,
					RemainingCredits: remainingCredits,
				})
				return
			case requests.CogFailed:
				if err := c.Repo.WithTx(func(tx *ent.Tx) error {
					DB := tx.Client()
					err := c.Repo.SetGenerationFailed(requestId, cogMsg.Error, cogMsg.NSFWCount, DB)
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
					go c.Track.GenerationFailed(user, cogMsg.Input, duration, cogMsg.Error, "system")
					// Refund credits
					_, err = c.Repo.RefundCreditsToUser(user.ID, generateReq.NumOutputs, DB)
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
					Error: cogMsg.Error,
				})
				return
			}
		case <-time.After(shared.REQUEST_COG_TIMEOUT):
			if err := c.Repo.WithTx(func(tx *ent.Tx) error {
				DB := tx.Client()
				err := c.Repo.SetGenerationFailed(requestId, shared.TIMEOUT_ERROR, 0, DB)
				if err != nil {
					log.Error("Failed to set generation failed", "id", upscale.ID, "err", err)
				}
				// Refund credits
				_, err = c.Repo.RefundCreditsToUser(user.ID, generateReq.NumOutputs, DB)
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
				Error: shared.TIMEOUT_ERROR,
			})
			return
		}
	}
}
