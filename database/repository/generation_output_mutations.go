package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/ent/generation"
	"github.com/stablecog/sc-go/database/ent/generationoutput"
	"github.com/stablecog/sc-go/server/requests"
)

// Marks generations for deletions by setting deleted_at
func (r *Repository) MarkGenerationOutputsForDeletion(generationOutputIDs []uuid.UUID) (int, error) {
	var deleted int
	var err error
	// Start transaction
	if err := r.WithTx(func(tx *ent.Tx) error {
		db := tx.Client()
		deleted, err = db.GenerationOutput.Update().Where(generationoutput.IDIn(generationOutputIDs...)).SetDeletedAt(time.Now()).Save(r.Ctx)
		if err != nil {
			return err
		}
		_, err := r.MarkUpscaleOutputForDeletionBasedOnGenerationOutputIDs(generationOutputIDs, db)
		if err != nil {
			return err
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
	outputs, err := r.DB.Generation.Query().Select().Where(generation.UserIDEQ(userID)).QueryGenerationOutputs().Select(generationoutput.FieldID).Where(generationoutput.IDIn(generationOutputIDs...)).All(r.Ctx)
	if err != nil {
		return 0, err
	}

	// get IDs only
	var idsOnly []uuid.UUID
	for _, output := range outputs {
		idsOnly = append(idsOnly, output.ID)
	}

	removeFavorites := false
	if action == requests.RemoveFavoriteAction {
		removeFavorites = true
	}

	// Execute delete
	return r.DB.GenerationOutput.Update().SetIsFavorited(!removeFavorites).Where(generationoutput.IDIn(idsOnly...)).Save(r.Ctx)
}
