package repository

import (
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/ent/generation"
	"github.com/stablecog/sc-go/database/ent/generationoutput"
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

// Take user ID and limit, return array of output_ids they have liked in descending order of created_at
func (r *Repository) GetGenerationOutputIDsLikedByUser(userID uuid.UUID, limit int) ([]uuid.UUID, error) {
	likedByUser, err := r.DB.GenerationOutputLike.Query().Where(generationoutputlike.LikedByUserID(userID)).Order(ent.Desc(generationoutputlike.FieldCreatedAt)).Limit(limit).All(r.Ctx)
	if err != nil {
		return nil, err
	}
	outputIDs := make([]uuid.UUID, len(likedByUser))
	for i, like := range likedByUser {
		outputIDs[i] = like.OutputID
	}
	return outputIDs, nil
}

// Get how many likes a user has received for their generation outputs
func (r *Repository) GetGenerationOutputLikeCountForUser(userID uuid.UUID) (int, error) {
	var v []struct {
		Sum *int
	}
	err := r.DB.GenerationOutput.Query().
		Where(generationoutput.HasGenerationsWith(generation.UserID(userID))).
		Aggregate(
			ent.Sum(generationoutput.FieldLikeCount),
		).Scan(r.Ctx, &v)
	if err != nil {
		return 0, err
	}
	if len(v) == 0 || v[0].Sum == nil {
		return 0, nil
	}
	return *v[0].Sum, nil
}

// Get how many likes a user has received for their generation outputs excluding self
func (r *Repository) GetGenerationOutputLikeCountForUserExcludingSelfLikes(userID uuid.UUID) (int, error) {
	return r.DB.GenerationOutputLike.Query().Where(generationoutputlike.LikedByUserIDNEQ(userID),
		generationoutputlike.HasGenerationOutputsWith(generationoutput.HasGenerationsWith(generation.UserID(userID)))).Count(r.Ctx)
}
