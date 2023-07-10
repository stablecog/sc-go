package shared

import (
	"time"

	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/enttypes"
)

type LivePageStatus string

const (
	LivePageQueued     LivePageStatus = "queued"
	LivePageProcessing LivePageStatus = "processing"
	LivePageSucceeded  LivePageStatus = "succeeded"
	LivePageFailed     LivePageStatus = "failed"
)

type LivePageMessage struct {
	ProcessType      ProcessType         `json:"process_type"`
	ID               string              `json:"id"`
	CountryCode      string              `json:"country_code"`
	Status           LivePageStatus      `json:"status"`
	FailureReason    string              `json:"failure_reason,omitempty"`
	Width            *int32              `json:"width,omitempty"`
	Height           *int32              `json:"height,omitempty"`
	TargetNumOutputs int32               `json:"target_num_outputs"`
	ActualNumOutputs int                 `json:"actual_num_outputs"`
	NSFWCount        *int32              `json:"nsfw_count,omitempty"`
	CreatedAt        time.Time           `json:"created_at"`
	StartedAt        *time.Time          `json:"started_at,omitempty"`
	CompletedAt      *time.Time          `json:"completed_at,omitempty"`
	ProductID        *string             `json:"product_id,omitempty"`
	SystemGenerated  bool                `json:"system_generated"`
	Source           enttypes.SourceType `json:"source,omitempty"`
	Temperature      *float32            `json:"temperature,omitempty"`
	RemoveSilence    *bool               `json:"remove_silence,omitempty"`
	DenoiseAudio     *bool               `json:"denoise_audio,omitempty"`
	SpeakerID        *uuid.UUID          `json:"speaker_id,omitempty"`
}
