package database

import (
	"context"
	"fmt"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/stablecog/sc-go/shared"
	"github.com/stablecog/sc-go/utils"
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

// Keep track of request ID to cog, with stream ID of the client
func (r *RedisWrapper) SetCogRequestStreamID(ctx context.Context, requestID string, streamID string) error {
	// We set 2 keys since we expect 2 responses from the cog, started and failed/succeeded
	// These keys are basically used to make sure only 1 instance of the cog takes these requests
	// TODO: We should probably use a queue to get responses from the cog, or go back to webhook
	_, err := r.Client.Set(ctx, fmt.Sprintf("first:%s", requestID), streamID, 1*time.Hour).Result()
	if err != nil {
		return err
	}
	_, err = r.Client.Set(ctx, fmt.Sprintf("second:%s", requestID), streamID, 1*time.Hour).Result()
	return err
}

// Get the stream ID of the client for a given request ID
func (r *RedisWrapper) GetCogRequestStreamID(ctx context.Context, requestID string) (string, error) {
	return r.Client.Get(ctx, requestID).Result()
}

// Delete the stream ID of the client for a given request ID
func (r *RedisWrapper) DeleteCogRequestStreamID(ctx context.Context, requestID string) (int64, error) {
	return r.Client.Del(ctx, requestID).Result()
}
