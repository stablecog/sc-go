package responses

import "github.com/google/uuid"

type AvailableScheduler struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type SettingsResponseItem struct {
	ID                  uuid.UUID            `json:"id"`
	Name                string               `json:"name"`
	Default             *bool                `json:"default,omitempty"`
	Active              *bool                `json:"active,omitempty"`
	AvailableSchedulers []AvailableScheduler `json:"available_schedulers,omitempty"`
	DefaultWidth        *int32               `json:"default_width,omitempty"`
	DefaultHeight       *int32               `json:"default_height,omitempty"`
}

type ImageGenerationSettingsResponse struct {
	Model          uuid.UUID `json:"model"`
	Scheduler      uuid.UUID `json:"scheduler"`
	Width          int32     `json:"width"`
	Height         int32     `json:"height"`
	NumOutputs     int32     `json:"num_outputs"`
	GuidanceScale  float32   `json:"guidance_scale"`
	InferenceSteps int32     `json:"inference_steps"`
	Seed           *int      `json:"seed,omitempty"`
}

type SettingsResponse struct {
	GenerationDefaults ImageGenerationSettingsResponse `json:"generation_defaults"`
	GenerationModels   []SettingsResponseItem          `json:"generation_models"`
	UpscaleModels      []SettingsResponseItem          `json:"upscale_models"`
}
