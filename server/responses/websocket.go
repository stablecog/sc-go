package responses

import (
	"github.com/stablecog/go-apps/database/ent"
	"github.com/stablecog/go-apps/server/requests"
)

type WebsocketStatusUpdateResponse struct {
	Status  requests.WebhookStatus  `json:"status"`
	Id      string                  `json:"id"`
	Outputs []*ent.GenerationOutput `json:"outputs,omitempty"`
}
