package responses

import (
	"time"

	"github.com/stablecog/go-apps/database/ent/generation"
)

// API response for retrieving user generations
type UserGenerationsResponse struct {
	Width             int32             `json:"width"`
	Height            int32             `json:"height"`
	NumInferenceSteps int32             `json:"num_inference_steps"`
	GuidanceScale     float32           `json:"guidance_scale"`
	Prompt            string            `json:"prompt"`
	NegativePrompt    string            `json:"negative_prompt,omitempty"`
	Model             string            `json:"model"`
	Scheduler         string            `json:"scheduler"`
	Seed              int               `json:"seed"`
	Outputs           []string          `json:"outputs"`
	Status            generation.Status `json:"status"`
	CreatedAt         time.Time         `json:"created_at"`
	StartedAt         *time.Time        `json:"started_at,omitempty"`
	CompletedAt       *time.Time        `json:"completed_at,omitempty"`
}
