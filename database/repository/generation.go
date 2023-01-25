package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/stablecog/go-apps/database/ent"
	"github.com/stablecog/go-apps/database/ent/generation"
	"github.com/stablecog/go-apps/server/requests"
	"k8s.io/klog/v2"
)

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
		SetNumInterferenceSteps(req.NumInferenceSteps).
		SetSeed(req.Seed).
		SetModelID(req.ModelId).
		SetSchedulerID(req.SchedulerId).
		SetPromptID(promptId).
		SetDeviceInfoID(deviceInfoId).
		SetCountryCode(countryCode).
		SetUserID(userID)
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
