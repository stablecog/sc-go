package requests

import "github.com/google/uuid"

type DeleteGenerationRequest struct {
	GenerationOutputIDs []uuid.UUID `json:"generation_output_ids"`
}
