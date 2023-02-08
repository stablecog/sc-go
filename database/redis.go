package database

import (
	"context"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/stablecog/go-apps/shared"
	"github.com/stablecog/go-apps/utils"
	"k8s.io/klog/v2"
)

var klogInfof = klog.Infof
var klogErrorf = klog.Errorf

type RedisWrapper struct {
	Client *redis.Client
}

// Should return render redis url if render is set
func getRedisURL() string {
	if utils.GetEnv("RENDER", "") != "" {
		return utils.GetEnv("REDIS_CONNECTION_STRING_RENDER", "")
	}
	return utils.GetEnv("REDIS_CONNECTION_STRING", "")
}

// Returns our *RedisWrapper, since we wrap some useful methods with the redis client
func NewRedis(ctx context.Context) (*RedisWrapper, error) {
	var opts *redis.Options
	var err error
	if utils.GetEnv("MOCK_REDIS", "false") == "true" {
		klogInfof("Using mock redis client because MOCK_REDIS=true is set in environment")
		mr, _ := miniredis.Run()
		opts = &redis.Options{
			Addr: mr.Addr(),
		}
	} else {
		opts, err = redis.ParseURL(getRedisURL())
		if err != nil {
			klogErrorf("Error parsing REDIS_CONNECTION_STRING: %v", err)
			return nil, err
		}
	}
	redis := redis.NewClient(opts)
	_, err = redis.Ping(ctx).Result()
	if err != nil {
		klogErrorf("Error pinging Redis: %v", err)
		return nil, err
	}
	return &RedisWrapper{
		Client: redis,
	}, nil
}

// Enqueues a request for the cog
func (r *RedisWrapper) EnqueueCogRequest(ctx context.Context, request interface{}) error {
	_, err := r.Client.XAdd(ctx, &redis.XAddArgs{
		Stream: shared.COG_REDIS_QUEUE,
		ID:     "*", // Asterisk auto-generates an ID for the item on the stream
		Values: []interface{}{"value", request},
	}).Result()
	return err
}
