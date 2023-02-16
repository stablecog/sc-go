package responses

import "time"

type LivePageMessageType string

const (
	LivePageMessageGeneration LivePageMessageType = "live_generation"
	LivePageMessageUpscale    LivePageMessageType = "live_upscale"
)

type LivePageStatus string

const (
	LivePageQueued     LivePageStatus = "queued"
	LivePageProcessing LivePageStatus = "processing"
	LivePageSucceeded  LivePageStatus = "succeeded"
	LivePageFailed     LivePageStatus = "failed"
)

type LivePageMessage struct {
	Type        LivePageMessageType `json:"type"`
	ID          string              `json:"id"`
	CountryCode string              `json:"country_code"`
	Status      LivePageStatus      `json:"status"`
	Width       int32               `json:"width"`
	Height      int32               `json:"height"`
	CreatedAt   time.Time           `json:"created_at"`
	StartedAt   *time.Time          `json:"started_at,omitempty"`
	CompletedAt *time.Time          `json:"completed_at,omitempty"`
}
