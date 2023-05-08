package requests

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/stablecog/sc-go/shared"
	"github.com/stablecog/sc-go/utils"
)

// ! From our application to sc-worker

// Filters specify what events we want sc-worker to send to our webhook
type CogEventFilter string

const (
	CogEventFilterStart     CogEventFilter = "start"
	CogEventFilterOutput    CogEventFilter = "output"
	CogEventFilterCompleted CogEventFilter = "completed"
)

// Base request data sc-worker uses to process request
type BaseCogRequest struct {
	// These fields are irrelevant to sc-worker, just used to identify the request when it comes back
	ID                 string                  `json:"id"`
	UserID             *uuid.UUID              `json:"user_id,omitempty"`
	IP                 string                  `json:"ip,omitempty"`
	UIId               string                  `json:"ui_id,omitempty"`
	GenerationOutputID string                  `json:"generation_output_id,omitempty"` // Specific to upscale requests
	LivePageData       *shared.LivePageMessage `json:"live_page_data,omitempty"`
	StreamID           string                  `json:"stream_id,omitempty"`
	DeviceInfo         utils.ClientDeviceInfo  `json:"device_info,omitempty"`
	Internal           bool                    `json:"internal,omitempty"`    // Used to indicate if the request is internal or not
	APIRequest         bool                    `json:"api_request,omitempty"` // Used to indicate if the request is from token or not
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
	ModelId              uuid.UUID          `json:"model_id,omitempty"`
	Scheduler            string             `json:"scheduler,omitempty"`
	SchedulerId          uuid.UUID          `json:"scheduler_id,omitempty"`
	InitImageUrl         string             `json:"init_image_url,omitempty"`
	InitImageUrlS3       string             `json:"init_image_url_s3,omitempty"`
	PromptStrength       string             `json:"prompt_strength,omitempty"`
	Mask                 string             `json:"mask,omitempty"`
	Seed                 string             `json:"seed,omitempty"`
	NumOutputs           string             `json:"num_outputs,omitempty"`
	ProcessType          shared.ProcessType `json:"process_type"`
	PromptFlores         string             `json:"prompt_flores_200_code,omitempty"`
	NegativePromptFlores string             `json:"negative_prompt_flores_200_code,omitempty"`
	SubmitToGallery      bool               `json:"submit_to_gallery,omitempty"`
	// Upscale specific
	Image        string             `json:"image_to_upscale,omitempty"`
	Type         UpscaleRequestType `json:"type,omitempty"`
	UpscaleModel string             `json:"upscale_model,omitempty"`
}

// Data type is what we actually send to the cog, includes some additional metadata beyond BaseCogRequest
type CogQueueRequest struct {
	WebhookUrl          string           `json:"webhook_url,omitempty"`
	WebhookEventsFilter []CogEventFilter `json:"webhook_events_filter"`
	Input               BaseCogRequest   `json:"input"`
}

func (i CogQueueRequest) MarshalBinary() (data []byte, err error) {
	return json.Marshal(i)
}

// ! From sc-worker to our application

type CogTaskStatus string

const (
	CogSucceeded  CogTaskStatus = "succeeded"
	CogFailed     CogTaskStatus = "failed"
	CogProcessing CogTaskStatus = "processing"
)

// Msg from sc-worker to redis channel
type CogWebhookOutputImage struct {
	Image      string    `json:"image"`
	ImageEmbed []float32 `json:"image_embed"`
}

type CogWebhookOutput struct {
	PromptEmbed []float32               `json:"prompt_embed"`
	Images      []CogWebhookOutputImage `json:"images"`
}

type CogWebhookMessage struct {
	Webhook   string           `json:"webhook"`
	Input     BaseCogRequest   `json:"input"`
	Status    CogTaskStatus    `json:"status"`
	Error     string           `json:"error"`
	Output    CogWebhookOutput `json:"output"`
	NSFWCount int32            `json:"nsfw_count"`
}
