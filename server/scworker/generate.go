package scworker

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
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

func (w *SCWorker) CreateGeneration(source enttypes.SourceType,
	r *http.Request,
	user *ent.User,
	apiTokenId *uuid.UUID,
	generateReq requests.CreateGenerationRequest) (*responses.ApiSucceededResponse, *responses.ImageGenerationSettingsResponse, *WorkerError) {
	free := user.ActiveProductID == nil
	if free {
		// Re-evaluate if they have paid credits
		count, err := w.Repo.HasPaidCredits(user.ID)
		if err != nil {
			log.Error("Error getting paid credit sum for users", "err", err)
			return nil, nil, WorkerInternalServerError()
		}
		free = count <= 0
	}

	var qMax int
	roles, err := w.Repo.GetRoles(user.ID)
	if err != nil {
		log.Error("Error getting roles for user", "err", err)
		return nil, nil, WorkerInternalServerError()
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

	if user.BannedAt != nil {
		return nil, nil, &WorkerError{http.StatusForbidden, fmt.Errorf("user_banned"), ""}
	}

	// Validation
	if !isSuperAdmin {
		err = generateReq.Validate(source != enttypes.SourceTypeWebUI)
		if err != nil {
			return nil, nil, &WorkerError{http.StatusBadRequest, err, ""}
		}
	} else {
		generateReq.ApplyDefaults()
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
				return nil, &initSettings, &WorkerError{http.StatusBadRequest, fmt.Errorf("image_url_width_height_error"), ""}
			}
			signedInitImageUrl = generateReq.InitImageUrl
		} else if w.S3 != nil {
			// Remove s3 prefix
			signedInitImageUrl = strings.TrimPrefix(generateReq.InitImageUrl, "s3://")
			// Hash user ID to see if it belongs to this user
			uidHash := utils.Sha256(user.ID.String())
			if !strings.HasPrefix(signedInitImageUrl, fmt.Sprintf("%s/", uidHash)) {
				return nil, &initSettings, &WorkerError{http.StatusUnauthorized, fmt.Errorf("init_image_not_owned"), ""}
			}
			// Verify exists in bucket
			_, err := w.S3.HeadObject(&s3.HeadObjectInput{
				Bucket: aws.String(os.Getenv("S3_IMG2IMG_BUCKET_NAME")),
				Key:    aws.String(signedInitImageUrl),
			})
			if err != nil {
				if aerr, ok := err.(awserr.Error); ok {
					switch aerr.Code() {
					case "NotFound": // s3.ErrCodeNoSuchKey does not work, aws is missing this error code so we hardwire a string
						return nil, &initSettings, &WorkerError{http.StatusBadRequest, fmt.Errorf("init_image_not_found"), ""}
					default:
						log.Error("Error checking if init image exists in bucket", "err", err)
						return nil, &initSettings, &WorkerError{http.StatusInternalServerError, fmt.Errorf("unknown_error"), ""}
					}
				}
				return nil, &initSettings, &WorkerError{http.StatusBadRequest, fmt.Errorf("init_image_not_found"), ""}
			}
			// Sign object URL to pass to worker
			req, _ := w.S3.GetObjectRequest(&s3.GetObjectInput{
				Bucket: aws.String(os.Getenv("S3_IMG2IMG_BUCKET_NAME")),
				Key:    aws.String(signedInitImageUrl),
			})
			urlStr, err := req.Presign(5 * time.Minute)
			if err != nil {
				log.Error("Error signing init image URL", "err", err)
				return nil, &initSettings, &WorkerError{http.StatusInternalServerError, fmt.Errorf("unknown_error"), ""}
			}
			signedInitImageUrl = urlStr
		}

		if signedInitImageUrl == "" {
			return nil, &initSettings, &WorkerError{http.StatusBadRequest, fmt.Errorf("invalid_image_url"), ""}
		}
	}

	// Get queue count
	nq, err := w.QueueThrottler.NumQueued(fmt.Sprintf("g:%s", user.ID.String()))
	if err != nil {
		log.Warn("Error getting queue count", "err", err, "user_id", user.ID.String())
	}
	if err == nil && nq > qMax {
		// Get queue overflow size
		overflowSize, err := w.QueueThrottler.NumQueued(fmt.Sprintf("of:%s", user.ID.String()))
		if err != nil {
			log.Warn("Error getting queue overflow count", "err", err, "user_id", user.ID.String())
		}
		// If overflow size is greater than max, return error
		if overflowSize > shared.QUEUE_OVERFLOW_MAX {
			return nil, &initSettings, &WorkerError{http.StatusBadRequest, fmt.Errorf("queue_limit_reached"), ""}
		}
		// Overflow size can be 0 so we need to add 1
		overflowSize++
		w.QueueThrottler.IncrementBy(1, fmt.Sprintf("of:%s", user.ID.String()))
		for {
			time.Sleep(time.Duration(shared.QUEUE_OVERFLOW_PENALTY_MS*overflowSize) * time.Millisecond)
			nq, err = w.QueueThrottler.NumQueued(fmt.Sprintf("g:%s", user.ID.String()))
			if err != nil {
				log.Warn("Error getting queue count", "err", err, "user_id", user.ID.String())
			}
			if err == nil && nq <= qMax {
				w.QueueThrottler.DecrementBy(1, fmt.Sprintf("of:%s", user.ID.String()))
				break
			}
			// Update overflow size
			overflowSize, err = w.QueueThrottler.NumQueued(fmt.Sprintf("of:%s", user.ID.String()))
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
	var countryCode string
	var deviceInfo utils.ClientDeviceInfo
	ipAddress := "system"
	if r != nil {
		countryCode = utils.GetCountryCode(r)
		deviceInfo = utils.GetClientDeviceInfo(r)
		ipAddress = utils.GetIPAddress(r)
	} else {
		countryCode = "US"
		deviceInfo = utils.ClientDeviceInfo{
			DeviceType:    utils.Bot,
			DeviceOs:      "Linux",
			DeviceBrowser: "Discord",
		}
	}
	// Get model and scheduler name for cog
	modelName := shared.GetCache().GetGenerationModelNameFromID(*generateReq.ModelId)
	schedulerName := shared.GetCache().GetSchedulerNameFromID(*generateReq.SchedulerId)
	if modelName == "" || schedulerName == "" {
		log.Error("Error getting model or scheduler name: %s - %s", modelName, schedulerName)
		return nil, &initSettings, &WorkerError{http.StatusBadRequest, fmt.Errorf("invalid_model_or_scheduler"), ""}
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
	if err := w.Repo.WithTx(func(tx *ent.Tx) error {
		if source == enttypes.SourceTypeAPI {
			log.Warn("1")
		}
		// Bind a client to the transaction
		DB := tx.Client()
		if source == enttypes.SourceTypeAPI {
			log.Warn("2")
		}
		// Deduct credits from user
		deducted, err := w.Repo.DeductCreditsFromUser(user.ID, *generateReq.NumOutputs, DB)
		if source == enttypes.SourceTypeAPI {
			log.Warn("3")
		}
		if err != nil {
			log.Error("Error deducting credits", "err", err)
			return err
		} else if !deducted {
			if source == enttypes.SourceTypeAPI {
				log.Warn("4")
			}
			return responses.InsufficientCreditsErr
		}

		// Translate prompts
		translatedPrompt, translatedNegativePrompt, err := w.SafetyChecker.TranslatePrompt(generateReq.Prompt, generateReq.NegativePrompt)
		if err != nil {
			log.Error("Error translating prompt", "err", err)
			return err
		}
		if source == enttypes.SourceTypeAPI {
			log.Warn("5")
		}
		generateReq.Prompt = translatedPrompt
		generateReq.NegativePrompt = translatedNegativePrompt
		// Check NSFW
		isNSFW, reason, err := w.SafetyChecker.IsPromptNSFW(generateReq.Prompt)
		if err != nil {
			log.Error("Error checking prompt NSFW", "err", err)
			return err
		}
		if source == enttypes.SourceTypeAPI {
			log.Warn("6")
		}
		if isNSFW {
			return fmt.Errorf("nsfw: %s", reason)
		}
		if source == enttypes.SourceTypeAPI {
			log.Warn("7")
		}

		remainingCredits, err = w.Repo.GetNonExpiredCreditTotalForUser(user.ID, DB)
		if err != nil {
			log.Error("Error getting remaining credits", "err", err)
			return err
		}

		if source == enttypes.SourceTypeAPI {
			log.Warn("8")
		}
		// Create generation
		g, err := w.Repo.CreateGeneration(
			user.ID,
			string(deviceInfo.DeviceType),
			deviceInfo.DeviceOs,
			deviceInfo.DeviceBrowser,
			countryCode,
			generateReq,
			user.ActiveProductID,
			apiTokenId,
			source,
			DB)
		if err != nil {
			log.Error("Error creating generation", "err", err)
			return err
		}
		if source == enttypes.SourceTypeAPI {
			log.Warn("9")
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
			Source:           source,
		}

		if source == enttypes.SourceTypeAPI {
			log.Warnf("Setting cog request body!")
		}
		cogReqBody = requests.CogQueueRequest{
			WebhookEventsFilter: []requests.CogEventFilter{requests.CogEventFilterStart, requests.CogEventFilterStart},
			WebhookUrl:          fmt.Sprintf("%s/v1/worker/webhook", utils.GetEnv("PUBLIC_API_URL", "")),
			Input: requests.BaseCogRequest{
				SkipSafetyChecker:    true,
				SkipTranslation:      true,
				APIRequest:           source != enttypes.SourceTypeWebUI,
				ID:                   requestId,
				IP:                   ipAddress,
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

		if source == enttypes.SourceTypeWebUI {
			cogReqBody.Input.UIId = generateReq.UIId
			cogReqBody.Input.StreamID = generateReq.StreamID
		}

		if cogReqBody.Input.InitImageUrl != "" {
			cogReqBody.Input.InitImageUrlS3 = generateReq.InitImageUrl
		}

		err = w.Redis.EnqueueCogRequest(w.Redis.Ctx, shared.COG_REDIS_QUEUE, cogReqBody)
		if err != nil {
			log.Error("Failed to write request %s to queue: %v", requestId, err)
			return err
		}

		w.QueueThrottler.IncrementBy(1, fmt.Sprintf("g:%s", user.ID.String()))
		return nil
	}); err != nil {
		log.Error("Error in transaction", "err", err)
		if errors.Is(err, responses.InsufficientCreditsErr) {
			return nil, &initSettings, &WorkerError{http.StatusBadRequest, responses.InsufficientCreditsErr, ""}
		}
		return nil, &initSettings, WorkerInternalServerError()
	}

	// Add channel to sync array (basically a thread-safe map)
	if source != enttypes.SourceTypeWebUI {
		w.SMap.Put(requestId.String(), activeChl)
		defer w.SMap.Delete(requestId.String())
		defer w.QueueThrottler.DecrementBy(1, fmt.Sprintf("g:%s", user.ID.String()))
	}

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
		err = w.Redis.Client.Publish(w.Redis.Ctx, shared.REDIS_SSE_BROADCAST_CHANNEL, respBytes).Err()
		if err != nil {
			log.Error("Failed to publish live page update", "err", err)
		}
	}()

	// Analytics
	if source != enttypes.SourceTypeAPI {
		go w.Track.GenerationStarted(user, cogReqBody.Input, source, ipAddress)
	} else {
		log.Warnf("Guidance Scale %f", *generateReq.GuidanceScale)
		if cogReqBody.Input.GuidanceScale == nil {
			log.Warn("Guidance Scale is nil - cogReqBody")
		}
		log.Warnf("Prompt %s", cogReqBody.Input.Prompt)
		log.Warnf("URL %s", cogReqBody.WebhookUrl)
	}
	// Set timeout delay for UI
	if source == enttypes.SourceTypeWebUI {
		// Set timeout key
		err = w.Redis.SetCogRequestStreamID(w.Redis.Ctx, requestId.String(), generateReq.StreamID)
		if err != nil {
			// Don't time it out if this fails
			log.Error("Failed to set timeout key", "err", err)
		} else {
			// Start the timeout timer
			go func() {
				// sleep
				time.Sleep(shared.REQUEST_COG_TIMEOUT)
				// this will trigger timeout if it hasnt been finished
				w.Repo.FailCogMessageDueToTimeoutIfTimedOut(requests.CogWebhookMessage{
					Input:  cogReqBody.Input,
					Error:  shared.TIMEOUT_ERROR,
					Status: requests.CogFailed,
				})
			}()
		}

		// Return queued indication
		return &responses.ApiSucceededResponse{
			RemainingCredits: remainingCredits,
			Settings:         initSettings,
			QueuedResponse: &responses.TaskQueuedResponse{
				ID:               requestId.String(),
				UIId:             generateReq.UIId,
				RemainingCredits: remainingCredits,
			},
		}, &initSettings, nil
	}

	// Wait for result
	for {
		select {
		case cogMsg := <-activeChl:
			switch cogMsg.Status {
			case requests.CogProcessing:
				err := w.Repo.SetGenerationStarted(requestId.String())
				if err != nil {
					log.Error("Failed to set generate started", "id", requestId, "err", err)
					return nil, &initSettings, WorkerInternalServerError()
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
					err = w.Redis.Client.Publish(w.Redis.Ctx, shared.REDIS_SSE_BROADCAST_CHANNEL, respBytes).Err()
					if err != nil {
						log.Error("Failed to publish live page update", "err", err)
					}
				}()
			case requests.CogSucceeded:
				outputs, err := w.Repo.SetGenerationSucceeded(requestId.String(), generateReq.Prompt, generateReq.NegativePrompt, cogMsg.Output, cogMsg.NSFWCount)
				if err != nil {
					log.Error("Failed to set generation succeeded", "id", upscale.ID, "err", err)
					return nil, &initSettings, WorkerInternalServerError()
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
					err = w.Redis.Client.Publish(w.Redis.Ctx, shared.REDIS_SSE_BROADCAST_CHANNEL, respBytes).Err()
					if err != nil {
						log.Error("Failed to publish live page update", "err", err)
					}
				}()
				// Analytics
				generation, err := w.Repo.GetGeneration(requestId)
				if err != nil {
					log.Error("Error getting generation for analytics", "err", err)
				}
				// Get durations in seconds
				if generation.StartedAt == nil {
					log.Error("Generation started at is nil", "id", cogMsg.Input.ID)
				}

				// Analytics
				duration := time.Now().Sub(*generation.StartedAt).Seconds()
				qDuration := (*generation.StartedAt).Sub(generation.CreatedAt).Seconds()
				go w.Track.GenerationSucceeded(user, cogMsg.Input, duration, qDuration, source, ipAddress)

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
				if generation.APITokenID != nil {
					err = w.Repo.SetTokenUsedAndIncrementCreditsSpent(int(*generateReq.NumOutputs), *generation.APITokenID)
					if err != nil {
						log.Error("Failed to set token used", "err", err)
					}
				}

				return &responses.ApiSucceededResponse{
					Outputs:          resOutputs,
					RemainingCredits: remainingCredits,
					Settings:         initSettings,
				}, &initSettings, nil
			case requests.CogFailed:
				if err := w.Repo.WithTx(func(tx *ent.Tx) error {
					DB := tx.Client()
					err := w.Repo.SetGenerationFailed(requestId.String(), cogMsg.Error, cogMsg.NSFWCount, DB)
					if err != nil {
						log.Error("Failed to set generation failed", "id", upscale.ID, "err", err)
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
						err = w.Redis.Client.Publish(w.Redis.Ctx, shared.REDIS_SSE_BROADCAST_CHANNEL, respBytes).Err()
						if err != nil {
							log.Error("Failed to publish live page update", "err", err)
						}
					}()
					// Analytics
					duration := time.Now().Sub(cogMsg.Input.LivePageData.CreatedAt).Seconds()
					go w.Track.GenerationFailed(user, cogMsg.Input, duration, cogMsg.Error, source, ipAddress)
					// Refund credits
					_, err = w.Repo.RefundCreditsToUser(user.ID, *generateReq.NumOutputs, DB)
					if err != nil {
						log.Error("Failed to refund credits", "err", err)
						return err
					}
					return nil
				}); err != nil {
					log.Error("Failed to set generation failed", "id", requestId, "err", err)
					return nil, &initSettings, WorkerInternalServerError()
				}

				return nil, &initSettings, &WorkerError{http.StatusInternalServerError, fmt.Errorf(cogMsg.Error), ""}
			}
		case <-time.After(shared.REQUEST_COG_TIMEOUT):
			if err := w.Repo.WithTx(func(tx *ent.Tx) error {
				DB := tx.Client()
				err := w.Repo.SetGenerationFailed(requestId.String(), shared.TIMEOUT_ERROR, 0, DB)
				if err != nil {
					log.Error("Failed to set generation failed", "id", upscale.ID, "err", err)
				}
				// Refund credits
				_, err = w.Repo.RefundCreditsToUser(user.ID, *generateReq.NumOutputs, DB)
				if err != nil {
					log.Error("Failed to refund credits", "err", err)
					return err
				}
				return nil
			}); err != nil {
				log.Error("Failed to set generation failed", "id", requestId, "err", err)
				return nil, &initSettings, WorkerInternalServerError()
			}

			return nil, &initSettings, &WorkerError{http.StatusInternalServerError, fmt.Errorf(shared.TIMEOUT_ERROR), ""}
		}
	}
}
