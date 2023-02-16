package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent/generation"
	"github.com/stablecog/sc-go/database/ent/generationoutput"
)

// Marks generations for deletions by setting deleted_at
func (r *Repository) MarkGenerationOutputsForDeletion(generationOutputIDs []uuid.UUID) (int, error) {
	deleted, err := r.DB.GenerationOutput.Update().Where(generationoutput.IDIn(generationOutputIDs...)).SetDeletedAt(time.Now()).Save(r.Ctx)
	if err != nil {
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
