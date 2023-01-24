package models

import "sync"

// A singleton that caches the features available to free users
// Avoids having to query the database every time a user requests the features

type Cache struct {
	// Models and options available to free users
	FreeWidths            []int
	FreeHeights           []int
	FreeInterferenceSteps []int
	FreeModelIDs          []string
	FreeSchedulerIDs      []string
}

var lock = &sync.Mutex{}

var singleCache *Cache

func newCache() *Cache {
	return &Cache{
		FreeWidths:            []int{512},
		FreeHeights:           []int{512},
		FreeInterferenceSteps: []int{30},
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

func (f *Cache) UpdateFreeModelsAndSchedulers(freeModelIds []string, freeSchedulerIds []string) {
	lock.Lock()
	defer lock.Unlock()
	f.FreeModelIDs = freeModelIds
	f.FreeSchedulerIDs = freeSchedulerIds
}
