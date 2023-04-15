package shared

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/log"
	"github.com/hashicorp/go-multierror"
	"github.com/redis/go-redis/v9"
)

const REDIS_KEY = "queue_throttler"

// UserQueueThrottlerMap builds an thread-safe map
// For number of items enqueued by this user
type UserQueueThrottlerMap struct {
	redis *redis.Client
	ttl   time.Duration
	ctx   context.Context
}

func NewQueueThrottler(ctx context.Context, redis *redis.Client, ttl time.Duration) *UserQueueThrottlerMap {
	return &UserQueueThrottlerMap{
		redis: redis,
		ttl:   ttl,
		ctx:   ctx,
	}
}

// Increment the number of items in queue for this user
func (r *UserQueueThrottlerMap) IncrementBy(amount int, userID string) error {
	now := time.Now().Format(time.RFC3339Nano)
	// Create array of amount size
	arr := make([]interface{}, amount)
	for i := 0; i < amount; i++ {
		arr[i] = now
	}
	return r.redis.RPush(r.ctx, fmt.Sprintf("%s:%s", REDIS_KEY, userID), arr...).Err()
}

// Decrement the number of items in queue for this user, minimum 0
func (r *UserQueueThrottlerMap) DecrementBy(amount int, userID string) error {
	// Remove the oldest item for this user
	var mErr *multierror.Error
	for i := 0; i < amount; i++ {
		err := r.redis.LPop(r.ctx, fmt.Sprintf("%s:%s", REDIS_KEY, userID)).Err()
		if err != redis.Nil {
			mErr = multierror.Append(mErr, err)
		}
	}
	return mErr.ErrorOrNil()
}

// Get the number of items in queue for this user
func (r *UserQueueThrottlerMap) NumQueued(userID string) (int, error) {
	// Get all items for this user
	items, err := r.redis.LRange(r.ctx, fmt.Sprintf("%s:%s", REDIS_KEY, userID), 0, -1).Result()
	if err != nil {
		return 0, err
	}
	// Remove all items older than 1 minute
	for _, item := range items {
		putAt, err := time.Parse(time.RFC3339Nano, item)
		if err != nil {
			log.Errorf("Error parsing time: %s %s", err, item)
			// LPop anyway
			r.redis.LPop(r.ctx, fmt.Sprintf("%s:%s", REDIS_KEY, userID))
		}
		if time.Since(putAt) > r.ttl {
			// LPop this item
			r.redis.LPop(r.ctx, fmt.Sprintf("%s:%s", REDIS_KEY, userID))
		}
	}
	// Count the number of items for this user
	c, err := r.redis.LLen(r.ctx, fmt.Sprintf("%s:%s", REDIS_KEY, userID)).Result()
	if err != nil {
		log.Error("Error in redis llen", "err", err)
	}
	return int(c), err
}
