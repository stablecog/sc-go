package requests

import "github.com/google/uuid"

// Request for submitting outputs to gallery
type SubmitGalleryRequest struct {
	GenerationOutputIDs []uuid.UUID `json:"generation_output_ids"`
}
