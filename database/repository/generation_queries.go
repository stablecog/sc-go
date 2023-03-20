package repository

import (
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/ent/generation"
	"github.com/stablecog/sc-go/database/ent/generationoutput"
	"github.com/stablecog/sc-go/database/ent/negativeprompt"
	"github.com/stablecog/sc-go/database/ent/prompt"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/utils"
)

// Get generation by ID
func (r *Repository) GetGeneration(id uuid.UUID) (*ent.Generation, error) {
	return r.DB.Generation.Query().Where(generation.IDEQ(id)).First(r.Ctx)
}

// Get generation output by ID
func (r *Repository) GetGenerationOutput(id uuid.UUID) (*ent.GenerationOutput, error) {
	return r.DB.GenerationOutput.Query().Where(generationoutput.IDEQ(id)).First(r.Ctx)
}

// Get generation output for user
func (r *Repository) GetGenerationOutputForUser(id uuid.UUID, userID uuid.UUID) (*ent.GenerationOutput, error) {
	return r.DB.Generation.Query().Where(generation.UserIDEQ(userID)).QueryGenerationOutputs().Where(generationoutput.IDEQ(id)).First(r.Ctx)
}

// Get width/height for generation output
func (r *Repository) GetGenerationOutputWidthHeight(outputID uuid.UUID) (width, height int32, err error) {
	gen, err := r.DB.GenerationOutput.Query().Where(generationoutput.IDEQ(outputID)).QueryGenerations().Select(generation.FieldWidth, generation.FieldHeight).First(r.Ctx)
	if err != nil {
		return 0, 0, err
	}
	return gen.Width, gen.Height, nil
}

// Get last N generations, a basic view
func (r *Repository) GetGenerations(limit int) ([]*ent.Generation, error) {
	return r.DB.Generation.Query().Order(ent.Desc(generation.FieldCreatedAt)).Limit(limit).All(r.Ctx)
}

// Apply all filters to root ent query
func (r *Repository) ApplyUserGenerationsFilters(query *ent.GenerationQuery, filters *requests.QueryGenerationFilters, omitEdges bool) *ent.GenerationQuery {
	resQuery := query
	if filters != nil {
		// Apply filters
		if len(filters.ModelIDs) > 0 {
			resQuery = resQuery.Where(generation.ModelIDIn(filters.ModelIDs...))
		}
		if len(filters.SchedulerIDs) > 0 {
			resQuery = resQuery.Where(generation.SchedulerIDIn(filters.SchedulerIDs...))
		}
		// Apply OR if both are present
		// Confusing, but example of what we want to do:
		// If min_width=100, max_width=200, widths=[300,400]
		// We want to query like; WHERE (width >= 100 AND width <= 200) OR width IN (300,400)
		if (filters.MinWidth != 0 || filters.MaxWidth != 0) && len(filters.Widths) > 0 {
			if filters.MinWidth != 0 && filters.MaxWidth != 0 {
				resQuery = resQuery.Where(generation.Or(generation.And(generation.WidthGTE(filters.MinWidth), generation.WidthLTE(filters.MaxWidth)), generation.WidthIn(filters.Widths...)))
			} else if filters.MinWidth != 0 {
				resQuery = resQuery.Where(generation.Or(generation.WidthGTE(filters.MinWidth), generation.WidthIn(filters.Widths...)))
			} else {
				resQuery = resQuery.Where(generation.Or(generation.WidthLTE(filters.MaxWidth), generation.WidthIn(filters.Widths...)))
			}
		} else {
			if filters.MinWidth != 0 {
				resQuery = resQuery.Where(generation.WidthGTE(filters.MinWidth))
			}
			if filters.MaxWidth != 0 {
				resQuery = resQuery.Where(generation.WidthLTE(filters.MaxWidth))
			}
			if len(filters.Widths) > 0 {
				resQuery = resQuery.Where(generation.WidthIn(filters.Widths...))
			}
		}

		// Height
		if (filters.MinHeight != 0 || filters.MaxHeight != 0) && len(filters.Heights) > 0 {
			if filters.MinHeight != 0 && filters.MaxHeight != 0 {
				resQuery = resQuery.Where(generation.Or(generation.And(generation.HeightGTE(filters.MinHeight), generation.HeightLTE(filters.MaxHeight)), generation.HeightIn(filters.Heights...)))
			} else if filters.MinHeight != 0 {
				resQuery = resQuery.Where(generation.Or(generation.HeightGTE(filters.MinHeight), generation.HeightIn(filters.Heights...)))
			} else {
				resQuery = resQuery.Where(generation.Or(generation.HeightLTE(filters.MaxHeight), generation.HeightIn(filters.Heights...)))
			}
		} else {
			if len(filters.Heights) > 0 {
				resQuery = resQuery.Where(generation.HeightIn(filters.Heights...))
			}
			if filters.MaxHeight != 0 {
				resQuery = resQuery.Where(generation.HeightLTE(filters.MaxHeight))
			}
			if filters.MinHeight != 0 {
				resQuery = resQuery.Where(generation.HeightGTE(filters.MinHeight))
			}
		}

		// Inference steps
		if (filters.MinInferenceSteps != 0 || filters.MaxInferenceSteps != 0) && len(filters.InferenceSteps) > 0 {
			if filters.MinInferenceSteps != 0 && filters.MaxInferenceSteps != 0 {
				resQuery = resQuery.Where(generation.Or(generation.And(generation.InferenceStepsGTE(filters.MinInferenceSteps), generation.InferenceStepsLTE(filters.MaxInferenceSteps)), generation.InferenceStepsIn(filters.InferenceSteps...)))
			} else if filters.MinInferenceSteps != 0 {
				resQuery = resQuery.Where(generation.Or(generation.InferenceStepsGTE(filters.MinInferenceSteps), generation.InferenceStepsIn(filters.InferenceSteps...)))
			} else {
				resQuery = resQuery.Where(generation.Or(generation.InferenceStepsLTE(filters.MaxInferenceSteps), generation.InferenceStepsIn(filters.InferenceSteps...)))
			}
		} else {
			if len(filters.InferenceSteps) > 0 {
				resQuery = resQuery.Where(generation.InferenceStepsIn(filters.InferenceSteps...))
			}
			if filters.MaxInferenceSteps != 0 {
				resQuery = resQuery.Where(generation.InferenceStepsLTE(filters.MaxInferenceSteps))
			}
			if filters.MinInferenceSteps != 0 {
				resQuery = resQuery.Where(generation.InferenceStepsGTE(filters.MinInferenceSteps))
			}
		}

		// Guidance Scales
		if (filters.MinGuidanceScale != 0 || filters.MaxGuidanceScale != 0) && len(filters.GuidanceScales) > 0 {
			if filters.MinGuidanceScale != 0 && filters.MaxGuidanceScale != 0 {
				resQuery = resQuery.Where(generation.Or(generation.And(generation.GuidanceScaleGTE(filters.MinGuidanceScale), generation.GuidanceScaleLTE(filters.MaxGuidanceScale)), generation.GuidanceScaleIn(filters.GuidanceScales...)))
			} else if filters.MinGuidanceScale != 0 {
				resQuery = resQuery.Where(generation.Or(generation.GuidanceScaleGTE(filters.MinGuidanceScale), generation.GuidanceScaleIn(filters.GuidanceScales...)))
			} else {
				resQuery = resQuery.Where(generation.Or(generation.GuidanceScaleLTE(filters.MaxGuidanceScale), generation.GuidanceScaleIn(filters.GuidanceScales...)))
			}
		} else {
			if len(filters.GuidanceScales) > 0 {
				resQuery = resQuery.Where(generation.GuidanceScaleIn(filters.GuidanceScales...))
			}
			if filters.MaxGuidanceScale != 0 {
				resQuery = resQuery.Where(generation.GuidanceScaleLTE(filters.MaxGuidanceScale))
			}
			if filters.MinGuidanceScale != 0 {
				resQuery = resQuery.Where(generation.GuidanceScaleGTE(filters.MinGuidanceScale))
			}
		}

		if !omitEdges {
			// Upscaled
			if filters.UpscaleStatus == requests.UpscaleStatusNot {
				resQuery = resQuery.Where(func(s *sql.Selector) {
					s.Where(sql.IsNull("upscaled_image_path"))
				})
			} else if filters.UpscaleStatus == requests.UpscaleStatusOnly {
				resQuery = resQuery.Where(func(s *sql.Selector) {
					s.Where(sql.NotNull("upscaled_image_path"))
				})
			}

			if len(filters.GalleryStatus) > 0 {
				v := make([]any, len(filters.GalleryStatus))
				for i := range v {
					v[i] = filters.GalleryStatus[i]
				}
				resQuery = resQuery.Where(func(s *sql.Selector) {
					s.Where(sql.In(generationoutput.FieldGalleryStatus, v...))
				})
			}

			if filters.IsFavorited != nil {
				resQuery = resQuery.Where(func(s *sql.Selector) {
					s.Where(sql.EQ(generationoutput.FieldIsFavorited, *filters.IsFavorited))
				})
			}
		}

		// Start dt
		if filters.StartDt != nil {
			resQuery = resQuery.Where(generation.CreatedAtGTE(*filters.StartDt))
		}

		// End dt
		if filters.EndDt != nil {
			resQuery = resQuery.Where(generation.CreatedAtLTE(*filters.EndDt))
		}

		if filters.WasAutoSubmitted != nil {
			resQuery = resQuery.Where(generation.WasAutoSubmittedEQ(*filters.WasAutoSubmitted))
		}
	}
	return resQuery
}

// Gets the count of generations with outputs user has with filters
func (r *Repository) GetGenerationCount(filters *requests.QueryGenerationFilters) (int, error) {
	var query *ent.GenerationQuery

	query = r.DB.Generation.Query().
		Where(generation.StatusEQ(generation.StatusSucceeded))
	if filters.UserID != nil {
		query = query.Where(generation.UserID(*filters.UserID))
	}

	// Exclude deleted at always
	query = query.Where(func(s *sql.Selector) {
		s.Where(sql.IsNull("deleted_at"))
	})

	// Apply filters
	query = r.ApplyUserGenerationsFilters(query, filters, false)

	// Join other data
	var res []UserGenCount
	err := query.Modify(func(s *sql.Selector) {
		npt := sql.Table(negativeprompt.Table)
		pt := sql.Table(prompt.Table)
		got := sql.Table(generationoutput.Table)
		s.LeftJoin(npt).On(
			s.C(generation.FieldNegativePromptID), npt.C(negativeprompt.FieldID),
		).LeftJoin(pt).On(
			s.C(generation.FieldPromptID), pt.C(prompt.FieldID),
		).LeftJoin(got).On(
			s.C(generation.FieldID), got.C(generationoutput.FieldGenerationID),
		).Select(sql.As(sql.Count("*"), "total"))
	}).Scan(r.Ctx, &res)
	if err != nil {
		return 0, err
	} else if len(res) == 0 {
		return 0, nil
	}
	return res[0].Total, nil
}

type UserGenCount struct {
	Total int `json:"total" sql:"total"`
}

// Get user generations from the database using page options
// Cursor actually represents created_at, we paginate using this for performance reasons
// If present, we will get results after the cursor (anything before, represents previous pages)
// ! using ent .With... doesn't use joins, so we construct our own query to make it more efficient
func (r *Repository) QueryGenerations(per_page int, cursor *time.Time, filters *requests.QueryGenerationFilters) (*GenerationQueryWithOutputsMeta, error) {
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
	}
	var query *ent.GenerationQuery
	var gQueryResult []GenerationQueryWithOutputsResult

	// User can't order by updated_At
	if filters != nil && filters.OrderBy == requests.OrderByUpdatedAt {
		filters.OrderBy = requests.OrderByCreatedAt
	}

	query = r.DB.Generation.Query().Select(selectFields...).
		Where(generation.StatusEQ(generation.StatusSucceeded))
	if filters.UserID != nil {
		query = query.Where(generation.UserID(*filters.UserID))
	}
	if cursor != nil {
		query = query.Where(generation.CreatedAtLT(*cursor))
	}

	// Exclude deleted at always
	query = query.Where(func(s *sql.Selector) {
		s.Where(sql.IsNull("deleted_at"))
	})

	// Apply filters
	query = r.ApplyUserGenerationsFilters(query, filters, false)

	// Limits is + 1 so we can check if there are more pages
	query = query.Limit(per_page + 1)

	// Join other data
	err := query.Modify(func(s *sql.Selector) {
		gt := sql.Table(generation.Table)
		npt := sql.Table(negativeprompt.Table)
		pt := sql.Table(prompt.Table)
		got := sql.Table(generationoutput.Table)
		s.LeftJoin(npt).On(
			s.C(generation.FieldNegativePromptID), npt.C(negativeprompt.FieldID),
		).LeftJoin(pt).On(
			s.C(generation.FieldPromptID), pt.C(prompt.FieldID),
		).LeftJoin(got).On(
			s.C(generation.FieldID), got.C(generationoutput.FieldGenerationID),
		).AppendSelect(sql.As(npt.C(negativeprompt.FieldText), "negative_prompt_text"), sql.As(pt.C(prompt.FieldText), "prompt_text"), sql.As(got.C(generationoutput.FieldID), "output_id"), sql.As(got.C(generationoutput.FieldGalleryStatus), "output_gallery_status"), sql.As(got.C(generationoutput.FieldImagePath), "image_path"), sql.As(got.C(generationoutput.FieldUpscaledImagePath), "upscaled_image_path"), sql.As(got.C(generationoutput.FieldDeletedAt), "deleted_at"), sql.As(got.C(generationoutput.FieldIsFavorited), "is_favorited")).
			GroupBy(s.C(generation.FieldID), npt.C(negativeprompt.FieldText), pt.C(prompt.FieldText),
				got.C(generationoutput.FieldID), got.C(generationoutput.FieldGalleryStatus),
				got.C(generationoutput.FieldImagePath), got.C(generationoutput.FieldUpscaledImagePath))
		var orderByGeneration string
		var orderByOutput string
		orderByGeneration = generation.FieldCreatedAt
		orderByOutput = generationoutput.FieldCreatedAt
		// Order by generation, then output
		if filters == nil || (filters != nil && filters.Order == requests.SortOrderDescending) {
			s.OrderBy(sql.Desc(got.C(orderByOutput)), sql.Desc(gt.C(orderByGeneration)))
		} else {
			s.OrderBy(sql.Asc(got.C(orderByOutput)), sql.Asc(gt.C(orderByGeneration)))
		}
	}).Scan(r.Ctx, &gQueryResult)

	if err != nil {
		log.Error("Error getting user generations", "err", err)
		return nil, err
	}

	if len(gQueryResult) == 0 {
		meta := &GenerationQueryWithOutputsMeta{
			Outputs: []GenerationQueryWithOutputsResultFormatted{},
		}
		// Only give total if we have no cursor
		if cursor == nil {
			zero := 0
			meta.Total = &zero
		}
		return meta, nil
	}

	meta := &GenerationQueryWithOutputsMeta{}
	if len(gQueryResult) > per_page {
		// Remove last item
		gQueryResult = gQueryResult[:len(gQueryResult)-1]
		meta.Next = &gQueryResult[len(gQueryResult)-1].CreatedAt
	}

	// Get real image URLs for each
	for i, g := range gQueryResult {
		if g.ImageUrl != "" {
			parsed := utils.GetURLFromImagePath(g.ImageUrl)
			gQueryResult[i].ImageUrl = parsed
		}
		if g.UpscaledImageUrl != "" {
			parsed := utils.GetURLFromImagePath(g.UpscaledImageUrl)
			gQueryResult[i].UpscaledImageUrl = parsed
		}
	}

	// Format to GenerationQueryWithOutputsResultFormatted
	generationOutputMap := make(map[uuid.UUID][]GenerationUpscaleOutput)
	for _, g := range gQueryResult {
		if g.OutputID == nil {
			log.Warn("Output ID is nil for generation, cannot include in result", "id", g.ID)
			continue
		}
		gOutput := GenerationUpscaleOutput{
			ID:               *g.OutputID,
			ImageUrl:         g.ImageUrl,
			UpscaledImageUrl: g.UpscaledImageUrl,
			GalleryStatus:    g.GalleryStatus,
			WasAutoSubmitted: g.WasAutoSubmitted,
			IsFavorited:      g.IsFavorited,
		}
		output := GenerationQueryWithOutputsResultFormatted{
			GenerationUpscaleOutput: gOutput,
			Generation: GenerationQueryWithOutputsData{
				ID:               g.ID,
				Height:           g.Height,
				Width:            g.Width,
				InferenceSteps:   g.InferenceSteps,
				Seed:             g.Seed,
				Status:           g.Status,
				GuidanceScale:    g.GuidanceScale,
				SchedulerID:      g.SchedulerID,
				ModelID:          g.ModelID,
				PromptID:         g.PromptID,
				NegativePromptID: g.NegativePromptID,
				CreatedAt:        g.CreatedAt,
				UpdatedAt:        g.UpdatedAt,
				StartedAt:        g.StartedAt,
				CompletedAt:      g.CompletedAt,
				Prompt: PromptType{
					Text: g.PromptText,
					ID:   *g.PromptID,
				},
				IsFavorited: gOutput.IsFavorited,
			},
		}
		if g.NegativePromptID != nil {
			output.Generation.NegativePrompt = &PromptType{
				Text: g.NegativePromptText,
				ID:   *g.NegativePromptID,
			}
		}
		generationOutputMap[g.ID] = append(generationOutputMap[g.ID], gOutput)
		meta.Outputs = append(meta.Outputs, output)
	}
	// Now loop through and add outputs to each generation
	for i, g := range meta.Outputs {
		meta.Outputs[i].Generation.Outputs = generationOutputMap[g.Generation.ID]
	}

	if cursor == nil {
		total, err := r.GetGenerationCount(filters)
		if err != nil {
			log.Error("Error getting user generation count", "err", err)
			return nil, err
		}
		meta.Total = &total
	}

	return meta, err
}

// Separate count function
func (r *Repository) GetGenerationCountAdmin(filters *requests.QueryGenerationFilters) (int, error) {
	var query *ent.GenerationOutputQuery

	query = r.DB.GenerationOutput.Query().Where(
		generationoutput.DeletedAtIsNil(),
	)
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
		s.Where(generation.StatusEQ(generation.StatusSucceeded))
		s = r.ApplyUserGenerationsFilters(s, filters, true)
		s.WithPrompt()
		s.WithNegativePrompt()
		s.WithGenerationOutputs()
	})

	return query.Count(r.Ctx)
}

// Alternate version for performance when we can't index by user_id
func (r *Repository) QueryGenerationsAdmin(per_page int, dtCursor *time.Time, offsetCursor *int, filters *requests.QueryGenerationFilters) (*GenerationQueryWithOutputsMeta, error) {
	var gQueryResult []GenerationQueryWithOutputsResult

	// Figure out order bys
	var orderByGeneration string
	var orderByOutput string
	if filters == nil || (filters != nil && filters.OrderBy == requests.OrderByCreatedAt) {
		orderByGeneration = generation.FieldCreatedAt
		orderByOutput = generationoutput.FieldCreatedAt
	} else {
		orderByGeneration = generation.FieldUpdatedAt
		orderByOutput = generationoutput.FieldUpdatedAt
	}

	query := r.ApplyUserGenerationsFilters(r.DB.Generation.Query(), filters, true).QueryGenerationOutputs().Where(
		generationoutput.DeletedAtIsNil(),
	)
	if dtCursor != nil {
		query = query.Where(generationoutput.CreatedAtLT(*dtCursor))
	} else if offsetCursor != nil {
		query = query.Offset(*offsetCursor)
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
		s.WithGenerationOutputs(func(goq *ent.GenerationOutputQuery) {
			if filters == nil || (filters != nil && filters.Order == requests.SortOrderDescending) {
				goq = goq.Order(ent.Desc(orderByOutput))
			} else {
				goq = goq.Order(ent.Asc(orderByOutput))
			}
		})
		if filters == nil || (filters != nil && filters.Order == requests.SortOrderDescending) {
			s = s.Order(ent.Desc(orderByGeneration))
		} else {
			s = s.Order(ent.Asc(orderByGeneration))
		}
	})

	if filters == nil || (filters != nil && filters.Order == requests.SortOrderDescending) {
		query = query.Order(ent.Desc(orderByOutput))
	} else {
		query = query.Order(ent.Asc(orderByOutput))
	}

	// Limit
	query = query.Limit(per_page + 1)

	res, err := query.All(r.Ctx)

	if err != nil {
		log.Error("Error getting admin generations", "err", err)
		return nil, err
	}

	meta := &GenerationQueryWithOutputsMeta{}

	if len(res) == 0 {
		meta := &GenerationQueryWithOutputsMeta{
			Outputs: []GenerationQueryWithOutputsResultFormatted{},
		}
		// Only give total if we have no cursor
		if dtCursor == nil || offsetCursor == nil {
			zero := 0
			meta.Total = &zero
		}
		return meta, nil
	}

	if len(res) > per_page {
		// Remove last item
		res = res[:len(res)-1]
		if filters == nil || (filters != nil && filters.OrderBy == requests.OrderByCreatedAt) {
			meta.Next = &res[len(res)-1].CreatedAt
		} else {
			// Use offset otherwise
			meta.Next = len(res)
		}
	}

	// Get real image URLs for each
	// Since we do this weird format need to cap at per page
	for _, g := range res {
		// root generation data
		generationRoot := GenerationQueryWithOutputsData{
			ID:               g.ID,
			Width:            g.Edges.Generations.Width,
			Height:           g.Edges.Generations.Height,
			InferenceSteps:   g.Edges.Generations.InferenceSteps,
			Seed:             g.Edges.Generations.Seed,
			Status:           string(g.Edges.Generations.Status),
			GuidanceScale:    g.Edges.Generations.GuidanceScale,
			SchedulerID:      g.Edges.Generations.SchedulerID,
			ModelID:          g.Edges.Generations.ModelID,
			PromptID:         g.Edges.Generations.PromptID,
			NegativePromptID: g.Edges.Generations.NegativePromptID,
			CreatedAt:        g.Edges.Generations.CreatedAt,
			UpdatedAt:        g.Edges.Generations.UpdatedAt,
			StartedAt:        g.Edges.Generations.StartedAt,
			CompletedAt:      g.Edges.Generations.CompletedAt,
			WasAutoSubmitted: g.Edges.Generations.WasAutoSubmitted,
			IsFavorited:      g.IsFavorited,
		}
		if g.Edges.Generations.Edges.NegativePrompt != nil {
			generationRoot.NegativePrompt = &PromptType{
				Text: g.Edges.Generations.Edges.NegativePrompt.Text,
				ID:   *g.Edges.Generations.NegativePromptID,
			}
		}
		if g.Edges.Generations.Edges.Prompt != nil {
			generationRoot.Prompt = PromptType{
				Text: g.Edges.Generations.Edges.Prompt.Text,
				ID:   *g.Edges.Generations.PromptID,
			}
		}

		// Add outputs
		for _, o := range g.Edges.Generations.Edges.GenerationOutputs {
			output := GenerationUpscaleOutput{
				ID:               o.ID,
				ImageUrl:         utils.GetURLFromImagePath(o.ImagePath),
				GalleryStatus:    o.GalleryStatus,
				CreatedAt:        &o.CreatedAt,
				IsFavorited:      o.IsFavorited,
				WasAutoSubmitted: generationRoot.WasAutoSubmitted,
			}
			if o.UpscaledImagePath != nil {
				output.UpscaledImageUrl = utils.GetURLFromImagePath(*o.UpscaledImagePath)
			}
			generationRoot.Outputs = append(generationRoot.Outputs, output)
		}

		ret := GenerationQueryWithOutputsResult{
			OutputID:                       &g.ID,
			ImageUrl:                       utils.GetURLFromImagePath(g.ImagePath),
			GalleryStatus:                  g.GalleryStatus,
			DeletedAt:                      g.DeletedAt,
			GenerationQueryWithOutputsData: generationRoot,
		}
		if g.UpscaledImagePath != nil {
			ret.UpscaledImageUrl = utils.GetURLFromImagePath(*g.UpscaledImagePath)
		}

		gQueryResult = append(gQueryResult, ret)
	}

	// Format to GenerationQueryWithOutputsResultFormatted
	for _, g := range gQueryResult {
		if g.OutputID == nil {
			log.Warn("Output ID is nil for generation, cannot include in result", "id", g.ID)
			continue
		}
		gOutput := GenerationUpscaleOutput{
			ID:               *g.OutputID,
			ImageUrl:         g.ImageUrl,
			UpscaledImageUrl: g.UpscaledImageUrl,
			GalleryStatus:    g.GalleryStatus,
		}
		for _, o := range g.Outputs {
			if o.ID == *g.OutputID {
				gOutput.CreatedAt = o.CreatedAt
				break
			}
		}
		output := GenerationQueryWithOutputsResultFormatted{
			GenerationUpscaleOutput: gOutput,
			Generation:              g.GenerationQueryWithOutputsData,
		}

		meta.Outputs = append(meta.Outputs, output)
	}

	// Get count when no cursor
	if dtCursor == nil && offsetCursor == nil {
		total, err := r.GetGenerationCountAdmin(filters)
		if err != nil {
			log.Error("Error getting user generation count", "err", err)
			return nil, err
		}
		meta.Total = &total
	}

	return meta, err
}

type GenerationUpscaleOutput struct {
	ID               uuid.UUID                      `json:"id"`
	ImageUrl         string                         `json:"image_url"`
	UpscaledImageUrl string                         `json:"upscaled_image_url,omitempty"`
	GalleryStatus    generationoutput.GalleryStatus `json:"gallery_status,omitempty"`
	InputImageUrl    string                         `json:"input_image_url,omitempty"`
	OutputID         *uuid.UUID                     `json:"output_id,omitempty"`
	CreatedAt        *time.Time                     `json:"created_at,omitempty"`
	IsFavorited      bool                           `json:"is_favorited"`
	WasAutoSubmitted bool                           `json:"was_auto_submitted"`
}

// Paginated meta for querying generations
type GenerationQueryWithOutputsMeta struct {
	Total   *int                                        `json:"total_count,omitempty"`
	Outputs []GenerationQueryWithOutputsResultFormatted `json:"outputs"`
	Next    interface{}                                 `json:"next,omitempty"`
}

type PromptType struct {
	ID   uuid.UUID `json:"id"`
	Text string    `json:"text"`
}

type GenerationQueryWithOutputsData struct {
	ID                 uuid.UUID                 `json:"id" sql:"id"`
	Height             int32                     `json:"height" sql:"height"`
	Width              int32                     `json:"width" sql:"width"`
	InferenceSteps     int32                     `json:"inference_steps" sql:"inference_steps"`
	Seed               int                       `json:"seed" sql:"seed"`
	Status             string                    `json:"status" sql:"status"`
	GuidanceScale      float32                   `json:"guidance_scale" sql:"guidance_scale"`
	SchedulerID        uuid.UUID                 `json:"scheduler_id" sql:"scheduler_id"`
	ModelID            uuid.UUID                 `json:"model_id" sql:"model_id"`
	PromptID           *uuid.UUID                `json:"prompt_id,omitempty" sql:"prompt_id"`
	NegativePromptID   *uuid.UUID                `json:"negative_prompt_id,omitempty" sql:"negative_prompt_id"`
	CreatedAt          time.Time                 `json:"created_at" sql:"created_at"`
	UpdatedAt          time.Time                 `json:"updated_at" sql:"updated_at"`
	StartedAt          *time.Time                `json:"started_at,omitempty" sql:"started_at"`
	CompletedAt        *time.Time                `json:"completed_at,omitempty" sql:"completed_at"`
	NegativePromptText string                    `json:"negative_prompt_text,omitempty" sql:"negative_prompt_text"`
	PromptText         string                    `json:"prompt_text,omitempty" sql:"prompt_text"`
	IsFavorited        bool                      `json:"is_favorited" sql:"is_favorited"`
	Outputs            []GenerationUpscaleOutput `json:"outputs"`
	Prompt             PromptType                `json:"prompt"`
	NegativePrompt     *PromptType               `json:"negative_prompt,omitempty"`
	WasAutoSubmitted   bool                      `json:"was_auto_submitted" sql:"was_auto_submitted"`
}

type GenerationQueryWithOutputsResult struct {
	OutputID         *uuid.UUID                     `json:"output_id,omitempty" sql:"output_id"`
	ImageUrl         string                         `json:"image_url,omitempty" sql:"image_path"`
	UpscaledImageUrl string                         `json:"upscaled_image_url,omitempty" sql:"upscaled_image_path"`
	GalleryStatus    generationoutput.GalleryStatus `json:"gallery_status,omitempty" sql:"output_gallery_status"`
	DeletedAt        *time.Time                     `json:"deleted_at,omitempty" sql:"deleted_at"`
	GenerationQueryWithOutputsData
}

type GenerationQueryWithOutputsResultFormatted struct {
	GenerationUpscaleOutput
	Generation GenerationQueryWithOutputsData `json:"generation"`
}
