package repository

import (
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/ent/generation"
	"github.com/stablecog/sc-go/database/ent/generationmodel"
	"github.com/stablecog/sc-go/database/ent/generationoutput"
	"github.com/stablecog/sc-go/database/ent/negativeprompt"
	"github.com/stablecog/sc-go/database/ent/prompt"
	"github.com/stablecog/sc-go/database/ent/scheduler"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/utils"
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

// Retrieved a single generation output by ID, in GalleryData format
func (r *Repository) RetrieveGalleryDataByID(id uuid.UUID) (*GalleryData, error) {
	output, err := r.DB.GenerationOutput.Query().Where(generationoutput.IDEQ(id), generationoutput.GalleryStatusEQ(generationoutput.GalleryStatusApproved)).WithGenerations(func(gq *ent.GenerationQuery) {
		gq.WithPrompt()
		gq.WithNegativePrompt()
	}).Only(r.Ctx)
	if err != nil {
		return nil, err
	}
	data := GalleryData{
		ID:             output.ID,
		ImagePath:      output.ImagePath,
		ImageURL:       utils.GetURLFromImagePath(output.ImagePath),
		CreatedAt:      output.CreatedAt,
		UpdatedAt:      output.UpdatedAt,
		Width:          output.Edges.Generations.Width,
		Height:         output.Edges.Generations.Height,
		InferenceSteps: output.Edges.Generations.InferenceSteps,
		GuidanceScale:  output.Edges.Generations.GuidanceScale,
		Seed:           output.Edges.Generations.Seed,
		ModelID:        output.Edges.Generations.ModelID,
		SchedulerID:    output.Edges.Generations.SchedulerID,
		PromptID:       output.Edges.Generations.Edges.Prompt.ID,
		PromptText:     output.Edges.Generations.Edges.Prompt.Text,
	}
	if output.Edges.Generations.Edges.NegativePrompt != nil {
		data.NegativePromptID = &output.Edges.Generations.Edges.NegativePrompt.ID
		data.NegativePromptText = output.Edges.Generations.Edges.NegativePrompt.Text
	}
	if output.UpscaledImagePath != nil {
		data.UpscaledImagePath = *output.UpscaledImagePath
		data.UpscaledImageURL = utils.GetURLFromImagePath(*output.UpscaledImagePath)
	}
	return &data, nil
}

// Retrieves data in gallery format given  output IDs
// Returns data, next cursor, error
func (r *Repository) RetrieveMostRecentGalleryData(filters *requests.QueryGenerationFilters, per_page int, cursor *time.Time) ([]GalleryData, *time.Time, error) {
	// Apply filters
	queryG := r.DB.Generation.Query().Where(
		generation.StatusEQ(generation.StatusSucceeded),
	)
	queryG = r.ApplyUserGenerationsFilters(queryG, filters, true)
	query := queryG.QueryGenerationOutputs().Where(
		generationoutput.DeletedAtIsNil(),
	)
	if cursor != nil {
		query = query.Where(generationoutput.CreatedAtLT(*cursor))
	}
	if filters != nil {
		if filters.UpscaleStatus == requests.UpscaleStatusNot {
			query = query.Where(generationoutput.UpscaledImagePathIsNil())
		}
		if filters.UpscaleStatus == requests.UpscaleStatusOnly {
			query = query.Where(generationoutput.UpscaledImagePathNotNil())
		}
		if len(filters.GalleryStatus) > 0 {
			query = query.Where(generationoutput.GalleryStatusIn(filters.GalleryStatus...))
		}
	}
	query = query.WithGenerations(func(s *ent.GenerationQuery) {
		s.WithPrompt()
		s.WithNegativePrompt()
		s.WithGenerationOutputs()
	})

	// Limit
	query = query.Order(ent.Desc(generationoutput.FieldCreatedAt)).Limit(per_page + 1)

	res, err := query.All(r.Ctx)

	if err != nil {
		log.Errorf("Error retrieving gallery data: %v", err)
		return nil, nil, err
	}

	var nextCursor *time.Time
	if len(res) > per_page {
		res = res[:len(res)-1]
		nextCursor = &res[len(res)-1].CreatedAt
	}

	galleryData := make([]GalleryData, len(res))
	for i, output := range res {
		data := GalleryData{
			ID:             output.ID,
			ImagePath:      output.ImagePath,
			ImageURL:       utils.GetURLFromImagePath(output.ImagePath),
			CreatedAt:      output.CreatedAt,
			UpdatedAt:      output.UpdatedAt,
			Width:          output.Edges.Generations.Width,
			Height:         output.Edges.Generations.Height,
			InferenceSteps: output.Edges.Generations.InferenceSteps,
			GuidanceScale:  output.Edges.Generations.GuidanceScale,
			Seed:           output.Edges.Generations.Seed,
			ModelID:        output.Edges.Generations.ModelID,
			SchedulerID:    output.Edges.Generations.SchedulerID,
			PromptText:     output.Edges.Generations.Edges.Prompt.Text,
			PromptID:       output.Edges.Generations.Edges.Prompt.ID,
			UserID:         &output.Edges.Generations.UserID,
		}
		if output.UpscaledImagePath != nil {
			data.UpscaledImagePath = *output.UpscaledImagePath
			data.UpscaledImageURL = utils.GetURLFromImagePath(data.UpscaledImagePath)
		}
		if output.Edges.Generations.Edges.NegativePrompt != nil {
			data.NegativePromptText = output.Edges.Generations.Edges.NegativePrompt.Text
			data.NegativePromptID = &output.Edges.Generations.Edges.NegativePrompt.ID
		}
		galleryData[i] = data
	}

	return galleryData, nextCursor, nil
}

// Retrieves data in gallery format given  output IDs
func (r *Repository) RetrieveGalleryDataWithOutputIDs(outputIDs []uuid.UUID) ([]GalleryData, error) {
	res, err := r.DB.GenerationOutput.Query().Where(generationoutput.IDIn(outputIDs...), generationoutput.GalleryStatusEQ(generationoutput.GalleryStatusApproved)).
		WithGenerations(func(gq *ent.GenerationQuery) {
			gq.WithPrompt()
			gq.WithNegativePrompt()
		},
		).All(r.Ctx)
	if err != nil {
		return nil, err
	}

	galleryData := make([]GalleryData, len(res))
	for i, output := range res {
		data := GalleryData{
			ID:             output.ID,
			ImagePath:      output.ImagePath,
			ImageURL:       utils.GetURLFromImagePath(output.ImagePath),
			CreatedAt:      output.CreatedAt,
			UpdatedAt:      output.UpdatedAt,
			Width:          output.Edges.Generations.Width,
			Height:         output.Edges.Generations.Height,
			InferenceSteps: output.Edges.Generations.InferenceSteps,
			GuidanceScale:  output.Edges.Generations.GuidanceScale,
			Seed:           output.Edges.Generations.Seed,
			ModelID:        output.Edges.Generations.ModelID,
			SchedulerID:    output.Edges.Generations.SchedulerID,
			PromptText:     output.Edges.Generations.Edges.Prompt.Text,
			PromptID:       output.Edges.Generations.Edges.Prompt.ID,
			UserID:         &output.Edges.Generations.UserID,
		}
		if output.UpscaledImagePath != nil {
			data.UpscaledImagePath = *output.UpscaledImagePath
			data.UpscaledImageURL = utils.GetURLFromImagePath(data.UpscaledImagePath)
		}
		if output.Edges.Generations.Edges.NegativePrompt != nil {
			data.NegativePromptText = output.Edges.Generations.Edges.NegativePrompt.Text
			data.NegativePromptID = &output.Edges.Generations.Edges.NegativePrompt.ID
		}
		galleryData[i] = data
	}
	return galleryData, nil
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
	Score              *float32   `json:"score,omitempty" sql:"score"`
}
