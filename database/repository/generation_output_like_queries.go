package repository

import (
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent/generationoutputlike"
)

// Take user id and generation output ID array, return a map containing uuids for faster lookup.
func (r *Repository) GetGenerationOutputsLikedByUser(userID uuid.UUID, generationOutputIDs []uuid.UUID) (map[uuid.UUID]struct{}, error) {
	likedByUser, err := r.DB.GenerationOutputLike.Query().Where(generationoutputlike.OutputIDIn(generationOutputIDs...), generationoutputlike.LikedByUserID(userID)).All(r.Ctx)
	if err != nil {
		return nil, err
	}
	liked := make(map[uuid.UUID]struct{})
	for _, like := range likedByUser {
		liked[like.OutputID] = struct{}{}
	}
	return liked, nil
}
