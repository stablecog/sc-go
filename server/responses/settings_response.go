package responses

import "github.com/google/uuid"

type SettingsResponseItem struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type SettingsResponse struct {
	GenerationModels []SettingsResponseItem `json:"generation_models"`
	UpscaleModels    []SettingsResponseItem `json:"upscale_models"`
	Schedulers       []SettingsResponseItem `json:"schedulers"`
}
