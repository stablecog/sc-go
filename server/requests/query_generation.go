// * Requests initiated by logged in users
package requests

import (
	"time"

	"github.com/google/uuid"
)

// For filtering user's generations
type SortOrder string

const (
	SortOrderAscending  SortOrder = "asc"
	SortOrderDescending SortOrder = "desc"
)

type UpscaleStatus string

const (
	// Include upscaled and not upscaled
	UpscaleStatusAny UpscaleStatus = "any"
	// Only upscaled
	UpscaleStatusOnly UpscaleStatus = "only"
	// Not upscaled
	UpscaleStatusNot UpscaleStatus = "not"
)

type QueryGenerationFilters struct {
	ModelIDs          []uuid.UUID   `json:"model_ids"`
	SchedulerIDs      []uuid.UUID   `json:"scheduler_ids"`
	MinHeight         int32         `json:"min_height"`
	MaxHeight         int32         `json:"max_height"`
	MinWidth          int32         `json:"min_width"`
	MaxWidth          int32         `json:"max_width"`
	Widths            []int32       `json:"widths"`
	Heights           []int32       `json:"heights"`
	MaxInferenceSteps int32         `json:"max_inference_steps"`
	MinInferenceSteps int32         `json:"min_inference_steps"`
	InferenceSteps    []int32       `json:"inference_steps"`
	MaxGuidanceScale  float32       `json:"max_guidance_scale"`
	MinGuidanceScale  float32       `json:"min_guidance_scale"`
	GuidanceScales    []float32     `json:"guidance_scales"`
	UpscaleStatus     UpscaleStatus `json:"upscale_status"`
	Order             SortOrder     `json:"order"`
	StartDt           *time.Time    `json:"start_dt"`
	EndDt             *time.Time    `json:"end_dt"`
}
