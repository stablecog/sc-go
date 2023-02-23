package repository

import (
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent/generation"
	"github.com/stablecog/sc-go/database/ent/generationmodel"
	"github.com/stablecog/sc-go/database/ent/generationoutput"
	"github.com/stablecog/sc-go/database/ent/negativeprompt"
	"github.com/stablecog/sc-go/database/ent/prompt"
	"github.com/stablecog/sc-go/database/ent/scheduler"
)

// Retrieves data for meilisearch
func (r *Repository) RetrieveGalleryData(limit int, updatedAtGT *time.Time) ([]GalleryData, error) {
	if limit <= 0 {
		limit = 100
	}
	var res []GalleryData
	query := r.DB.GenerationOutput.Query().Select(generationoutput.FieldID, generationoutput.FieldImagePath, generationoutput.FieldUpscaledImagePath, generationoutput.FieldCreatedAt, generationoutput.FieldUpdatedAt).
		Where(generationoutput.GalleryStatusEQ(generationoutput.GalleryStatusApproved))
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
				sql.As(pt.C(prompt.FieldText), "prompt_text"), sql.As(pt.C(prompt.FieldID), "prompt_id"), sql.As(npt.C(negativeprompt.FieldText), "negative_prompt_text"),
				sql.As(npt.C(negativeprompt.FieldID), "negative_prompt_id"), sql.As(g.C(generation.FieldUserID), "user_id"))
			s.OrderBy(sql.Desc(s.C(generationoutput.FieldCreatedAt)), sql.Desc(g.C(generation.FieldCreatedAt)))
		}).Scan(r.Ctx, &res)
	return res, err
}

type GalleryData struct {
	ID                 uuid.UUID  `json:"id,omitempty" sql:"id"`
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
	Seed               int        `json:"seed,omitempty" sql:"generation_seed"`
	ModelID            uuid.UUID  `json:"model_id" sql:"model_id"`
	SchedulerID        uuid.UUID  `json:"scheduler_id" sql:"scheduler_id"`
	PromptText         string     `json:"prompt_text" sql:"prompt_text"`
	PromptID           uuid.UUID  `json:"prompt_id" sql:"prompt_id"`
	NegativePromptText string     `json:"negative_prompt_text,omitempty" sql:"negative_prompt_text"`
	NegativePromptID   *uuid.UUID `json:"negative_prompt_id,omitempty" sql:"negative_prompt_id"`
	UserID             *uuid.UUID `json:"user_id,omitempty" sql:"user_id"`
}
