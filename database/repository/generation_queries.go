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
	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/utils"
	"k8s.io/klog/v2"
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

// Apply all filters to root ent query
func (r *Repository) ApplyUserGenerationsFilters(query *ent.GenerationQuery, filters *requests.QueryGenerationFilters) *ent.GenerationQuery {
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

		// Start dt
		if filters.StartDt != nil {
			resQuery = resQuery.Where(generation.CreatedAtGTE(*filters.StartDt))
		}

		// End dt
		if filters.EndDt != nil {
			resQuery = resQuery.Where(generation.CreatedAtLTE(*filters.EndDt))
		}
	}
	return resQuery
}

// Gets the count of generations with outputs user has with filters
func (r *Repository) GetUserGenerationCountWithFilters(userID uuid.UUID, filters *requests.QueryGenerationFilters) (int, error) {
	var query *ent.GenerationQuery

	query = r.DB.Generation.Query().
		Where(generation.UserID(userID), generation.StatusEQ(generation.StatusSucceeded))

	// Exclude deleted at always
	query = query.Where(func(s *sql.Selector) {
		s.Where(sql.IsNull("deleted_at"))
	})

	// Apply filters
	query = r.ApplyUserGenerationsFilters(query, filters)

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
func (r *Repository) GetUserGenerations(userID uuid.UUID, per_page int, cursor *time.Time, filters *requests.QueryGenerationFilters) (*GenerationQueryWithOutputsMeta, error) {
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
		generation.FieldCreatedAt,
		generation.FieldStartedAt,
		generation.FieldCompletedAt,
	}
	var query *ent.GenerationQuery
	var gQueryResult []GenerationQueryWithOutputsResult

	query = r.DB.Generation.Query().Select(selectFields...).
		Where(generation.UserID(userID), generation.StatusEQ(generation.StatusSucceeded))
	if cursor != nil {
		query = query.Where(generation.CreatedAtLT(*cursor))
	}

	// Exclude deleted at always
	query = query.Where(func(s *sql.Selector) {
		s.Where(sql.IsNull("deleted_at"))
	})

	// Apply filters
	query = r.ApplyUserGenerationsFilters(query, filters)

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
		).AppendSelect(sql.As(npt.C(negativeprompt.FieldText), "negative_prompt_text"), sql.As(pt.C(prompt.FieldText), "prompt_text"), sql.As(got.C(generationoutput.FieldID), "output_id"), sql.As(got.C(generationoutput.FieldGalleryStatus), "output_gallery_status"), sql.As(got.C(generationoutput.FieldImagePath), "image_path"), sql.As(got.C(generationoutput.FieldUpscaledImagePath), "upscaled_image_path"), sql.As(got.C(generationoutput.FieldDeletedAt), "deleted_at")).
			GroupBy(s.C(generation.FieldID)).
			GroupBy(npt.C(negativeprompt.FieldText)).
			GroupBy(pt.C(prompt.FieldText)).
			GroupBy(got.C(generationoutput.FieldID)).
			GroupBy(got.C(generationoutput.FieldGalleryStatus)).
			GroupBy(got.C(generationoutput.FieldImagePath)).
			GroupBy(got.C(generationoutput.FieldUpscaledImagePath))
		// Order by generation, then output
		if filters == nil || (filters != nil && filters.Order == requests.SortOrderDescending) {
			s.OrderBy(sql.Desc(gt.C(generation.FieldCreatedAt)), sql.Desc(got.C(generationoutput.FieldCreatedAt)))
		} else {
			s.OrderBy(sql.Asc(gt.C(generation.FieldCreatedAt)), sql.Asc(got.C(generationoutput.FieldCreatedAt)))
		}
	}).Scan(r.Ctx, &gQueryResult)

	if err != nil {
		klog.Errorf("Error getting user generations: %v", err)
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
			klog.Warningf("Output ID is nil for generation, cannot include in result %v", g.ID)
			continue
		}
		gOutput := GenerationUpscaleOutput{
			ID:               *g.OutputID,
			ImageUrl:         g.ImageUrl,
			UpscaledImageUrl: g.UpscaledImageUrl,
			GalleryStatus:    g.GalleryStatus,
		}
		output := GenerationQueryWithOutputsResultFormatted{
			GenerationUpscaleOutput: gOutput,
			Generation: GenerationQueryWithOutputsData{
				ID:             g.ID,
				Height:         g.Height,
				Width:          g.Width,
				InferenceSteps: g.InferenceSteps,
				Seed:           g.Seed,
				Status:         g.Status,
				GuidanceScale:  g.GuidanceScale,
				SchedulerID:    g.SchedulerID,
				ModelID:        g.ModelID,
				CreatedAt:      g.CreatedAt,
				StartedAt:      g.StartedAt,
				CompletedAt:    g.CompletedAt,
				NegativePrompt: g.NegativePrompt,
				Prompt:         g.Prompt,
			},
		}
		generationOutputMap[g.ID] = append(generationOutputMap[g.ID], gOutput)
		meta.Outputs = append(meta.Outputs, output)
	}
	// Now loop through and add outputs to each generation
	for i, g := range meta.Outputs {
		meta.Outputs[i].Generation.Outputs = generationOutputMap[g.Generation.ID]
	}

	if cursor == nil {
		total, err := r.GetUserGenerationCountWithFilters(userID, filters)
		if err != nil {
			klog.Errorf("Error getting user generation count: %v", err)
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
}

// Paginated meta for querying generations
type GenerationQueryWithOutputsMeta struct {
	Total   *int                                        `json:"total_count,omitempty"`
	Outputs []GenerationQueryWithOutputsResultFormatted `json:"outputs"`
	Next    *time.Time                                  `json:"next,omitempty"`
}

type GenerationQueryWithOutputsData struct {
	ID             uuid.UUID                 `json:"id" sql:"id"`
	Height         int32                     `json:"height" sql:"height"`
	Width          int32                     `json:"width" sql:"width"`
	InferenceSteps int32                     `json:"inference_steps" sql:"inference_steps"`
	Seed           int                       `json:"seed" sql:"seed"`
	Status         string                    `json:"status" sql:"status"`
	GuidanceScale  float32                   `json:"guidance_scale" sql:"guidance_scale"`
	SchedulerID    uuid.UUID                 `json:"scheduler_id" sql:"scheduler_id"`
	ModelID        uuid.UUID                 `json:"model_id" sql:"model_id"`
	CreatedAt      time.Time                 `json:"created_at" sql:"created_at"`
	StartedAt      *time.Time                `json:"started_at,omitempty" sql:"started_at"`
	CompletedAt    *time.Time                `json:"completed_at,omitempty" sql:"completed_at"`
	NegativePrompt string                    `json:"negative_prompt" sql:"negative_prompt_text"`
	Prompt         string                    `json:"prompt" sql:"prompt_text"`
	Outputs        []GenerationUpscaleOutput `json:"outputs"`
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
