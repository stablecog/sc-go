package responses

import (
	"github.com/google/uuid"
)

type AvailableScheduler struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type SettingsResponseItem struct {
	ID                  uuid.UUID            `json:"id"`
	Name                string               `json:"name"`
	IsDefault           *bool                `json:"is_default,omitempty"`
	Active              *bool                `json:"active,omitempty"`
	AvailableSchedulers []AvailableScheduler `json:"available_schedulers,omitempty"`
	DefaultWidth        *int32               `json:"default_width,omitempty"`
	DefaultHeight       *int32               `json:"default_height,omitempty"`
}

type ImageGenerationSettingsResponse struct {
	ModelId        uuid.UUID `json:"model_id"`
	SchedulerId    uuid.UUID `json:"scheduler_id"`
	Width          int32     `json:"width"`
	Height         int32     `json:"height"`
	NumOutputs     int32     `json:"num_outputs"`
	GuidanceScale  float32   `json:"guidance_scale"`
	InferenceSteps int32     `json:"inference_steps"`
	Seed           *int      `json:"seed,omitempty"`
	InitImageURL   string    `json:"init_image_url,omitempty"`
	PromptStrength *float32  `json:"prompt_strength,omitempty"`
}

type ImageUpscaleSettingsResponse struct {
	ModelId uuid.UUID `json:"model_id"`
	Input   string    `json:"input,omitempty"`
}

type ImageModelsResponse struct {
	Models []SettingsResponseItem `json:"models"`
}
