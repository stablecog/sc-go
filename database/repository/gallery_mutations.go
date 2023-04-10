package repository

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/ent/generation"
	"github.com/stablecog/sc-go/database/ent/generationoutput"
	"github.com/stablecog/sc-go/log"
)

// Submits all generation outputs to gallery for review, from user
// Verifies that all outputs belong to the user
// Only submits generation outputs that are not already submitted, accepted, or rejected
// Returns count of how many were submitted
func (r *Repository) SubmitGenerationOutputsToGalleryForUser(outputIDs []uuid.UUID, userID uuid.UUID) (int, error) {
	// Make sure the user owns all the generations of these outputs
	count, err := r.DB.GenerationOutput.Query().Where(generationoutput.IDIn(outputIDs...)).QueryGenerations().Where(generation.UserIDNEQ(userID)).Count(r.Ctx)
	if err != nil {
		log.Error("Error getting generation count", "err", err)
		return 0, err
	}
	if count > 0 {
		log.Warn("MALICIOUS USER tried to submit generation outputs that don't belong to them", "user_id", userID.String())
		return 0, fmt.Errorf("Not all outputs belong to user")
	}

	// Get the IDs to update exactly so we can accurately sync with qdrant
	outputs, err := r.DB.GenerationOutput.Query().Where(generationoutput.IDIn(outputIDs...), generationoutput.GalleryStatusNotIn(generationoutput.GalleryStatusApproved, generationoutput.GalleryStatusRejected, generationoutput.GalleryStatusSubmitted)).All(r.Ctx)
	if err != nil {
		log.Error("Error getting generation outputs SubmitGenerationOutputsToGalleryForUser", "err", err)
		return 0, err
	}

	var ids []uuid.UUID
	var qdrantIds []uuid.UUID
	qdrantPayload := map[string]interface{}{
		"gallery_status": generationoutput.GalleryStatusSubmitted,
	}
	for _, output := range outputs {
		ids = append(ids, output.ID)
		if output.HasEmbeddings {
			qdrantIds = append(qdrantIds, output.ID)
		}
	}

	var updated int
	if err := r.WithTx(func(tx *ent.Tx) error {
		u, err := r.DB.GenerationOutput.Update().
			Where(generationoutput.IDIn(ids...)).
			SetGalleryStatus(generationoutput.GalleryStatusSubmitted).Save(r.Ctx)
		if err != nil {
			log.Error("Error updating generation outputs to gallery", "err", err)
			return err
		}
		updated = u

		if r.Qdrant != nil && len(qdrantIds) > 0 {
			err = r.Qdrant.SetPayload(qdrantPayload, qdrantIds, false)
			if err != nil {
				log.Error("Error updating generation outputs to gallery qdrant", "err", err)
				return err
			}
		}
		return nil
	}); err != nil {
		return 0, err
	}

	return updated, nil
}

// Approve or reject a generation outputs for gallery
func (r *Repository) ApproveOrRejectGenerationOutputs(outputIDs []uuid.UUID, approved bool) (int, error) {
	var status generationoutput.GalleryStatus
	if approved {
		status = generationoutput.GalleryStatusApproved
	} else {
		status = generationoutput.GalleryStatusRejected
	}

	// ! TODO temporary query since some stuff doesnt have embeddings
	outputs, err := r.DB.GenerationOutput.Query().Where(generationoutput.IDIn(outputIDs...)).All(r.Ctx)
	if err != nil {
		log.Error("Error getting generation outputs ApproveOrRejectGenerationOutputs", "err", err)
		return 0, err
	}

	qdrantPayload := map[string]interface{}{
		"gallery_status": status,
	}
	var qdrantIds []uuid.UUID
	for _, o := range outputs {
		if o.HasEmbeddings {
			qdrantIds = append(qdrantIds, o.ID)
		}
	}

	var updated int
	if err := r.WithTx(func(tx *ent.Tx) error {
		u, err := r.DB.GenerationOutput.Update().Where(generationoutput.IDIn(outputIDs...)).SetGalleryStatus(status).Save(r.Ctx)
		if err != nil {
			log.Error("Error updating generation outputs to gallery", "err", err)
			return err
		}
		updated = u
		if r.Qdrant != nil && len(qdrantIds) > 0 {
			err = r.Qdrant.SetPayload(qdrantPayload, qdrantIds, false)
			if err != nil {
				log.Error("Error updating generation outputs to gallery qdrant", "err", err)
				return err
			}
		}
		return nil
	}); err != nil {
		return 0, err
	}

	return updated, nil
}
