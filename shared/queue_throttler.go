package shared

import (
	"sync"
	"time"
)

type queueItem struct {
	Put    time.Time
	ID     string
	UserID string
}

// UserQueueThrottlerMap builds an thread-safe map
// For number of items enqueued by this user
type UserQueueThrottlerMap struct {
	mu sync.Mutex
	// User ID -> Number of items in queue
	sMap []queueItem
	ttl  time.Duration
}

func NewQueueThrottler(ttl time.Duration) *UserQueueThrottlerMap {
	return &UserQueueThrottlerMap{
		sMap: []queueItem{},
		ttl:  ttl,
	}
}

// Increment the number of items in queue for this user
func (r *UserQueueThrottlerMap) Increment(uniqueId string, userID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sMap = append(r.sMap, queueItem{
		Put:    time.Now(),
		ID:     uniqueId,
		UserID: userID,
	})
}

// Decrement the number of items in queue for this user, minimum 0
func (r *UserQueueThrottlerMap) Decrement(uniqueId string, userID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	// Remove the oldest item for this user
	for i, item := range r.sMap {
		if item.UserID == userID && item.ID == uniqueId {
			r.sMap = append(r.sMap[:i], r.sMap[i+1:]...)
			return
		}
	}
}

// Get the number of items in queue for this user
func (r *UserQueueThrottlerMap) NumQueued(userID string) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	// Remove all items older than 1 minute
	newMap := []queueItem{}
	for i, item := range r.sMap {
		if time.Since(item.Put) < r.ttl {
			newMap = append(newMap, r.sMap[i])
		}
	}
	r.sMap = newMap
	// Count the number of items for this user
	count := 0
	for _, item := range r.sMap {
		if item.UserID == userID {
			count++
		}
	}
	return count
}
