package responses

import "github.com/stablecog/sc-go/shared"

type SSEStatusUpdateResponse struct {
	Status      CogTaskStatus              `json:"status"`
	ProcessType shared.ProcessType         `json:"process_type"`
	Id          string                     `json:"id"`
	StreamId    string                     `json:"stream_id"`
	Error       string                     `json:"error,omitempty"`
	NSFWCount   int32                      `json:"nsfw_count,omitempty"`
	Outputs     []GenerationOutputResponse `json:"outputs,omitempty"`
}
