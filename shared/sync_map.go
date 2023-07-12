package shared

import (
	"sync"
)

// SyncMap builds an thread-safe map
type SyncMap[T any] struct {
	mu   sync.Mutex
	sMap map[string]T
}

func NewSyncMap[T any]() *SyncMap[T] {
	return &SyncMap[T]{
		sMap: make(map[string]T),
	}
}

// See if element exists
func (r *SyncMap[T]) Exists(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, ok := r.sMap[key]
	return ok
}

// Put value into map - synchronized
func (r *SyncMap[T]) Put(key string, value T) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sMap[key] = value
}

// Gets a value from the map - synchronized
func (r *SyncMap[T]) Get(key string) T {
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.sMap[key]
	if ok {
		return s
	}
	var d T
	return d
}

// Removes specified hash - synchronized
func (r *SyncMap[T]) Delete(key string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.sMap, key)
}

// Get all keys and values
func (r *SyncMap[T]) GetAll() map[string]T {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.sMap
}
