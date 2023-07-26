package rest

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/render"
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/enttypes"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/server/responses"
	"github.com/stablecog/sc-go/shared"
	"github.com/stablecog/sc-go/utils"
)

// POST generate expand (zooom-out)
// Adds generate to queue, if authenticated, returns the ID of the generation
func (c *RestAPI) HandleCreateGenerationZoomOutWebUI(w http.ResponseWriter, r *http.Request) {
	var user *ent.User
	if user = c.GetUserIfAuthenticated(w, r); user == nil {
		return
	}

	// Parse request body
	reqBody, _ := io.ReadAll(r.Body)
	var generateReq requests.CreateGenerationRequest
	err := json.Unmarshal(reqBody, &generateReq)
	if err != nil {
		responses.ErrUnableToParseJson(w, r)
		return
	}

	if generateReq.OutputID == nil {
		responses.ErrBadRequest(w, r, "output_id_required", "")
		return
	}

	if user.BannedAt != nil {
		remainingCredits, _ := c.Repo.GetNonExpiredCreditTotalForUser(user.ID, nil)
		render.Status(r, http.StatusOK)
		render.JSON(w, r, &responses.TaskQueuedResponse{
			ID:               uuid.NewString(),
			UIId:             generateReq.UIId,
			RemainingCredits: remainingCredits,
		})
		return
	}

	// Get output
	output, err := c.Repo.GetGenerationOutputForUser(*generateReq.OutputID, user.ID)
	if err != nil {
		if ent.IsNotFound(err) {
			responses.ErrNotFound(w, r, "output_not_found")
			return
		}
		log.Error("Error getting generation output", "err", err)
		responses.ErrInternalServerError(w, r, "An unknown error has occured")
		return
	}

	// Get generation
	generation, err := c.Repo.GetGenerationWithPrompts(output.GenerationID)
	if err != nil {
		log.Error("Error getting generation with output", "err", err)
		responses.ErrInternalServerError(w, r, "An unknown error has occured")
		return
	}

	// Create a new generation request
	expandGenerateReq := requests.CreateGenerationRequest{
		Prompt:                generation.Edges.Prompt.Text,
		Width:                 &generation.Width,
		Height:                &generation.Height,
		InferenceSteps:        &generation.InferenceSteps,
		GuidanceScale:         &generation.GuidanceScale,
		ModelId:               &generation.ModelID,
		SchedulerId:           &generation.SchedulerID,
		NumOutputs:            utils.ToPtr[int32](1),
		StreamID:              generateReq.StreamID,
		UIId:                  generateReq.UIId,
		PromptStrength:        generation.PromptStrength,
		ZoomedOutFromOutputID: utils.ToPtr(output.ID),
		ZoomOutScale:          utils.ToPtr(float32(shared.DEFAULT_GENERATE_ZOOM_OUT_SCALE)),
	}

	if generation.Edges.NegativePrompt != nil {
		expandGenerateReq.NegativePrompt = generation.Edges.NegativePrompt.Text
	}

	// Get bg/mask url
	bgUrlStr, maskUrlStr, wErr := c.SCWorker.GetExpandImageUrlsFromOutput(user.ID, output, shared.DEFAULT_GENERATE_ZOOM_OUT_SCALE)
	if wErr != nil {
		errResp := responses.ApiFailedResponse{
			Error: wErr.Err.Error(),
		}
		render.Status(r, wErr.StatusCode)
		render.JSON(w, r, errResp)
		return
	}

	expandGenerateReq.InitImageUrl = bgUrlStr
	expandGenerateReq.MaskImageUrl = maskUrlStr

	newGeneration, initSettings, workerErr := c.SCWorker.CreateGeneration(
		enttypes.SourceTypeWebUI,
		r,
		user,
		nil,
		expandGenerateReq,
	)

	if workerErr != nil {
		errResp := responses.ApiFailedResponse{
			Error: workerErr.Err.Error(),
		}
		if initSettings != nil {
			errResp.Settings = initSettings
		}
		render.Status(r, workerErr.StatusCode)
		render.JSON(w, r, errResp)
		return
	}

	// Return response
	render.Status(r, http.StatusOK)
	render.JSON(w, r, newGeneration.QueuedResponse)
}

// POST generate expand (zooom-out)
// Adds generate to queue, if authenticated, returns the ID of the generation
func (c *RestAPI) HandleCreateGenerationZoomOutAPI(w http.ResponseWriter, r *http.Request) {
	var user *ent.User
	if user = c.GetUserIfAuthenticated(w, r); user == nil {
		return
	}
	var apiToken *ent.ApiToken
	if apiToken = c.GetApiToken(w, r); apiToken == nil {
		return
	}

	// Parse request body
	reqBody, _ := io.ReadAll(r.Body)
	var generateReq requests.CreateGenerationRequest
	err := json.Unmarshal(reqBody, &generateReq)
	if err != nil {
		responses.ErrUnableToParseJson(w, r)
		return
	}

	if generateReq.OutputID == nil {
		responses.ErrBadRequest(w, r, "output_id_required", "")
		return
	}

	if user.BannedAt != nil {
		remainingCredits, _ := c.Repo.GetNonExpiredCreditTotalForUser(user.ID, nil)
		render.Status(r, http.StatusOK)
		render.JSON(w, r, &responses.TaskQueuedResponse{
			ID:               uuid.NewString(),
			UIId:             generateReq.UIId,
			RemainingCredits: remainingCredits,
		})
		return
	}

	// Get output
	output, err := c.Repo.GetGenerationOutputForUser(*generateReq.OutputID, user.ID)
	if err != nil {
		if ent.IsNotFound(err) {
			responses.ErrNotFound(w, r, "output_not_found")
			return
		}
		log.Error("Error getting generation output", "err", err)
		responses.ErrInternalServerError(w, r, "An unknown error has occured")
		return
	}

	// Get generation
	generation, err := c.Repo.GetGenerationWithPrompts(output.GenerationID)
	if err != nil {
		log.Error("Error getting generation with output", "err", err)
		responses.ErrInternalServerError(w, r, "An unknown error has occured")
		return
	}

	// Create a new generation request
	expandGenerateReq := requests.CreateGenerationRequest{
		Prompt:                generation.Edges.Prompt.Text,
		Width:                 &generation.Width,
		Height:                &generation.Height,
		InferenceSteps:        &generation.InferenceSteps,
		GuidanceScale:         &generation.GuidanceScale,
		ModelId:               &generation.ModelID,
		SchedulerId:           &generation.SchedulerID,
		NumOutputs:            utils.ToPtr[int32](1),
		StreamID:              generateReq.StreamID,
		UIId:                  generateReq.UIId,
		PromptStrength:        generation.PromptStrength,
		ZoomedOutFromOutputID: utils.ToPtr(output.ID),
		ZoomOutScale:          utils.ToPtr(float32(shared.DEFAULT_GENERATE_ZOOM_OUT_SCALE)),
	}

	if generation.Edges.NegativePrompt != nil {
		expandGenerateReq.NegativePrompt = generation.Edges.NegativePrompt.Text
	}

	// Get bg/mask url
	bgUrlStr, maskUrlStr, wErr := c.SCWorker.GetExpandImageUrlsFromOutput(user.ID, output, shared.DEFAULT_GENERATE_ZOOM_OUT_SCALE)
	if wErr != nil {
		errResp := responses.ApiFailedResponse{
			Error: wErr.Err.Error(),
		}
		render.Status(r, wErr.StatusCode)
		render.JSON(w, r, errResp)
		return
	}

	expandGenerateReq.InitImageUrl = bgUrlStr
	expandGenerateReq.MaskImageUrl = maskUrlStr

	newGeneration, initSettings, workerErr := c.SCWorker.CreateGeneration(
		enttypes.SourceTypeAPI,
		r,
		user,
		nil,
		expandGenerateReq,
	)

	if workerErr != nil {
		errResp := responses.ApiFailedResponse{
			Error: workerErr.Err.Error(),
		}
		if initSettings != nil {
			errResp.Settings = initSettings
		}
		render.Status(r, workerErr.StatusCode)
		render.JSON(w, r, errResp)
		return
	}

	// Return response
	render.Status(r, http.StatusOK)
	render.JSON(w, r, newGeneration)
}
