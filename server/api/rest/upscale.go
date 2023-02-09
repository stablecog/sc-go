package rest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/render"
	"github.com/google/uuid"
	"github.com/stablecog/go-apps/database/ent"
	"github.com/stablecog/go-apps/server/requests"
	"github.com/stablecog/go-apps/server/responses"
	"github.com/stablecog/go-apps/shared"
	"github.com/stablecog/go-apps/utils"
	"k8s.io/klog/v2"
)

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
	if !utils.IsSha256Hash(upscaleReq.StreamID) || c.Hub.GetClientByUid(upscaleReq.StreamID) == nil {
		responses.ErrBadRequest(w, r, "Invalid stream ID")
		return
	}

	if upscaleReq.Type != requests.UpscaleRequestTypeImage && upscaleReq.Type != requests.UpscaleRequestTypeOutput {
		responses.ErrBadRequest(w, r, fmt.Sprintf("Invalid upscale type, should be %s or %s", requests.UpscaleRequestTypeImage, requests.UpscaleRequestTypeOutput))
		return
	}

	var outputID uuid.UUID
	if upscaleReq.Type == requests.UpscaleRequestTypeImage && !utils.IsValidHTTPURL(upscaleReq.Input) {
		responses.ErrBadRequest(w, r, "Invalid image URL")
		return
	} else if upscaleReq.Type == requests.UpscaleRequestTypeOutput {
		outputID, err = uuid.Parse(upscaleReq.Input)
		if err != nil {
			responses.ErrBadRequest(w, r, "Invalid output ID")
			return
		}
	}

	if !shared.GetCache().IsValidUpscaleModelID(upscaleReq.ModelId) {
		klog.Infof("Invalid model ID: %s", upscaleReq.ModelId)
		responses.ErrBadRequest(w, r, "Invalid model ID")
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

	// Pro user check
	isProUser, err := c.Repo.IsProUser(*userID)
	if err != nil {
		klog.Errorf("Error checking if user is pro: %v", err)
		responses.ErrInternalServerError(w, r, "Error retrieving user")
		return
	}

	if !isProUser {
		responses.ErrBadRequest(w, r, "Upscale feature isn't available on the free plan :(")
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
			responses.ErrBadRequest(w, r, "Unable to retrieve width/height for upscale")
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
				responses.ErrBadRequest(w, r, "Output not found")
				return
			}
			klog.Errorf("Error getting output: %v", err)
			responses.ErrInternalServerError(w, r, "Error getting output")
			return
		}
		if output.UpscaledImageURL != nil {
			responses.ErrBadRequest(w, r, "Image already upscaled")
			return
		}
		imageUrl = output.ImageURL
		// Get width/height of generation
		width, height, err = c.Repo.GetGenerationOutputWidthHeight(outputID)
		if err != nil {
			responses.ErrBadRequest(w, r, "Unable to retrieve width/height for upscale")
			return
		}
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
		upscaleReq)
	if err != nil {
		klog.Errorf("Error creating upscale: %v", err)
		responses.ErrInternalServerError(w, r, "Error creating upscale")
		return
	}

	// Request ID matches upscale ID
	requestId := upscale.ID.String()

	// Send to the cog
	cogReqBody := requests.CogUpscaleQueueRequest{
		BaseCogRequestQueue: requests.BaseCogRequestQueue{
			WebhookEventsFilter: []requests.WebhookEventFilterOption{requests.WebhookEventFilterStart, requests.WebhookEventFilterStart},
			RedisPubsubKey:      shared.COG_REDIS_UPSCALE_EVENT_CHANNEL,
		},
		Input: requests.BaseCogUpscaleRequest{
			ID:                 requestId,
			GenerationOutputID: outputIDStr,
			Image:              imageUrl,
			ProcessType:        string(shared.UPSCALE),
		},
	}

	err = c.Redis.EnqueueCogRequest(r.Context(), cogReqBody)
	if err != nil {
		klog.Errorf("Failed to write request %s to queue: %v", requestId, err)
		responses.ErrInternalServerError(w, r, "Failed to queue upscale request")
		return
	}

	// Track the request in our internal map
	c.CogRequestSSEConnMap.Put(requestId, upscaleReq.StreamID)

	render.Status(r, http.StatusOK)
	render.JSON(w, r, &responses.QueuedResponse{
		ID: requestId,
	})
}
