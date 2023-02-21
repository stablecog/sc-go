package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/ent/generation"
	"github.com/stablecog/sc-go/database/ent/generationoutput"
	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/utils"
	"k8s.io/klog/v2"
)

// CreateGeneration creates the initial generation in the database
// Takes in a userID (creator),  device info, countryCode, and a request body
func (r *Repository) CreateGeneration(userID uuid.UUID, deviceType, deviceOs, deviceBrowser, countryCode string, req requests.CreateGenerationRequest, DB *ent.Client) (*ent.Generation, error) {
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

func (r *Repository) SetGenerationFailed(generationID string, reason string, nsfwCount int32, db *ent.Client) error {
	if db == nil {
		db = r.DB
	}

	uid, err := uuid.Parse(generationID)
	if err != nil {
		klog.Errorf("Error parsing generation id in SetGenerationFailed %s: %v", generationID, err)
		return err
	}
	_, err = db.Generation.UpdateOneID(uid).SetStatus(generation.StatusFailed).SetFailureReason(reason).SetNsfwCount(nsfwCount).SetCompletedAt(time.Now()).Save(r.Ctx)
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
			parsedS3, err := utils.GetPathFromS3URL(output)
			if err != nil {
				klog.Errorf("Error parsing s3 url %s: %v", output, err)
				parsedS3 = output
			}
			gOutput, err := tx.GenerationOutput.Create().SetGenerationID(uid).SetImagePath(parsedS3).SetGalleryStatus(galleryStatus).Save(r.Ctx)
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
