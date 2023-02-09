// * Requests that go from our application to the cog
package requests

import (
	"encoding/json"

	"github.com/stablecog/go-apps/shared"
)

// Filters specify what events we want the cog to send to our webhook
type WebhookEventFilterOption string

const (
	WebhookEventFilterStart     WebhookEventFilterOption = "start"
	WebhookEventFilterOutput    WebhookEventFilterOption = "output"
	WebhookEventFilterCompleted WebhookEventFilterOption = "completed"
)

// Base request data cog used to process request
type BaseCogRequest struct {
	// These fields are irrelevant to cog, just used to identify the request when it comes back
	ID                 string `json:"id"`
	GenerationOutputID string `json:"generation_output_id,omitempty"` // Specific to upscale requests
	// Generate specific
	UploadPathPrefix     string             `json:"upload_path_prefix,omitempty"`
	Prompt               string             `json:"prompt,omitempty"`
	NegativePrompt       string             `json:"negative_prompt,omitempty"`
	Width                string             `json:"width,omitempty"`
	Height               string             `json:"height,omitempty"`
	OutputImageExtension string             `json:"output_image_extension,omitempty"`
	OutputImageQuality   string             `json:"output_image_quality,omitempty"`
	NumInferenceSteps    string             `json:"num_inference_steps,omitempty"`
	GuidanceScale        string             `json:"guidance_scale,omitempty"`
	Model                string             `json:"model,omitempty"`
	Scheduler            string             `json:"scheduler,omitempty"`
	InitImage            string             `json:"init_image,omitempty"`
	PromptStrength       string             `json:"prompt_strength,omitempty"`
	Mask                 string             `json:"mask,omitempty"`
	Seed                 string             `json:"seed,omitempty"`
	NumOutputs           string             `json:"num_outputs,omitempty"`
	ProcessType          shared.ProcessType `json:"process_type"`
	PromptFlores         string             `json:"prompt_flores_200_code,omitempty"`
	NegativePromptFlores string             `json:"negative_prompt_flores_200_code,omitempty"`
	// Upscale specific
	Image string `json:"image_to_upscale,omitempty"`
}

// Data type is what we actually send to the cog, includes some additional metadata beyond BaseCogRequest
type CogQueueRequest struct {
	RedisPubsubKey      string                     `json:"redis_pubsub_key,omitempty"`
	WebhookEventsFilter []WebhookEventFilterOption `json:"webhook_events_filter"`
	Input               BaseCogRequest             `json:"input"`
}

func (i CogQueueRequest) MarshalBinary() (data []byte, err error) {
	return json.Marshal(i)
}
