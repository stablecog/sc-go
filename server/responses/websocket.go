package responses

import (
	"github.com/stablecog/go-apps/database/ent"
)

type WebsocketStatusUpdateResponse struct {
	Status  CogTaskStatus           `json:"status"`
	Id      string                  `json:"id"`
	Error   string                  `json:"error,omitempty"`
	Outputs []*ent.GenerationOutput `json:"outputs,omitempty"`
}
