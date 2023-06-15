package rest

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/render"
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/repository"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/server/responses"
	"github.com/stablecog/sc-go/shared"
	"github.com/stablecog/sc-go/utils"
	"golang.org/x/exp/slices"
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

	free := user.ActiveProductID == nil
	if free {
		// Re-evaluate if they have paid credits
		count, err := c.Repo.HasPaidCredits(user.ID)
		if err != nil {
			log.Error("Error getting paid credit sum for users", "err", err)
			responses.ErrInternalServerError(w, r, "An unknown error has occurred")
			return
		}
		free = count <= 0
	}

	var qMax int
	roles, err := c.Repo.GetRoles(user.ID)
	if err != nil {
		log.Error("Error getting roles for user", "err", err)
		responses.ErrInternalServerError(w, r, "An unknown error has occurred")
		return
	}
	isSuperAdmin := slices.Contains(roles, "SUPER_ADMIN")
	if isSuperAdmin {
		qMax = math.MaxInt64
	} else {
		qMax = shared.MAX_QUEUED_ITEMS_VOICEOVER
	}

	for _, role := range roles {
		switch role {
		case "ULTIMATE":
			free = false
		case "PRO":
			free = false
		case "STARTER":
			free = false
		}
	}

	// Validation
	if !isSuperAdmin {
		err = voiceoverReq.Validate(false)
		if err != nil {
			responses.ErrBadRequest(w, r, err.Error(), "")
			return
		}
	}

	// Get queue count
	nq, err := c.QueueThrottler.NumQueued(fmt.Sprintf("v:%s", user.ID.String()))
	if err != nil {
		log.Warn("Error getting queue count for user", "err", err, "user_id", user.ID)
	}
	if err == nil && nq >= qMax {
		responses.ErrBadRequest(w, r, "queue_limit_reached", "")
		return
	}

	// Enforce submit to gallery
	if free {
		voiceoverReq.SubmitToGallery = true
	}

	// Parse request headers
	countryCode := utils.GetCountryCode(r)
	deviceInfo := utils.GetClientDeviceInfo(r)

	// Get model name for cog
	modelName := shared.GetCache().GetVoiceoverModelNameFromID(*voiceoverReq.ModelId)
	if modelName == "" {
		log.Error("Error getting model name", "model_name", modelName)
		responses.ErrInternalServerError(w, r, "An unknown error has occurred")
		return
	}

	// Get speaker name for cog
	speakerName := shared.GetCache().GetVoiceoverSpeakerNameFromID(*voiceoverReq.SpeakerId)
	if speakerName == "" {
		log.Error("Error getting speaker name", "speaker_name", speakerName)
		responses.ErrInternalServerError(w, r, "An unknown error has occurred")
		return
	}

	// For live page update
	var livePageMsg shared.LivePageMessage
	// For keeping track of this request as it gets sent to the worker
	var requestId string
	// The cog request body
	var cogReqBody requests.CogQueueRequest
	// The total remaining credits
	var remainingCredits int

	// Wrap everything in a DB transaction
	// We do this since we want our credit deduction to be atomic with the whole process
	if err := c.Repo.WithTx(func(tx *ent.Tx) error {
		// Bind transaction to client
		DB := tx.Client()

		// Charge credits
		creditAmount := utils.CalculateVoiceoverCredits(voiceoverReq.Prompt)
		deducted, err := c.Repo.DeductCreditsFromUser(user.ID, creditAmount, DB)
		if err != nil {
			log.Error("Error deducting credits", "err", err)
			responses.ErrInternalServerError(w, r, "Error deducting credits from user")
			return err
		} else if !deducted {
			responses.ErrInsufficientCredits(w, r)
			return responses.InsufficientCreditsErr
		}

		remainingCredits, err = c.Repo.GetNonExpiredCreditTotalForUser(user.ID, DB)
		if err != nil {
			log.Error("Error getting remaining credits", "err", err)
			responses.ErrInternalServerError(w, r, "An unknown error has occurred")
			return err
		}

		// Create voiceover
		voiceover, err := c.Repo.CreateVoiceover(
			user.ID,
			string(deviceInfo.DeviceType),
			deviceInfo.DeviceOs,
			deviceInfo.DeviceBrowser,
			countryCode,
			voiceoverReq,
			user.ActiveProductID,
			nil,
			DB)
		if err != nil {
			log.Error("Error creating voiceover", "err", err)
			responses.ErrInternalServerError(w, r, "Error creating voiceover")
			return err
		}

		// Request ID matches upscale ID
		requestId = voiceover.ID.String()

		// For live page update
		livePageMsg = shared.LivePageMessage{
			ProcessType:      shared.VOICEOVER,
			ID:               utils.Sha256(requestId),
			CountryCode:      countryCode,
			Status:           shared.LivePageQueued,
			TargetNumOutputs: 1,
			CreatedAt:        voiceover.CreatedAt,
			ProductID:        user.ActiveProductID,
			Source:           shared.OperationSourceTypeWebUI,
		}

		// Send to the cog
		cogReqBody = requests.CogQueueRequest{
			WebhookEventsFilter: []requests.CogEventFilter{requests.CogEventFilterStart, requests.CogEventFilterStart},
			WebhookUrl:          fmt.Sprintf("%s/v1/worker/webhook", utils.GetEnv("PUBLIC_API_URL", "")),
			Input: requests.BaseCogRequest{
				ID:            requestId,
				IP:            utils.GetIPAddress(r),
				UIId:          voiceoverReq.UIId,
				UserID:        &user.ID,
				DeviceInfo:    deviceInfo,
				StreamID:      voiceoverReq.StreamID,
				LivePageData:  &livePageMsg,
				ProcessType:   shared.VOICEOVER,
				Model:         modelName,
				Speaker:       speakerName,
				ModelId:       *voiceoverReq.ModelId,
				Prompt:        voiceoverReq.Prompt,
				Temp:          fmt.Sprint(*voiceoverReq.Temperature),
				Seed:          fmt.Sprint(*voiceoverReq.Seed),
				RemoveSilence: fmt.Sprint(*voiceoverReq.RemoveSilence),
				DenoiseAudio:  fmt.Sprint(*voiceoverReq.DenoiseAudio),
			},
		}

		err = c.Redis.EnqueueCogRequest(r.Context(), shared.COG_REDIS_VOICEOVER_QUEUE, cogReqBody)
		if err != nil {
			log.Error("Failed to write request to queue", "id", requestId, "err", err)
			responses.ErrInternalServerError(w, r, "Failed to queue upscale request")
			return err
		}

		c.QueueThrottler.IncrementBy(1, fmt.Sprintf("v:%s", user.ID.String()))

		return nil
	}); err != nil {
		log.Error("Error with transaction", "err", err)
		return
	}

	// Send live page update
	go func() {
		liveResp := repository.TaskStatusUpdateResponse{
			ForLivePage:     true,
			LivePageMessage: &livePageMsg,
		}
		respBytes, err := json.Marshal(liveResp)
		if err != nil {
			log.Error("Error marshalling sse live response", "err", err)
			return
		}
		err = c.Redis.Client.Publish(c.Redis.Ctx, shared.REDIS_SSE_BROADCAST_CHANNEL, respBytes).Err()
		if err != nil {
			log.Error("Failed to publish live page update", "err", err)
		}
	}()

	// Set timeout key
	err = c.Redis.SetCogRequestStreamID(c.Redis.Ctx, requestId, voiceoverReq.StreamID)
	if err != nil {
		// Don't time it out if this fails
		log.Error("Failed to set timeout key", "err", err)
	} else {
		// Start the timeout timer
		go func() {
			// sleep
			time.Sleep(shared.REQUEST_COG_TIMEOUT)
			// this will trigger timeout if it hasnt been finished
			c.Repo.FailCogMessageDueToTimeoutIfTimedOut(requests.CogWebhookMessage{
				Input:  cogReqBody.Input,
				Error:  shared.TIMEOUT_ERROR,
				Status: requests.CogFailed,
			})
		}()
	}

	go c.Track.UpscaleStarted(user, cogReqBody.Input, utils.GetIPAddress(r))

	render.Status(r, http.StatusOK)
	render.JSON(w, r, &responses.TaskQueuedResponse{
		ID:               requestId,
		UIId:             voiceoverReq.UIId,
		RemainingCredits: remainingCredits,
	})
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
