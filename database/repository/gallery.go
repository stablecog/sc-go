package repository

import (
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent/generation"
	"github.com/stablecog/sc-go/database/ent/generationmodel"
	"github.com/stablecog/sc-go/database/ent/generationoutput"
	"github.com/stablecog/sc-go/database/ent/negativeprompt"
	"github.com/stablecog/sc-go/database/ent/prompt"
	"github.com/stablecog/sc-go/database/ent/scheduler"
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

// Retrieves data for meilisearch
func (r *Repository) RetrieveGalleryData(limit int, updatedAtGT *time.Time) ([]GalleryData, error) {
	if limit <= 0 {
		limit = 100
	}
	var res []GalleryData
	query := r.DB.GenerationOutput.Query().Select(generationoutput.FieldID, generationoutput.FieldImagePath, generationoutput.FieldUpscaledImagePath, generationoutput.FieldCreatedAt, generationoutput.FieldUpdatedAt).
		Where(generationoutput.GalleryStatusEQ(generationoutput.GalleryStatusAccepted))
	if updatedAtGT != nil {
		query = query.Where(generationoutput.UpdatedAtGT(*updatedAtGT))
	}
	err := query.Limit(limit).
		Modify(func(s *sql.Selector) {
			g := sql.Table(generation.Table)
			pt := sql.Table(prompt.Table)
			npt := sql.Table(negativeprompt.Table)
			mt := sql.Table(generationmodel.Table)
			st := sql.Table(scheduler.Table)
			s.LeftJoin(g).On(
				s.C(generationoutput.FieldGenerationID), g.C(generation.FieldID),
			).LeftJoin(pt).On(
				g.C(generation.FieldPromptID), pt.C(prompt.FieldID),
			).LeftJoin(npt).On(
				g.C(generation.FieldNegativePromptID), npt.C(negativeprompt.FieldID),
			).LeftJoin(mt).On(
				g.C(generation.FieldModelID), mt.C(generationmodel.FieldID),
			).LeftJoin(st).On(
				g.C(generation.FieldSchedulerID), st.C(scheduler.FieldID),
			).AppendSelect(sql.As(g.C(generation.FieldWidth), "generation_width"), sql.As(g.C(generation.FieldHeight), "generation_height"),
				sql.As(g.C(generation.FieldInferenceSteps), "generation_inference_steps"), sql.As(g.C(generation.FieldGuidanceScale), "generation_guidance_scale"),
				sql.As(g.C(generation.FieldSeed), "generation_seed"), sql.As(mt.C(generationmodel.FieldID), "model_id"), sql.As(st.C(scheduler.FieldID), "scheduler_id"),
				sql.As(pt.C(prompt.FieldText), "prompt_text"), sql.As(npt.C(negativeprompt.FieldText), "negative_prompt_text"), sql.As(g.C(generation.FieldUserID), "user_id"))
			s.OrderBy(sql.Desc(s.C(generationoutput.FieldCreatedAt)), sql.Desc(g.C(generation.FieldCreatedAt)))
		}).Scan(r.Ctx, &res)
	return res, err
}

type GalleryData struct {
	ID                 *uuid.UUID `json:"id,omitempty" sql:"id"`
	ImagePath          string     `json:"image_path,omitempty" sql:"image_path"`
	UpscaledImagePath  string     `json:"upscaled_image_path,omitempty" sql:"upscaled_image_path"`
	ImageURL           string     `json:"image_url"`
	UpscaledImageURL   string     `json:"upscaled_image_url,omitempty"`
	CreatedAt          time.Time  `json:"created_at" sql:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at" sql:"updated_at"`
	Width              int32      `json:"width" sql:"generation_width"`
	Height             int32      `json:"height" sql:"generation_height"`
	InferenceSteps     int32      `json:"inference_steps" sql:"generation_inference_steps"`
	GuidanceScale      float32    `json:"guidance_scale" sql:"generation_guidance_scale"`
	Seed               int        `json:"seed" sql:"generation_seed"`
	ModelID            uuid.UUID  `json:"model_id" sql:"model_id"`
	SchedulerID        uuid.UUID  `json:"scheduler_id" sql:"scheduler_id"`
	PromptText         string     `json:"prompt_text" sql:"prompt_text"`
	NegativePromptText string     `json:"negative_prompt_text,omitempty" sql:"negative_prompt_text"`
	UserID             *uuid.UUID `json:"user_id,omitempty" sql:"user_id"`
}
