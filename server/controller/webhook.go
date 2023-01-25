package controller

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/render"
	"github.com/stablecog/go-apps/server/requests"
	"github.com/stablecog/go-apps/server/responses"
	"github.com/stablecog/go-apps/shared"
	"k8s.io/klog/v2"
)

// HTTP Post for cog webhook
func (c *HttpController) PostWebhook(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	reqBody, _ := io.ReadAll(r.Body)
	var req requests.WebhookRequest
	err := json.Unmarshal(reqBody, &req)
	if err != nil {
		responses.ErrUnableToParseJson(w, r)
		return
	}

	klog.Infof("-- Webhook request received: %v --", req)
	if req.Status == requests.WebhookSucceeded || req.Status == requests.WebhookFailed || req.Status == requests.WebhookProcessing {
		// Publish to redis channel
		marshalled, err := json.Marshal(req)
		if err != nil {
			klog.Errorf("-- Error marshalling webhook request: %v --", err)
			render.Status(r, http.StatusInternalServerError)
			return
		}
		err = c.Redis.Client.Publish(r.Context(), shared.COG_REDIS_WEBHOOK_QUEUE_CHANNEL, string(marshalled)).Err()
		if err != nil {
			klog.Errorf("-- Error publishing to redis channel: %v --", err)
			render.Status(r, http.StatusInternalServerError)
			return
		}
	}

	render.Status(r, http.StatusOK)
}
