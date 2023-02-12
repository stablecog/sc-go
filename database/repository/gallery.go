package repository

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent/generation"
	"github.com/stablecog/sc-go/database/ent/generationoutput"
	"k8s.io/klog/v2"
)

// Submits all generation outputs to gallery for review, from user
// Verifies that all outputs belong to the user
// Only submits generation outputs that are not already submitted, accepted, or rejected
// Returns count of how many were submitted
func (r *Repository) SubmitGenerationOutputsToGalleryForUser(outputIDs []uuid.UUID, userID uuid.UUID) (int, error) {
	// Make sure the user owns all the generations of these outputs
	count, err := r.DB.GenerationOutput.Query().Where(generationoutput.IDIn(outputIDs...)).QueryGenerations().Where(generation.UserIDNEQ(userID)).Count(r.Ctx)
	if err != nil {
		klog.Errorf("Error getting generation count: %v", err)
		return 0, err
	}
	if count > 0 {
		klog.Warningf("User %s tried to submit generation outputs that don't belong to them", userID.String())
		return 0, fmt.Errorf("Not all outputs belong to user")
	}

	updated, err := r.DB.GenerationOutput.Update().
		Where(generationoutput.IDIn(outputIDs...), generationoutput.GalleryStatusNotIn(generationoutput.GalleryStatusAccepted, generationoutput.GalleryStatusRejected, generationoutput.GalleryStatusSubmitted)).
		SetGalleryStatus(generationoutput.GalleryStatusSubmitted).Save(r.Ctx)

	if err != nil {
		klog.Errorf("Error submitting generation outputs to gallery: %v", err)
		return 0, err
	}

	return updated, nil
}

// Approve or reject a generation outputs for gallery
func (r *Repository) ApproveOrRejectGenerationOutputs(outputIDs []uuid.UUID, approved bool) (int, error) {
	var status generationoutput.GalleryStatus
	if approved {
		status = generationoutput.GalleryStatusAccepted
	} else {
		status = generationoutput.GalleryStatusRejected
	}
	updated, err := r.DB.GenerationOutput.Update().Where(generationoutput.IDIn(outputIDs...)).SetGalleryStatus(status).Save(r.Ctx)
	if err != nil {
		return 0, err
	}
	return updated, nil
}

// Delete a generation
func (r *Repository) DeleteGenerations(generationIDs []uuid.UUID) (int, error) {
	deleted, err := r.DB.Generation.Delete().Where(generation.IDIn(generationIDs...)).Exec(r.Ctx)
	if err != nil {
		return 0, err
	}
	return deleted, nil
}
