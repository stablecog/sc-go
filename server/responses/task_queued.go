package responses

// For queue details
type TaskQueueInfo struct {
	ID       string `json:"id"`
	Position int    `json:"position"`
	Size     int    `json:"size"`
}

// When UI queues a request
type TaskQueuedResponse struct {
	ID               string         `json:"id"`
	UIId             string         `json:"ui_id,omitempty"`
	RemainingCredits int            `json:"total_remaining_credits"`
	WasAutoSubmitted bool           `json:"was_auto_submitted,omitempty"`
	IsPublic         bool           `json:"is_public,omitempty"`
	QueueInfo        *TaskQueueInfo `json:"queue_info,omitempty"`
}
