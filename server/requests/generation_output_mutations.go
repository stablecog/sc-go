package requests

import "github.com/google/uuid"

type GenerationDeleteRequest struct {
	GenerationOutputIDs []uuid.UUID `json:"generation_output_ids"`
}
