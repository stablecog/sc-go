// * Requests initiated by logged in users
package requests

import "github.com/google/uuid"

// Request for submitting outputs to gallery
type GenerateSubmitToGalleryRequestBody struct {
	GenerationOutputIDs []uuid.UUID `json:"generation_output_ids"`
}

// Request for creating a new generation
type GenerateRequestBody struct {
	Prompt                string    `json:"prompt"`
	NegativePrompt        string    `json:"negative_prompt,omitempty"`
	Width                 int32     `json:"width"`
	Height                int32     `json:"height"`
	InferenceSteps        int32     `json:"inference_steps"`
	GuidanceScale         float32   `json:"guidance_scale"`
	ModelId               uuid.UUID `json:"model_id"`
	SchedulerId           uuid.UUID `json:"scheduler_id"`
	Seed                  int       `json:"seed"`
	NumOutputs            int       `json:"num_outputs,omitempty"`
	WebsocketId           string    `json:"websocket_id"`
	ShouldSubmitToGallery bool      `json:"should_submit_to_gallery"`
}
