package rest

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"

	"github.com/go-chi/render"
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/ent/upscale"
	"github.com/stablecog/sc-go/database/repository"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/server/responses"
	"github.com/stablecog/sc-go/shared"
	"github.com/stablecog/sc-go/utils"
	"golang.org/x/exp/slices"
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

	if user.BannedAt != nil {
		remainingCredits, _ := c.Repo.GetNonExpiredCreditTotalForUser(user.ID, nil)
		render.Status(r, http.StatusOK)
		render.JSON(w, r, &responses.TaskQueuedResponse{
			ID:               uuid.NewString(),
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
		err = voiceoverReq.Validate(true)
		if err != nil {
			responses.ErrBadRequest(w, r, err.Error(), "")
			return
		}
	} else {
		voiceoverReq.ApplyDefaults()
	}

	// Set settings resp
	initSettings := responses.VoiceoverSettingsResponse{
		ModelId:       *voiceoverReq.ModelId,
		SpeakerId:     *voiceoverReq.SpeakerId,
		Temperature:   *voiceoverReq.Temperature,
		Seed:          voiceoverReq.Seed,
		DenoiseAudio:  *voiceoverReq.DenoiseAudio,
		RemoveSilence: *voiceoverReq.RemoveSilence,
	}

	// Get queue count
	nq, err := c.QueueThrottler.NumQueued(fmt.Sprintf("v:%s", user.ID.String()))
	if err != nil {
		log.Warn("Error getting queue count for user", "err", err, "user_id", user.ID)
	}
	if err == nil && nq > qMax {
		// Get queue overflow size
		overflowSize, err := c.QueueThrottler.NumQueued(fmt.Sprintf("of:%s", user.ID.String()))
		if err != nil {
			log.Warn("Error getting queue overflow count", "err", err, "user_id", user.ID.String())
		}
		// If overflow size is greater than max, return error
		if overflowSize > shared.QUEUE_OVERFLOW_MAX {
			responses.ErrBadRequest(w, r, "queue_limit_reached", "")
			return
		}
		// Overflow size can be 0 so we need to add 1
		overflowSize++
		c.QueueThrottler.IncrementBy(1, fmt.Sprintf("of:%s", user.ID.String()))
		for {
			time.Sleep(time.Duration(shared.QUEUE_OVERFLOW_PENALTY_MS*overflowSize) * time.Millisecond)
			nq, err = c.QueueThrottler.NumQueued(fmt.Sprintf("v:%s", user.ID.String()))
			if err != nil {
				log.Warn("Error getting queue count", "err", err, "user_id", user.ID.String())
			}
			if err == nil && nq <= qMax {
				c.QueueThrottler.DecrementBy(1, fmt.Sprintf("of:%s", user.ID.String()))
				break
			}
			// Update overflow size
			overflowSize, err = c.QueueThrottler.NumQueued(fmt.Sprintf("of:%s", user.ID.String()))
			if err != nil {
				log.Warn("Error getting queue overflow count", "err", err, "user_id", user.ID.String())
			}
			overflowSize++
		}
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
	var requestId uuid.UUID
	// The cog request body
	var cogReqBody requests.CogQueueRequest
	// The total remaining credits
	var remainingCredits int

	// Create channel to track request
	// Create channel
	activeChl := make(chan requests.CogWebhookMessage)
	// Cleanup
	defer close(activeChl)

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
			&apiToken.ID,
			DB)
		if err != nil {
			log.Error("Error creating voiceover", "err", err)
			responses.ErrInternalServerError(w, r, "Error creating voiceover")
			return err
		}

		// Request ID matches upscale ID
		requestId = voiceover.ID

		// For live page update
		livePageMsg = shared.LivePageMessage{
			ProcessType:      shared.VOICEOVER,
			ID:               utils.Sha256(requestId.String()),
			CountryCode:      countryCode,
			Status:           shared.LivePageQueued,
			TargetNumOutputs: 1,
			CreatedAt:        voiceover.CreatedAt,
			ProductID:        user.ActiveProductID,
			Source:           shared.OperationSourceTypeAPI,
		}

		// Send to the cog
		cogReqBody = requests.CogQueueRequest{
			WebhookEventsFilter: []requests.CogEventFilter{requests.CogEventFilterStart, requests.CogEventFilterStart},
			WebhookUrl:          fmt.Sprintf("%s/v1/worker/webhook", utils.GetEnv("PUBLIC_API_URL", "")),
			Input: requests.BaseCogRequest{
				APIRequest:    true,
				ID:            requestId,
				IP:            utils.GetIPAddress(r),
				UserID:        &user.ID,
				DeviceInfo:    deviceInfo,
				LivePageData:  &livePageMsg,
				ProcessType:   shared.VOICEOVER,
				Model:         modelName,
				Speaker:       speakerName,
				ModelId:       *voiceoverReq.ModelId,
				Prompt:        voiceoverReq.Prompt,
				Temp:          voiceoverReq.Temperature,
				Seed:          voiceoverReq.Seed,
				RemoveSilence: voiceoverReq.RemoveSilence,
				DenoiseAudio:  voiceoverReq.DenoiseAudio,
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

	// Add channel to sync array (basically a thread-safe map)
	c.SMap.Put(requestId.String(), activeChl)
	defer c.SMap.Delete(requestId.String())

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

	// Analytics
	go c.Track.VoiceoverStarted(user, cogReqBody.Input, utils.GetIPAddress(r))

	// Wait for result
	for {
		select {
		case cogMsg := <-activeChl:
			switch cogMsg.Status {
			case requests.CogProcessing:
				err := c.Repo.SetVoiceoverStarted(requestId.String())
				if err != nil {
					log.Error("Failed to set voiceover started", "id", requestId, "err", err)
					responses.ErrInternalServerError(w, r, "An unknown error occurred")
					return
				}
				// Send live page update
				go func() {
					cogMsg.Input.LivePageData.Status = shared.LivePageProcessing
					now := time.Now()
					cogMsg.Input.LivePageData.StartedAt = &now
					liveResp := repository.TaskStatusUpdateResponse{
						ForLivePage:     true,
						LivePageMessage: cogMsg.Input.LivePageData,
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
			case requests.CogSucceeded:
				outputs, err := c.Repo.SetVoiceoverSucceeded(requestId.String(), voiceoverReq.Prompt, cogMsg.Output)
				if err != nil {
					log.Error("Failed to set voiceover succeeded", "id", upscale.ID, "err", err)
					responses.ErrInternalServerError(w, r, "An unknown error occurred")
					return
				}
				// Send live page update
				go func() {
					cogMsg.Input.LivePageData.Status = shared.LivePageSucceeded
					now := time.Now()
					cogMsg.Input.LivePageData.CompletedAt = &now
					cogMsg.Input.LivePageData.ActualNumOutputs = 1
					liveResp := repository.TaskStatusUpdateResponse{
						ForLivePage:     true,
						LivePageMessage: cogMsg.Input.LivePageData,
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
				// Analytics
				voiceover, err := c.Repo.GetVoiceover(requestId)
				if err != nil {
					log.Error("Error getting voiceover for analytics", "err", err)
				}
				// Get durations in seconds
				if voiceover.StartedAt == nil {
					log.Error("Voiceover started at is nil", "id", cogMsg.Input.ID)
				}
				duration := time.Now().Sub(*voiceover.StartedAt).Seconds()
				qDuration := (*voiceover.StartedAt).Sub(voiceover.CreatedAt).Seconds()
				go c.Track.VoiceoverSucceeded(user, cogMsg.Input, duration, qDuration, utils.GetIPAddress(r))

				// Format response
				resOutputs := make([]responses.ApiOutput, 1)
				resOutputs[0] = responses.ApiOutput{
					URL:           utils.GetURLFromAudioFilePath(outputs.AudioPath),
					ID:            outputs.ID,
					AudioDuration: utils.ToPtr(outputs.AudioDuration),
				}

				// Set token used
				err = c.Repo.SetTokenUsedAndIncrementCreditsSpent(int(utils.CalculateVoiceoverCredits(voiceoverReq.Prompt)), *voiceover.APITokenID)
				if err != nil {
					log.Error("Failed to set token used", "err", err)
				}

				render.Status(r, http.StatusOK)
				render.JSON(w, r, responses.ApiSucceededResponse{
					Outputs:          resOutputs,
					RemainingCredits: remainingCredits,
					Settings:         initSettings,
				})
				return
			case requests.CogFailed:
				if err := c.Repo.WithTx(func(tx *ent.Tx) error {
					DB := tx.Client()
					err := c.Repo.SetVoiceoverFailed(requestId.String(), cogMsg.Error, DB)
					if err != nil {
						log.Error("Failed to set voiceover failed", "id", upscale.ID, "err", err)
						responses.ErrInternalServerError(w, r, "An unknown error occurred")
						return err
					}
					// Send live page update
					go func() {
						cogMsg.Input.LivePageData.Status = shared.LivePageFailed
						now := time.Now()
						cogMsg.Input.LivePageData.CompletedAt = &now
						liveResp := repository.TaskStatusUpdateResponse{
							ForLivePage:     true,
							LivePageMessage: cogMsg.Input.LivePageData,
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
					// Analytics
					duration := time.Now().Sub(cogMsg.Input.LivePageData.CreatedAt).Seconds()
					go c.Track.VoiceoverFailed(user, cogMsg.Input, duration, cogMsg.Error, utils.GetIPAddress(r))
					// Refund credits
					_, err = c.Repo.RefundCreditsToUser(user.ID, utils.CalculateVoiceoverCredits(voiceoverReq.Prompt), DB)
					if err != nil {
						log.Error("Failed to refund credits", "err", err)
						return err
					}
					return nil
				}); err != nil {
					log.Error("Failed to set voiceover failed", "id", requestId, "err", err)
					responses.ErrInternalServerError(w, r, "An unknown error occurred")
					return
				}

				render.Status(r, http.StatusInternalServerError)
				render.JSON(w, r, responses.ApiFailedResponse{
					Error:    cogMsg.Error,
					Settings: initSettings,
				})
				return
			}
		case <-time.After(shared.REQUEST_COG_TIMEOUT_VOICEOVER):
			if err := c.Repo.WithTx(func(tx *ent.Tx) error {
				DB := tx.Client()
				err := c.Repo.SetVoiceoverFailed(requestId.String(), shared.TIMEOUT_ERROR, DB)
				if err != nil {
					log.Error("Failed to set voiceover failed", "id", upscale.ID, "err", err)
				}
				// Refund credits
				_, err = c.Repo.RefundCreditsToUser(user.ID, utils.CalculateVoiceoverCredits(voiceoverReq.Prompt), DB)
				if err != nil {
					log.Error("Failed to refund credits", "err", err)
					return err
				}
				return nil
			}); err != nil {
				log.Error("Failed to set voiceover failed", "id", requestId, "err", err)
				responses.ErrInternalServerError(w, r, "An unknown error occurred")
				return
			}

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, responses.ApiFailedResponse{
				Error:    shared.TIMEOUT_ERROR,
				Settings: initSettings,
			})
			return
		}
	}
}
