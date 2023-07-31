package rest

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/render"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/enttypes"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/server/responses"
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

	// Parse request body
	reqBody, _ := io.ReadAll(r.Body)
	var upscaleReq requests.CreateUpscaleRequest
	err := json.Unmarshal(reqBody, &upscaleReq)
	if err != nil {
		responses.ErrUnableToParseJson(w, r)
		return
	}

	// Create upscale
	upscale, initSettings, workerErr := c.SCWorker.CreateUpscale(
		enttypes.SourceTypeAPI,
		r,
		user,
		&apiToken.ID,
		upscaleReq,
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

	err = c.Repo.UpdateLastSeenAt(user.ID)
	if err != nil {
		log.Warn("Error updating last seen at", "err", err, "user", user.ID.String())
	}

	// Return response
	render.Status(r, http.StatusOK)
	render.JSON(w, r, upscale)
}
