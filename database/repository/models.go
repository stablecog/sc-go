package repository

import (
	"github.com/stablecog/go-apps/database/ent"
	"github.com/stablecog/go-apps/database/ent/generationmodel"
)

func (r *Repository) GetAllGenerationModels() ([]*ent.GenerationModel, error) {
	models, err := r.DB.GenerationModel.Query().Select(generationmodel.FieldID, generationmodel.FieldIsFree, generationmodel.FieldName).All(r.Ctx)
	if err != nil {
		return nil, err
	}

	return models, nil
}
