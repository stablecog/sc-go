package repository

import (
	"github.com/stablecog/go-apps/database/ent"
	"github.com/stablecog/go-apps/database/ent/generationmodel"
	"github.com/stablecog/go-apps/database/ent/upscalemodel"
)

func (r *Repository) GetAllGenerationModels() ([]*ent.GenerationModel, error) {
	models, err := r.DB.GenerationModel.Query().Select(generationmodel.FieldID, generationmodel.FieldNameInWorker).All(r.Ctx)
	if err != nil {
		return nil, err
	}

	return models, nil
}

func (r *Repository) GetAllUpscaleModels() ([]*ent.UpscaleModel, error) {
	models, err := r.DB.UpscaleModel.Query().Select(upscalemodel.FieldID, upscalemodel.FieldNameInWorker).All(r.Ctx)
	if err != nil {
		return nil, err
	}

	return models, nil
}
