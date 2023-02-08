// * Requests that go from our application to the cog
package requests

import "encoding/json"

// Filters specify what events we want the cog to send to our webhook
type WebhookEventFilterOption string

const (
	WebhookEventFilterStart     WebhookEventFilterOption = "start"
	WebhookEventFilterOutput    WebhookEventFilterOption = "output"
	WebhookEventFilterCompleted WebhookEventFilterOption = "completed"
)

// Common fields for all requests using cog's redis queue
type BaseCogRequestQueue struct {
	Webhook             string                     `json:"webhook,omitempty"`
	RedisPubsubKey      string                     `json:"redis_pubsub_key,omitempty"`
	WebhookEventsFilter []WebhookEventFilterOption `json:"webhook_events_filter"`
}

// ! Generate

// Base request
type BaseCogGenerateRequest struct {
	ID                   string `json:"id"`
	UploadPathPrefix     string `json:"upload_path_prefix,omitempty"`
	Prompt               string `json:"prompt"`
	NegativePrompt       string `json:"negative_prompt,omitempty"`
	Width                string `json:"width"`
	Height               string `json:"height"`
	OutputImageExtension string `json:"output_image_extension"`
	OutputImageQuality   string `json:"output_image_quality"`
	NumInferenceSteps    string `json:"num_inference_steps"`
	GuidanceScale        string `json:"guidance_scale"`
	Model                string `json:"model"`
	Scheduler            string `json:"scheduler"`
	InitImage            string `json:"init_image,omitempty"`
	PromptStrength       string `json:"prompt_strength,omitempty"`
	Mask                 string `json:"mask,omitempty"`
	Seed                 string `json:"seed"`
	NumOutputs           string `json:"num_outputs"`
	ProcessType          string `json:"process_type"`
	PromptFlores         string `json:"prompt_flores_200_code,omitempty"`
	NegativePromptFlores string `json:"negative_prompt_flores_200_code,omitempty"`
}

// ! Upscale

// Base request
type BaseCogUpscaleRequest struct {
	// These are irrelevant to the cog, just used in our return messages
	ID                 string `json:"id"`
	GenerationOutputID string `json:"generation_output_id,omitempty"`
	// These fields actually go to the cog
	Image       string `json:"image_u"`
	Task        string `json:"task_u,omitempty"`
	ProcessType string `json:"process_type"`
}

// Redis queue requests
type CogGenerateQueueRequest struct {
	BaseCogRequestQueue
	Input BaseCogGenerateRequest `json:"input"`
}

func (i CogGenerateQueueRequest) MarshalBinary() (data []byte, err error) {
	return json.Marshal(i)
}

type CogUpscaleQueueRequest struct {
	BaseCogRequestQueue
	Input BaseCogUpscaleRequest `json:"input"`
}

func (i CogUpscaleQueueRequest) MarshalBinary() (data []byte, err error) {
	return json.Marshal(i)
}
