package scworker

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/ent/upscale"
	"github.com/stablecog/sc-go/database/enttypes"
	"github.com/stablecog/sc-go/database/repository"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/server/analytics"
	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/server/responses"
	"github.com/stablecog/sc-go/shared"
	"github.com/stablecog/sc-go/utils"
	"golang.org/x/exp/slices"
)

func CreateVoiceover(source enttypes.SourceType,
	r *http.Request,
	repo *repository.Repository,
	redis *database.RedisWrapper,
	SMap *shared.SyncMap[chan requests.CogWebhookMessage],
	qThrottler *shared.UserQueueThrottlerMap,
	user *ent.User,
	track *analytics.AnalyticsService,
	apiTokenId *uuid.UUID,
	voiceoverReq requests.CreateVoiceoverRequest) (*responses.ApiSucceededResponse, *responses.VoiceoverSettingsResponse, *WorkerError) {
	free := user.ActiveProductID == nil
	if free {
		// Re-evaluate if they have paid credits
		count, err := repo.HasPaidCredits(user.ID)
		if err != nil {
			log.Error("Error getting paid credit sum for users", "err", err)
			return nil, nil, WorkerInternalServerError()
		}
		free = count <= 0
	}

	var qMax int
	roles, err := repo.GetRoles(user.ID)
	if err != nil {
		log.Error("Error getting roles for user", "err", err)
		return nil, nil, WorkerInternalServerError()
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

	if user.BannedAt != nil {
		return nil, nil, &WorkerError{http.StatusForbidden, fmt.Errorf("user_banned"), ""}
	}

	// Validation
	if !isSuperAdmin {
		err = voiceoverReq.Validate(true)
		if err != nil {
			return nil, nil, &WorkerError{http.StatusBadRequest, err, ""}
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
	nq, err := qThrottler.NumQueued(fmt.Sprintf("v:%s", user.ID.String()))
	if err != nil {
		log.Warn("Error getting queue count for user", "err", err, "user_id", user.ID)
	}
	if err == nil && nq > qMax {
		// Get queue overflow size
		overflowSize, err := qThrottler.NumQueued(fmt.Sprintf("of:%s", user.ID.String()))
		if err != nil {
			log.Warn("Error getting queue overflow count", "err", err, "user_id", user.ID.String())
		}
		// If overflow size is greater than max, return error
		if overflowSize > shared.QUEUE_OVERFLOW_MAX {
			return nil, nil, &WorkerError{http.StatusBadRequest, fmt.Errorf("queue_limit_reached"), ""}
		}
		// Overflow size can be 0 so we need to add 1
		overflowSize++
		qThrottler.IncrementBy(1, fmt.Sprintf("of:%s", user.ID.String()))
		for {
			time.Sleep(time.Duration(shared.QUEUE_OVERFLOW_PENALTY_MS*overflowSize) * time.Millisecond)
			nq, err = qThrottler.NumQueued(fmt.Sprintf("v:%s", user.ID.String()))
			if err != nil {
				log.Warn("Error getting queue count", "err", err, "user_id", user.ID.String())
			}
			if err == nil && nq <= qMax {
				qThrottler.DecrementBy(1, fmt.Sprintf("of:%s", user.ID.String()))
				break
			}
			// Update overflow size
			overflowSize, err = qThrottler.NumQueued(fmt.Sprintf("of:%s", user.ID.String()))
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
	var countryCode string
	var deviceInfo utils.ClientDeviceInfo
	ipAddress := "system"
	if r != nil {
		countryCode = utils.GetCountryCode(r)
		deviceInfo = utils.GetClientDeviceInfo(r)
		ipAddress = utils.GetIPAddress(r)
	} else {
		countryCode = "US"
		deviceInfo = utils.ClientDeviceInfo{
			DeviceType:    utils.Bot,
			DeviceOs:      "Linux",
			DeviceBrowser: "Discord",
		}
	}

	// Get model name for cog
	modelName := shared.GetCache().GetVoiceoverModelNameFromID(*voiceoverReq.ModelId)
	if modelName == "" {
		log.Error("Error getting model name", "model_name", modelName)
		return nil, nil, WorkerInternalServerError()
	}

	// Get speaker name for cog
	speakerName := shared.GetCache().GetVoiceoverSpeakerNameFromID(*voiceoverReq.SpeakerId)
	if speakerName == "" {
		log.Error("Error getting speaker name", "speaker_name", speakerName)
		return nil, nil, WorkerInternalServerError()
	}

	// For live page update
	var livePageMsg shared.LivePageMessage
	// For keeping track of this request as it gets sent to the worker
	var requestId uuid.UUID
	// Cog request
	var cogReqBody requests.CogQueueRequest

	// Credits left after this operation
	var remainingCredits int

	// Create channel to track request
	// Create channel
	activeChl := make(chan requests.CogWebhookMessage)
	// Cleanup
	defer close(activeChl)

	// Wrap everything in a DB transaction
	// We do this since we want our credit deduction to be atomic with the whole process
	if err := repo.WithTx(func(tx *ent.Tx) error {
		// Bind transaction to client
		DB := tx.Client()

		// Charge credits
		creditAmount := utils.CalculateVoiceoverCredits(voiceoverReq.Prompt)
		deducted, err := repo.DeductCreditsFromUser(user.ID, creditAmount, DB)
		if err != nil {
			log.Error("Error deducting credits", "err", err)
			return err
		} else if !deducted {
			return responses.InsufficientCreditsErr
		}

		remainingCredits, err = repo.GetNonExpiredCreditTotalForUser(user.ID, DB)
		if err != nil {
			log.Error("Error getting remaining credits", "err", err)
			return err
		}

		// Create voiceover
		voiceover, err := repo.CreateVoiceover(
			user.ID,
			string(deviceInfo.DeviceType),
			deviceInfo.DeviceOs,
			deviceInfo.DeviceBrowser,
			countryCode,
			voiceoverReq,
			user.ActiveProductID,
			apiTokenId,
			source,
			DB)
		if err != nil {
			log.Error("Error creating voiceover", "err", err)
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
			Source:           source,
			SpeakerID:        voiceoverReq.SpeakerId,
			RemoveSilence:    voiceoverReq.RemoveSilence,
			DenoiseAudio:     voiceoverReq.DenoiseAudio,
			Temperature:      voiceoverReq.Temperature,
		}

		// Send to the cog
		cogReqBody = requests.CogQueueRequest{
			WebhookEventsFilter: []requests.CogEventFilter{requests.CogEventFilterStart, requests.CogEventFilterStart},
			WebhookUrl:          fmt.Sprintf("%s/v1/worker/webhook", utils.GetEnv("PUBLIC_API_URL", "")),
			Input: requests.BaseCogRequest{
				APIRequest:    true,
				ID:            requestId,
				IP:            ipAddress,
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

		err = redis.EnqueueCogRequest(redis.Ctx, shared.COG_REDIS_VOICEOVER_QUEUE, cogReqBody)
		if err != nil {
			log.Error("Failed to write request to queue", "id", requestId, "err", err)
			return err
		}

		qThrottler.IncrementBy(1, fmt.Sprintf("v:%s", user.ID.String()))

		return nil
	}); err != nil {
		log.Error("Error in transaction", "err", err)
		if errors.Is(err, responses.InsufficientCreditsErr) {
			return nil, nil, &WorkerError{http.StatusBadRequest, responses.InsufficientCreditsErr, ""}
		}
		return nil, nil, WorkerInternalServerError()
	}

	// Add channel to sync array (basically a thread-safe map)
	SMap.Put(requestId.String(), activeChl)
	defer SMap.Delete(requestId.String())

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
		err = redis.Client.Publish(redis.Ctx, shared.REDIS_SSE_BROADCAST_CHANNEL, respBytes).Err()
		if err != nil {
			log.Error("Failed to publish live page update", "err", err)
		}
	}()

	// Analytics
	go track.VoiceoverStarted(user, cogReqBody.Input, source, ipAddress)

	// Wait for result
	for {
		select {
		case cogMsg := <-activeChl:
			switch cogMsg.Status {
			case requests.CogProcessing:
				err := repo.SetVoiceoverStarted(requestId.String())
				if err != nil {
					log.Error("Failed to set voiceover started", "id", requestId, "err", err)
					return nil, nil, WorkerInternalServerError()
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
					err = redis.Client.Publish(redis.Ctx, shared.REDIS_SSE_BROADCAST_CHANNEL, respBytes).Err()
					if err != nil {
						log.Error("Failed to publish live page update", "err", err)
					}
				}()
			case requests.CogSucceeded:
				outputs, err := repo.SetVoiceoverSucceeded(requestId.String(), voiceoverReq.Prompt, cogMsg.Output)
				if err != nil {
					log.Error("Failed to set voiceover succeeded", "id", upscale.ID, "err", err)
					return nil, nil, WorkerInternalServerError()
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
					err = redis.Client.Publish(redis.Ctx, shared.REDIS_SSE_BROADCAST_CHANNEL, respBytes).Err()
					if err != nil {
						log.Error("Failed to publish live page update", "err", err)
					}
				}()
				// Analytics
				voiceover, err := repo.GetVoiceover(requestId)
				if err != nil {
					log.Error("Error getting voiceover for analytics", "err", err)
				}
				// Get durations in seconds
				if voiceover.StartedAt == nil {
					log.Error("Voiceover started at is nil", "id", cogMsg.Input.ID)
				}
				// Analytics
				duration := time.Now().Sub(*voiceover.StartedAt).Seconds()
				qDuration := (*voiceover.StartedAt).Sub(voiceover.CreatedAt).Seconds()
				go track.VoiceoverSucceeded(user, cogMsg.Input, duration, qDuration, source, ipAddress)

				// Format response
				resOutputs := make([]responses.ApiOutput, 1)
				resOutputs[0] = responses.ApiOutput{
					URL:           utils.GetURLFromAudioFilePath(outputs.AudioPath),
					AudioFileURL:  utils.ToPtr(utils.GetURLFromAudioFilePath(outputs.AudioPath)),
					ID:            outputs.ID,
					AudioDuration: utils.ToPtr(outputs.AudioDuration),
				}
				if outputs.VideoPath != nil {
					resOutputs[0].VideoFileURL = utils.ToPtr(utils.GetURLFromAudioFilePath(*outputs.VideoPath))
				}

				// Set token used
				if voiceover.APITokenID != nil {
					err = repo.SetTokenUsedAndIncrementCreditsSpent(int(utils.CalculateVoiceoverCredits(voiceoverReq.Prompt)), *voiceover.APITokenID)
					if err != nil {
						log.Error("Failed to set token used", "err", err)
					}
				}

				return &responses.ApiSucceededResponse{
					Outputs:          resOutputs,
					RemainingCredits: remainingCredits,
					Settings:         initSettings,
				}, &initSettings, nil
			case requests.CogFailed:
				if err := repo.WithTx(func(tx *ent.Tx) error {
					DB := tx.Client()
					err := repo.SetVoiceoverFailed(requestId.String(), cogMsg.Error, DB)
					if err != nil {
						log.Error("Failed to set voiceover failed", "id", upscale.ID, "err", err)
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
						err = redis.Client.Publish(redis.Ctx, shared.REDIS_SSE_BROADCAST_CHANNEL, respBytes).Err()
						if err != nil {
							log.Error("Failed to publish live page update", "err", err)
						}
					}()
					// Analytics
					duration := time.Now().Sub(cogMsg.Input.LivePageData.CreatedAt).Seconds()
					go track.VoiceoverFailed(user, cogMsg.Input, duration, cogMsg.Error, source, ipAddress)
					// Refund credits
					_, err = repo.RefundCreditsToUser(user.ID, utils.CalculateVoiceoverCredits(voiceoverReq.Prompt), DB)
					if err != nil {
						log.Error("Failed to refund credits", "err", err)
						return err
					}
					return nil
				}); err != nil {
					log.Error("Failed to set voiceover failed", "id", requestId, "err", err)
					return nil, nil, WorkerInternalServerError()
				}

				return nil, &initSettings, &WorkerError{http.StatusInternalServerError, fmt.Errorf(cogMsg.Error), ""}
			}
		case <-time.After(shared.REQUEST_COG_TIMEOUT_VOICEOVER):
			if err := repo.WithTx(func(tx *ent.Tx) error {
				DB := tx.Client()
				err := repo.SetVoiceoverFailed(requestId.String(), shared.TIMEOUT_ERROR, DB)
				if err != nil {
					log.Error("Failed to set voiceover failed", "id", upscale.ID, "err", err)
				}
				// Refund credits
				_, err = repo.RefundCreditsToUser(user.ID, utils.CalculateVoiceoverCredits(voiceoverReq.Prompt), DB)
				if err != nil {
					log.Error("Failed to refund credits", "err", err)
					return err
				}
				return nil
			}); err != nil {
				log.Error("Failed to set voiceover failed", "id", requestId, "err", err)
				return nil, nil, WorkerInternalServerError()
			}

			return nil, nil, &WorkerError{http.StatusInternalServerError, fmt.Errorf(shared.TIMEOUT_ERROR), ""}
		}
	}
}
