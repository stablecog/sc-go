package repository

import (
	"github.com/stablecog/sc-go/database/ent/generationmodel"
	"github.com/stablecog/sc-go/database/ent/upscalemodel"
)

func (r *Repository) EnableRunpodServerless() error {
	errGen := r.DB.GenerationModel.Update().Where(
		generationmodel.RunpodEndpointNotNil(),
		generationmodel.RunpodActive(false),
	).SetRunpodActive(true).Exec(r.Ctx)

	if errGen != nil {
		return errGen
	}

	errUps := r.DB.UpscaleModel.Update().Where(
		upscalemodel.RunpodEndpointNotNil(),
		upscalemodel.RunpodActive(false),
	).SetRunpodActive(true).Exec(r.Ctx)

	if errUps != nil {
		return errUps
	}

	return nil
}

func (r *Repository) DisableRunpodServerless() error {
	errGen := r.DB.GenerationModel.Update().Where(
		generationmodel.RunpodEndpointNotNil(),
		generationmodel.RunpodActive(true),
	).SetRunpodActive(false).Exec(r.Ctx)

	if errGen != nil {
		return errGen
	}

	errUps := r.DB.UpscaleModel.Update().Where(
		upscalemodel.RunpodEndpointNotNil(),
		upscalemodel.RunpodActive(true),
	).SetRunpodActive(false).Exec(r.Ctx)

	if errUps != nil {
		return errUps
	}

	return nil
}

func (r *Repository) IsRunpodServerlessActive() (bool, error) {
	genCount, errGen := r.DB.GenerationModel.Query().Where(
		generationmodel.RunpodEndpointNotNil(),
		generationmodel.RunpodActive(false),
	).Count(r.Ctx)

	if errGen != nil {
		return false, errGen
	}

	upsCount, errUps := r.DB.UpscaleModel.Query().Where(
		upscalemodel.RunpodEndpointNotNil(),
		upscalemodel.RunpodActive(false),
	).Count(r.Ctx)

	if errUps != nil {
		return false, errUps
	}

	return genCount == 0 && upsCount == 0, nil
}
