package repository

import (
	"database/sql/driver"
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

// CreateGeneration creates the initial generation in the database
// Takes in a userID (creator),  device info, countryCode, and a request body
func (r *Repository) CreateGeneration(userID uuid.UUID, deviceType, deviceOs, deviceBrowser, countryCode string, req requests.GenerateRequestBody, DB *ent.Client) (*ent.Generation, error) {
	if DB == nil {
		DB = r.DB
	}
	// Get prompt, negative prompt, device info
	promptId, negativePromptId, deviceInfoId, err := r.GetOrCreateDeviceInfoAndPrompts(req.Prompt, req.NegativePrompt, deviceType, deviceOs, deviceBrowser, DB)
	if err != nil {
		return nil, err
	}
	insert := DB.Generation.Create().
		SetStatus(generation.StatusQueued).
		SetWidth(req.Width).
		SetHeight(req.Height).
		SetGuidanceScale(req.GuidanceScale).
		SetInferenceSteps(req.InferenceSteps).
		SetSeed(req.Seed).
		SetModelID(req.ModelId).
		SetSchedulerID(req.SchedulerId).
		SetPromptID(promptId).
		SetDeviceInfoID(deviceInfoId).
		SetCountryCode(countryCode).
		SetUserID(userID).
		SetSubmitToGallery(req.SubmitToGallery).
		SetNumOutputs(req.NumOutputs)
	if negativePromptId != nil {
		insert.SetNegativePromptID(*negativePromptId)
	}
	return insert.Save(r.Ctx)
}

func (r *Repository) SetGenerationStarted(generationID string) error {
	uid, err := uuid.Parse(generationID)
	if err != nil {
		klog.Errorf("Error parsing generation id in SetGenerationStarted %s: %v", generationID, err)
		return err
	}
	_, err = r.DB.Generation.Update().Where(generation.IDEQ(uid), generation.StatusEQ(generation.StatusQueued)).SetStatus(generation.StatusStarted).SetStartedAt(time.Now()).Save(r.Ctx)
	if err != nil {
		// Log error here since this might be happening in a goroutine
		klog.Errorf("Error setting generation started %s: %v", generationID, err)
	}
	return err
}

func (r *Repository) SetGenerationFailed(generationID string, reason string, nsfwCount int32) error {
	uid, err := uuid.Parse(generationID)
	if err != nil {
		klog.Errorf("Error parsing generation id in SetGenerationFailed %s: %v", generationID, err)
		return err
	}
	_, err = r.DB.Generation.UpdateOneID(uid).SetStatus(generation.StatusFailed).SetFailureReason(reason).SetNsfwCount(nsfwCount).Save(r.Ctx)
	if err != nil {
		klog.Errorf("Error setting generation failed %s: %v", generationID, err)
	}
	return err
}

func (r *Repository) SetGenerationSucceeded(generationID string, outputs []string, nsfwCount int32) ([]*ent.GenerationOutput, error) {
	uid, err := uuid.Parse(generationID)
	if err != nil {
		klog.Errorf("Error parsing generation id in SetGenerationSucceeded %s: %v", generationID, err)
		return nil, err
	}

	var outputRet []*ent.GenerationOutput

	// Wrap in transaction
	if err := r.WithTx(func(tx *ent.Tx) error {
		if err != nil {
			klog.Errorf("Error starting transaction in SetGenerationSucceeded %s: %v", generationID, err)
			return err
		}

		// Retrieve the generation
		g, err := r.GetGeneration(uid)
		if err != nil {
			klog.Errorf("Error retrieving generation %s: %v", generationID, err)
			return err
		}

		// Update the generation
		_, err = tx.Generation.UpdateOneID(uid).SetStatus(generation.StatusSucceeded).SetCompletedAt(time.Now()).SetNsfwCount(nsfwCount).Save(r.Ctx)
		if err != nil {
			klog.Errorf("Error setting generation succeeded %s: %v", generationID, err)
			return err
		}

		// If this generation was created with "submit_to_gallery", then submit all outputs to gallery
		var galleryStatus generationoutput.GalleryStatus
		if g.SubmitToGallery {
			galleryStatus = generationoutput.GalleryStatusSubmitted
		} else {
			galleryStatus = generationoutput.GalleryStatusNotSubmitted
		}

		// Insert all generation outputs
		for _, output := range outputs {
			gOutput, err := tx.GenerationOutput.Create().SetGenerationID(uid).SetImagePath(output).SetGalleryStatus(galleryStatus).Save(r.Ctx)
			if err != nil {
				klog.Errorf("Error inserting generation output %s: %v", generationID, err)
				return err
			}
			outputRet = append(outputRet, gOutput)
		}

		return nil
	}); err != nil {
		klog.Errorf("Error starting transaction in SetGenerationSucceeded %s: %v", generationID, err)
		return nil, err
	}

	return outputRet, nil
}

// Apply all filters to root ent query
func (r *Repository) ApplyUserGenerationsFilters(query *ent.GenerationQuery, filters *requests.UserGenerationFilters) *ent.GenerationQuery {
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
	}
	return resQuery
}

// Gets the count of generations with outputs user has with filters
func (r *Repository) GetUserGenerationCountWithFilters(userID uuid.UUID, filters *requests.UserGenerationFilters) (int, error) {
	var query *ent.GenerationQuery

	// Parse status from query
	status := []generation.Status{}
	if filters != nil && filters.SucceededOnly {
		status = append(status, generation.StatusSucceeded)
	} else {
		status = append(status, generation.StatusSucceeded, generation.StatusFailed, generation.StatusQueued, generation.StatusStarted)
	}

	query = r.DB.Generation.Query().
		Where(func(s *sql.Selector) {
			t := sql.Table(generation.Table)
			statusValues := make([]driver.Value, 0, len(status))
			for _, v := range status {
				statusValues = append(statusValues, v)
			}
			predicates := []*sql.Predicate{
				sql.EQ(t.C(generation.FieldUserID), userID),
				sql.InValues(t.C(generation.FieldStatus), statusValues...),
			}
			// Also filter if necessary
			if filters != nil && filters.UpscaleStatus == requests.UserGenerationQueryUpscaleStatusNot {
				predicates = append(predicates, sql.IsNull("upscaled_image_path"))
			} else if filters != nil && filters.UpscaleStatus == requests.UserGenerationQueryUpscaleStatusOnly {
				predicates = append(predicates, sql.NotNull("upscaled_image_path"))
			}
			// Apply
			s.Where(sql.And(predicates...))
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
func (r *Repository) GetUserGenerations(userID uuid.UUID, per_page int, cursor *time.Time, filters *requests.UserGenerationFilters) (*UserGenerationQueryMeta, error) {
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
	var gQueryResult []UserGenerationQueryResult

	// Parse status from query
	status := []generation.Status{}
	if filters != nil && filters.SucceededOnly {
		status = append(status, generation.StatusSucceeded)
	} else {
		status = append(status, generation.StatusSucceeded, generation.StatusFailed, generation.StatusQueued, generation.StatusStarted)
	}

	query = r.DB.Generation.Query().Select(selectFields...).Where(func(s *sql.Selector) {
		t := sql.Table(generation.Table)
		statusValues := make([]driver.Value, 0, len(status))
		for _, v := range status {
			statusValues = append(statusValues, v)
		}
		predicates := []*sql.Predicate{
			sql.EQ(t.C(generation.FieldUserID), userID),
			sql.InValues(t.C(generation.FieldStatus), statusValues...),
		}
		// Apply cursor if necessary
		if cursor != nil {
			predicates = append(predicates, sql.LT(t.C(generation.FieldCreatedAt), *cursor))
		}
		// Also filter if necessary
		if filters != nil && filters.UpscaleStatus == requests.UserGenerationQueryUpscaleStatusNot {
			predicates = append(predicates, sql.IsNull("upscaled_image_path"))
		} else if filters != nil && filters.UpscaleStatus == requests.UserGenerationQueryUpscaleStatusOnly {
			predicates = append(predicates, sql.NotNull("upscaled_image_path"))
		}
		// Apply
		s.Where(sql.And(predicates...))
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
		).AppendSelect(sql.As(npt.C(negativeprompt.FieldText), "negative_prompt_text"), sql.As(pt.C(prompt.FieldText), "prompt_text"), sql.As(got.C(generationoutput.FieldID), "output_id"), sql.As(got.C(generationoutput.FieldGalleryStatus), "output_gallery_status"), sql.As(got.C(generationoutput.FieldImagePath), "image_path"), sql.As(got.C(generationoutput.FieldUpscaledImagePath), "upscaled_image_path")).GroupBy(s.C(generation.FieldID)).
			GroupBy(npt.C(negativeprompt.FieldText)).
			GroupBy(pt.C(prompt.FieldText)).
			GroupBy(got.C(generationoutput.FieldID)).
			GroupBy(got.C(generationoutput.FieldGalleryStatus)).
			GroupBy(got.C(generationoutput.FieldImagePath)).
			GroupBy(got.C(generationoutput.FieldUpscaledImagePath))
		// Order by generation, then output
		if filters == nil || (filters != nil && filters.Order == requests.UserGenerationQueryOrderDescending) {
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
		meta := &UserGenerationQueryMeta{
			Outputs: []UserGenerationQueryResult{},
		}
		// Only give total if we have no cursor
		if cursor == nil {
			zero := 0
			meta.Total = &zero
		}
		return meta, nil
	}

	meta := &UserGenerationQueryMeta{}
	if len(gQueryResult) > per_page {
		// Remove last item
		gQueryResult = gQueryResult[:len(gQueryResult)-1]
		meta.Next = &gQueryResult[len(gQueryResult)-1].CreatedAt
	}

	// Get real image URLs for each
	for i, g := range gQueryResult {
		if g.ImageUrl != "" {
			parsed, err := utils.ParseS3UrlToURL(g.ImageUrl)
			if err != nil {
				parsed = g.ImageUrl
			}
			gQueryResult[i].ImageUrl = parsed
		}
		if g.UpscaledImageUrl != "" {
			parsed, err := utils.ParseS3UrlToURL(g.UpscaledImageUrl)
			if err != nil {
				parsed = g.UpscaledImageUrl
			}
			gQueryResult[i].UpscaledImageUrl = parsed
		}
	}

	meta.Outputs = gQueryResult

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

type UserGenerationQueryMeta struct {
	Total   *int                        `json:"total_count,omitempty"`
	Outputs []UserGenerationQueryResult `json:"outputs"`
	Next    *time.Time                  `json:"next,omitempty"`
}

type UserGenerationQueryResult struct {
	ID               uuid.UUID                      `json:"id" sql:"id"`
	Height           int32                          `json:"height" sql:"height"`
	Width            int32                          `json:"width" sql:"width"`
	InferenceSteps   int32                          `json:"inference_steps" sql:"inference_steps"`
	Seed             int                            `json:"seed" sql:"seed"`
	Status           string                         `json:"status" sql:"status"`
	GuidanceScale    float32                        `json:"guidance_scale" sql:"guidance_scale"`
	SchedulerID      uuid.UUID                      `json:"scheduler_id" sql:"scheduler_id"`
	ModelID          uuid.UUID                      `json:"model_id" sql:"model_id"`
	CreatedAt        time.Time                      `json:"created_at" sql:"created_at"`
	StartedAt        *time.Time                     `json:"started_at,omitempty" sql:"started_at"`
	CompletedAt      *time.Time                     `json:"completed_at,omitempty" sql:"completed_at"`
	NegativePrompt   string                         `json:"negative_prompt" sql:"negative_prompt_text"`
	Prompt           string                         `json:"prompt" sql:"prompt_text"`
	OutputID         *uuid.UUID                     `json:"output_id,omitempty" sql:"output_id"`
	ImageUrl         string                         `json:"image_url,omitempty" sql:"image_path"`
	UpscaledImageUrl string                         `json:"upscaled_image_url,omitempty" sql:"upscaled_image_path"`
	GalleryStatus    generationoutput.GalleryStatus `json:"gallery_status,omitempty" sql:"output_gallery_status"`
}
