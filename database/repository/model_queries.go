package repository

import (
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/ent/generationmodel"
	"github.com/stablecog/sc-go/database/ent/upscalemodel"
)

func (r *Repository) GetAllGenerationModels() ([]*ent.GenerationModel, error) {
	models, err := r.DB.GenerationModel.Query().Select(generationmodel.FieldID, generationmodel.FieldNameInWorker, generationmodel.FieldIsActive, generationmodel.FieldIsDefault, generationmodel.FieldIsHidden).All(r.Ctx)
	if err != nil {
		return nil, err
	}

	return models, nil
}

func (r *Repository) GetAllUpscaleModels() ([]*ent.UpscaleModel, error) {
	models, err := r.DB.UpscaleModel.Query().Select(upscalemodel.FieldID, upscalemodel.FieldNameInWorker, upscalemodel.FieldIsActive, upscalemodel.FieldIsDefault, upscalemodel.FieldIsHidden).All(r.Ctx)
	if err != nil {
		return nil, err
	}

	return models, nil
}
