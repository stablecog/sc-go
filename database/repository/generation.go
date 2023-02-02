package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/stablecog/go-apps/database/ent"
	"github.com/stablecog/go-apps/database/ent/generation"
	"github.com/stablecog/go-apps/server/requests"
	"k8s.io/klog/v2"
)

// Get generation by ID
func (r *Repository) GetGeneration(id uuid.UUID) (*ent.Generation, error) {
	return r.DB.Generation.Query().Where(generation.IDEQ(id)).First(r.Ctx)
}

// CreateGeneration creates the initial generation in the database
// Takes in a userID (creator),  device info, countryCode, and a request body
func (r *Repository) CreateGeneration(userID uuid.UUID, deviceType, deviceOs, deviceBrowser, countryCode string, req requests.GenerateRequestBody) (*ent.Generation, error) {
	// Get prompt, negative prompt, device info
	promptId, negativePromptId, deviceInfoId, err := r.GetOrCreateDeviceInfoAndPrompts(req.Prompt, req.NegativePrompt, deviceType, deviceOs, deviceBrowser)
	if err != nil {
		return nil, err
	}
	// Gallery status depends on req body
	galleryStatus := generation.GalleryStatusNotSubmitted
	if req.ShouldSubmitToGallery {
		galleryStatus = generation.GalleryStatusSubmitted
	}
	insert := r.DB.Generation.Create().
		SetStatus(generation.StatusQueued).
		SetWidth(req.Width).
		SetHeight(req.Height).
		SetGuidanceScale(req.GuidanceScale).
		SetInferenceSteps(req.NumInferenceSteps).
		SetSeed(req.Seed).
		SetModelID(req.ModelId).
		SetSchedulerID(req.SchedulerId).
		SetPromptID(promptId).
		SetDeviceInfoID(deviceInfoId).
		SetCountryCode(countryCode).
		SetUserID(userID).
		SetGalleryStatus(galleryStatus)
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
	_, err = r.DB.Generation.UpdateOneID(uid).SetStatus(generation.StatusStarted).SetStartedAt(time.Now()).Save(r.Ctx)
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

func (r *Repository) SetGenerationSucceeded(generationID string, outputs []string) error {
	uid, err := uuid.Parse(generationID)
	if err != nil {
		klog.Errorf("Error parsing generation id in SetGenerationFailed %s: %v", generationID, err)
		return err
	}
	// Start a transaction
	tx, err := r.DB.Tx(r.Ctx)
	if err != nil {
		klog.Errorf("Error starting transaction in SetGenerationSucceeded %s: %v", generationID, err)
		return err
	}
	// Update the generation
	_, err = tx.Generation.UpdateOneID(uid).SetStatus(generation.StatusSucceeded).SetCompletedAt(time.Now()).Save(r.Ctx)
	if err != nil {
		tx.Rollback()
		klog.Errorf("Error setting generation succeeded %s: %v", generationID, err)
		return err
	}

	// Insert all generation outputs
	for _, output := range outputs {
		_, err = tx.GenerationOutput.Create().SetGenerationID(uid).SetImageURL(output).Save(r.Ctx)
		if err != nil {
			tx.Rollback()
			klog.Errorf("Error inserting generation output %s: %v", generationID, err)
			return err
		}
	}
	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		klog.Errorf("Error committing transaction in SetGenerationSucceeded %s: %v", generationID, err)
		return err
	}
	return nil
}

// Get user generations from the database using page options
// Offset actually represents created_at, we paginate using this for performance reasons
// If present, we will get results after the offset (anything before, represents previous pages)
func (r *Repository) GetUserGenerations(userID uuid.UUID, per_page int, offset *time.Time) ([]*ent.Generation, error) {
	var query *ent.GenerationQuery
	if offset != nil {
		query = r.DB.Generation.Query().
			Where(generation.UserID(userID), generation.CreatedAtLT(*offset))
	} else {
		query = r.DB.Generation.Query().
			Where(generation.UserID(userID))
	}
	return query.WithPrompt().
		WithNegativePrompt().
		WithScheduler().
		WithGenerationOutputs().
		WithGenerationModel().
		Order(ent.Desc(generation.FieldCreatedAt)).
		Limit(per_page).
		All(r.Ctx)
}
