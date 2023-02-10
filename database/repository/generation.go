package repository

import (
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/stablecog/go-apps/database/ent"
	"github.com/stablecog/go-apps/database/ent/generation"
	"github.com/stablecog/go-apps/database/ent/generationmodel"
	"github.com/stablecog/go-apps/database/ent/generationoutput"
	"github.com/stablecog/go-apps/database/ent/negativeprompt"
	"github.com/stablecog/go-apps/database/ent/prompt"
	"github.com/stablecog/go-apps/database/ent/scheduler"
	"github.com/stablecog/go-apps/server/requests"
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
func (r *Repository) CreateGeneration(userID uuid.UUID, deviceType, deviceOs, deviceBrowser, countryCode string, req requests.GenerateRequestBody) (*ent.Generation, error) {
	// Get prompt, negative prompt, device info
	promptId, negativePromptId, deviceInfoId, err := r.GetOrCreateDeviceInfoAndPrompts(req.Prompt, req.NegativePrompt, deviceType, deviceOs, deviceBrowser)
	if err != nil {
		return nil, err
	}
	insert := r.DB.Generation.Create().
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
		SetShouldSubmitToGallery(req.ShouldSubmitToGallery)
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

func (r *Repository) SetGenerationFailed(generationID string, reason string) error {
	uid, err := uuid.Parse(generationID)
	if err != nil {
		klog.Errorf("Error parsing generation id in SetGenerationFailed %s: %v", generationID, err)
		return err
	}
	_, err = r.DB.Generation.UpdateOneID(uid).SetStatus(generation.StatusFailed).SetFailureReason(reason).Save(r.Ctx)
	if err != nil {
		klog.Errorf("Error setting generation failed %s: %v", generationID, err)
	}
	return err
}

func (r *Repository) SetGenerationSucceeded(generationID string, outputs []string) ([]*ent.GenerationOutput, error) {
	uid, err := uuid.Parse(generationID)
	if err != nil {
		klog.Errorf("Error parsing generation id in SetGenerationFailed %s: %v", generationID, err)
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
	_, err = tx.Generation.UpdateOneID(uid).SetStatus(generation.StatusSucceeded).SetCompletedAt(time.Now()).Save(r.Ctx)
	if err != nil {
		tx.Rollback()
		klog.Errorf("Error setting generation succeeded %s: %v", generationID, err)
		return nil, err
	}

	// If this generation was created with "submit_to_gallery", then submit all outputs to gallery
	var galleryStatus generationoutput.GalleryStatus
	if g.ShouldSubmitToGallery {
		galleryStatus = generationoutput.GalleryStatusSubmitted
	} else {
		galleryStatus = generationoutput.GalleryStatusNotSubmitted
	}

	// Insert all generation outputs
	var outputRet []*ent.GenerationOutput
	for _, output := range outputs {
		gOutput, err := tx.GenerationOutput.Create().SetGenerationID(uid).SetImageURL(output).SetGalleryStatus(galleryStatus).Save(r.Ctx)
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

// Get user generations from the database using page options
// Offset actually represents created_at, we paginate using this for performance reasons
// If present, we will get results after the offset (anything before, represents previous pages)
// ! TODO - this is currently two queries, 1 for outputs, 1 for everything else - figure out how to make it only 1
// ! using ent .With... doesn't use joins, so we construct our own query to make it more efficient
func (r *Repository) GetUserGenerations(userID uuid.UUID, per_page int, offset *time.Time) ([]UserGenerationQueryResult, error) {
	// Base fields to select in our query
	selectFields := []string{
		generation.FieldID,
		generation.FieldWidth,
		generation.FieldHeight,
		generation.FieldInferenceSteps,
		generation.FieldSeed,
		generation.FieldStatus,
		generation.FieldGuidanceScale,
		generation.FieldCreatedAt,
		generation.FieldStartedAt,
		generation.FieldCompletedAt,
	}
	var query *ent.GenerationQuery
	var res []UserGenerationQueryResult
	if offset != nil {
		query = r.DB.Generation.Query().Select(selectFields...).
			Where(generation.UserID(userID), generation.CreatedAtLT(*offset))
	} else {
		query = r.DB.Generation.Query().Select(selectFields...).
			Where(generation.UserID(userID))
	}
	query = query.Order(ent.Desc(generation.FieldCreatedAt)).
		Limit(per_page)

	// Join other data
	err := query.Modify(func(s *sql.Selector) {
		npt := sql.Table(negativeprompt.Table)
		pt := sql.Table(prompt.Table)
		st := sql.Table(scheduler.Table)
		got := sql.Table(generationoutput.Table)
		mt := sql.Table(generationmodel.Table)
		s.LeftJoin(npt).On(
			s.C(generation.FieldNegativePromptID), npt.C(negativeprompt.FieldID),
		).LeftJoin(pt).On(
			s.C(generation.FieldPromptID), pt.C(prompt.FieldID),
		).LeftJoin(st).On(
			s.C(generation.FieldSchedulerID), st.C(scheduler.FieldID),
		).LeftJoin(got).On(
			s.C(generation.FieldID), got.C(generationoutput.FieldGenerationID),
		).LeftJoin(mt).On(
			s.C(generation.FieldModelID), mt.C(generationmodel.FieldID),
		).AppendSelect(sql.As(npt.C(negativeprompt.FieldText), "negative_prompt_text"), sql.As(pt.C(prompt.FieldText), "prompt_text"), sql.As(st.C(scheduler.FieldName), "scheduler_name"), sql.As(mt.C(generationmodel.FieldName), "model_name")).
			GroupBy(s.C(generation.FieldID)).
			GroupBy(npt.C(negativeprompt.FieldText)).
			GroupBy(pt.C(prompt.FieldText)).
			GroupBy(st.C(scheduler.FieldName)).
			GroupBy(mt.C(generationmodel.FieldName))
	}).Scan(r.Ctx, &res)

	if err != nil {
		klog.Errorf("Error getting user generations: %v", err)
		return nil, err
	}

	if len(res) == 0 {
		return res, nil
	}

	// We want all IDs of the generations, then get their outputs, then append them to our resulting array
	var generationIDs []uuid.UUID
	for _, gen := range res {
		generationIDs = append(generationIDs, gen.ID)
	}
	outputs, err := r.DB.GenerationOutput.Query().Where(generationoutput.GenerationIDIn(generationIDs...)).All(r.Ctx)
	if err != nil {
		klog.Errorf("Error getting generation outputs: %v", err)
		return nil, err
	}

	// Append outputs to response matching generation ID
	for i, gen := range res {
		for _, output := range outputs {
			if gen.ID == output.GenerationID {
				res[i].Outputs = append(res[i].Outputs, UserGenerationOutputResult{
					ID:               output.ID,
					ImageUrl:         output.ImageURL,
					UpscaledImageUrl: output.UpscaledImageURL,
					GalleryStatus:    output.GalleryStatus,
				})
			}
		}
	}

	return res, err
}

type UserGenerationOutputResult struct {
	ID               uuid.UUID                      `json:"id"`
	ImageUrl         string                         `json:"image_url"`
	UpscaledImageUrl *string                        `json:"upscaled_image_url,omitempty"`
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
	CreatedAt      time.Time                    `json:"created_at" sql:"created_at"`
	StartedAt      *time.Time                   `json:"started_at,omitempty" sql:"started_at"`
	CompletedAt    *time.Time                   `json:"completed_at,omitempty" sql:"completed_at"`
	NegativePrompt string                       `json:"negative_prompt" sql:"negative_prompt_text"`
	Prompt         string                       `json:"prompt" sql:"prompt_text"`
	Scheduler      string                       `json:"scheduler" sql:"scheduler_name"`
	Model          string                       `json:"model" sql:"model_name"`
	Outputs        []UserGenerationOutputResult `json:"outputs"`
}
