package requests

import "github.com/google/uuid"

type LikeUnlikeAction string

const (
	LikeAction   LikeUnlikeAction = "like"
	UnlikeAction LikeUnlikeAction = "unlike"
)

type LikeUnlikeActionRequest struct {
	GenerationOutputIDs []uuid.UUID      `json:"generation_output_ids"`
	Action              LikeUnlikeAction `json:"action"`
}
