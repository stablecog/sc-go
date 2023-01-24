package repository

import (
	"github.com/stablecog/go-apps/database/ent"
	"github.com/stablecog/go-apps/database/ent/scheduler"
)

func (r *Repository) GetAllSchedulers() ([]*ent.Scheduler, error) {
	schedulers, err := r.DB.Scheduler.Query().Select(scheduler.FieldID, scheduler.FieldIsFree, scheduler.FieldName).All(r.Ctx)
	if err != nil {
		return nil, err
	}

	return schedulers, nil
}
