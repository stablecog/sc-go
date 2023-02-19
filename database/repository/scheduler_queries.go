package repository

import (
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/ent/scheduler"
)

func (r *Repository) GetAllSchedulers() ([]*ent.Scheduler, error) {
	schedulers, err := r.DB.Scheduler.Query().Select(scheduler.FieldID, scheduler.FieldNameInWorker).All(r.Ctx)
	if err != nil {
		return nil, err
	}

	return schedulers, nil
}
