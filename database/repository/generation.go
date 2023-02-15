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
	// Start a transaction
	tx, err := r.DB.Tx(r.Ctx)
	if err != nil {
		klog.Errorf("Error starting transaction in SetGenerationSucceeded %s: %v", generationID, err)
		return nil, err
	}

	// Retrieve the generation
	g, err := r.GetGeneration(uid)
	if err != nil {
		tx.Rollback()
		klog.Errorf("Error retrieving generation %s: %v", generationID, err)
		return nil, err
	}

	// Update the generation
	_, err = tx.Generation.UpdateOneID(uid).SetStatus(generation.StatusSucceeded).SetCompletedAt(time.Now()).SetNsfwCount(nsfwCount).Save(r.Ctx)
	if err != nil {
		tx.Rollback()
		klog.Errorf("Error setting generation succeeded %s: %v", generationID, err)
		return nil, err
	}

	// If this generation was created with "submit_to_gallery", then submit all outputs to gallery
	var galleryStatus generationoutput.GalleryStatus
	if g.SubmitToGallery {
		galleryStatus = generationoutput.GalleryStatusSubmitted
	} else {
		galleryStatus = generationoutput.GalleryStatusNotSubmitted
	}

	// Insert all generation outputs
	var outputRet []*ent.GenerationOutput
	for _, output := range outputs {
		gOutput, err := tx.GenerationOutput.Create().SetGenerationID(uid).SetImagePath(output).SetGalleryStatus(galleryStatus).Save(r.Ctx)
		if err != nil {
			tx.Rollback()
			klog.Errorf("Error inserting generation output %s: %v", generationID, err)
			return nil, err
		}
		outputRet = append(outputRet, gOutput)
	}
	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		klog.Errorf("Error committing transaction in SetGenerationSucceeded %s: %v", generationID, err)
		return nil, err
	}
	return outputRet, nil
}

func (r *Repository) GetUserGenerationsCount(userID uuid.UUID) (int, error) {
	return r.DB.Generation.Query().Where(generation.UserIDEQ(userID)).Count(r.Ctx)
}

// Get user generations from the database using page options
// Cursot actually represents created_at, we paginate using this for performance reasons
// If present, we will get results after the cursor (anything before, represents previous pages)
// TODO - this is currently two queries, 1 for outputs, 1 for everything else - figure out how to make it only 1
// ! using ent .With... doesn't use joins, so we construct our own query to make it more efficient
// TODO - Define indexes for this query
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

	if cursor != nil {
		query = r.DB.Generation.Query().Select(selectFields...).
			Where(generation.UserID(userID), generation.CreatedAtLT(*cursor), generation.StatusIn(status...))
	} else {
		query = r.DB.Generation.Query().Select(selectFields...).
			Where(generation.UserID(userID), generation.StatusIn(status...))
	}

	if filters != nil {
		// Apply filters
		if len(filters.ModelIDs) > 0 {
			query = query.Where(generation.ModelIDIn(filters.ModelIDs...))
		}
		if len(filters.SchedulerIDs) > 0 {
			query = query.Where(generation.SchedulerIDIn(filters.SchedulerIDs...))
		}
		// Apply OR if both are present
		// Confusing, but example of what we want to do:
		// If min_width=100, max_width=200, widths=[300,400]
		// We want to query like; WHERE (width >= 100 AND width <= 200) OR width IN (300,400)
		if (filters.MinWidth != 0 || filters.MaxWidth != 0) && len(filters.Widths) > 0 {
			if filters.MinWidth != 0 && filters.MaxWidth != 0 {
				query = query.Where(generation.Or(generation.And(generation.WidthGTE(filters.MinWidth), generation.WidthLTE(filters.MaxWidth)), generation.WidthIn(filters.Widths...)))
			} else if filters.MinWidth != 0 {
				query = query.Where(generation.Or(generation.WidthGTE(filters.MinWidth), generation.WidthIn(filters.Widths...)))
			} else {
				query = query.Where(generation.Or(generation.WidthLTE(filters.MaxWidth), generation.WidthIn(filters.Widths...)))
			}
		} else {
			if filters.MinWidth != 0 {
				query = query.Where(generation.WidthGTE(filters.MinWidth))
			}
			if filters.MaxWidth != 0 {
				query = query.Where(generation.WidthLTE(filters.MaxWidth))
			}
			if len(filters.Widths) > 0 {
				query = query.Where(generation.WidthIn(filters.Widths...))
			}
		}

		// Height
		if (filters.MinHeight != 0 || filters.MaxHeight != 0) && len(filters.Heights) > 0 {
			if filters.MinHeight != 0 && filters.MaxHeight != 0 {
				query = query.Where(generation.Or(generation.And(generation.HeightGTE(filters.MinHeight), generation.HeightLTE(filters.MaxHeight)), generation.HeightIn(filters.Heights...)))
			} else if filters.MinHeight != 0 {
				query = query.Where(generation.Or(generation.HeightGTE(filters.MinHeight), generation.HeightIn(filters.Heights...)))
			} else {
				query = query.Where(generation.Or(generation.HeightLTE(filters.MaxHeight), generation.HeightIn(filters.Heights...)))
			}
		} else {
			if len(filters.Heights) > 0 {
				query = query.Where(generation.HeightIn(filters.Heights...))
			}
			if filters.MaxHeight != 0 {
				query = query.Where(generation.HeightLTE(filters.MaxHeight))
			}
			if filters.MinHeight != 0 {
				query = query.Where(generation.HeightGTE(filters.MinHeight))
			}
		}

		// Inference steps
		if (filters.MinInferenceSteps != 0 || filters.MaxInferenceSteps != 0) && len(filters.InferenceSteps) > 0 {
			if filters.MinInferenceSteps != 0 && filters.MaxInferenceSteps != 0 {
				query = query.Where(generation.Or(generation.And(generation.InferenceStepsGTE(filters.MinInferenceSteps), generation.InferenceStepsLTE(filters.MaxInferenceSteps)), generation.InferenceStepsIn(filters.InferenceSteps...)))
			} else if filters.MinInferenceSteps != 0 {
				query = query.Where(generation.Or(generation.InferenceStepsGTE(filters.MinInferenceSteps), generation.InferenceStepsIn(filters.InferenceSteps...)))
			} else {
				query = query.Where(generation.Or(generation.InferenceStepsLTE(filters.MaxInferenceSteps), generation.InferenceStepsIn(filters.InferenceSteps...)))
			}
		} else {
			if len(filters.InferenceSteps) > 0 {
				query = query.Where(generation.InferenceStepsIn(filters.InferenceSteps...))
			}
			if filters.MaxInferenceSteps != 0 {
				query = query.Where(generation.InferenceStepsLTE(filters.MaxInferenceSteps))
			}
			if filters.MinInferenceSteps != 0 {
				query = query.Where(generation.InferenceStepsGTE(filters.MinInferenceSteps))
			}
		}

		// Guidance Scales
		if (filters.MinGuidanceScale != 0 || filters.MaxGuidanceScale != 0) && len(filters.GuidanceScales) > 0 {
			if filters.MinGuidanceScale != 0 && filters.MaxGuidanceScale != 0 {
				query = query.Where(generation.Or(generation.And(generation.GuidanceScaleGTE(filters.MinGuidanceScale), generation.GuidanceScaleLTE(filters.MaxGuidanceScale)), generation.GuidanceScaleIn(filters.GuidanceScales...)))
			} else if filters.MinGuidanceScale != 0 {
				query = query.Where(generation.Or(generation.GuidanceScaleGTE(filters.MinGuidanceScale), generation.GuidanceScaleIn(filters.GuidanceScales...)))
			} else {
				query = query.Where(generation.Or(generation.GuidanceScaleLTE(filters.MaxGuidanceScale), generation.GuidanceScaleIn(filters.GuidanceScales...)))
			}
		} else {
			if len(filters.GuidanceScales) > 0 {
				query = query.Where(generation.GuidanceScaleIn(filters.GuidanceScales...))
			}
			if filters.MaxGuidanceScale != 0 {
				query = query.Where(generation.GuidanceScaleLTE(filters.MaxGuidanceScale))
			}
			if filters.MinGuidanceScale != 0 {
				query = query.Where(generation.GuidanceScaleGTE(filters.MinGuidanceScale))
			}
		}
	}

	if filters == nil || (filters != nil && filters.Order == requests.UserGenerationQueryOrderDescending) {
		query = query.Order(ent.Desc(generation.FieldCreatedAt)).
			Limit(per_page + 1)
	} else {
		query = query.Order(ent.Asc(generation.FieldCreatedAt)).Limit(per_page + 1)
	}

	// Join other data
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
		).AppendSelect(sql.As(npt.C(negativeprompt.FieldText), "negative_prompt_text"), sql.As(pt.C(prompt.FieldText), "prompt_text")).
			GroupBy(s.C(generation.FieldID)).
			GroupBy(npt.C(negativeprompt.FieldText)).
			GroupBy(pt.C(prompt.FieldText))
	}).Scan(r.Ctx, &gQueryResult)

	if err != nil {
		klog.Errorf("Error getting user generations: %v", err)
		return nil, err
	}

	if len(gQueryResult) == 0 {
		meta := &UserGenerationQueryMeta{
			Generations: []UserGenerationQueryResult{},
		}
		// Only give total if we have no cursor
		if cursor == nil {
			zero := 0
			meta.Total = &zero
		}
		return meta, nil
	}

	// We want all IDs of the generations, then get their outputs, then append them to our resulting array
	var generationIDs []uuid.UUID
	for _, gen := range gQueryResult {
		generationIDs = append(generationIDs, gen.ID)
	}
	outputs, err := r.DB.GenerationOutput.Query().Where(generationoutput.GenerationIDIn(generationIDs...)).All(r.Ctx)
	if err != nil {
		klog.Errorf("Error getting generation outputs: %v", err)
		return nil, err
	}

	meta := &UserGenerationQueryMeta{}
	if len(gQueryResult) > per_page {
		// Remove last item
		gQueryResult = gQueryResult[:len(gQueryResult)-1]
		meta.Next = &gQueryResult[len(gQueryResult)-1].CreatedAt
	}

	// Append outputs to response matching generation ID
	for i, gen := range gQueryResult {
		for _, output := range outputs {
			if gen.ID == output.GenerationID {
				// Parse S3 URLs to usable URLs
				imageUrl, err := utils.ParseS3UrlToURL(output.ImagePath)
				if err != nil {
					klog.Errorf("Error parsing image url %s: %v", output.ImagePath, err)
					imageUrl = output.ImagePath
				}
				var upscaledImageUrl string
				if output.UpscaledImagePath != nil {
					upscaledImageUrl, err = utils.ParseS3UrlToURL(*output.UpscaledImagePath)
					if err != nil {
						klog.Errorf("Error parsing upscaled image url %s: %v", *output.UpscaledImagePath, err)
						upscaledImageUrl = *output.UpscaledImagePath
					}
				}
				gQueryResult[i].Outputs = append(gQueryResult[i].Outputs, UserGenerationOutputResult{
					ID:               output.ID,
					ImageUrl:         imageUrl,
					UpscaledImageUrl: upscaledImageUrl,
					GalleryStatus:    output.GalleryStatus,
				})
			}
		}
	}

	meta.Generations = gQueryResult

	if cursor == nil {
		total, err := r.GetUserGenerationsCount(userID)
		if err != nil {
			klog.Errorf("Error getting user generation count: %v", err)
			return nil, err
		}
		meta.Total = &total
	}

	return meta, err
}

type UserGenerationQueryMeta struct {
	Total       *int                        `json:"total_count,omitempty"`
	Generations []UserGenerationQueryResult `json:"generations"`
	Next        *time.Time                  `json:"next,omitempty"`
}

type UserGenerationOutputResult struct {
	ID               uuid.UUID                      `json:"id"`
	ImageUrl         string                         `json:"image_url"`
	UpscaledImageUrl string                         `json:"upscaled_image_url,omitempty"`
	GalleryStatus    generationoutput.GalleryStatus `json:"gallery_status"`
}

type UserGenerationQueryResult struct {
	ID             uuid.UUID                    `json:"id" sql:"id"`
	Height         int32                        `json:"height" sql:"height"`
	Width          int32                        `json:"width" sql:"width"`
	InferenceSteps int32                        `json:"inference_steps" sql:"inference_steps"`
	Seed           int                          `json:"seed" sql:"seed"`
	Status         string                       `json:"status" sql:"status"`
	GuidanceScale  float32                      `json:"guidance_scale" sql:"guidance_scale"`
	SchedulerID    uuid.UUID                    `json:"scheduler_id" sql:"scheduler_id"`
	ModelID        uuid.UUID                    `json:"model_id" sql:"model_id"`
	CreatedAt      time.Time                    `json:"created_at" sql:"created_at"`
	StartedAt      *time.Time                   `json:"started_at,omitempty" sql:"started_at"`
	CompletedAt    *time.Time                   `json:"completed_at,omitempty" sql:"completed_at"`
	NegativePrompt string                       `json:"negative_prompt" sql:"negative_prompt_text"`
	Prompt         string                       `json:"prompt" sql:"prompt_text"`
	Outputs        []UserGenerationOutputResult `json:"outputs"`
}
