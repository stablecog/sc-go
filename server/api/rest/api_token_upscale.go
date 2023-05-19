package rest

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"

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

// POST upscale endpoint
// Handles creating a upscale with API token
func (c *RestAPI) HandleCreateUpscaleToken(w http.ResponseWriter, r *http.Request) {
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
	}

	// Parse request body
	reqBody, _ := io.ReadAll(r.Body)
	var upscaleReq requests.CreateUpscaleRequest
	err := json.Unmarshal(reqBody, &upscaleReq)
	if err != nil {
		responses.ErrUnableToParseJson(w, r)
		return
	}

	// Validation
	err = upscaleReq.Validate(true)
	if err != nil {
		responses.ErrBadRequest(w, r, err.Error(), "")
		return
	}

	// Get queue count
	nq, err := c.QueueThrottler.NumQueued(fmt.Sprintf("u:%s", user.ID.String()))
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
			nq, err = c.QueueThrottler.NumQueued(fmt.Sprintf("u:%s", user.ID.String()))
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

	// Parse request headers
	countryCode := utils.GetCountryCode(r)
	deviceInfo := utils.GetClientDeviceInfo(r)

	// Get model name for cog
	modelName := shared.GetCache().GetUpscaleModelNameFromID(upscaleReq.ModelId)
	if modelName == "" {
		log.Error("Error getting model name", "model_name", modelName)
		responses.ErrInternalServerError(w, r, "An unknown error has occurred")
		return
	}

	// Initiate upscale
	// We need to get width/height, from our database if output otherwise from the external image
	var width int32
	var height int32

	// Image Type
	imageUrl := upscaleReq.Input
	if upscaleReq.Type == requests.UpscaleRequestTypeImage {
		width, height, err = utils.GetImageWidthHeightFromUrl(imageUrl, shared.MAX_UPSCALE_IMAGE_SIZE)
		if err != nil {
			responses.ErrBadRequest(w, r, "image_url_width_height_error", "")
			return
		}
	}

	// Output Type
	var outputIDStr string
	if upscaleReq.Type == requests.UpscaleRequestTypeOutput {
		outputIDStr = upscaleReq.OutputID.String()
		output, err := c.Repo.GetGenerationOutputForUser(upscaleReq.OutputID, user.ID)
		if err != nil {
			if ent.IsNotFound(err) {
				responses.ErrBadRequest(w, r, "output_not_found", "")
				return
			}
			log.Error("Error getting output", "err", err)
			responses.ErrInternalServerError(w, r, "Error getting output")
			return
		}
		if output.UpscaledImagePath != nil {
			responses.ErrBadRequest(w, r, "image_already_upscaled", "")
			return
		}
		imageUrl = utils.GetURLFromImagePath(output.ImagePath)

		// Get width/height of generation
		width, height, err = c.Repo.GetGenerationOutputWidthHeight(upscaleReq.OutputID)
		if err != nil {
			responses.ErrBadRequest(w, r, "Unable to retrieve width/height for upscale", "")
			return
		}
	}

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
		deducted, err := c.Repo.DeductCreditsFromUser(user.ID, 1, DB)
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

		// Create upscale
		upscale, err := c.Repo.CreateUpscale(
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
			&apiToken.ID,
			DB)
		if err != nil {
			log.Error("Error creating upscale", "err", err)
			responses.ErrInternalServerError(w, r, "Error creating upscale")
			return err
		}

		// Request Id matches upscale ID
		requestId = upscale.ID.String()

		// For live page update
		livePageMsg = shared.LivePageMessage{
			ProcessType:      shared.UPSCALE,
			ID:               utils.Sha256(requestId),
			CountryCode:      countryCode,
			Status:           shared.LivePageQueued,
			TargetNumOutputs: 1,
			Width:            width,
			Height:           height,
			CreatedAt:        upscale.CreatedAt,
			ProductID:        user.ActiveProductID,
			Source:           shared.OperationSourceTypeAPI,
		}

		// Send to the cog
		cogReqBody = requests.CogQueueRequest{
			WebhookEventsFilter: []requests.CogEventFilter{requests.CogEventFilterStart, requests.CogEventFilterStart},
			WebhookUrl:          fmt.Sprintf("%s/v1/worker/webhook", utils.GetEnv("PUBLIC_API_URL", "")),
			Input: requests.BaseCogRequest{
				APIRequest:           true,
				ID:                   requestId,
				IP:                   utils.GetIPAddress(r),
				UIId:                 upscaleReq.UIId,
				UserID:               &user.ID,
				DeviceInfo:           deviceInfo,
				StreamID:             upscaleReq.StreamID,
				LivePageData:         &livePageMsg,
				GenerationOutputID:   outputIDStr,
				Image:                imageUrl,
				ProcessType:          shared.UPSCALE,
				Width:                fmt.Sprint(width),
				Height:               fmt.Sprint(height),
				UpscaleModel:         modelName,
				ModelId:              upscaleReq.ModelId,
				OutputImageExtension: string(shared.DEFAULT_UPSCALE_OUTPUT_EXTENSION),
				OutputImageQuality:   fmt.Sprint(shared.DEFAULT_UPSCALE_OUTPUT_QUALITY),
				Type:                 upscaleReq.Type,
			},
		}

		// Add channel to sync array (basically a thread-safe map)
		c.SMap.Put(requestId, activeChl)

		err = c.Redis.EnqueueCogRequest(r.Context(), cogReqBody)
		if err != nil {
			log.Error("Failed to write request %s to queue: %v", requestId, err)
			responses.ErrInternalServerError(w, r, "Failed to queue upscale request")
			return err
		}

		c.QueueThrottler.IncrementBy(1, fmt.Sprintf("u:%s", user.ID.String()))
		return nil
	}); err != nil {
		log.Error("Error in transaction", "err", err)
		return
	}
	defer c.QueueThrottler.DecrementBy(1, fmt.Sprintf("u:%s", user.ID.String()))

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
	go c.Track.UpscaleStarted(user, cogReqBody.Input, utils.GetIPAddress(r))

	// Wait for result
	for {
		select {
		case cogMsg := <-activeChl:
			switch cogMsg.Status {
			case requests.CogProcessing:
				err := c.Repo.SetUpscaleStarted(requestId)
				if err != nil {
					log.Error("Failed to set upscale started", "id", requestId, "err", err)
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
				output, err := c.Repo.SetUpscaleSucceeded(requestId, outputIDStr, imageUrl, cogMsg.Output)
				if err != nil {
					log.Error("Failed to set upscale succeeded", "id", upscale.ID, "err", err)
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
				upscale, err := c.Repo.GetUpscale(uuid.MustParse(requestId))
				if err != nil {
					log.Error("Error getting upscale for analytics", "err", err)
				}
				// Get durations in seconds
				if upscale.StartedAt == nil {
					log.Error("Upscale started at is nil", "id", cogMsg.Input.ID)
				}
				duration := time.Now().Sub(*upscale.StartedAt).Seconds()
				qDuration := (*upscale.StartedAt).Sub(upscale.CreatedAt).Seconds()
				go c.Track.UpscaleSucceeded(user, cogMsg.Input, duration, qDuration, "system")

				// Format response
				resOutputs := []responses.ApiOutput{
					{
						URL: utils.GetURLFromImagePath(output.ImagePath),
						ID:  output.ID,
					},
				}

				// Set token used
				err = c.Repo.SetTokenUsedAndIncrementCreditsSpent(1, *upscale.APITokenID)
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
					err := c.Repo.SetUpscaleFailed(requestId, cogMsg.Error, DB)
					if err != nil {
						log.Error("Failed to set upscale failed", "id", upscale.ID, "err", err)
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
					go c.Track.UpscaleFailed(user, cogMsg.Input, duration, cogMsg.Error, "system")
					// Refund credits
					_, err = c.Repo.RefundCreditsToUser(user.ID, int32(1), DB)
					if err != nil {
						log.Error("Failed to refund credits", "err", err)
						return err
					}
					return nil
				}); err != nil {
					log.Error("Failed to set upscale failed", "id", requestId, "err", err)
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
				err := c.Repo.SetUpscaleFailed(requestId, shared.TIMEOUT_ERROR, DB)
				if err != nil {
					log.Error("Failed to set upscale failed", "id", upscale.ID, "err", err)
				}
				// Refund credits
				_, err = c.Repo.RefundCreditsToUser(user.ID, int32(1), DB)
				if err != nil {
					log.Error("Failed to refund credits", "err", err)
					return err
				}
				return nil
			}); err != nil {
				log.Error("Failed to set upscale failed", "id", requestId, "err", err)
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
