package responses

import "github.com/google/uuid"

type SettingsResponseItem struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	Default *bool     `json:"default,omitempty"`
	Active  *bool     `json:"active,omitempty"`
}

type ImageGenerationSettingsResponse struct {
	Model          uuid.UUID `json:"model"`
	Scheduler      uuid.UUID `json:"scheduler"`
	Width          int32     `json:"width"`
	Height         int32     `json:"height"`
	NumImages      int32     `json:"num_images"`
	GuidanceScale  float32   `json:"guidance_scale"`
	InferenceSteps int32     `json:"inference_steps"`
	Seed           *int      `json:"seed,omitempty"`
}

type SettingsResponse struct {
	GenerationDefaults ImageGenerationSettingsResponse `json:"generation_defaults"`
	GenerationModels   []SettingsResponseItem          `json:"generation_models"`
	UpscaleModels      []SettingsResponseItem          `json:"upscale_models"`
	Schedulers         []SettingsResponseItem          `json:"schedulers"`
}
