package rest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/render"
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/server/responses"
	"github.com/stablecog/sc-go/shared"
	"github.com/stablecog/sc-go/utils"
	"k8s.io/klog/v2"
)

// ! TODO - we need some type of timeout functionality
// ! If we don't get a response from cog within a certain amount of time, we should update generation as failed
// ! and refund user credits

func (c *RestAPI) HandleUpscale(w http.ResponseWriter, r *http.Request) {
	userID := c.GetUserIDIfAuthenticated(w, r)
	if userID == nil {
		return
	}

	// Parse request body
	reqBody, _ := io.ReadAll(r.Body)
	var upscaleReq requests.UpscaleRequestBody
	err := json.Unmarshal(reqBody, &upscaleReq)
	if err != nil {
		responses.ErrUnableToParseJson(w, r)
		return
	}

	// Validation
	if !utils.IsSha256Hash(upscaleReq.StreamID) {
		responses.ErrInvalidStreamID(w, r)
		return
	}

	if upscaleReq.Type != requests.UpscaleRequestTypeImage && upscaleReq.Type != requests.UpscaleRequestTypeOutput {
		responses.ErrBadRequest(w, r, fmt.Sprintf("Invalid upscale type, should be %s or %s", requests.UpscaleRequestTypeImage, requests.UpscaleRequestTypeOutput))
		return
	}

	var outputID uuid.UUID
	if upscaleReq.Type == requests.UpscaleRequestTypeImage && !utils.IsValidHTTPURL(upscaleReq.Input) {
		responses.ErrBadRequest(w, r, "invalid_image_url")
		return
	} else if upscaleReq.Type == requests.UpscaleRequestTypeOutput {
		outputID, err = uuid.Parse(upscaleReq.Input)
		if err != nil {
			responses.ErrBadRequest(w, r, "invalid_output_id")
			return
		}
	}

	if !shared.GetCache().IsValidUpscaleModelID(upscaleReq.ModelId) {
		klog.Infof("invalid_model_id: %s", upscaleReq.ModelId)
		responses.ErrBadRequest(w, r, "invalid_model_id")
		return
	}

	// Parse request headers
	countryCode := utils.GetCountryCode(r)
	deviceInfo := utils.GetClientDeviceInfo(r)

	// Get model name for cog
	modelName := shared.GetCache().GetUpscaleModelNameFromID(upscaleReq.ModelId)
	if modelName == "" {
		klog.Errorf("Error getting model name: %s", modelName)
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
			responses.ErrBadRequest(w, r, "image_url_width_height_error")
			return
		}
	}

	// Output Type
	var outputIDStr string
	if upscaleReq.Type == requests.UpscaleRequestTypeOutput {
		outputIDStr = outputID.String()
		output, err := c.Repo.GetGenerationOutputForUser(outputID, *userID)
		if err != nil {
			if ent.IsNotFound(err) {
				responses.ErrBadRequest(w, r, "output_not_found")
				return
			}
			klog.Errorf("Error getting output: %v", err)
			responses.ErrInternalServerError(w, r, "Error getting output")
			return
		}
		if output.UpscaledImagePath != nil {
			responses.ErrBadRequest(w, r, "image_already_upscaled")
			return
		}
		imageUrl = output.ImagePath
		// Get width/height of generation
		width, height, err = c.Repo.GetGenerationOutputWidthHeight(outputID)
		if err != nil {
			responses.ErrBadRequest(w, r, "Unable to retrieve width/height for upscale")
			return
		}
	}

	// For live page update
	var livePageMsg responses.LivePageMessage
	// For keeping track of this request as it gets sent to the worker
	var requestId string

	// Wrap everything in a DB transaction
	// We do this since we want our credit deduction to be atomic with the whole process
	if err := c.Repo.WithTx(func(tx *ent.Tx) error {
		// Bind transaction to client
		DB := tx.Client()

		// Charge credits
		deducted, err := c.Repo.DeductCreditsFromUser(*userID, 1, DB)
		if err != nil {
			klog.Errorf("Error deducting credits: %v", err)
			responses.ErrInternalServerError(w, r, "Error deducting credits from user")
			return err
		} else if !deducted {
			responses.ErrInsufficientCredits(w, r)
			return responses.InsufficientCreditsErr
		}

		// Create upscale
		upscale, err := c.Repo.CreateUpscale(
			*userID,
			width,
			height,
			string(deviceInfo.DeviceType),
			deviceInfo.DeviceOs,
			deviceInfo.DeviceBrowser,
			countryCode,
			upscaleReq,
			DB)
		if err != nil {
			klog.Errorf("Error creating upscale: %v", err)
			responses.ErrInternalServerError(w, r, "Error creating upscale")
			return err
		}

		// Request ID matches upscale ID
		requestId = upscale.ID.String()

		// For live page update
		livePageMsg = responses.LivePageMessage{
			Type:        responses.LivePageMessageUpscale,
			ID:          utils.Sha256(requestId),
			CountryCode: countryCode,
			Status:      responses.LivePageQueued,
			Width:       width,
			Height:      height,
			CreatedAt:   upscale.CreatedAt,
		}

		// Send to the cog
		cogReqBody := requests.CogQueueRequest{
			WebhookEventsFilter: []requests.WebhookEventFilterOption{requests.WebhookEventFilterStart, requests.WebhookEventFilterStart},
			RedisPubsubKey:      shared.COG_REDIS_EVENT_CHANNEL,
			Input: requests.BaseCogRequest{
				ID:                 requestId,
				LivePageData:       livePageMsg,
				GenerationOutputID: outputIDStr,
				Image:              imageUrl,
				ProcessType:        shared.UPSCALE,
			},
		}

		err = c.Redis.EnqueueCogRequest(r.Context(), cogReqBody)
		if err != nil {
			klog.Errorf("Failed to write request %s to queue: %v", requestId, err)
			responses.ErrInternalServerError(w, r, "Failed to queue upscale request")
			return err
		}

		return nil
	}); err != nil {
		klog.Errorf("Error with transaction: %v", err)
		return
	}

	// Track the request in our internal map
	c.Redis.SetCogRequestStreamID(r.Context(), requestId, upscaleReq.StreamID)

	// Deal with live page update
	go c.Hub.BroadcastLivePageMessage(livePageMsg)

	render.Status(r, http.StatusOK)
	render.JSON(w, r, &responses.QueuedResponse{
		ID: requestId,
	})
}
