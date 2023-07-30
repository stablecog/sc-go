package requests

import "github.com/google/uuid"

type DeleteVoiceoverRequest struct {
	OutputIDs []uuid.UUID `json:"output_ids"`
}
