package responses

import "github.com/stablecog/sc-go/server/requests"

type RunpodStatus string

const (
	RunpodStatusCompleted  RunpodStatus = "COMPLETED"
	RunpodStatusFailed     RunpodStatus = "FAILED"
	RunpodStatusInQueue    RunpodStatus = "IN_QUEUE"
	RunpodStatusInProgress RunpodStatus = "IN_PROGRESS"
)

// Runpod returns {"output": {"input": ..., "output": {"images": []}}} where "images" is a list of image URLs
type RunpodOutputOutput struct {
	Images []string `json:"images"`
}

type RunpodBaseOutput struct {
	Input  requests.BaseCogRequest `json:"input"`
	Output RunpodOutputOutput      `json:"output"`
}

type RunpodOutput struct {
	ID     string           `json:"id"`
	Output RunpodBaseOutput `json:"output"`
	Status RunpodStatus     `json:"status"`
	Error  string           `json:"error,omitempty"`
}
