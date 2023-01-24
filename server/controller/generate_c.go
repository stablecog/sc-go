package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"time"

	"github.com/go-chi/render"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/stablecog/go-apps/models"
	"github.com/stablecog/go-apps/models/constants"
	"github.com/stablecog/go-apps/utils"
	"k8s.io/klog/v2"
)

// POST generate endpoint
// Adds generate to queue, if authenticated, returns the ID of the generation
func (c *HttpController) PostGenerate(w http.ResponseWriter, r *http.Request) {
	// See if authenticated
	userIDStr, authenticated := r.Context().Value("user_id").(string)
	// This should always be true because of the auth middleware, but check it anyway
	if !authenticated || userIDStr == "" {
		models.ErrUnauthorized(w, r)
		return
	}
	// Ensure valid uuid
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		models.ErrUnauthorized(w, r)
		return
	}

	// Parse request body
	reqBody, _ := io.ReadAll(r.Body)
	var generateReq models.GenerateRequestBody
	err = json.Unmarshal(reqBody, &generateReq)
	if err != nil {
		models.ErrUnableToParseJson(w, r)
		return
	}

	// Validate request body
	if generateReq.Height > constants.MaxGenerateHeight {
		models.ErrBadRequest(w, r, fmt.Sprintf("Height is too large, max is: %d", constants.MaxGenerateHeight))
		return
	}

	if generateReq.Width > constants.MaxGenerateWidth {
		models.ErrBadRequest(w, r, fmt.Sprintf("Width is too large, max is: %d", constants.MaxGenerateWidth))
		return
	}

	if generateReq.Width*generateReq.Height*generateReq.NumInferenceSteps >= constants.MaxProPixelSteps {
		klog.Infof(
			"Pick fewer inference steps or smaller dimensions: %d - %d - %d",
			generateReq.Width,
			generateReq.Height,
			generateReq.NumInferenceSteps,
		)
		models.ErrBadRequest(w, r, "Pick fewer inference steps or smaller dimensions")
		return
	}

	// Validate model and scheduler IDs in request are valid
	if !models.GetCache().IsValidGenerationModelID(generateReq.ModelId) {
		klog.Infof("Invalid model ID: %s", generateReq.ModelId)
		models.ErrBadRequest(w, r, "Invalid model ID")
		return
	}

	if !models.GetCache().IsValidShedulerID(generateReq.SchedulerId) {
		klog.Infof("Invalid scheduler ID: %s", generateReq.SchedulerId)
		models.ErrBadRequest(w, r, "Invalid scheduler ID")
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

	isProUser, err := c.Repo.IsProUser(userID)
	if err != nil {
		klog.Errorf("Error checking if user is pro: %v", err)
		models.ErrInternalServerError(w, r, "Error retrieving user")
		return
	}

	// If not pro user, they are restricted from some features
	if !isProUser {
		if !models.GetCache().IsGenerationModelAvailableForFree(generateReq.ModelId) {
			models.ErrBadRequest(w, r, "That model is not available on the free plan :(")
			return
		}
		if !models.GetCache().IsSchedulerAvailableForFree(generateReq.SchedulerId) {
			models.ErrBadRequest(w, r, "That scheduler is not available on the free plan :(")
			return
		}
		if !models.GetCache().IsHeightAvailableForFree(generateReq.Height) {
			models.ErrBadRequest(w, r, "That generation height is not available on the free plan :(")
			return
		}
		if !models.GetCache().IsWidthAvailableForFree(generateReq.Width) {
			models.ErrBadRequest(w, r, "That generation width is not available on the free plan :(")
			return
		}
		if !models.GetCache().IsNumInterferenceStepsAvailableForFree(generateReq.NumInferenceSteps) {
			models.ErrBadRequest(w, r, "That number of inference steps is not available on the free plan :(")
			return
		}
	}

	// ! TODO - rate limit free

	// ! TODO - parallel generation toggle

	// Get model and scheduler name for cog
	modelName := models.GetCache().GetGenerationModelNameFromID(generateReq.ModelId)
	schedulerName := models.GetCache().GetSchedulerNameFromID(generateReq.SchedulerId)
	if modelName == "" || schedulerName == "" {
		klog.Errorf("Error getting model or scheduler name: %s - %s", modelName, schedulerName)
		models.ErrInternalServerError(w, r, "An unknown error has occured")
		return
	}

	// Format prompts
	generateReq.Prompt = utils.FormatPrompt(generateReq.Prompt)
	generateReq.NegativePrompt = utils.FormatPrompt(generateReq.NegativePrompt)

	// Create generation
	_, err = c.Repo.CreateGeneration(
		userID,
		string(deviceInfo.DeviceType),
		deviceInfo.DeviceOs,
		deviceInfo.DeviceBrowser,
		countryCode,
		generateReq)
	if err != nil {
		klog.Errorf("Error creating generation: %v", err)
		models.ErrInternalServerError(w, r, "Error creating generation")
		return
	}

	// Get language codes
	promptFlores, negativePromptFlores := utils.GetPromptFloresCodes(generateReq.Prompt, generateReq.NegativePrompt)
	// Generate a unique request ID for the cog
	requestId := uuid.NewString()

	cogReqBody := models.CogGenerateQueueRequest{
		BaseCogRequestQueue: models.BaseCogRequestQueue{
			WebhookEventsFilter: []models.WebhookEventFilterOption{models.WebhookEventStart, models.WebhookEventCompleted},
			// ! TODO
			Webhook: "TODO",
		},
		BaseCogGenerateRequest: models.BaseCogGenerateRequest{
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
			OutputImageExt:       string(models.DefaultOutputImageExtension),
		},
	}

	_, err = c.Redis.XAdd(r.Context(), &redis.XAddArgs{
		Stream: "input_queue",
		ID:     "*", // Auto generate ID
		Values: []interface{}{"value", cogReqBody},
	}).Result()
	if err != nil {
		klog.Errorf("Failed to write request %s to queue: %v", requestId, err)
		models.ErrInternalServerError(w, r, "Failed to queue generate request")
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, map[string]string{
		"id": requestId,
	})
}
