package requests

import "github.com/google/uuid"

type FavoriteGenerationRequest struct {
	GenerationOutputIDs []uuid.UUID `json:"generation_output_ids"`
	RemoveFavorites     bool        `json:"remove_favorites"`
}
