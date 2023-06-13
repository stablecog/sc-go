package rest

import (
	"encoding/json"
	"io"
	"net/http"
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
)

// Webhook for worker results
func (c *RestAPI) HandleSCWorkerWebhook(w http.ResponseWriter, r *http.Request) {
	// Verify signature of request
	sig := r.Header.Get("signature")
	expectedSig := utils.GetEnv("SC_WORKER_WEBHOOK_SECRET", "invalid")
	if sig != expectedSig {
		responses.ErrUnauthorized(w, r)
		return
	}

	// Parse request body
	reqBody, _ := io.ReadAll(r.Body)
	var cogMessage requests.CogWebhookMessage
	err := json.Unmarshal(reqBody, &cogMessage)
	if err != nil {
		log.Errorf("Failed to parse COG webhook message, %v", err)
		responses.ErrUnableToParseJson(w, r)
		return
	} else if cogMessage.Input.APIRequest {
		// API request handled in a separate flow
		err = c.Redis.Client.Publish(c.Redis.Ctx, shared.REDIS_APITOKEN_COG_CHANNEL, reqBody).Err()
		if err != nil {
			log.Error("Failed to publish API worker msg", "err", err)
		}
		render.Status(r, http.StatusOK)
		render.PlainText(w, r, "OK")
		return
	}

	if cogMessage.Input.Internal {
		// Internal request handled in a separate flow
		err = c.Redis.Client.Publish(c.Redis.Ctx, shared.REDIS_INTERNAL_COG_CHANNEL, reqBody).Err()
		if err != nil {
			log.Error("Failed to publish internal worker msg", "err", err)
		}
		render.Status(r, http.StatusOK)
		render.PlainText(w, r, "OK")
		return
	}

	// Process live page message and analytics
	go func() {
		// Live page update
		livePageMsg := cogMessage.Input.LivePageData
		if cogMessage.Status == requests.CogProcessing {
			livePageMsg.Status = shared.LivePageProcessing
		} else if cogMessage.Status == requests.CogSucceeded && len(cogMessage.Output.Images) > 0 {
			livePageMsg.Status = shared.LivePageSucceeded
		} else if cogMessage.Status == requests.CogSucceeded && cogMessage.NSFWCount > 0 {
			livePageMsg.Status = shared.LivePageFailed
			livePageMsg.FailureReason = shared.NSFW_ERROR
		} else {
			livePageMsg.Status = shared.LivePageFailed
		}

		now := time.Now()
		if cogMessage.Status == requests.CogProcessing {
			livePageMsg.StartedAt = &now
		}
		if cogMessage.Status == requests.CogSucceeded || cogMessage.Status == requests.CogFailed {
			livePageMsg.CompletedAt = &now
			livePageMsg.ActualNumOutputs = len(cogMessage.Output.Images)
			livePageMsg.NSFWCount = cogMessage.NSFWCount
		}
		// Send live page update
		liveResp := repository.TaskStatusUpdateResponse{
			ForLivePage:     true,
			LivePageMessage: livePageMsg,
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
	go func() {
		if cogMessage.Input.UserID == nil {
			return
		}
		if cogMessage.Status == requests.CogSucceeded && len(cogMessage.Output.Images) > 0 {
			u, err := c.Repo.GetUser(*cogMessage.Input.UserID)
			if err != nil {
				log.Error("Error getting user for analytics", "err", err)
				return
			}
			if cogMessage.Input.ProcessType == shared.GENERATE || cogMessage.Input.ProcessType == shared.GENERATE_AND_UPSCALE {
				// Get generation
				uid, _ := uuid.Parse(cogMessage.Input.ID)
				generation, err := c.Repo.GetGeneration(uid)
				if err != nil {
					log.Error("Error getting generation for analytics", "err", err)
					return
				}
				// Get durations in seconds
				if generation.StartedAt == nil {
					log.Error("Generation started at is nil", "id", cogMessage.Input.ID)
					return
				}
				duration := time.Now().Sub(*generation.StartedAt).Seconds()
				qDuration := (*generation.StartedAt).Sub(generation.CreatedAt).Seconds()
				c.Track.GenerationSucceeded(u, cogMessage.Input, duration, qDuration, cogMessage.Input.IP)
			} else if cogMessage.Input.ProcessType == shared.UPSCALE {
				// Get upscale
				uid, _ := uuid.Parse(cogMessage.Input.ID)
				upscale, err := c.Repo.GetUpscale(uid)
				if err != nil {
					log.Error("Error getting upscale for analytics", "err", err)
					return
				}
				// Get durations in seconds
				if upscale.StartedAt == nil {
					log.Error("Upscale started at is nil", "id", cogMessage.Input.ID)
					return
				}
				duration := time.Now().Sub(*upscale.StartedAt).Seconds()
				qDuration := (*upscale.StartedAt).Sub(upscale.CreatedAt).Seconds()
				c.Track.UpscaleSucceeded(u, cogMessage.Input, duration, qDuration, cogMessage.Input.IP)
			}
		}

		if cogMessage.Status == requests.CogSucceeded && cogMessage.NSFWCount > 0 {
			u, err := c.Repo.GetUser(*cogMessage.Input.UserID)
			if err != nil {
				log.Error("Error getting user for analytics", "err", err)
				return
			}
			// Get duration in seconds
			duration := time.Now().Sub(cogMessage.Input.LivePageData.CreatedAt).Seconds()
			if cogMessage.Input.ProcessType == shared.GENERATE || cogMessage.Input.ProcessType == shared.GENERATE_AND_UPSCALE {
				c.Track.GenerationFailedNSFW(u, cogMessage.Input, duration, cogMessage.Input.IP)
			}
		}

		if cogMessage.Status == requests.CogFailed {
			u, err := c.Repo.GetUser(*cogMessage.Input.UserID)
			if err != nil {
				log.Error("Error getting user for analytics", "err", err)
				return
			}
			// Get duration in seconds
			duration := time.Now().Sub(cogMessage.Input.LivePageData.CreatedAt).Seconds()
			if cogMessage.Input.ProcessType == shared.GENERATE || cogMessage.Input.ProcessType == shared.GENERATE_AND_UPSCALE {
				c.Track.GenerationFailed(u, cogMessage.Input, duration, cogMessage.Error, cogMessage.Input.IP)
			} else if cogMessage.Input.ProcessType == shared.UPSCALE {
				c.Track.UpscaleFailed(u, cogMessage.Input, duration, cogMessage.Error, cogMessage.Input.IP)
			}
		}
	}()

	// Process in database
	err = c.Repo.ProcessCogMessage(cogMessage)
	if err != nil {
		log.Error("Error processing COG message", "err", err)
		if ent.IsConstraintError(err) {
			// Squish
			render.Status(r, http.StatusOK)
			render.PlainText(w, r, "OK")
			return
		}
		responses.ErrInternalServerError(w, r, "server error")
		return
	}

	render.Status(r, http.StatusOK)
	render.PlainText(w, r, "OK")
}
