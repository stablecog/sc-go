package responses

// API generate simply returns a UUID to track the request to our compute while its in flight
type TaskQueuedResponse struct {
	ID               string `json:"id"`
	UIId             string `json:"ui_id,omitempty"`
	RemainingCredits int    `json:"total_remaining_credits"`
}
