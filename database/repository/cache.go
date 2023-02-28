package repository

import (
	"log"

	"github.com/stablecog/sc-go/shared"
)

// Update the cache from the database
func (r *Repository) UpdateCache() error {
	generationModels, err := r.GetAllGenerationModels()
	if err != nil {
		log.Fatal("Failed to get generation_models", "err", err)
		return err
	}
	shared.GetCache().UpdateGenerationModels(generationModels)

	upscaleModels, err := r.GetAllUpscaleModels()
	if err != nil {
		log.Fatal("Failed to get upscale_models", "err", err)
		return err
	}
	shared.GetCache().UpdateUpscaleModels(upscaleModels)

	schedulers, err := r.GetAllSchedulers()
	if err != nil {
		log.Fatal("Failed to get schedulers", "err", err)
		return err
	}
	shared.GetCache().UpdateSchedulers(schedulers)

	admins, err := r.GetSuperAdminUserIDs()
	if err != nil {
		log.Fatal("Failed to get super admins", "err", err)
		return err
	}
	shared.GetCache().SetAdminUUIDs(admins)
	return nil
}
