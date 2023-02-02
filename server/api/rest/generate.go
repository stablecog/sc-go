package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"time"

	"github.com/go-chi/render"
	"github.com/google/uuid"
	"github.com/stablecog/go-apps/database/ent"
	"github.com/stablecog/go-apps/database/repository"
	"github.com/stablecog/go-apps/server/requests"
	"github.com/stablecog/go-apps/server/responses"
	"github.com/stablecog/go-apps/shared"
	"github.com/stablecog/go-apps/utils"
	"k8s.io/klog/v2"
)

// POST generate endpoint
// Adds generate to queue, if authenticated, returns the ID of the generation
func (c *RestAPI) HandleCreateGeneration(w http.ResponseWriter, r *http.Request) {
	userID := c.GetUserIDIfAuthenticated(w, r)
	if userID == nil {
		return
	}

	// Parse request body
	reqBody, _ := io.ReadAll(r.Body)
	var generateReq requests.GenerateRequestBody
	err := json.Unmarshal(reqBody, &generateReq)
	if err != nil {
		responses.ErrUnableToParseJson(w, r)
		return
	}

	// Make sure the websocket ID is valid
	if !utils.IsSha256Hash(generateReq.WebsocketId) || c.Hub.GetClientByUid(generateReq.WebsocketId) == nil {
		responses.ErrBadRequest(w, r, "Invalid websocket ID")
		return
	}

	// Validate request body
	if generateReq.Height > shared.MAX_GENERATE_HEIGHT {
		responses.ErrBadRequest(w, r, fmt.Sprintf("Height is too large, max is: %d", shared.MAX_GENERATE_HEIGHT))
		return
	}

	if generateReq.Width > shared.MAX_GENERATE_WIDTH {
		responses.ErrBadRequest(w, r, fmt.Sprintf("Width is too large, max is: %d", shared.MAX_GENERATE_WIDTH))
		return
	}

	if generateReq.Width*generateReq.Height*generateReq.NumInferenceSteps >= shared.MAX_PRO_PIXEL_STEPS {
		klog.Infof(
			"Pick fewer inference steps or smaller dimensions: %d - %d - %d",
			generateReq.Width,
			generateReq.Height,
			generateReq.NumInferenceSteps,
		)
		responses.ErrBadRequest(w, r, "Pick fewer inference steps or smaller dimensions")
		return
	}

	// Validate model and scheduler IDs in request are valid
	if !shared.GetCache().IsValidGenerationModelID(generateReq.ModelId) {
		klog.Infof("Invalid model ID: %s", generateReq.ModelId)
		responses.ErrBadRequest(w, r, "Invalid model ID")
		return
	}

	if !shared.GetCache().IsValidShedulerID(generateReq.SchedulerId) {
		klog.Infof("Invalid scheduler ID: %s", generateReq.SchedulerId)
		responses.ErrBadRequest(w, r, "Invalid scheduler ID")
		return
	}

	// Generate seed if not provided
	if generateReq.Seed < 0 {
		rand.Seed(time.Now().Unix())
		generateReq.Seed = rand.Intn(math.MaxInt32)
	}

	// Parse request headers
	countryCode := utils.GetCountryCode(r)
	deviceInfo := utils.GetClientDeviceInfo(r)

	isProUser, err := c.Repo.IsProUser(*userID)
	if err != nil {
		klog.Errorf("Error checking if user is pro: %v", err)
		responses.ErrInternalServerError(w, r, "Error retrieving user")
		return
	}

	// If not pro user, they are restricted from some features
	if !isProUser {
		if !shared.GetCache().IsGenerationModelAvailableForFree(generateReq.ModelId) {
			responses.ErrBadRequest(w, r, "That model is not available on the free plan :(")
			return
		}
		if !shared.GetCache().IsSchedulerAvailableForFree(generateReq.SchedulerId) {
			responses.ErrBadRequest(w, r, "That scheduler is not available on the free plan :(")
			return
		}
		if !shared.GetCache().IsHeightAvailableForFree(generateReq.Height) {
			responses.ErrBadRequest(w, r, "That generation height is not available on the free plan :(")
			return
		}
		if !shared.GetCache().IsWidthAvailableForFree(generateReq.Width) {
			responses.ErrBadRequest(w, r, "That generation width is not available on the free plan :(")
			return
		}
		if !shared.GetCache().IsNumInterferenceStepsAvailableForFree(generateReq.NumInferenceSteps) {
			responses.ErrBadRequest(w, r, "That number of inference steps is not available on the free plan :(")
			return
		}
	}

	// ! TODO - rate limit free

	// ! TODO - parallel generation toggle

	// Get model and scheduler name for cog
	modelName := shared.GetCache().GetGenerationModelNameFromID(generateReq.ModelId)
	schedulerName := shared.GetCache().GetSchedulerNameFromID(generateReq.SchedulerId)
	if modelName == "" || schedulerName == "" {
		klog.Errorf("Error getting model or scheduler name: %s - %s", modelName, schedulerName)
		responses.ErrInternalServerError(w, r, "An unknown error has occured")
		return
	}

	// Format prompts
	generateReq.Prompt = utils.FormatPrompt(generateReq.Prompt)
	generateReq.NegativePrompt = utils.FormatPrompt(generateReq.NegativePrompt)

	// Create generation
	_, err = c.Repo.CreateGeneration(
		*userID,
		string(deviceInfo.DeviceType),
		deviceInfo.DeviceOs,
		deviceInfo.DeviceBrowser,
		countryCode,
		generateReq)
	if err != nil {
		klog.Errorf("Error creating generation: %v", err)
		responses.ErrInternalServerError(w, r, "Error creating generation")
		return
	}

	// Get language codes
	promptFlores, negativePromptFlores := utils.GetPromptFloresCodes(generateReq.Prompt, generateReq.NegativePrompt)
	// Generate a unique request ID for the cog
	requestId := uuid.NewString()

	cogReqBody := requests.CogGenerateQueueRequest{
		BaseCogRequestQueue: requests.BaseCogRequestQueue{
			WebhookEventsFilter: []requests.WebhookEventFilterOption{requests.WebhookEventFilterStart, requests.WebhookEventFilterStart},
			Webhook:             fmt.Sprintf("%s/v1/queue/webhook/%s", utils.GetEnv("PUBLIC_API_URL", "https://api.stablecog.com"), utils.GetEnv("QUEUE_SECRET", "")),
		},
		BaseCogGenerateRequest: requests.BaseCogGenerateRequest{
			ID:                   requestId,
			Prompt:               generateReq.Prompt,
			NegativePrompt:       generateReq.NegativePrompt,
			PromptFlores:         promptFlores,
			NegativePromptFlores: negativePromptFlores,
			Width:                fmt.Sprint(generateReq.Width),
			Height:               fmt.Sprint(generateReq.Height),
			NumInferenceSteps:    fmt.Sprint(generateReq.NumInferenceSteps),
			GuidanceScale:        fmt.Sprint(generateReq.GuidanceScale),
			Model:                modelName,
			Scheduler:            schedulerName,
			Seed:                 fmt.Sprint(generateReq.Seed),
			OutputImageExt:       string(shared.DEFAULT_GENERATE_OUTPUT_IMAGE_EXTENSION),
		},
	}

	err = c.Redis.EnqueueCogGenerateRequest(r.Context(), cogReqBody)
	if err != nil {
		klog.Errorf("Failed to write request %s to queue: %v", requestId, err)
		responses.ErrInternalServerError(w, r, "Failed to queue generate request")
		return
	}

	// Track the request in our internal map
	c.CogRequestWebsocketConnMap.Put(requestId, generateReq.WebsocketId)

	render.Status(r, http.StatusOK)
	render.JSON(w, r, &responses.GenerateResponse{
		ID: requestId,
	})
}

// HTTP POST submit a generation to gallery
func (c *RestAPI) HandleSubmitGenerationToGallery(w http.ResponseWriter, r *http.Request) {
	userID := c.GetUserIDIfAuthenticated(w, r)
	if userID == nil {
		return
	}

	// Parse request body
	reqBody, _ := io.ReadAll(r.Body)
	var submitToGalleryReq requests.GenerateSubmitToGalleryRequestBody
	err := json.Unmarshal(reqBody, &submitToGalleryReq)
	if err != nil {
		responses.ErrUnableToParseJson(w, r)
		return
	}

	// See if generation already exists
	err = c.Repo.SubmitGenerationToGalleryForUser(submitToGalleryReq.GenerationID, *userID)
	if err != nil {
		if ent.IsNotFound(err) {
			responses.ErrBadRequest(w, r, "Generation not found")
			return
		}
		if errors.Is(err, repository.ErrAlreadySubmitted) {
			responses.ErrBadRequest(w, r, "Generation already submitted to gallery")
			return
		}
		klog.Errorf("Error submitting generation to gallery: %v", err)
		responses.ErrInternalServerError(w, r, "Error submitting generation to gallery")
		return
	}

	// Empty response body for successful
	render.Status(r, http.StatusOK)
}
