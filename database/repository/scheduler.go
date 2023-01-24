package repository

import (
	"github.com/stablecog/go-apps/database/ent/scheduler"
)

func (r *Repository) GetFreeSchedulerIDs() ([]string, error) {
	models, err := r.DB.Scheduler.Query().Where(scheduler.IsFree(true)).All(r.Ctx)
	if err != nil {
		return nil, err
	}

	ids := make([]string, len(models))
	for i, model := range models {
		ids[i] = model.ID.String()
	}

	return ids, nil
}
