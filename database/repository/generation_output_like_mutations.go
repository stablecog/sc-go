package repository

import (
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/ent/generationoutputlike"
	"github.com/stablecog/sc-go/server/requests"
)

// Inserts like
func (r *Repository) SetOutputsLikedForUser(generationOutputIDs []uuid.UUID, userID uuid.UUID, action requests.LikeUnlikeAction) error {
	removeLikes := false
	if action == requests.UnlikeAction {
		removeLikes = true
	}

	// Execute in TX
	if err := r.WithTx(func(tx *ent.Tx) error {
		// Inserting likes
		if !removeLikes {
			for _, id := range generationOutputIDs {
				if !removeLikes {
					err := tx.GenerationOutputLike.Create().SetOutputID(id).SetLikedByUserID(userID).OnConflict().DoNothing().Exec(r.Ctx)
					if err != nil {
						return err
					}
				}
			}
		} else {
			// Unliking
			_, err := tx.GenerationOutputLike.Delete().Where(generationoutputlike.OutputIDIn(generationOutputIDs...), generationoutputlike.LikedByUserID(userID)).Exec(r.Ctx)
			if err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return err
	}
	return nil
}
