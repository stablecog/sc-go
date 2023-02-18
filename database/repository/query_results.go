package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent/generationoutput"
)

// Query results
// For if we want to return a different format than the ent generated models
// Or where we have a join/subquery

type GenerationOutput struct {
	ID               uuid.UUID                      `json:"id"`
	ImageUrl         string                         `json:"image_url"`
	UpscaledImageUrl string                         `json:"upscaled_image_url,omitempty"`
	GalleryStatus    generationoutput.GalleryStatus `json:"gallery_status,omitempty"`
}

// Paginated meta for querying generations
type GenerationQueryWithOutputsMeta struct {
	Total   *int                                        `json:"total_count,omitempty"`
	Outputs []GenerationQueryWithOutputsResultFormatted `json:"outputs"`
	Next    *time.Time                                  `json:"next,omitempty"`
}

type GenerationQueryWithOutputsData struct {
	ID             uuid.UUID          `json:"id" sql:"id"`
	Height         int32              `json:"height" sql:"height"`
	Width          int32              `json:"width" sql:"width"`
	InferenceSteps int32              `json:"inference_steps" sql:"inference_steps"`
	Seed           int                `json:"seed" sql:"seed"`
	Status         string             `json:"status" sql:"status"`
	GuidanceScale  float32            `json:"guidance_scale" sql:"guidance_scale"`
	SchedulerID    uuid.UUID          `json:"scheduler_id" sql:"scheduler_id"`
	ModelID        uuid.UUID          `json:"model_id" sql:"model_id"`
	CreatedAt      time.Time          `json:"created_at" sql:"created_at"`
	StartedAt      *time.Time         `json:"started_at,omitempty" sql:"started_at"`
	CompletedAt    *time.Time         `json:"completed_at,omitempty" sql:"completed_at"`
	NegativePrompt string             `json:"negative_prompt" sql:"negative_prompt_text"`
	Prompt         string             `json:"prompt" sql:"prompt_text"`
	Outputs        []GenerationOutput `json:"outputs"`
}

type GenerationQueryWithOutputsResult struct {
	OutputID         *uuid.UUID                     `json:"output_id,omitempty" sql:"output_id"`
	ImageUrl         string                         `json:"image_url,omitempty" sql:"image_path"`
	UpscaledImageUrl string                         `json:"upscaled_image_url,omitempty" sql:"upscaled_image_path"`
	GalleryStatus    generationoutput.GalleryStatus `json:"gallery_status,omitempty" sql:"output_gallery_status"`
	DeletedAt        *time.Time                     `json:"deleted_at,omitempty" sql:"deleted_at"`
	GenerationQueryWithOutputsData
}

type GenerationQueryWithOutputsResultFormatted struct {
	GenerationOutput
	Generation GenerationQueryWithOutputsData `json:"generation"`
}
