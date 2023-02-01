package requests

import "github.com/google/uuid"

type GenerateSubmitToGalleryRequestBody struct {
	GenerationID uuid.UUID `json:"generation_id"`
}
