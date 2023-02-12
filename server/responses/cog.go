package responses

import "github.com/stablecog/go-apps/server/requests"

// Messages sent from the cog to our application

type CogTaskStatus string

const (
	CogSucceeded  CogTaskStatus = "succeeded"
	CogFailed     CogTaskStatus = "failed"
	CogProcessing CogTaskStatus = "processing"
)

// Msg from cog to redis
type CogStatusUpdate struct {
	Webhook   string                  `json:"webhook"`
	Input     requests.BaseCogRequest `json:"input"`
	Status    CogTaskStatus           `json:"status"`
	Error     string                  `json:"error"`
	Outputs   []string                `json:"outputs"`
	NSFWCount int32                   `json:"nsfw_count"`
}
