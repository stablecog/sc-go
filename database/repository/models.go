package repository

import "github.com/stablecog/go-apps/database/ent/generationmodel"

func (r *Repository) GetFreeGeneratedModelIDs() ([]string, error) {
	models, err := r.DB.GenerationModel.Query().Where(generationmodel.IsFree(true)).All(r.Ctx)
	if err != nil {
		return nil, err
	}

	ids := make([]string, len(models))
	for i, model := range models {
		ids[i] = model.ID.String()
	}

	return ids, nil
}
