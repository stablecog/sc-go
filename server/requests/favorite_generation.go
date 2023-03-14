package requests

import "github.com/google/uuid"

type FavoriteAction string

const (
	AddFavoriteAction    FavoriteAction = "add"
	RemoveFavoriteAction FavoriteAction = "remove"
)

type FavoriteGenerationRequest struct {
	GenerationOutputIDs []uuid.UUID    `json:"generation_output_ids"`
	Action              FavoriteAction `json:"action"`
}
