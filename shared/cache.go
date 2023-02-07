package shared

import (
	"sync"

	"github.com/google/uuid"
	"github.com/stablecog/go-apps/database/ent"
	"golang.org/x/exp/slices"
)

// A singleton that caches the features available to free users
// Avoids having to query the database every time a user requests the features

type Cache struct {
	// Models and options available to free users
	FreeWidths            []int32
	FreeHeights           []int32
	FreeInterferenceSteps []int32
	GenerateModels        []*ent.GenerationModel
	UpscaleModels         []*ent.UpscaleModel
	Schedulers            []*ent.Scheduler
}

var lock = &sync.Mutex{}

var singleCache *Cache

func newCache() *Cache {
	return &Cache{
		FreeWidths:            []int32{MAX_GENERATE_WIDTH_FREE},
		FreeHeights:           []int32{MAX_GENERATE_HEIGHT_FREE},
		FreeInterferenceSteps: []int32{MAX_GENERATE_INTERFERENCE_STEPS_FREE},
	}
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

func (f *Cache) IsGenerationModelAvailableForFree(id uuid.UUID) bool {
	for _, model := range f.GenerateModels {
		if model.ID == id && model.IsFree {
			return true
		}
	}
	return false
}

func (f *Cache) IsSchedulerAvailableForFree(id uuid.UUID) bool {
	for _, scheduler := range f.Schedulers {
		if scheduler.ID == id && scheduler.IsFree {
			return true
		}
	}
	return false
}

func (f *Cache) IsWidthAvailableForFree(width int32) bool {
	return slices.Contains(f.FreeWidths, width)
}

func (f *Cache) IsHeightAvailableForFree(width int32) bool {
	return slices.Contains(f.FreeHeights, width)
}

func (f *Cache) IsNumInterferenceStepsAvailableForFree(width int32) bool {
	return slices.Contains(f.FreeInterferenceSteps, width)
}

func (f *Cache) GetGenerationModelNameFromID(id uuid.UUID) string {
	for _, model := range f.GenerateModels {
		if model.ID == id {
			return model.Name
		}
	}
	return ""
}

func (f *Cache) GetUpscaleModelNameFromID(id uuid.UUID) string {
	for _, model := range f.UpscaleModels {
		if model.ID == id {
			return model.Name
		}
	}
	return ""
}

func (f *Cache) GetSchedulerNameFromID(id uuid.UUID) string {
	for _, scheduler := range f.Schedulers {
		if scheduler.ID == id {
			return scheduler.Name
		}
	}
	return ""
}
