package responses

import (
	"github.com/stablecog/go-apps/database/ent"
)

type WebsocketStatusUpdateResponse struct {
	Status    CogTaskStatus           `json:"status"`
	Id        string                  `json:"id"`
	Error     string                  `json:"error,omitempty"`
	NSFWCount int                     `json:"nsfw_count,omitempty"`
	Outputs   []*ent.GenerationOutput `json:"outputs,omitempty"`
}
