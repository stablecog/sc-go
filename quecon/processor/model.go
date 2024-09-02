package processor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hibiken/asynq"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/server/responses"
	"github.com/stablecog/sc-go/utils"
)

func (p *QueueProcessor) HandleImageJob(ctx context.Context, t *asynq.Task) error {
	start := time.Now()

	var payload requests.RunpodInput
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	// Informative logging
	log.Infof("Processing image job %s, model %s", payload.Input.ID.String(), payload.Input.Model)

	if payload.Input.RunpodEndpoint == nil {
		log.Errorf("Received job with no runpod endpoint %s", payload.Input.ID.String())
		go func() {
			p.IssueSCWebhook(requests.CogWebhookMessage{
				Status: requests.CogFailed,
				Input:  payload.Input,
				Error:  "runpod_endpoint_not_set",
			}, 0)
		}()
		return fmt.Errorf("runpod_endpoint_not_set: %w", asynq.SkipRetry)
	}

	// Post processing to webhook
	go func() {
		// Retry webhook
		p.IssueSCWebhook(requests.CogWebhookMessage{
			Status: requests.CogProcessing,
			Input:  payload.Input,
		}, 0)
	}()

	// Issue task to runpod
	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Errorf("Error marshaling payload: %v", err)
		return fmt.Errorf("json.Marshal failed: %v: %w", err, asynq.SkipRetry)
	}

	// Create a new request
	req, err := http.NewRequest("POST", *payload.Input.RunpodEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("http.NewRequest failed: %v: %w", err, asynq.SkipRetry)
	}

	// Set the content type and signature header
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", utils.GetEnv().RunpodApiToken))

	// Perform the request
	resp, err := p.Client.Do(req)
	if err != nil {
		return fmt.Errorf("http.Client.Do failed: %v", err)
	}
	defer resp.Body.Close()

	// Unmarshal response
	var runpodResponse responses.RunpodOutput
	if err := json.NewDecoder(resp.Body).Decode(&runpodResponse); err != nil {
		log.Errorf("Error decoding runpod response: %v", err)
		// Send error to webhook
		go func() {
			p.IssueSCWebhook(requests.CogWebhookMessage{
				Status: requests.CogFailed,
				Input:  payload.Input,
				Error:  "error_decoding_runpod_response",
			}, 0)
		}()
		return fmt.Errorf("error_decoding_runpod_response: %w", asynq.SkipRetry)
	}

	if runpodResponse.Status != responses.RunpodStatusCompleted || len(runpodResponse.Output.Output.Images) == 0 {
		errorMsg := runpodResponse.Error
		if errorMsg == "" {
			errorMsg = "runpod_failed"
		} else if len(runpodResponse.Output.Output.Images) == 0 {
			errorMsg = "no_outputs"
		}
		log.Errorf("Runpod failed for task %s: %s", payload.Input.ID, errorMsg)
		// Send error to webhook
		go func() {
			p.IssueSCWebhook(requests.CogWebhookMessage{
				Status: requests.CogFailed,
				Input:  payload.Input,
				Error:  errorMsg,
			}, 0)
		}()
		return fmt.Errorf("runpod_failed: %w", asynq.SkipRetry)
	}

	// Send success to webhook
	go func() {
		// Convert shape of images array for compatibility
		images := make([]requests.CogWebhookOutputImage, len(runpodResponse.Output.Output.Images))
		for i, url := range runpodResponse.Output.Output.Images {
			images[i] = requests.CogWebhookOutputImage{Image: url}
		}

		p.IssueSCWebhook(requests.CogWebhookMessage{
			Status: requests.CogSucceeded,
			Input:  payload.Input,
			Output: requests.CogWebhookOutput{
				Images: images,
			},
		}, 0)
	}()

	end := time.Now()
	//Log duration in seconds
	log.Infof("Generated %d outputs of %s in %f seconds", len(runpodResponse.Output.Output.Images), payload.Input.Model, end.Sub(start).Seconds())

	return nil
}
