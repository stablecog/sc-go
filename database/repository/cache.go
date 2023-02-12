package repository

import (
	"github.com/stablecog/sc-go/shared"
	"k8s.io/klog/v2"
)

// Update the cache from the database
func (r *Repository) UpdateCache() error {
	generationModels, err := r.GetAllGenerationModels()
	if err != nil {
		klog.Fatalf("Failed to get generation_models: %v", err)
		return err
	}
	shared.GetCache().UpdateGenerationModels(generationModels)

	upscaleModels, err := r.GetAllUpscaleModels()
	if err != nil {
		klog.Fatalf("Failed to get upscale_models: %v", err)
		return err
	}
	shared.GetCache().UpdateUpscaleModels(upscaleModels)

	schedulers, err := r.GetAllSchedulers()
	if err != nil {
		klog.Fatalf("Failed to get schedulers: %v", err)
		return err
	}
	shared.GetCache().UpdateSchedulers(schedulers)
	return nil
}
