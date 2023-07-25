package rest

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/render"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/enttypes"
	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/server/responses"
	"github.com/stablecog/sc-go/server/scworker"
)

func (c *RestAPI) HandleCreateVoiceoverToken(w http.ResponseWriter, r *http.Request) {
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
	var voiceoverReq requests.CreateVoiceoverRequest
	err := json.Unmarshal(reqBody, &voiceoverReq)
	if err != nil {
		responses.ErrUnableToParseJson(w, r)
		return
	}

	// Create voiceover

	voiceover, initSettings, workerErr := scworker.CreateVoiceover(
		enttypes.SourceTypeAPI,
		r,
		c.Repo,
		c.Redis,
		c.SMap,
		c.QueueThrottler,
		user,
		c.Track,
		&apiToken.ID,
		voiceoverReq,
	)

	if err != nil {
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
	render.JSON(w, r, voiceover)
}
