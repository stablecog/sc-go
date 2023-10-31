package repository

import (
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/ent/generation"
	"github.com/stablecog/sc-go/database/ent/generationoutput"
	"github.com/stablecog/sc-go/database/ent/generationoutputlike"
	"github.com/stablecog/sc-go/database/ent/negativeprompt"
	"github.com/stablecog/sc-go/database/ent/prompt"
	"github.com/stablecog/sc-go/database/ent/user"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/utils"
)

// Retrieved a single generation output by ID, in GalleryData format
func (r *Repository) RetrieveGalleryDataByID(id uuid.UUID, userId *uuid.UUID, callingUserId *uuid.UUID, all bool) (*GalleryData, error) {
	var q *ent.GenerationOutputQuery
	if userId != nil {
		q = r.DB.Generation.Query().Where(generation.UserIDEQ(*userId)).QueryGenerationOutputs()
	} else {
		q = r.DB.GenerationOutput.Query()
	}
	q = q.Where(generationoutput.IDEQ(id))
	if !all {
		q = q.Where(generationoutput.GalleryStatusEQ(generationoutput.GalleryStatusApproved))
	}
	if callingUserId != nil {
		q = q.WithGenerationOutputLikes(func(gql *ent.GenerationOutputLikeQuery) {
			gql.Where(generationoutputlike.LikedByUserID(*callingUserId))
		})
	}
	output, err := q.WithGenerations(func(gq *ent.GenerationQuery) {
		gq.WithPrompt()
		gq.WithNegativePrompt()
		gq.WithUser()
	}).Only(r.Ctx)
	if err != nil {
		return nil, err
	}
	data := GalleryData{
		ID:             output.ID,
		ImageURL:       utils.GetEnv().GetURLFromImagePath(output.ImagePath),
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
		PromptStrength: output.Edges.Generations.PromptStrength,
		User: &UserType{
			Username: output.Edges.Generations.Edges.User.Username,
		},
		LikeCount:   output.LikeCount,
		LikedByUser: utils.ToPtr(len(output.Edges.GenerationOutputLikes) > 0),
	}
	if all {
		data.IsPublic = output.IsPublic
		data.WasAutoSubmitted = output.Edges.Generations.WasAutoSubmitted
	}
	if output.Edges.Generations.Edges.NegativePrompt != nil {
		data.NegativePromptID = &output.Edges.Generations.Edges.NegativePrompt.ID
		data.NegativePromptText = output.Edges.Generations.Edges.NegativePrompt.Text
	}
	if output.UpscaledImagePath != nil {
		data.UpscaledImageURL = utils.GetEnv().GetURLFromImagePath(*output.UpscaledImagePath)
	}
	return &data, nil
}

func (r *Repository) RetrieveMostRecentGalleryDataV2(filters *requests.QueryGenerationFilters, callingUserId *uuid.UUID, per_page int, cursor *time.Time) ([]GalleryData, *time.Time, error) {
	// Base fields to select in our query
	selectFields := []string{
		generation.FieldID,
		generation.FieldWidth,
		generation.FieldHeight,
		generation.FieldInferenceSteps,
		generation.FieldSeed,
		generation.FieldStatus,
		generation.FieldGuidanceScale,
		generation.FieldSchedulerID,
		generation.FieldModelID,
		generation.FieldPromptID,
		generation.FieldNegativePromptID,
		generation.FieldCreatedAt,
		generation.FieldUpdatedAt,
		generation.FieldStartedAt,
		generation.FieldCompletedAt,
		generation.FieldWasAutoSubmitted,
		generation.FieldInitImageURL,
		generation.FieldPromptStrength,
	}
	var query *ent.GenerationQuery
	var gQueryResult []GenerationQueryWithOutputsResult

	// Figure out order bys
	var orderByGeneration []string
	var orderByOutput []string
	if filters == nil || (filters != nil && filters.OrderBy == requests.OrderByCreatedAt) {
		orderByGeneration = []string{generation.FieldCreatedAt}
		orderByOutput = []string{generationoutput.FieldCreatedAt}
	} else {
		orderByGeneration = []string{generation.FieldCreatedAt, generation.FieldUpdatedAt}
		orderByOutput = []string{generationoutput.FieldCreatedAt, generationoutput.FieldUpdatedAt}
	}

	query = r.DB.Generation.Query().Select(selectFields...).
		Where(generation.StatusEQ(generation.StatusSucceeded))
	if cursor != nil {
		query = query.Where(generation.CreatedAtLT(*cursor))
	}

	// Apply filters
	query = r.ApplyUserGenerationsFilters(query, filters, false)

	// Limits is + 1 so we can check if there are more pages
	query = query.Limit(per_page + 1)

	// Join other data
	err := query.Modify(func(s *sql.Selector) {
		gt := sql.Table(generation.Table)
		got := sql.Table(generationoutput.Table)
		ut := sql.Table(user.Table)
		ltj := s.Join(got).OnP(
			sql.And(
				sql.ColumnsEQ(gt.C(generation.FieldID), got.C(generationoutput.FieldGenerationID)),
				sql.IsNull(got.C(generationoutput.FieldDeletedAt)),
			),
		)
		if filters != nil && filters.UserID != nil {
			ltj.Join(ut).OnP(
				sql.And(
					sql.ColumnsEQ(s.C(generation.FieldUserID), ut.C(user.FieldID)),
					sql.EQ(ut.C(user.FieldID), *filters.UserID),
				),
			)
		} else {
			ltj.LeftJoin(ut).On(
				s.C(generation.FieldUserID), ut.C(user.FieldID),
			)
		}
		ltj.AppendSelect(sql.As(got.C(generationoutput.FieldID), "output_id"), sql.As(got.C(generationoutput.FieldLikeCount), "like_count"), sql.As(got.C(generationoutput.FieldGalleryStatus), "output_gallery_status"), sql.As(got.C(generationoutput.FieldImagePath), "image_path"), sql.As(got.C(generationoutput.FieldUpscaledImagePath), "upscaled_image_path"), sql.As(got.C(generationoutput.FieldDeletedAt), "deleted_at"), sql.As(got.C(generationoutput.FieldIsFavorited), "is_favorited"), sql.As(ut.C(user.FieldUsername), "username"), sql.As(got.C(generationoutput.FieldIsPublic), "is_public")).
			GroupBy(s.C(generation.FieldID),
				got.C(generationoutput.FieldID), got.C(generationoutput.FieldGalleryStatus),
				got.C(generationoutput.FieldImagePath), got.C(generationoutput.FieldUpscaledImagePath),
				ut.C(user.FieldUsername))
		orderDir := "asc"
		if filters == nil || (filters != nil && filters.Order == requests.SortOrderDescending) {
			orderDir = "desc"
		}
		var orderByGeneration2 []string
		var orderByOutput2 []string
		for _, o := range orderByGeneration {
			if orderDir == "desc" {
				orderByGeneration2 = append(orderByGeneration2, sql.Desc(gt.C(o)))
			} else {
				orderByGeneration2 = append(orderByGeneration2, sql.Asc(gt.C(o)))
			}
		}
		for _, o := range orderByOutput {
			if orderDir == "desc" {
				orderByOutput2 = append(orderByOutput2, sql.Desc(got.C(o)))
			} else {
				orderByOutput2 = append(orderByOutput2, sql.Asc(got.C(o)))
			}
		}
		// Order by generation, then output
		orderByCombined := append(orderByGeneration2, orderByOutput2...)
		s.OrderBy(orderByCombined...)
	}).Scan(r.Ctx, &gQueryResult)

	if err != nil {
		log.Error("Error retrieving generations", "err", err)
		return nil, nil, err
	}

	if len(gQueryResult) == 0 {
		return []GalleryData{}, nil, nil
	}

	// Get prompt texts
	promptIDsMap := make(map[uuid.UUID]string)
	negativePromptIdsMap := make(map[uuid.UUID]string)
	for _, g := range gQueryResult {
		if g.PromptID != nil {
			promptIDsMap[*g.PromptID] = ""
		}
		if g.NegativePromptID != nil {
			negativePromptIdsMap[*g.NegativePromptID] = ""
		}
	}
	promptIDs := make([]uuid.UUID, len(promptIDsMap))
	negativePromptId := make([]uuid.UUID, len(negativePromptIdsMap))

	i := 0
	for k := range promptIDsMap {
		promptIDs[i] = k
		i++
	}
	i = 0
	for k := range negativePromptIdsMap {
		negativePromptId[i] = k
		i++
	}

	prompts, err := r.DB.Prompt.Query().Select(prompt.FieldText).Where(prompt.IDIn(promptIDs...)).All(r.Ctx)
	if err != nil {
		log.Error("Error retrieving prompts", "err", err)
		return nil, nil, err
	}
	negativePrompts, err := r.DB.NegativePrompt.Query().Select(negativeprompt.FieldText).Where(negativeprompt.IDIn(negativePromptId...)).All(r.Ctx)
	if err != nil {
		log.Error("Error retrieving prompts", "err", err)
		return nil, nil, err
	}
	for _, p := range prompts {
		promptIDsMap[p.ID] = p.Text
	}
	for _, p := range negativePrompts {
		negativePromptIdsMap[p.ID] = p.Text
	}

	var nextCursor *time.Time
	if len(gQueryResult) > per_page {
		gQueryResult = gQueryResult[:len(gQueryResult)-1]
		nextCursor = &gQueryResult[len(gQueryResult)-1].CreatedAt
	}

	// Figure out liked by in another query, if calling user is provided
	likedByMap := make(map[uuid.UUID]struct{})
	if callingUserId != nil {
		outputIds := make([]uuid.UUID, len(gQueryResult))
		for i, g := range gQueryResult {
			outputIds[i] = *g.OutputID
		}
		likedByMap, err = r.GetGenerationOutputsLikedByUser(*callingUserId, outputIds)
		if err != nil {
			log.Error("Error getting liked by map", "err", err)
			return nil, nil, err
		}
	}

	galleryData := make([]GalleryData, len(gQueryResult))
	for i, g := range gQueryResult {
		likedByUser := false
		if _, ok := likedByMap[*g.OutputID]; ok {
			likedByUser = true
		}
		promptText, _ := promptIDsMap[*g.PromptID]
		galleryData[i] = GalleryData{
			ID:             *g.OutputID,
			ImageURL:       utils.GetEnv().GetURLFromImagePath(g.ImageUrl),
			CreatedAt:      g.CreatedAt,
			UpdatedAt:      g.UpdatedAt,
			Width:          g.Width,
			Height:         g.Height,
			InferenceSteps: g.InferenceSteps,
			GuidanceScale:  g.GuidanceScale,
			Seed:           g.Seed,
			ModelID:        g.ModelID,
			SchedulerID:    g.SchedulerID,
			PromptText:     promptText,
			PromptID:       *g.PromptID,
			PromptStrength: g.PromptStrength,
			User: &UserType{
				Username: g.Username,
			},
			WasAutoSubmitted: g.WasAutoSubmitted,
			IsPublic:         g.IsPublic,
			LikeCount:        g.LikeCount,
			LikedByUser:      utils.ToPtr(likedByUser),
		}

		if g.NegativePromptID != nil {
			galleryData[i].NegativePromptText, _ = negativePromptIdsMap[*g.NegativePromptID]
			galleryData[i].NegativePromptID = g.NegativePromptID
		}

		if g.UpscaledImageUrl != "" {
			galleryData[i].UpscaledImageURL = utils.GetEnv().GetURLFromImagePath(g.UpscaledImageUrl)
		}
	}

	return galleryData, nextCursor, nil
}

// Retrieves data in gallery format given  output IDs
// Returns data, next cursor, error
func (r *Repository) RetrieveMostRecentGalleryData(filters *requests.QueryGenerationFilters, callingUserId *uuid.UUID, per_page int, cursor *time.Time) ([]GalleryData, *time.Time, error) {
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
		if filters.IsPublic != nil {
			query = query.Where(generationoutput.IsPublic(*filters.IsPublic))
		}
	}
	if callingUserId != nil {
		query = query.WithGenerationOutputLikes(func(gql *ent.GenerationOutputLikeQuery) {
			gql.Where(generationoutputlike.LikedByUserID(*callingUserId))
		})
	}
	query = query.WithGenerations(func(s *ent.GenerationQuery) {
		s.WithPrompt()
		s.WithNegativePrompt()
		s.WithGenerationOutputs()
		s.WithUser()
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
			ImageURL:       utils.GetEnv().GetURLFromImagePath(output.ImagePath),
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
			User: &UserType{
				Username: output.Edges.Generations.Edges.User.Username,
			},
			LikeCount:   output.LikeCount,
			LikedByUser: utils.ToPtr(len(output.Edges.GenerationOutputLikes) > 0),
		}
		if output.UpscaledImagePath != nil {
			data.UpscaledImageURL = utils.GetEnv().GetURLFromImagePath(*output.UpscaledImagePath)
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
func (r *Repository) RetrieveGalleryDataWithOutputIDs(outputIDs []uuid.UUID, callingUserId *uuid.UUID, allIsPublic bool) ([]GalleryData, error) {
	q := r.DB.GenerationOutput.Query().Where(generationoutput.IDIn(outputIDs...))
	if allIsPublic {
		q = q.Where(generationoutput.IsPublic(true))
	} else {
		q = q.Where(generationoutput.GalleryStatusEQ(generationoutput.GalleryStatusApproved))
	}
	if callingUserId != nil {
		q = q.WithGenerationOutputLikes(func(gql *ent.GenerationOutputLikeQuery) {
			gql.Where(generationoutputlike.LikedByUserID(*callingUserId))
		})
	}
	res, err := q.
		WithGenerations(func(gq *ent.GenerationQuery) {
			gq.WithPrompt()
			gq.WithNegativePrompt()
			gq.WithUser()
		},
		).All(r.Ctx)
	if err != nil {
		return nil, err
	}

	galleryData := make([]GalleryData, len(res))
	for i, output := range res {
		data := GalleryData{
			ID:             output.ID,
			ImageURL:       utils.GetEnv().GetURLFromImagePath(output.ImagePath),
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
			User: &UserType{
				Username: output.Edges.Generations.Edges.User.Username,
			},
			LikeCount:   output.LikeCount,
			LikedByUser: utils.ToPtr(len(output.Edges.GenerationOutputLikes) > 0),
		}
		if output.UpscaledImagePath != nil {
			data.UpscaledImageURL = utils.GetEnv().GetURLFromImagePath(*output.UpscaledImagePath)
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
	Username           *string    `json:"username,omitempty" sql:"username"`
	User               *UserType  `json:"user,omitempty" sql:"user"`
	PromptStrength     *float32   `json:"prompt_strength,omitempty" sql:"prompt_strength"`
	WasAutoSubmitted   bool       `json:"was_auto_submitted" sql:"was_auto_submitted"`
	IsPublic           bool       `json:"is_public" sql:"is_public"`
	LikeCount          int        `json:"like_count" sql:"like_count"`
	LikedByUser        *bool      `json:"liked_by_user,omitempty" sql:"liked_by_user"`
}
