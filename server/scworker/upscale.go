package scworker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database"
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

func CreateUpscale(ctx context.Context,
	source enttypes.SourceType,
	r *http.Request,
	repo *repository.Repository,
	redis *database.RedisWrapper,
	SMap *shared.SyncMap[chan requests.CogWebhookMessage],
	qThrottler *shared.UserQueueThrottlerMap,
	user *ent.User,
	upscaleReq requests.CreateUpscaleRequest) (*responses.ApiSucceededResponse, error) {
	free := user.ActiveProductID == nil
	if free {
		// Re-evaluate if they have paid credits
		count, err := repo.HasPaidCredits(user.ID)
		if err != nil {
			log.Error("Error getting paid credit sum for users", "err", err)
			return nil, err
		}
		free = count <= 0
	}

	var qMax int
	roles, err := repo.GetRoles(user.ID)
	if err != nil {
		log.Error("Error getting roles for user", "err", err)
		return nil, err
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
		return nil, &WorkerError{http.StatusForbidden, fmt.Errorf("user_banned"), ""}
	}

	// Validation
	err = upscaleReq.Validate(true)
	if err != nil {
		return nil, &WorkerError{http.StatusBadRequest, err, ""}
	}

	// Set settings resp
	initSettings := responses.ImageUpscaleSettingsResponse{
		ModelId: *upscaleReq.ModelId,
		Input:   upscaleReq.Input,
	}

	// Get queue count
	nq, err := qThrottler.NumQueued(fmt.Sprintf("u:%s", user.ID.String()))
	if err != nil {
		log.Warn("Error getting queue count", "err", err, "user_id", user.ID.String())
	}
	if err == nil && nq > qMax {
		// Get queue overflow size
		overflowSize, err := qThrottler.NumQueued(fmt.Sprintf("of:%s", user.ID.String()))
		if err != nil {
			log.Warn("Error getting queue overflow count", "err", err, "user_id", user.ID.String())
		}
		// If overflow size is greater than max, return error
		if overflowSize > shared.QUEUE_OVERFLOW_MAX {
			return nil, &WorkerError{http.StatusBadRequest, fmt.Errorf("queue_limit_reached"), ""}
		}
		// Overflow size can be 0 so we need to add 1
		overflowSize++
		qThrottler.IncrementBy(1, fmt.Sprintf("of:%s", user.ID.String()))
		for {
			time.Sleep(time.Duration(shared.QUEUE_OVERFLOW_PENALTY_MS*overflowSize) * time.Millisecond)
			nq, err = qThrottler.NumQueued(fmt.Sprintf("u:%s", user.ID.String()))
			if err != nil {
				log.Warn("Error getting queue count", "err", err, "user_id", user.ID.String())
			}
			if err == nil && nq <= qMax {
				qThrottler.DecrementBy(1, fmt.Sprintf("of:%s", user.ID.String()))
				break
			}
			// Update overflow size
			overflowSize, err = qThrottler.NumQueued(fmt.Sprintf("of:%s", user.ID.String()))
			if err != nil {
				log.Warn("Error getting queue overflow count", "err", err, "user_id", user.ID.String())
			}
			overflowSize++
		}
	}

	// Parse request headers
	var countryCode string
	var deviceInfo utils.ClientDeviceInfo
	ipAddress := "internal"
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
		return nil, WorkerInternalServerError()
	}

	// Initiate upscale
	// We need to get width/height, from our database if output otherwise from the external image
	var width int32
	var height int32

	// Image Type
	imageUrl := upscaleReq.Input
	if *upscaleReq.Type == requests.UpscaleRequestTypeImage {
		width, height, err = utils.GetImageWidthHeightFromUrl(imageUrl, shared.MAX_UPSCALE_IMAGE_SIZE)
		if err != nil {
			return nil, &WorkerError{http.StatusBadRequest, fmt.Errorf("image_url_width_height_error"), ""}
		}
		if width*height > shared.MAX_UPSCALE_MEGAPIXELS {
			return nil, &WorkerError{http.StatusBadRequest, fmt.Errorf("image_url_width_height_error"), fmt.Sprintf("Image cannot exceed %d megapixels", shared.MAX_UPSCALE_MEGAPIXELS/1000000)}
		}
	}

	// Output Type
	var outputIDStr string
	if *upscaleReq.Type == requests.UpscaleRequestTypeOutput {
		outputIDStr = upscaleReq.OutputID.String()
		output, err := repo.GetPublicGenerationOutput(*upscaleReq.OutputID)
		if err != nil {
			if ent.IsNotFound(err) {
				return nil, &WorkerError{http.StatusBadRequest, fmt.Errorf("output_not_found"), ""}
			}
			log.Error("Error getting output", "err", err)
			return nil, WorkerInternalServerError()
		}
		if output.UpscaledImagePath != nil {
			// Format response
			resOutputs := []responses.ApiOutput{
				{
					URL:              utils.GetURLFromImagePath(output.ImagePath),
					UpscaledImageURL: utils.ToPtr(utils.GetURLFromImagePath(*output.UpscaledImagePath)),
					ID:               output.ID,
				},
			}

			remainingCredits, err := repo.GetNonExpiredCreditTotalForUser(user.ID, nil)
			if err != nil {
				log.Error("Error getting remaining credits", "err", err)
				return nil, WorkerInternalServerError()
			}

			return &responses.ApiSucceededResponse{
				Outputs:          resOutputs,
				RemainingCredits: remainingCredits,
				Settings:         initSettings,
			}, nil
		}
		imageUrl = utils.GetURLFromImagePath(output.ImagePath)

		// Get width/height of generation
		width, height, err = repo.GetGenerationOutputWidthHeight(*upscaleReq.OutputID)
		if err != nil {
			return nil, WorkerInternalServerError()
		}
	}

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
	if err := repo.WithTx(func(tx *ent.Tx) error {
		// Bind a client to the transaction
		DB := tx.Client()
		// Deduct credits from user
		deducted, err := repo.DeductCreditsFromUser(user.ID, 1, DB)
		if err != nil {
			log.Error("Error deducting credits", "err", err)
			return err
		} else if !deducted {
			return responses.InsufficientCreditsErr
		}

		remainingCredits, err = repo.GetNonExpiredCreditTotalForUser(user.ID, DB)
		if err != nil {
			log.Error("Error getting remaining credits", "err", err)
			return err
		}

		// Create upscale
		upscale, err := repo.CreateUpscale(
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
			nil,
			source,
			DB)
		if err != nil {
			log.Error("Error creating upscale", "err", err)
			return err
		}

		// Request Id matches upscale ID
		requestId = upscale.ID

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
			WebhookUrl:          fmt.Sprintf("%s/v1/worker/webhook", utils.GetEnv("PUBLIC_API_URL", "")),
			Input: requests.BaseCogRequest{
				APIRequest:           true,
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

		err = redis.EnqueueCogRequest(ctx, shared.COG_REDIS_QUEUE, cogReqBody)
		if err != nil {
			log.Error("Failed to write request %s to queue: %v", requestId, err)
			return err
		}

		qThrottler.IncrementBy(1, fmt.Sprintf("u:%s", user.ID.String()))
		return nil
	}); err != nil {
		log.Error("Error in transaction", "err", err)
		if errors.Is(err, responses.InsufficientCreditsErr) {
			return nil, responses.InsufficientCreditsErr
		}
		return nil, WorkerInternalServerError()
	}
	// Add channel to sync array (basically a thread-safe map)
	SMap.Put(requestId.String(), activeChl)
	defer SMap.Delete(requestId.String())
	defer qThrottler.DecrementBy(1, fmt.Sprintf("u:%s", user.ID.String()))

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
		err = redis.Client.Publish(redis.Ctx, shared.REDIS_SSE_BROADCAST_CHANNEL, respBytes).Err()
		if err != nil {
			log.Error("Failed to publish live page update", "err", err)
		}
	}()

	// Analytics
	// ! TODO
	// go c.Track.UpscaleStarted(user, cogReqBody.Input, utils.GetIPAddress(r))

	// Wait for result
	for {
		select {
		case cogMsg := <-activeChl:
			switch cogMsg.Status {
			case requests.CogProcessing:
				err := repo.SetUpscaleStarted(requestId.String())
				if err != nil {
					log.Error("Failed to set upscale started", "id", requestId, "err", err)
					return nil, WorkerInternalServerError()
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
					err = redis.Client.Publish(redis.Ctx, shared.REDIS_SSE_BROADCAST_CHANNEL, respBytes).Err()
					if err != nil {
						log.Error("Failed to publish live page update", "err", err)
					}
				}()
			case requests.CogSucceeded:
				output, err := repo.SetUpscaleSucceeded(requestId.String(), outputIDStr, imageUrl, cogMsg.Output)
				if err != nil {
					log.Error("Failed to set upscale succeeded", "id", upscale.ID, "err", err)
					return nil, WorkerInternalServerError()
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
					err = redis.Client.Publish(redis.Ctx, shared.REDIS_SSE_BROADCAST_CHANNEL, respBytes).Err()
					if err != nil {
						log.Error("Failed to publish live page update", "err", err)
					}
				}()
				// Analytics
				upscale, err := repo.GetUpscale(requestId)
				if err != nil {
					log.Error("Error getting upscale for analytics", "err", err)
				}
				// Get durations in seconds
				if upscale.StartedAt == nil {
					log.Error("Upscale started at is nil", "id", cogMsg.Input.ID)
				}
				// ! TODO
				// duration := time.Now().Sub(*upscale.StartedAt).Seconds()
				// qDuration := (*upscale.StartedAt).Sub(upscale.CreatedAt).Seconds()
				// go c.Track.UpscaleSucceeded(user, cogMsg.Input, duration, qDuration, utils.GetIPAddress(r))

				// Format response
				resOutputs := []responses.ApiOutput{
					{
						URL:              utils.GetURLFromImagePath(output.ImagePath),
						UpscaledImageURL: utils.ToPtr(utils.GetURLFromImagePath(output.ImagePath)),
						ID:               output.ID,
					},
				}

				// ! TODO Set token used
				// err = c.Repo.SetTokenUsedAndIncrementCreditsSpent(1, *upscale.APITokenID)
				// if err != nil {
				// 	log.Error("Failed to set token used", "err", err)
				// }

				return &responses.ApiSucceededResponse{
					Outputs:          resOutputs,
					RemainingCredits: remainingCredits,
					Settings:         initSettings,
				}, nil
			case requests.CogFailed:
				if err := repo.WithTx(func(tx *ent.Tx) error {
					DB := tx.Client()
					err := repo.SetUpscaleFailed(requestId.String(), cogMsg.Error, DB)
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
						err = redis.Client.Publish(redis.Ctx, shared.REDIS_SSE_BROADCAST_CHANNEL, respBytes).Err()
						if err != nil {
							log.Error("Failed to publish live page update", "err", err)
						}
					}()
					// ! TODO Analytics
					// duration := time.Now().Sub(cogMsg.Input.LivePageData.CreatedAt).Seconds()
					// go c.Track.UpscaleFailed(user, cogMsg.Input, duration, cogMsg.Error, utils.GetIPAddress(r))
					// Refund credits
					_, err = repo.RefundCreditsToUser(user.ID, int32(1), DB)
					if err != nil {
						log.Error("Failed to refund credits", "err", err)
						return err
					}
					return nil
				}); err != nil {
					log.Error("Failed to set upscale failed", "id", requestId, "err", err)
					return nil, WorkerInternalServerError()
				}

				return nil, WorkerInternalServerError()
			}
		case <-time.After(shared.REQUEST_COG_TIMEOUT):
			if err := repo.WithTx(func(tx *ent.Tx) error {
				DB := tx.Client()
				err := repo.SetUpscaleFailed(requestId.String(), shared.TIMEOUT_ERROR, DB)
				if err != nil {
					log.Error("Failed to set upscale failed", "id", upscale.ID, "err", err)
				}
				// Refund credits
				_, err = repo.RefundCreditsToUser(user.ID, int32(1), DB)
				if err != nil {
					log.Error("Failed to refund credits", "err", err)
					return err
				}
				return nil
			}); err != nil {
				log.Error("Failed to set upscale failed", "id", requestId, "err", err)
				return nil, WorkerInternalServerError()
			}

			return nil, &WorkerError{http.StatusInternalServerError, fmt.Errorf(shared.TIMEOUT_ERROR), ""}
		}
	}
}
