package repository

import (
	"github.com/google/uuid"
	"github.com/stablecog/go-apps/database/ent"
	"github.com/stablecog/go-apps/database/ent/generation"
	"github.com/stablecog/go-apps/models"
)

// CreateGeneration creates the initial generation in the database
// Takes in a userID (creator), and a request body
func (r *Repository) CreateGeneration(userID uuid.UUID, req models.GenerateRequestBody) (*ent.Generation, error) {
	return r.DB.Generation.Create().
		SetStatus(generation.StatusStarted).
		SetWidth(req.Width).
		SetHeight(req.Height).
		SetGuidanceScale(req.GuidanceScale).
		SetNumInterferenceSteps(req.NumInferenceSteps).
		SetSeed(req.Seed).
		SetModelID(req.ModelId).
		SetSchedulerID(req.SchedulerId).
		SetUserID(userID).Save(r.Ctx)
}
