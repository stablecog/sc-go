package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/ent/generationoutput"
	"github.com/stablecog/sc-go/database/ent/upscale"
	"github.com/stablecog/sc-go/database/ent/upscaleoutput"
)

// Marks upscales for deletions by setting deleted_at
func (r *Repository) MarkUpscaleOutputForDeletion(upscaleOutputIDs []uuid.UUID) (int, error) {
	deleted, err := r.DB.UpscaleOutput.Update().Where(upscaleoutput.IDIn(upscaleOutputIDs...)).SetDeletedAt(time.Now()).Save(r.Ctx)
	if err != nil {
		return 0, err
	}
	return deleted, nil
}

// Marks upscales for deletions by setting deleted_at, only if they belong to the user with ID userID
func (r *Repository) MarkUpscaleOutputForDeletionForUser(upscaleOutputIDs []uuid.UUID, userID uuid.UUID) (int, error) {
	// Get outputs belonging to this user
	outputs, err := r.DB.Upscale.Query().Select().Where(upscale.UserIDEQ(userID)).QueryUpscaleOutputs().Select(upscaleoutput.FieldID).Where(upscaleoutput.IDIn(upscaleOutputIDs...)).All(r.Ctx)
	if err != nil {
		return 0, err
	}

	// Filter out outputs that don't belong to the user
	var userUpscaleOutputIDs []uuid.UUID
	for _, output := range outputs {
		userUpscaleOutputIDs = append(userUpscaleOutputIDs, output.ID)
	}

	// Execute delete
	return r.MarkUpscaleOutputForDeletion(userUpscaleOutputIDs)
}

// Marks upscales for deletion if they are associated with the given generation outputs
func (r *Repository) MarkUpscaleOutputForDeletionBasedOnGenerationOutputIDs(generationOutputIDs []uuid.UUID, DB *ent.Client) (int, error) {
	if DB == nil {
		DB = r.DB
	}

	// Get outputs
	outputs, err := DB.GenerationOutput.Query().Select(generationoutput.FieldUpscaledImagePath).Where(generationoutput.IDIn(generationOutputIDs...), generationoutput.UpscaledImagePathNotNil()).All(r.Ctx)
	if err != nil && !ent.IsNotFound(err) {
		return 0, err
	}

	if len(outputs) == 0 {
		return 0, nil
	}

	var paths []string
	for _, output := range outputs {
		paths = append(paths, *output.UpscaledImagePath)
	}

	// Mark upscale outputs for deletion
	return DB.UpscaleOutput.Update().Where(upscaleoutput.ImagePathIn(paths...)).SetDeletedAt(time.Now()).Save(r.Ctx)
}
