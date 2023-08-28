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
)

// POST generate endpoint
// Adds generate to queue, if authenticated, returns the ID of the generation
func (c *RestAPI) HandleCreateGeneration(w http.ResponseWriter, r *http.Request) {
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

	generation, initSettings, workerErr := c.SCWorker.CreateGeneration(
		enttypes.SourceTypeWebUI,
		r,
		user,
		nil,
		c.Clip,
		generateReq,
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
	render.JSON(w, r, generation.QueuedResponse)
}

// POST generate endpoint
// Handles creating a generation with API token
func (c *RestAPI) HandleCreateGenerationToken(w http.ResponseWriter, r *http.Request) {
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

	// Create generation
	generation, initSettings, workerErr := c.SCWorker.CreateGeneration(
		enttypes.SourceTypeAPI,
		r,
		user,
		&apiToken.ID,
		c.Clip,
		generateReq,
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

	// Update last seen at
	err = c.Repo.UpdateLastSeenAt(user.ID)
	if err != nil {
		log.Warn("Error updating last seen at", "err", err, "user", user.ID.String())
	}

	// Return response
	render.Status(r, http.StatusOK)
	render.JSON(w, r, generation)
}
