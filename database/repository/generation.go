package repository

import (
	"github.com/google/uuid"
	"github.com/stablecog/go-apps/database/ent"
	"github.com/stablecog/go-apps/database/ent/generation"
	"github.com/stablecog/go-apps/models"
)

// CreateGeneration creates the initial generation in the database
// Takes in a userID (creator),  device info, countryCode, and a request body
func (r *Repository) CreateGeneration(userID uuid.UUID, deviceType, deviceOs, deviceBrowser, countryCode string, req models.GenerateRequestBody) (*ent.Generation, error) {
	// Get prompt, negative prompt, device info
	promptId, negativePromptId, deviceInfoId, err := r.GetOrCreateDeviceInfoAndPrompts(req.Prompt, req.NegativePrompt, deviceType, deviceOs, deviceBrowser)
	if err != nil {
		return nil, err
	}
	insert := r.DB.Generation.Create().
		SetStatus(generation.StatusStarted).
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
