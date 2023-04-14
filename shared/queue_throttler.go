package shared

import (
	"sync"
	"time"
)

type queueItem struct {
	Put    time.Time
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
func (r *UserQueueThrottlerMap) IncrementBy(amount int, userID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := 0; i < amount; i++ {
		r.sMap = append(r.sMap, queueItem{
			Put:    time.Now(),
			UserID: userID,
		})
	}
}

// Decrement the number of items in queue for this user, minimum 0
func (r *UserQueueThrottlerMap) DecrementBy(amount int, userID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	// Remove the oldest item for this user
	removed := 0
	for i, item := range r.sMap {
		if item.UserID == userID {
			r.sMap = append(r.sMap[:i], r.sMap[i+1:]...)
			removed++
			i--
			if removed >= amount {
				break
			}
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
