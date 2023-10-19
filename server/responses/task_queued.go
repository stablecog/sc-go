package responses

import (
	"time"
)

type QueuedItem struct {
	Id        string    `json:"id"`
	Priority  int       `json:"priority"`
	CreatedAt time.Time `json:"created_at"`
}

// When UI queues a request
type TaskQueuedResponse struct {
	ID               string        `json:"id"`
	UIId             string        `json:"ui_id,omitempty"`
	RemainingCredits int           `json:"total_remaining_credits"`
	WasAutoSubmitted bool          `json:"was_auto_submitted,omitempty"`
	IsPublic         bool          `json:"is_public,omitempty"`
	QueuedId         string        `json:"queued_id,omitempty"`
	QueueItems       []*QueuedItem `json:"queue_items,omitempty"`
}
