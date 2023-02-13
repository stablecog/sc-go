// * Requests initiated by logged in users
package requests

import (
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/shared"
)

// Request for submitting outputs to gallery
type GenerateSubmitToGalleryRequestBody struct {
	GenerationOutputIDs []uuid.UUID `json:"generation_output_ids"`
}

// Request for creating a new generation
type GenerateRequestBody struct {
	Prompt               string             `json:"prompt"`
	NegativePrompt       string             `json:"negative_prompt,omitempty"`
	Width                int32              `json:"width"`
	Height               int32              `json:"height"`
	InferenceSteps       int32              `json:"inference_steps"`
	GuidanceScale        float32            `json:"guidance_scale"`
	ModelId              uuid.UUID          `json:"model_id"`
	SchedulerId          uuid.UUID          `json:"scheduler_id"`
	Seed                 int                `json:"seed"`
	NumOutputs           int32              `json:"num_outputs,omitempty"`
	StreamID             string             `json:"stream_id"` // Corresponds to SSE stream
	SubmitToGallery      bool               `json:"submit_to_gallery"`
	ProcessType          shared.ProcessType `json:"process_type"`
	OutputImageExtension string             `json:"output_image_extension"`
}

// Request for initiationg an upscale
type UpscaleRequestType string

const (
	UpscaleRequestTypeImage  UpscaleRequestType = "from_image"
	UpscaleRequestTypeOutput UpscaleRequestType = "from_output"
)

// Can be initiated with either an image_url or a generation_output_id
type UpscaleRequestBody struct {
	Type     UpscaleRequestType `json:"type"`
	Input    string             `json:"input"`
	ModelId  uuid.UUID          `json:"model_id"`
	StreamID string             `json:"stream_id"`
}
