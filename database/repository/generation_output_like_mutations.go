package repository

import (
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/server/requests"
)

// Inserts like
func (r *Repository) SetOutputsLikedForUser(generationOutputIDs []uuid.UUID, userID uuid.UUID, action requests.LikeUnlikeAction) (int, error) {
	return 0, nil
	// // Get outputs belonging to this user
	// outputs, err := r.DB.Generation.Query().Select().Where(generation.UserIDEQ(userID)).QueryGenerationOutputs().Select(generationoutput.FieldID, generationoutput.FieldHasEmbeddings).Where(generationoutput.IDIn(generationOutputIDs...)).All(r.Ctx)
	// if err != nil {
	// 	return 0, err
	// }

	// removeLikes := false
	// if action == requests.UnlikeAction {
	// 	removeLikes = true
	// }

	// // get IDs only
	// var idsOnly []uuid.UUID
	// // Separate array for qdrant
	// var qdrantIds []uuid.UUID
	// qdrantPayload := map[string]interface{}{
	// 	"is_favorited": !removeFavorites,
	// }
	// for _, output := range outputs {
	// 	idsOnly = append(idsOnly, output.ID)
	// 	if output.HasEmbeddings {
	// 		qdrantIds = append(qdrantIds, output.ID)
	// 	}
	// }

	// // Execute in TX
	// var updated int
	// if err := r.WithTx(func(tx *ent.Tx) error {
	// 	ud, err := r.DB.GenerationOutput.Update().SetIsFavorited(!removeFavorites).Where(generationoutput.IDIn(idsOnly...)).Save(r.Ctx)
	// 	if err != nil {
	// 		log.Error("Error updating generation outputs", "err", err)
	// 		return err
	// 	}
	// 	updated = ud
	// 	// Update qdrant
	// 	if r.Qdrant != nil && len(qdrantIds) > 0 {
	// 		err = r.Qdrant.SetPayload(qdrantPayload, qdrantIds, false)
	// 		if err != nil {
	// 			log.Error("Error updating qdrant", "err", err)
	// 			return err
	// 		}
	// 	}
	// 	return nil
	// }); err != nil {
	// 	return 0, err
	// }
	// return updated, nil
}
