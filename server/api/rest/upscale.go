package rest

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"

	"github.com/go-chi/render"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/repository"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/server/responses"
	"github.com/stablecog/sc-go/shared"
	"github.com/stablecog/sc-go/utils"
)

func (c *RestAPI) HandleUpscale(w http.ResponseWriter, r *http.Request) {
	var user *ent.User
	if user = c.GetUserIfAuthenticated(w, r); user == nil {
		return
	}

	// Parse request body
	reqBody, _ := io.ReadAll(r.Body)
	var upscaleReq requests.CreateUpscaleRequest
	err := json.Unmarshal(reqBody, &upscaleReq)
	if err != nil {
		responses.ErrUnableToParseJson(w, r)
		return
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

	// Validation
	err = upscaleReq.Validate()
	if err != nil {
		responses.ErrBadRequest(w, r, err.Error(), "")
		return
	}

	// Get queue count
	nq, err := c.QueueThrottler.NumQueued(user.ID.String())
	if err != nil {
		log.Warn("Error getting queue count for user", "err", err, "user_id", user.ID)
	}
	if err == nil && nq > qMax {
		responses.ErrBadRequest(w, r, "queue_limit_reached", "")
		return
	}

	// Parse request headers
	countryCode := utils.GetCountryCode(r)
	deviceInfo := utils.GetClientDeviceInfo(r)

	// Get model name for cog
	modelName := shared.GetCache().GetUpscaleModelNameFromID(upscaleReq.ModelId)
	if modelName == "" {
		log.Error("Error getting model name", "model_name", modelName)
		responses.ErrInternalServerError(w, r, "An unknown error has occured")
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
	// The cog request body
	var cogReqBody requests.CogQueueRequest
	// The total remaining credits
	var remainingCredits int

	// Wrap everything in a DB transaction
	// We do this since we want our credit deduction to be atomic with the whole process
	if err := c.Repo.WithTx(func(tx *ent.Tx) error {
		// Bind transaction to client
		DB := tx.Client()

		// Charge credits
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
			DB)
		if err != nil {
			log.Error("Error creating upscale", "err", err)
			responses.ErrInternalServerError(w, r, "Error creating upscale")
			return err
		}

		// Request ID matches upscale ID
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
		}

		// Send to the cog
		cogReqBody = requests.CogQueueRequest{
			WebhookEventsFilter: []requests.CogEventFilter{requests.CogEventFilterStart, requests.CogEventFilterStart},
			WebhookUrl:          fmt.Sprintf("%s/v1/worker/webhook", utils.GetEnv("PUBLIC_API_URL", "")),
			Input: requests.BaseCogRequest{
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

		err = c.Redis.EnqueueCogRequest(r.Context(), cogReqBody)
		if err != nil {
			log.Error("Failed to write request to queue", "id", requestId, "err", err)
			responses.ErrInternalServerError(w, r, "Failed to queue upscale request")
			return err
		}

		c.QueueThrottler.IncrementBy(1, user.ID.String())

		return nil
	}); err != nil {
		log.Error("Error with transaction", "err", err)
		return
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
		err = c.Redis.Client.Publish(c.Redis.Ctx, shared.REDIS_SSE_BROADCAST_CHANNEL, respBytes).Err()
		if err != nil {
			log.Error("Failed to publish live page update", "err", err)
		}
	}()

	// Set timeout key
	err = c.Redis.SetCogRequestStreamID(c.Redis.Ctx, requestId, upscaleReq.StreamID)
	if err != nil {
		// Don't time it out if this fails
		log.Error("Failed to set timeout key", "err", err)
	} else {
		// Start the timeout timer
		go func() {
			// sleep
			time.Sleep(shared.REQUEST_COG_TIMEOUT)
			// this will trigger timeout if it hasnt been finished
			c.Repo.FailCogMessageDueToTimeoutIfTimedOut(requests.CogWebhookMessage{
				Input:  cogReqBody.Input,
				Error:  shared.TIMEOUT_ERROR,
				Status: requests.CogFailed,
			})
		}()
	}

	go c.Track.UpscaleStarted(user, cogReqBody.Input, utils.GetIPAddress(r))

	render.Status(r, http.StatusOK)
	render.JSON(w, r, &responses.TaskQueuedResponse{
		ID:               requestId,
		UIId:             upscaleReq.UIId,
		RemainingCredits: remainingCredits,
	})
}
