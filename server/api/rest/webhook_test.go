package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/stablecog/go-apps/server/requests"
	"github.com/stablecog/go-apps/shared"
	"github.com/stretchr/testify/assert"
)

// Ensures webhook broadcasts to redis when status is one we want
func TestHandleCogWebhookValid(t *testing.T) {
	/* Setup */
	pubsub := MockController.Redis.Client.Subscribe(context.Background(), shared.COG_REDIS_WEBHOOK_QUEUE_CHANNEL)
	reqBody := requests.WebhookRequest{
		Input: requests.WebhookRequestInput{
			Id: "123",
		},
		Status: requests.WebhookSucceeded,
	}
	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	// Build request
	req := httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	MockController.HandleCogWebhook(w, req)
	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)

	// Read message
	msg, err := pubsub.ReceiveMessage(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, string(body), msg.Payload)
}

// ! Idk how to test this one, ReceiveMessage panics
// ! ReceiveTimeout doesn't return an error after the timeout even though it says it will
// func TestHandleCogWebhookInvalid(t *testing.T) {
// 	/* Setup */
// 	pubsub := MockController.Redis.Client.Subscribe(context.Background(), shared.COG_REDIS_WEBHOOK_QUEUE_CHANNEL)
// 	reqBody := requests.WebhookRequest{
// 		Input: requests.WebhookRequestInput{
// 			Id: "123",
// 		},
// 		Status: "idkwhatthisstatusis",
// 	}
// 	body, _ := json.Marshal(reqBody)
// 	w := httptest.NewRecorder()
// 	// Build request
// 	req := httptest.NewRequest("POST", "/", bytes.NewReader(body))
// 	req.Header.Set("Content-Type", "application/json")

// 	MockController.HandleCogWebhook(w, req)
// 	resp := w.Result()
// 	defer resp.Body.Close()
// 	assert.Equal(t, 200, resp.StatusCode)

// 	// Read message
// 	_, err := pubsub.ReceiveMessage(context.Background())
// 	assert.NotNil(t, err)
// }
