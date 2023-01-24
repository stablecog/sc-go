package models

// Represents requests that go directly from our app to the cog

type WebhookEventFilterOption string

const (
	WebhookEventStart     WebhookEventFilterOption = "start"
	WebhookEventOutput    WebhookEventFilterOption = "output"
	WebhookEventCompleted WebhookEventFilterOption = "completed"
)

// Common fields for all requests using cog's redis queue
type BaseCogRequestQueue struct {
	Webhook             string                     `json:"webhook"`
	WebhookEventsFilter []WebhookEventFilterOption `json:"webhook_events_filter"`
}

// ! Generate

// Base request
type BaseCogGenerateRequest struct {
	ID                   string  `json:"id"`
	Prompt               string  `json:"prompt"`
	NegativePrompt       string  `json:"negative_prompt,omitempty"`
	Width                string  `json:"width"`
	Height               string  `json:"height"`
	OutputImageExt       string  `json:"output_image_ext"`
	NumInferenceSteps    string  `json:"num_inference_steps"`
	GuidanceScale        string  `json:"guidance_scale"`
	Model                string  `json:"model"`
	Scheduler            string  `json:"scheduler"`
	InitImage            string  `json:"init_image,omitempty"`
	PromptStrength       float32 `json:"prompt_strength,omitempty"`
	Mask                 string  `json:"mask,omitempty"`
	Seed                 string  `json:"seed"`
	PromptFlores         string  `json:"prompt_flores_200_code,omitempty"`
	NegativePromptFlores string  `json:"negative_prompt_flores_200_code,omitempty"`
}

// Redis queue request
type CogGenerateQueueRequest struct {
	BaseCogRequestQueue
	BaseCogGenerateRequest
}
