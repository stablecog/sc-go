package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/ent/generation"
	"github.com/stablecog/sc-go/database/ent/generationoutput"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/server/requests"
)

// Marks generations for deletions by setting deleted_at
func (r *Repository) MarkGenerationOutputsForDeletion(generationOutputIDs []uuid.UUID) (int, error) {
	var deleted int
	var err error
	deletedAt := time.Now()
	// qdrant payload
	qdrantPayload := map[string]interface{}{
		"deleted_at": deletedAt.Unix(),
	}
	// Figure out which IDs have embeddings
	embeddingOutputs, err := r.DB.GenerationOutput.Query().Select(generationoutput.FieldID, generationoutput.FieldHasEmbeddings).Where(generationoutput.IDIn(generationOutputIDs...)).All(r.Ctx)
	if err != nil {
		log.Error("Error getting generation outputs has_embeddings", "err", err)
		return 0, err
	}
	// Separate array for qdrant
	var qdrantIds []uuid.UUID
	for _, output := range embeddingOutputs {
		if output.HasEmbeddings {
			qdrantIds = append(qdrantIds, output.ID)
		}
	}
	// Start transaction
	if err := r.WithTx(func(tx *ent.Tx) error {
		db := tx.Client()
		deleted, err = db.GenerationOutput.Update().Where(generationoutput.IDIn(generationOutputIDs...)).SetDeletedAt(deletedAt).Save(r.Ctx)
		if err != nil {
			return err
		}
		_, err := r.MarkUpscaleOutputForDeletionBasedOnGenerationOutputIDs(generationOutputIDs, db)
		if err != nil {
			return err
		}
		// Update qdrant
		if r.Qdrant != nil && len(qdrantIds) > 0 {
			err = r.Qdrant.SetPayload(qdrantPayload, qdrantIds, false)
			if err != nil {
				log.Error("Error updating qdrant deleted_at", "err", err)
				return err
			}
		}
		return nil
	}); err != nil {
		return 0, err
	}
	return deleted, nil
}

// Marks generations for deletions by setting deleted_at, only if they belong to the user with ID userID
func (r *Repository) MarkGenerationOutputsForDeletionForUser(generationOutputIDs []uuid.UUID, userID uuid.UUID) (int, error) {
	// Get outputs belonging to this user
	outputs, err := r.DB.Generation.Query().Select().Where(generation.UserIDEQ(userID)).QueryGenerationOutputs().Select(generationoutput.FieldID).Where(generationoutput.IDIn(generationOutputIDs...)).All(r.Ctx)
	if err != nil {
		return 0, err
	}

	// Filter out outputs that don't belong to the user
	var userGenerationOutputIds []uuid.UUID
	for _, output := range outputs {
		userGenerationOutputIds = append(userGenerationOutputIds, output.ID)
	}

	// Execute delete
	return r.MarkGenerationOutputsForDeletion(userGenerationOutputIds)
}

// Marks generations for deletions by setting deleted_at, only if they belong to the user with ID userID
func (r *Repository) SetFavoriteGenerationOutputsForUser(generationOutputIDs []uuid.UUID, userID uuid.UUID, action requests.FavoriteAction) (int, error) {
	// Get outputs belonging to this user
	outputs, err := r.DB.Generation.Query().Select().Where(generation.UserIDEQ(userID)).QueryGenerationOutputs().Select(generationoutput.FieldID, generationoutput.FieldHasEmbeddings).Where(generationoutput.IDIn(generationOutputIDs...)).All(r.Ctx)
	if err != nil {
		return 0, err
	}

	removeFavorites := false
	if action == requests.RemoveFavoriteAction {
		removeFavorites = true
	}

	// get IDs only
	var idsOnly []uuid.UUID
	// Separate array for qdrant
	var qdrantIds []uuid.UUID
	qdrantPayload := map[string]interface{}{
		"is_favorited": !removeFavorites,
	}
	for _, output := range outputs {
		idsOnly = append(idsOnly, output.ID)
		if output.HasEmbeddings {
			qdrantIds = append(qdrantIds, output.ID)
		}
	}

	// Execute in TX
	var updated int
	if err := r.WithTx(func(tx *ent.Tx) error {
		ud, err := r.DB.GenerationOutput.Update().SetIsFavorited(!removeFavorites).Where(generationoutput.IDIn(idsOnly...)).Save(r.Ctx)
		if err != nil {
			log.Error("Error updating generation outputs", "err", err)
			return err
		}
		updated = ud
		// Update qdrant
		if r.Qdrant != nil && len(qdrantIds) > 0 {
			err = r.Qdrant.SetPayload(qdrantPayload, qdrantIds, false)
			if err != nil {
				log.Error("Error updating qdrant", "err", err)
				return err
			}
		}
		return nil
	}); err != nil {
		return 0, err
	}
	return updated, nil
}
