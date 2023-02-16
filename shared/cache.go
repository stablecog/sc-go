package shared

import (
	"sync"

	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
)

// A singleton that caches the features available to free users
// Avoids having to query the database every time a user requests the features

type Cache struct {
	// Models and options available to free users
	GenerateModels []*ent.GenerationModel
	UpscaleModels  []*ent.UpscaleModel
	Schedulers     []*ent.Scheduler
	AdminIDs       []uuid.UUID
}

var lock = &sync.Mutex{}

var singleCache *Cache

func newCache() *Cache {
	return &Cache{}
}

func GetCache() *Cache {
	if singleCache == nil {
		lock.Lock()
		defer lock.Unlock()
		if singleCache == nil {
			singleCache = newCache()
		}
	}
	return singleCache
}

func (f *Cache) UpdateGenerationModels(models []*ent.GenerationModel) {
	lock.Lock()
	defer lock.Unlock()
	f.GenerateModels = models
}

func (f *Cache) UpdateUpscaleModels(models []*ent.UpscaleModel) {
	lock.Lock()
	defer lock.Unlock()
	f.UpscaleModels = models
}

func (f *Cache) UpdateSchedulers(schedulers []*ent.Scheduler) {
	lock.Lock()
	defer lock.Unlock()
	f.Schedulers = schedulers
}

func (f *Cache) IsValidGenerationModelID(id uuid.UUID) bool {
	for _, model := range f.GenerateModels {
		if model.ID == id {
			return true
		}
	}
	return false
}

func (f *Cache) IsValidUpscaleModelID(id uuid.UUID) bool {
	for _, model := range f.UpscaleModels {
		if model.ID == id {
			return true
		}
	}
	return false
}

func (f *Cache) IsValidShedulerID(id uuid.UUID) bool {
	for _, scheduler := range f.Schedulers {
		if scheduler.ID == id {
			return true
		}
	}
	return false
}

func (f *Cache) GetGenerationModelNameFromID(id uuid.UUID) string {
	for _, model := range f.GenerateModels {
		if model.ID == id {
			return model.NameInWorker
		}
	}
	return ""
}

func (f *Cache) GetUpscaleModelNameFromID(id uuid.UUID) string {
	for _, model := range f.UpscaleModels {
		if model.ID == id {
			return model.NameInWorker
		}
	}
	return ""
}

func (f *Cache) GetSchedulerNameFromID(id uuid.UUID) string {
	for _, scheduler := range f.Schedulers {
		if scheduler.ID == id {
			return scheduler.NameInWorker
		}
	}
	return ""
}

func (f *Cache) IsAdmin(id uuid.UUID) bool {
	for _, adminID := range f.AdminIDs {
		if adminID == id {
			return true
		}
	}
	return false
}

func (f *Cache) SetAdminUUIDs(ids []uuid.UUID) {
	lock.Lock()
	defer lock.Unlock()
	f.AdminIDs = ids
}
