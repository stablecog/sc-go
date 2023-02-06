package rest

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"time"

	"github.com/go-chi/render"
	"github.com/stablecog/go-apps/server/requests"
	"github.com/stablecog/go-apps/server/responses"
	"github.com/stablecog/go-apps/shared"
	"github.com/stablecog/go-apps/utils"
	"k8s.io/klog/v2"
)

// POST generate endpoint
// Adds generate to queue, if authenticated, returns the ID of the generation
func (c *RestAPI) HandleCreateGeneration(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	userID := c.GetUserIDIfAuthenticated(w, r)
	if userID == nil {
		return
	}
	fmt.Printf("--- GetUserIDIFAuthenticated took: %s\n", time.Now().Sub(start))

	// Parse request body
	start = time.Now()
	reqBody, _ := io.ReadAll(r.Body)
	var generateReq requests.GenerateRequestBody
	err := json.Unmarshal(reqBody, &generateReq)
	if err != nil {
		responses.ErrUnableToParseJson(w, r)
		return
	}
	fmt.Printf("--- ParseRequestBody took: %s\n", time.Now().Sub(start))

	// Make sure the websocket ID is valid
	start = time.Now()
	if !utils.IsSha256Hash(generateReq.WebsocketId) || c.Hub.GetClientByUid(generateReq.WebsocketId) == nil {
		responses.ErrBadRequest(w, r, "Invalid websocket ID")
		return
	}
	fmt.Printf("--- Validate websocket ID took: %s\n", time.Now().Sub(start))

	// Validate request body
	if generateReq.Height > shared.MAX_GENERATE_HEIGHT {
		responses.ErrBadRequest(w, r, fmt.Sprintf("Height is too large, max is: %d", shared.MAX_GENERATE_HEIGHT))
		return
	}

	if generateReq.Width > shared.MAX_GENERATE_WIDTH {
		responses.ErrBadRequest(w, r, fmt.Sprintf("Width is too large, max is: %d", shared.MAX_GENERATE_WIDTH))
		return
	}

	if generateReq.Width*generateReq.Height*generateReq.InferenceSteps >= shared.MAX_PRO_PIXEL_STEPS {
		klog.Infof(
			"Pick fewer inference steps or smaller dimensions: %d - %d - %d",
			generateReq.Width,
			generateReq.Height,
			generateReq.InferenceSteps,
		)
		responses.ErrBadRequest(w, r, "Pick fewer inference steps or smaller dimensions")
		return
	}

	if generateReq.NumOutputs > shared.MAX_GENERATE_NUM_OUTPUTS {
		klog.Infof("Number of outputs can't be more than %d", shared.MAX_GENERATE_NUM_OUTPUTS)
		responses.ErrBadRequest(w, r, fmt.Sprintf("Number of outputs can't be more than %d", shared.MAX_GENERATE_NUM_OUTPUTS))
		return
	}

	if generateReq.NumOutputs < shared.MIN_GENERATE_NUM_OUTPUTS {
		klog.Infof("Number of outputs can't be less than %d", shared.MIN_GENERATE_NUM_OUTPUTS)
		responses.ErrBadRequest(w, r, fmt.Sprintf("Number of outputs can't be less than %d", shared.MIN_GENERATE_NUM_OUTPUTS))
		return
	}

	// Validate model and scheduler IDs in request are valid
	start = time.Now()
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
	fmt.Printf("--- Checking model and scheduler IDs took: %s\n", time.Now().Sub(start))

	// Generate seed if not provided
	if generateReq.Seed < 0 {
		rand.Seed(time.Now().Unix())
		generateReq.Seed = rand.Intn(math.MaxInt32)
	}

	// Parse request headers
	start = time.Now()
	countryCode := utils.GetCountryCode(r)
	deviceInfo := utils.GetClientDeviceInfo(r)
	fmt.Printf("--- Parse request headers took: %s\n", time.Now().Sub(start))

	start = time.Now()
	isProUser, err := c.Repo.IsProUser(*userID)
	if err != nil {
		klog.Errorf("Error checking if user is pro: %v", err)
		responses.ErrInternalServerError(w, r, "Error retrieving user")
		return
	}
	fmt.Printf("--- Check if user is pro took: %s\n", time.Now().Sub(start))

	// If not pro user, they are restricted from some features
	start = time.Now()
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
		if !shared.GetCache().IsNumInterferenceStepsAvailableForFree(generateReq.InferenceSteps) {
			responses.ErrBadRequest(w, r, "That number of inference steps is not available on the free plan :(")
			return
		}
	}
	fmt.Printf("--- Check pro features: %s\n", time.Now().Sub(start))

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
	start = time.Now()
	generateReq.Prompt = utils.FormatPrompt(generateReq.Prompt)
	generateReq.NegativePrompt = utils.FormatPrompt(generateReq.NegativePrompt)
	fmt.Printf("--- Format prompts took: %s\n", time.Now().Sub(start))

	// Create generation
	start = time.Now()
	g, err := c.Repo.CreateGeneration(
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
	fmt.Printf("--- Create generation took: %s\n", time.Now().Sub(start))

	// Get language codes
	promptFlores, negativePromptFlores := utils.GetPromptFloresCodes(generateReq.Prompt, generateReq.NegativePrompt)
	// Generate a unique request ID for the cog
	requestId := g.ID.String()

	cogReqBody := requests.CogGenerateQueueRequest{
		BaseCogRequestQueue: requests.BaseCogRequestQueue{
			WebhookEventsFilter: []requests.WebhookEventFilterOption{requests.WebhookEventFilterStart, requests.WebhookEventFilterStart},
			RedisPubsubKey:      shared.COG_REDIS_WEBHOOK_QUEUE_CHANNEL,
		},
		Input: requests.BaseCogGenerateRequest{
			ID:                   requestId,
			Prompt:               generateReq.Prompt,
			NegativePrompt:       generateReq.NegativePrompt,
			PromptFlores:         promptFlores,
			NegativePromptFlores: negativePromptFlores,
			Width:                fmt.Sprint(generateReq.Width),
			Height:               fmt.Sprint(generateReq.Height),
			NumInferenceSteps:    fmt.Sprint(generateReq.InferenceSteps),
			GuidanceScale:        fmt.Sprint(generateReq.GuidanceScale),
			Model:                modelName,
			Scheduler:            schedulerName,
			Seed:                 fmt.Sprint(generateReq.Seed),
			NumOutputs:           fmt.Sprint(generateReq.NumOutputs),
			OutputImageExtension: string(shared.DEFAULT_GENERATE_OUTPUT_EXTENSION),
			ProcessType:          string(shared.DEFAULT_PROCESS_TYPE),
		},
	}

	start = time.Now()
	err = c.Redis.EnqueueCogGenerateRequest(r.Context(), cogReqBody)
	if err != nil {
		klog.Errorf("Failed to write request %s to queue: %v", requestId, err)
		responses.ErrInternalServerError(w, r, "Failed to queue generate request")
		return
	}
	fmt.Printf("--- Enqueue cog request took: %s\n", time.Now().Sub(start))

	// Track the request in our internal map
	start = time.Now()
	c.CogRequestWebsocketConnMap.Put(requestId, generateReq.WebsocketId)
	fmt.Printf("--- Put request in map took: %s\n", time.Now().Sub(start))

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

	submitted, err := c.Repo.SubmitGenerationOutputsToGalleryForUser(submitToGalleryReq.GenerationOutputIDs, *userID)
	if err != nil {
		responses.ErrInternalServerError(w, r, "Error submitting generation outputs to gallery")
		return
	}

	res := responses.GenerateSubmitToGalleryResponse{
		Submitted: submitted,
	}

	render.JSON(w, r, res)
	render.Status(r, http.StatusOK)
}
