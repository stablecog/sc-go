package rest

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/render"
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
		responses.ErrUnableToParseJson(w, r)
		return
	}

	log.Infof("Received COG message, %v", cogMessage)

	// Process live page message and analytics
	go func() {
		// Live page update
		livePageMsg := cogMessage.Input.LivePageData
		if cogMessage.Status == requests.CogProcessing {
			livePageMsg.Status = shared.LivePageProcessing
		} else if cogMessage.Status == requests.CogSucceeded && len(cogMessage.Outputs) > 0 {
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
			livePageMsg.ActualNumOutputs = len(cogMessage.Outputs)
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

	// Process in database
	err = c.Repo.ProcessCogMessage(cogMessage)
	if err != nil {
		log.Error("Error processing COG message", "err", err)
		responses.ErrInternalServerError(w, r, "server error")
	}

	render.Status(r, http.StatusOK)
	render.PlainText(w, r, "OK")
}
