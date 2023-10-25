package scworker

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
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

func (w *SCWorker) CreateUpscale(source enttypes.SourceType,
	r *http.Request,
	user *ent.User,
	apiTokenId *uuid.UUID,
	upscaleReq requests.CreateUpscaleRequest) (*responses.ApiSucceededResponse, *responses.ImageUpscaleSettingsResponse, *WorkerError) {
	// Queue priority for MQ
	var queuePriority uint8 = shared.QUEUE_PRIORITY_2

	free := user.ActiveProductID == nil
	if free {
		// Re-evaluate if they have paid credits
		count, err := w.Repo.GetNonFreeCreditSum(user.ID)
		if err != nil {
			log.Error("Error getting paid credit sum for users", "err", err)
			return nil, nil, WorkerInternalServerError()
		}
		free = count <= 0
		if !free {
			queuePriority = shared.QUEUE_PRIORITY_3
		}
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
			queuePriority = shared.QUEUE_PRIORITY_5
			// Pro
		case stripe.GetProductIDs()[2]:
			qMax = shared.MAX_QUEUED_ITEMS_PRO
			queuePriority = shared.QUEUE_PRIORITY_5
		// Ultimate
		case stripe.GetProductIDs()[3]:
			qMax = shared.MAX_QUEUED_ITEMS_ULTIMATE
			queuePriority = shared.QUEUE_PRIORITY_5
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

	if isSuperAdmin {
		queuePriority = shared.QUEUE_PRIORITY_4
	}

	// With gift credits, give them a priority in between super admins and paid credits
	if !free && queuePriority < shared.QUEUE_PRIORITY_5 {
		// Re-evaluate if they have paid credits
		paidCount, err := w.Repo.GetPaidCreditSum(user.ID)
		if err != nil {
			log.Error("Error getting paid credit sum for users", "err", err)
			return nil, nil, WorkerInternalServerError()
		}
		if paidCount > 0 {
			queuePriority = shared.QUEUE_PRIORITY_5
		}
	}

	if user.BannedAt != nil {
		return nil, nil, &WorkerError{http.StatusForbidden, fmt.Errorf("user_banned"), ""}
	}

	// Validation
	err = upscaleReq.Validate(source != enttypes.SourceTypeWebUI)
	if err != nil {
		return nil, nil, &WorkerError{http.StatusBadRequest, err, ""}
	}

	// Set settings resp
	initSettings := responses.ImageUpscaleSettingsResponse{
		ModelId: *upscaleReq.ModelId,
		Input:   upscaleReq.Input,
	}

	// Get queue count
	// UI has no overflow so it's a different flow
	if source == enttypes.SourceTypeWebUI {
		// Get queue count
		nq, err := w.QueueThrottler.NumQueued(fmt.Sprintf("u:%s", user.ID.String()))
		if err != nil {
			log.Warn("Error getting queue count for user", "err", err, "user_id", user.ID)
		}
		if err == nil && nq >= qMax {
			return nil, nil, &WorkerError{http.StatusBadRequest, fmt.Errorf("queue_limit_reached"), ""}
		}
	} else {
		nq, err := w.QueueThrottler.NumQueued(fmt.Sprintf("u:%s", user.ID.String()))
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
				return nil, nil, &WorkerError{http.StatusBadRequest, fmt.Errorf("queue_limit_reached"), ""}
			}
			// Overflow size can be 0 so we need to add 1
			overflowSize++
			w.QueueThrottler.IncrementBy(1, fmt.Sprintf("of:%s", user.ID.String()))
			for {
				time.Sleep(time.Duration(shared.QUEUE_OVERFLOW_PENALTY_MS*overflowSize) * time.Millisecond)
				nq, err = w.QueueThrottler.NumQueued(fmt.Sprintf("u:%s", user.ID.String()))
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

	// Get model name for cog
	modelName := shared.GetCache().GetUpscaleModelNameFromID(*upscaleReq.ModelId)
	if modelName == "" {
		log.Error("Error getting model name", "model_name", modelName)
		return nil, nil, WorkerInternalServerError()
	}

	// Initiate upscale
	// We need to get width/height, from our database if output otherwise from the external image
	var width int32
	var height int32

	// Image Type
	var imageUrl string
	var headUrl string
	if strings.HasPrefix(upscaleReq.Input, "s3://") {
		// Remove prefix
		imageUrl = upscaleReq.Input[5:]
		// Get signed URL from s3
		// Hash user ID to see if it belongs to this user
		uidHash := utils.Sha256(user.ID.String())
		if !strings.HasPrefix(imageUrl, fmt.Sprintf("%s/", uidHash)) {
			return nil, &initSettings, &WorkerError{http.StatusUnauthorized, fmt.Errorf("image_not_owned"), ""}
		}
		// Verify exists in bucket
		_, err := w.S3.HeadObject(&s3.HeadObjectInput{
			Bucket: aws.String(utils.GetEnv().S3Img2ImgBucketName),
			Key:    aws.String(imageUrl),
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
			Bucket: aws.String(utils.GetEnv().S3Img2ImgBucketName),
			Key:    aws.String(imageUrl),
		})
		urlStr, err := req.Presign(168 * time.Hour)
		if err != nil {
			log.Error("Error signing init image URL", "err", err)
			return nil, &initSettings, &WorkerError{http.StatusInternalServerError, fmt.Errorf("unknown_error"), ""}
		}
		// Presign URL for head request
		headReq, _ := w.S3.HeadObjectRequest(&s3.HeadObjectInput{
			Bucket: aws.String(utils.GetEnv().S3Img2ImgBucketName),
			Key:    aws.String(imageUrl),
		})
		headUrl, err = headReq.Presign(168 * time.Hour)
		if err != nil {
			log.Error("Error signing init image URL", "err", err)
			return nil, &initSettings, &WorkerError{http.StatusInternalServerError, fmt.Errorf("unknown_error"), ""}
		}
		imageUrl = urlStr
	} else {
		imageUrl = upscaleReq.Input
	}

	if *upscaleReq.Type == requests.UpscaleRequestTypeImage {
		width, height, err = utils.GetImageWidthHeightFromUrl(imageUrl, headUrl, shared.MAX_UPSCALE_IMAGE_SIZE)
		if err != nil {
			return nil, &initSettings, &WorkerError{http.StatusBadRequest, fmt.Errorf("image_url_width_height_error"), ""}
		}
		if width*height > shared.MAX_UPSCALE_MEGAPIXELS {
			return nil, &initSettings, &WorkerError{http.StatusBadRequest, fmt.Errorf("image_url_width_height_error"), fmt.Sprintf("Image cannot exceed %d megapixels", shared.MAX_UPSCALE_MEGAPIXELS/1000000)}
		}
	}

	// Output Type
	var outputIDStr string
	if *upscaleReq.Type == requests.UpscaleRequestTypeOutput {
		outputIDStr = upscaleReq.OutputID.String()
		var output *ent.GenerationOutput
		if source == enttypes.SourceTypeDiscord {
			output, err = w.Repo.GetGenerationOutput(*upscaleReq.OutputID)
		} else {
			output, err = w.Repo.GetGenerationOutputForUser(*upscaleReq.OutputID, user.ID)
		}
		if err != nil {
			if ent.IsNotFound(err) {
				return nil, &initSettings, &WorkerError{http.StatusBadRequest, fmt.Errorf("output_not_found"), ""}
			}
			log.Error("Error getting output", "err", err)
			return nil, nil, WorkerInternalServerError()
		}
		if output.UpscaledImagePath != nil {
			// Format response
			resOutputs := []responses.ApiOutput{
				{
					URL:              utils.GetEnv().GetURLFromImagePath(output.ImagePath),
					UpscaledImageURL: utils.ToPtr(utils.GetEnv().GetURLFromImagePath(*output.UpscaledImagePath)),
					ID:               output.ID,
				},
			}

			remainingCredits, err := w.Repo.GetNonExpiredCreditTotalForUser(user.ID, nil)
			if err != nil {
				log.Error("Error getting remaining credits", "err", err)
				return nil, nil, WorkerInternalServerError()
			}

			return &responses.ApiSucceededResponse{
				Outputs:          resOutputs,
				RemainingCredits: remainingCredits,
				Settings:         initSettings,
			}, &initSettings, nil
		}
		imageUrl = utils.GetEnv().GetURLFromImagePath(output.ImagePath)

		// Get width/height of generation
		width, height, err = w.Repo.GetGenerationOutputWidthHeight(*upscaleReq.OutputID)
		if err != nil {
			log.Errorf("Error getting generation output width/height %v", err)
			return nil, &initSettings, WorkerInternalServerError()
		}
	}

	// For live page update
	var livePageMsg shared.LivePageMessage
	// For keeping track of this request as it gets sent to the worker
	var requestId uuid.UUID
	var queueId string
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
		// Bind a client to the transaction
		DB := tx.Client()
		// Deduct credits from user
		deducted, err := w.Repo.DeductCreditsFromUser(user.ID, 1, false, DB)
		if err != nil {
			log.Error("Error deducting credits", "err", err)
			return err
		} else if !deducted {
			return responses.InsufficientCreditsErr
		}

		remainingCredits, err = w.Repo.GetNonExpiredCreditTotalForUser(user.ID, DB)
		if err != nil {
			log.Error("Error getting remaining credits", "err", err)
			return err
		}

		// Create upscale
		upscale, err := w.Repo.CreateUpscale(
			user.ID,
			width,
			height,
			string(deviceInfo.DeviceType),
			deviceInfo.DeviceOs,
			deviceInfo.DeviceBrowser,
			countryCode,
			upscaleReq,
			user.ActiveProductID,
			false,
			apiTokenId,
			source,
			DB)
		if err != nil {
			log.Error("Error creating upscale", "err", err)
			return err
		}

		// Request Id matches upscale ID
		requestId = upscale.ID
		// queueId is message_id in amqp
		queueId = utils.Sha256(requestId.String())

		// For live page update
		livePageMsg = shared.LivePageMessage{
			ProcessType:      shared.UPSCALE,
			ID:               utils.Sha256(requestId.String()),
			CountryCode:      countryCode,
			Status:           shared.LivePageQueued,
			TargetNumOutputs: 1,
			Width:            utils.ToPtr(width),
			Height:           utils.ToPtr(height),
			CreatedAt:        upscale.CreatedAt,
			ProductID:        user.ActiveProductID,
			Source:           source,
		}

		// Send to the cog
		cogReqBody = requests.CogQueueRequest{
			WebhookEventsFilter: []requests.CogEventFilter{requests.CogEventFilterStart, requests.CogEventFilterStart},
			WebhookUrl:          fmt.Sprintf("%s/v1/worker/webhook", utils.GetEnv().PublicApiUrl),
			Input: requests.BaseCogRequest{
				APIRequest:           source != enttypes.SourceTypeWebUI,
				ID:                   requestId,
				IP:                   ipAddress,
				UIId:                 upscaleReq.UIId,
				UserID:               &user.ID,
				DeviceInfo:           deviceInfo,
				StreamID:             upscaleReq.StreamID,
				LivePageData:         &livePageMsg,
				GenerationOutputID:   outputIDStr,
				Image:                imageUrl,
				ProcessType:          shared.UPSCALE,
				Width:                utils.ToPtr(width),
				Height:               utils.ToPtr(height),
				UpscaleModel:         modelName,
				ModelId:              *upscaleReq.ModelId,
				OutputImageExtension: string(shared.DEFAULT_UPSCALE_OUTPUT_EXTENSION),
				OutputImageQuality:   utils.ToPtr(shared.DEFAULT_UPSCALE_OUTPUT_QUALITY),
				Type:                 *upscaleReq.Type,
			},
		}

		if source == enttypes.SourceTypeWebUI {
			cogReqBody.Input.UIId = upscaleReq.UIId
			cogReqBody.Input.StreamID = upscaleReq.StreamID
		}

		_, err = w.Repo.AddToQueueLog(queueId, int(queuePriority), DB)
		if err != nil {
			log.Error("Error adding to queue log", "err", err)
			return err
		}

		err = w.MQClient.Publish(queueId, cogReqBody, queuePriority)
		if err != nil {
			log.Error("Failed to write request %s to queue: %v", queueId, err)
			return err
		}

		w.QueueThrottler.IncrementBy(1, fmt.Sprintf("u:%s", user.ID.String()))
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
		defer w.QueueThrottler.DecrementBy(1, fmt.Sprintf("u:%s", user.ID.String()))
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
	go w.Track.UpscaleStarted(user, cogReqBody.Input, source, ipAddress)

	// Set timeout delay for UI
	if source == enttypes.SourceTypeWebUI {
		// Set timeout key
		err = w.Redis.SetCogRequestStreamID(w.Redis.Ctx, requestId.String(), upscaleReq.StreamID)
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

		// Get queued position
		queueLog, err := w.Repo.GetQueuedItems(nil)
		if err != nil {
			log.Error("Error getting queue log", "err", err)
		}

		// Return queued indication
		return &responses.ApiSucceededResponse{
			RemainingCredits: remainingCredits,
			Settings:         initSettings,
			QueuedResponse: &responses.TaskQueuedResponse{
				ID:               requestId.String(),
				UIId:             upscaleReq.UIId,
				RemainingCredits: remainingCredits,
				QueuedId:         queueId,
				QueueItems:       queueLog,
			},
		}, &initSettings, nil
	}

	// Wait for result
	for {
		select {
		case cogMsg := <-activeChl:
			_, err := w.Repo.DeleteFromQueueLog(queueId, nil)
			if err != nil {
				log.Error("Error deleting from queue log", "err", err)
			}
			switch cogMsg.Status {
			case requests.CogProcessing:
				err := w.Repo.SetUpscaleStarted(requestId.String())
				if err != nil {
					log.Error("Failed to set upscale started", "id", requestId, "err", err)
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
				output, err := w.Repo.SetUpscaleSucceeded(requestId.String(), outputIDStr, imageUrl, cogMsg.Output)
				if err != nil {
					log.Error("Failed to set upscale succeeded", "id", upscale.ID, "err", err)
					return nil, &initSettings, WorkerInternalServerError()
				}
				// Send live page update
				go func() {
					cogMsg.Input.LivePageData.Status = shared.LivePageSucceeded
					now := time.Now()
					cogMsg.Input.LivePageData.CompletedAt = &now
					cogMsg.Input.LivePageData.ActualNumOutputs = len(cogMsg.Output.Images)
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
				upscale, err := w.Repo.GetUpscale(requestId)
				if err != nil {
					log.Error("Error getting upscale for analytics", "err", err)
				}
				// Get durations in seconds
				if upscale.StartedAt == nil {
					log.Error("Upscale started at is nil", "id", cogMsg.Input.ID)
				}
				// Analytics
				duration := time.Now().Sub(*upscale.StartedAt).Seconds()
				qDuration := (*upscale.StartedAt).Sub(upscale.CreatedAt).Seconds()
				go w.Track.UpscaleSucceeded(user, cogMsg.Input, duration, qDuration, source, ipAddress)

				// Format response
				resOutputs := []responses.ApiOutput{
					{
						URL:              utils.GetEnv().GetURLFromImagePath(output.ImagePath),
						UpscaledImageURL: utils.ToPtr(utils.GetEnv().GetURLFromImagePath(output.ImagePath)),
						ID:               output.ID,
					},
				}

				// Set token used
				if upscale.APITokenID != nil {
					err = w.Repo.SetTokenUsedAndIncrementCreditsSpent(1, *upscale.APITokenID)
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
					err := w.Repo.SetUpscaleFailed(requestId.String(), cogMsg.Error, DB)
					if err != nil {
						log.Error("Failed to set upscale failed", "id", upscale.ID, "err", err)
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
					//  Analytics
					duration := time.Now().Sub(cogMsg.Input.LivePageData.CreatedAt).Seconds()
					go w.Track.UpscaleFailed(user, cogMsg.Input, duration, cogMsg.Error, source, ipAddress)
					// Refund credits
					_, err = w.Repo.RefundCreditsToUser(user.ID, int32(1), DB)
					if err != nil {
						log.Error("Failed to refund credits", "err", err)
						return err
					}
					return nil
				}); err != nil {
					log.Error("Failed to set upscale failed", "id", requestId, "err", err)
					return nil, &initSettings, WorkerInternalServerError()
				}

				return nil, &initSettings, WorkerInternalServerError()
			}
		case <-time.After(shared.REQUEST_COG_TIMEOUT):
			_, err := w.Repo.DeleteFromQueueLog(queueId, nil)
			if err != nil {
				log.Error("Error deleting from queue log", "err", err)
			}
			if err := w.Repo.WithTx(func(tx *ent.Tx) error {
				DB := tx.Client()
				err := w.Repo.SetUpscaleFailed(requestId.String(), shared.TIMEOUT_ERROR, DB)
				if err != nil {
					log.Error("Failed to set upscale failed", "id", upscale.ID, "err", err)
				}
				// Refund credits
				_, err = w.Repo.RefundCreditsToUser(user.ID, int32(1), DB)
				if err != nil {
					log.Error("Failed to refund credits", "err", err)
					return err
				}
				return nil
			}); err != nil {
				log.Error("Failed to set upscale failed", "id", requestId, "err", err)
				return nil, &initSettings, WorkerInternalServerError()
			}

			return nil, &initSettings, &WorkerError{http.StatusInternalServerError, fmt.Errorf(shared.TIMEOUT_ERROR), ""}
		}
	}
}
