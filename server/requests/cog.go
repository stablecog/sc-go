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
	Webhook             string                     `json:"webhook"`
	WebhookEventsFilter []WebhookEventFilterOption `json:"webhook_events_filter"`
}

// ! Generate

// Base request
type BaseCogGenerateRequest struct {
	ID                   string `json:"id"`
	Prompt               string `json:"prompt"`
	NegativePrompt       string `json:"negative_prompt,omitempty"`
	Width                string `json:"width"`
	Height               string `json:"height"`
	OutputImageExt       string `json:"output_image_ext"`
	NumInferenceSteps    string `json:"num_inference_steps"`
	GuidanceScale        string `json:"guidance_scale"`
	Model                string `json:"model"`
	Scheduler            string `json:"scheduler"`
	InitImage            string `json:"init_image,omitempty"`
	PromptStrength       string `json:"prompt_strength,omitempty"`
	Mask                 string `json:"mask,omitempty"`
	Seed                 string `json:"seed"`
	NumOutputs           string `json:"num_outputs"`
	PromptFlores         string `json:"prompt_flores_200_code,omitempty"`
	NegativePromptFlores string `json:"negative_prompt_flores_200_code,omitempty"`
}

// Redis queue request
type CogGenerateQueueRequest struct {
	BaseCogRequestQueue
	Input BaseCogGenerateRequest `json:"input"`
}

func (i CogGenerateQueueRequest) MarshalBinary() (data []byte, err error) {
	return json.Marshal(i)
}
