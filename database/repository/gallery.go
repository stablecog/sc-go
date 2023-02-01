package repository

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/stablecog/go-apps/database/ent/generation"
	"k8s.io/klog/v2"
)

// Submit a generation to gallery if not already submitted
var ErrAlreadySubmitted = fmt.Errorf("Already submitted to gallery")

func (r *Repository) SubmitGenerationToGalleryForUser(id uuid.UUID, userID uuid.UUID) error {
	g, err := r.DB.Generation.Query().Where(generation.IDEQ(id), generation.UserIDEQ(userID)).First(r.Ctx)
	if err != nil {
		klog.Errorf("Error getting generation %s: %v", id, err)
		return err
	}
	if g.GalleryStatus == generation.GalleryStatusSubmitted || g.GalleryStatus == generation.GalleryStatusAccepted || g.GalleryStatus == generation.GalleryStatusRejected {
		return ErrAlreadySubmitted
	}
	// Update status
	_, err = r.DB.Generation.UpdateOneID(id).SetGalleryStatus(generation.GalleryStatusSubmitted).Save(r.Ctx)
	if err != nil {
		klog.Errorf("Error submitting generation to gallery %s: %v", id, err)
	}
	return err
}

// Approve or reject a generation
func (r *Repository) ApproveOrRejectGeneration(id uuid.UUID, approved bool) error {
	var status generation.GalleryStatus
	if approved {
		status = generation.GalleryStatusAccepted
	} else {
		status = generation.GalleryStatusRejected
	}
	_, err := r.DB.Generation.UpdateOneID(id).SetGalleryStatus(status).Save(r.Ctx)
	if err != nil {
		return err
	}
	return nil
}

// Delete a generation
func (r *Repository) DeleteGeneration(id uuid.UUID) error {
	err := r.DB.Generation.DeleteOneID(id).Exec(r.Ctx)
	if err != nil {
		return err
	}
	return nil
}
