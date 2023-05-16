package scworker

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/favadi/osinfo"
	"github.com/stablecog/sc-go/database"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/repository"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/server/analytics"
	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/shared"
	"github.com/stablecog/sc-go/utils"
)

// Create an Upscale in sc-worker, wait for result
func CreateUpscale(Track *analytics.AnalyticsService, Repo *repository.Repository, Redis *database.RedisWrapper, sMap *shared.SyncMap[chan requests.CogWebhookMessage], generation *ent.Generation, output *ent.GenerationOutput) error {
	if len(shared.GetCache().UpscaleModels) == 0 {
		log.Error("No upscale models available")
		return fmt.Errorf("No upscale models available")
	}
	upscaleModel := shared.GetCache().UpscaleModels[0]
	// Create req
	upscaleReq := requests.CreateUpscaleRequest{
		Type:    requests.UpscaleRequestTypeOutput,
		Input:   output.ID.String(),
		ModelId: upscaleModel.ID,
	}

	var upscale *ent.Upscale
	var requestId string
	var user *ent.User
	var err error

	// Create channel
	activeChl := make(chan requests.CogWebhookMessage)
	// Cleanup
	defer close(activeChl)
	defer sMap.Delete(requestId)

	if err := Repo.WithTx(func(tx *ent.Tx) error {
		// Bind transaction to client
		DB := tx.Client()

		user, err = generation.QueryUser().Only(Repo.Ctx)
		if err != nil {
			log.Error("Error getting user", "err", err)
			return err
		}

		// Create upscale
		upscale, err = Repo.CreateUpscale(
			user.ID,
			generation.Width,
			generation.Height,
			"server",
			osinfo.New().String(),
			"",
			"US",
			upscaleReq,
			user.ActiveProductID,
			true,
			nil,
			DB)
		if err != nil {
			log.Error("Error creating upscale", "err", err)
			return err
		}

		// Send to the cog
		requestId = upscale.ID.String()

		// Live page
		livePageMsg := &shared.LivePageMessage{
			ProcessType:      shared.UPSCALE,
			ID:               utils.Sha256(requestId),
			CountryCode:      "US",
			Status:           shared.LivePageQueued,
			TargetNumOutputs: 1,
			Width:            generation.Width,
			Height:           generation.Height,
			CreatedAt:        upscale.CreatedAt,
			ProductID:        user.ActiveProductID,
			SystemGenerated:  true,
		}

		cogReqBody := requests.CogQueueRequest{
			WebhookEventsFilter: []requests.CogEventFilter{requests.CogEventFilterStart, requests.CogEventFilterStart},
			WebhookUrl:          fmt.Sprintf("%s/v1/worker/webhook", utils.GetEnv("PUBLIC_API_URL", "")),
			Input: requests.BaseCogRequest{
				Internal:             true,
				ID:                   requestId,
				UIId:                 upscaleReq.UIId,
				UserID:               &user.ID,
				StreamID:             upscaleReq.StreamID,
				GenerationOutputID:   output.ID.String(),
				LivePageData:         livePageMsg,
				Image:                utils.GetURLFromImagePath(output.ImagePath),
				ProcessType:          shared.UPSCALE,
				Width:                fmt.Sprint(generation.Width),
				Height:               fmt.Sprint(generation.Height),
				UpscaleModel:         upscaleModel.NameInWorker,
				ModelId:              upscaleReq.ModelId,
				OutputImageExtension: string(shared.DEFAULT_UPSCALE_OUTPUT_EXTENSION),
				OutputImageQuality:   fmt.Sprint(shared.DEFAULT_UPSCALE_OUTPUT_QUALITY),
				Type:                 upscaleReq.Type,
			},
		}

		// Add channel to sync array (basically a thread-safe map)
		sMap.Put(requestId, activeChl)

		err = Redis.EnqueueCogRequest(Redis.Ctx, cogReqBody)
		if err != nil {
			log.Error("Failed to write request to queue", "id", upscale.ID, "err", err)
			return err
		}

		// Analytics
		go Track.UpscaleStarted(user, cogReqBody.Input, "system")

		// Send live page update
		go func() {
			liveResp := repository.TaskStatusUpdateResponse{
				ForLivePage:     true,
				LivePageMessage: livePageMsg,
			}
			respBytes, err := json.Marshal(liveResp)
			if err != nil {
				log.Error("Error marshalling sse live response", "err", err)
				return
			}
			err = Redis.Client.Publish(Redis.Ctx, shared.REDIS_SSE_BROADCAST_CHANNEL, respBytes).Err()
			if err != nil {
				log.Error("Failed to publish live page update", "err", err)
			}
		}()

		return nil
	}); err != nil {
		return err
	}

	for {
		select {
		case cogMsg := <-activeChl:
			switch cogMsg.Status {
			case requests.CogProcessing:
				err := Repo.SetUpscaleStarted(upscale.ID.String())
				if err != nil {
					log.Error("Failed to set upscale started", "id", upscale.ID, "err", err)
					return err
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
					err = Redis.Client.Publish(Redis.Ctx, shared.REDIS_SSE_BROADCAST_CHANNEL, respBytes).Err()
					if err != nil {
						log.Error("Failed to publish live page update", "err", err)
					}
				}()
			case requests.CogSucceeded:
				_, err := Repo.SetUpscaleSucceeded(upscale.ID.String(), output.ID.String(), cogMsg.Input.Image, cogMsg.Output)
				if err != nil {
					log.Error("Failed to set upscale succeeded", "id", upscale.ID, "err", err)
				}
				// Send live page update
				go func() {
					cogMsg.Input.LivePageData.Status = shared.LivePageSucceeded
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
					err = Redis.Client.Publish(Redis.Ctx, shared.REDIS_SSE_BROADCAST_CHANNEL, respBytes).Err()
					if err != nil {
						log.Error("Failed to publish live page update", "err", err)
					}
				}()
				// Analytics
				upscale, err := Repo.GetUpscale(upscale.ID)
				if err != nil {
					log.Error("Error getting upscale for analytics", "err", err)
					return err
				}
				// Get durations in seconds
				if upscale.StartedAt == nil {
					log.Error("Upscale started at is nil", "id", cogMsg.Input.ID)
					return errors.New("Upscale started at is nil")
				}
				duration := time.Now().Sub(*upscale.StartedAt).Seconds()
				qDuration := (*upscale.StartedAt).Sub(upscale.CreatedAt).Seconds()
				go Track.UpscaleSucceeded(user, cogMsg.Input, duration, qDuration, "system")
				return err
			case requests.CogFailed:
				err := Repo.SetUpscaleFailed(upscale.ID.String(), cogMsg.Error, nil)
				if err != nil {
					log.Error("Failed to set upscale failed", "id", upscale.ID, "err", err)
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
					err = Redis.Client.Publish(Redis.Ctx, shared.REDIS_SSE_BROADCAST_CHANNEL, respBytes).Err()
					if err != nil {
						log.Error("Failed to publish live page update", "err", err)
					}
				}()
				// Analytics
				duration := time.Now().Sub(cogMsg.Input.LivePageData.CreatedAt).Seconds()
				go Track.UpscaleFailed(user, cogMsg.Input, duration, cogMsg.Error, "system")
				return err
			}
		case <-time.After(shared.REQUEST_COG_TIMEOUT):
			err := Repo.SetUpscaleFailed(upscale.ID.String(), shared.TIMEOUT_ERROR, nil)
			if err != nil {
				log.Error("Failed to set upscale failed", "id", upscale.ID, "err", err)
			}
			return fmt.Errorf(shared.TIMEOUT_ERROR)
		}
	}
}
