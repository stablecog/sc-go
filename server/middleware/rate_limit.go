package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/httprate"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stablecog/sc-go/database"
	"github.com/stablecog/sc-go/shared"
	"github.com/stablecog/sc-go/utils"
)

// Rate limit middleware
// @requestLimit: The number of requests they can make
// @windowLength: In this time window
func (m *Middleware) RateLimit(requestLimit int, customKey string, windowLength time.Duration) func(next http.Handler) http.Handler {
	// ! For some reason, it seems like we get half of what we set with chi httprate middleware
	// ! so we set requestLimit * 2 instead of requestLimit, which gives us what we actually want.
	return httprate.Limit(
		requestLimit*2,
		windowLength,
		httprate.WithKeyFuncs(func(r *http.Request) (string, error) {
			// Get user id from context
			userId, ok := r.Context().Value("user_id").(string)
			if ok {
				parsed, err := uuid.Parse(userId)
				if err == nil {
					// See if admin
					if shared.GetCache().IsAdmin(parsed) {
						// Rnadom UUID disables rate limit
						return fmt.Sprintf("%s:%s", customKey, uuid.NewString()), nil
					}
				}
			}
			return fmt.Sprintf("%s:%s", customKey, utils.GetIPAddress(r)), nil
		}),
		httprate.WithLimitCounter(&redisCounter{redis: m.Redis}),
	)
}

type redisCounter struct {
	redis        *database.RedisWrapper
	windowLength time.Duration
}

var _ httprate.LimitCounter = &redisCounter{}

func (c *redisCounter) Config(requestLimit int, windowLength time.Duration) {
	c.windowLength = windowLength
}

func (c *redisCounter) Increment(key string, currentWindow time.Time) error {
	hkey := limitCounterKey(key, currentWindow)

	c.redis.Client.Incr(c.redis.Ctx, hkey).Err()
	err := c.redis.Client.Incr(c.redis.Ctx, hkey).Err()
	if err != nil {
		return err
	}
	err = c.redis.Client.Expire(c.redis.Ctx, hkey, c.windowLength*3).Err()
	if err != nil {
		return err
	}

	return nil
}

func (c *redisCounter) Get(key string, currentWindow, previousWindow time.Time) (int, int, error) {
	currValue, err := c.redis.Client.Get(c.redis.Ctx, limitCounterKey(key, currentWindow)).Result()
	if err != nil && err != redis.Nil {
		return 0, 0, fmt.Errorf("redis get failed: %w", err)
	}

	var curr int
	if currValue != "" {
		curr, err = strconv.Atoi(currValue)
		if err != nil {
			return 0, 0, fmt.Errorf("redis int value: %w", err)
		}
	}

	prevValue, err := c.redis.Client.Get(c.redis.Ctx, limitCounterKey(key, previousWindow)).Result()
	if err != nil && err != redis.Nil {
		return 0, 0, fmt.Errorf("redis get failed: %w", err)
	}

	var prev int
	if prevValue != "" {
		prev, err = strconv.Atoi(prevValue)
		if err != nil {
			return 0, 0, fmt.Errorf("redis int value: %w", err)
		}
	}

	return curr, prev, nil
}

func limitCounterKey(key string, window time.Time) string {
	return fmt.Sprintf("httprate:%d", httprate.LimitCounterKey(key, window))
}
