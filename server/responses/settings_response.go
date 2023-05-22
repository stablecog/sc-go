package responses

import "github.com/google/uuid"

type SettingsResponseItem struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	Default *bool     `json:"default,omitempty"`
	Active  *bool     `json:"active,omitempty"`
}

type SettingsResponse struct {
	GenerationModels []SettingsResponseItem `json:"generation_models"`
	UpscaleModels    []SettingsResponseItem `json:"upscale_models"`
	Schedulers       []SettingsResponseItem `json:"schedulers"`
}
