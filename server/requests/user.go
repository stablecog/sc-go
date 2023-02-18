// * Requests initiated by logged in users
package requests

import (
	"time"

	"github.com/google/uuid"
)

// Request for submitting outputs to gallery
type GenerateSubmitToGalleryRequestBody struct {
	GenerationOutputIDs []uuid.UUID `json:"generation_output_ids"`
}

// For filtering user's generations
type UserGenerationQueryOrder string

const (
	UserGenerationQueryOrderAscending  UserGenerationQueryOrder = "asc"
	UserGenerationQueryOrderDescending UserGenerationQueryOrder = "desc"
)

type UserGenerationQueryUpscaleStatus string

const (
	// Include upscaled and not upscaled
	UserGenerationQueryUpscaleStatusAny UserGenerationQueryUpscaleStatus = "any"
	// Only upscaled
	UserGenerationQueryUpscaleStatusOnly UserGenerationQueryUpscaleStatus = "only"
	// Not upscaled
	UserGenerationQueryUpscaleStatusNot UserGenerationQueryUpscaleStatus = "not"
)

type UserGenerationFilters struct {
	ModelIDs          []uuid.UUID                      `json:"model_ids"`
	SchedulerIDs      []uuid.UUID                      `json:"scheduler_ids"`
	MinHeight         int32                            `json:"min_height"`
	MaxHeight         int32                            `json:"max_height"`
	MinWidth          int32                            `json:"min_width"`
	MaxWidth          int32                            `json:"max_width"`
	Widths            []int32                          `json:"widths"`
	Heights           []int32                          `json:"heights"`
	MaxInferenceSteps int32                            `json:"max_inference_steps"`
	MinInferenceSteps int32                            `json:"min_inference_steps"`
	InferenceSteps    []int32                          `json:"inference_steps"`
	MaxGuidanceScale  float32                          `json:"max_guidance_scale"`
	MinGuidanceScale  float32                          `json:"min_guidance_scale"`
	GuidanceScales    []float32                        `json:"guidance_scales"`
	UpscaleStatus     UserGenerationQueryUpscaleStatus `json:"upscale_status"`
	Order             UserGenerationQueryOrder         `json:"order"`
	StartDt           *time.Time                       `json:"start_dt"`
	EndDt             *time.Time                       `json:"end_dt"`
}
