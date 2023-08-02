package rest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/render"
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/enttypes"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/server/responses"
	"github.com/stablecog/sc-go/utils"
)

func (c *RestAPI) HandleVoiceover(w http.ResponseWriter, r *http.Request) {
	var user *ent.User
	if user = c.GetUserIfAuthenticated(w, r); user == nil {
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

	if user.BannedAt != nil {
		remainingCredits, _ := c.Repo.GetNonExpiredCreditTotalForUser(user.ID, nil)
		render.Status(r, http.StatusOK)
		render.JSON(w, r, &responses.TaskQueuedResponse{
			ID:               uuid.NewString(),
			UIId:             voiceoverReq.UIId,
			RemainingCredits: remainingCredits,
		})
		return
	}

	voiceover, initSettings, workerErr := c.SCWorker.CreateVoiceover(
		enttypes.SourceTypeWebUI,
		r,
		user,
		nil,
		voiceoverReq,
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
	render.JSON(w, r, voiceover.QueuedResponse)
}

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

	voiceover, initSettings, workerErr := c.SCWorker.CreateVoiceover(
		enttypes.SourceTypeAPI,
		r,
		user,
		&apiToken.ID,
		voiceoverReq,
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
	render.JSON(w, r, voiceover)
}

// HTTP Get - voiceovers for user
// Takes query paramers for pagination
// per_page: number of generations to return
// cursor: cursor for pagination, it is an iso time string in UTC
func (c *RestAPI) HandleQueryVoiceovers(w http.ResponseWriter, r *http.Request) {
	var user *ent.User
	if user = c.GetUserIfAuthenticated(w, r); user == nil {
		return
	}

	// Validate query parameters
	perPage := DEFAULT_PER_PAGE
	var err error
	if perPageStr := r.URL.Query().Get("per_page"); perPageStr != "" {
		perPage, err = strconv.Atoi(perPageStr)
		if err != nil {
			responses.ErrBadRequest(w, r, "per_page must be an integer", "")
			return
		} else if perPage < 1 || perPage > MAX_PER_PAGE {
			responses.ErrBadRequest(w, r, fmt.Sprintf("per_page must be between 1 and %d", MAX_PER_PAGE), "")
			return
		}
	}

	filters := &requests.QueryVoiceoverFilters{}
	err = filters.ParseURLQueryParameters(r.URL.Query())
	if err != nil {
		responses.ErrBadRequest(w, r, err.Error(), "")
		return
	}

	// query postgres
	var cursor *time.Time
	if cursorStr := r.URL.Query().Get("cursor"); cursorStr != "" {
		cursorTime, err := utils.ParseIsoTime(cursorStr)
		if err != nil {
			responses.ErrBadRequest(w, r, "cursor must be a valid iso time string", "")
			return
		}
		cursor = &cursorTime
	}

	// Ensure user ID is set to only include this users generations
	filters.UserID = &user.ID

	// Get generaions
	voiceovers, err := c.Repo.QueryVoiceovers(perPage, cursor, filters)
	if err != nil {
		log.Error("Error getting voiceovers for user", "err", err)
		responses.ErrInternalServerError(w, r, "Error getting voiceovers")
		return
	}

	// Return voiceovers
	render.Status(r, http.StatusOK)
	render.JSON(w, r, voiceovers)
}
