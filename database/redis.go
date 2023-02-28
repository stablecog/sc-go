package database

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/charmbracelet/log"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/shared"
	"github.com/stablecog/sc-go/utils"
)

var logInfo = log.Info
var logError = log.Error

type RedisWrapper struct {
	Client *redis.Client
}

// Should return render redis url if render is set
func getRedisURL() string {
	return utils.GetEnv("REDIS_CONNECTION_STRING", "")
}

// Returns our *RedisWrapper, since we wrap some useful methods with the redis client
func NewRedis(ctx context.Context) (*RedisWrapper, error) {
	var opts *redis.Options
	var err error
	if utils.GetEnv("MOCK_REDIS", "false") == "true" {
		logInfo("Using mock redis client because MOCK_REDIS=true is set in environment")
		mr, _ := miniredis.Run()
		opts = &redis.Options{
			Addr: mr.Addr(),
		}
	} else {
		opts, err = redis.ParseURL(getRedisURL())
		if err != nil {
			logError("Error parsing REDIS_CONNECTION_STRING", "err", err)
			return nil, err
		}
	}
	redis := redis.NewClient(opts)
	_, err = redis.Ping(ctx).Result()
	if err != nil {
		logError("Error pinging Redis", "err", err)
		return nil, err
	}
	return &RedisWrapper{
		Client: redis,
	}, nil
}

// Set generate and upscale count stats
func (r *RedisWrapper) SetGenerateUpscaleOutputCount(generationOutputCount, upscaleOutputCount int) error {
	stats := RedisStats{
		GenerationOutputCount: generationOutputCount,
		UpscaleOutputCount:    upscaleOutputCount,
	}
	statsJSON, err := json.Marshal(stats)
	if err != nil {
		return err
	}
	return r.Client.Set(context.Background(), "stats", statsJSON, 0).Err()
}

// Get generate and upscale count stats
func (r *RedisWrapper) GetGenerateUpscaleCount() (stats *RedisStats, err error) {
	statsJSON, err := r.Client.Get(context.Background(), "stats").Result()
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal([]byte(statsJSON), &stats)
	if err != nil {
		return nil, err
	}
	return stats, nil
}

type RedisStats struct {
	GenerationOutputCount int `json:"generation_output_count"`
	UpscaleOutputCount    int `json:"upscale_output_count"`
}

// Enqueues a request to sc-worker
func (r *RedisWrapper) EnqueueCogRequest(ctx context.Context, request interface{}) error {
	_, err := r.Client.XAdd(ctx, &redis.XAddArgs{
		Stream: shared.COG_REDIS_QUEUE,
		ID:     "*", // Asterisk auto-generates an ID for the item on the stream
		Values: []interface{}{"value", request},
	}).Result()
	return err
}

// Get pending request IDs on queue that are stale
// olderThan will be subtracted from the current time to return requests older than that
func (r *RedisWrapper) GetPendingGenerationAndUpscaleIDs(olderThan time.Duration) (generationOutputIDs, upscaleOutputIDs []PendingCogRequestRedis, err error) {
	// Get current time in MS since epoch, minus endAt
	to := time.Now().UnixNano()/int64(time.Millisecond) - int64(olderThan/time.Millisecond)
	// Get from the redis stream COG_REDIS_QUEUE, we want to read them without deleting them
	messages, err := r.Client.XRange(r.Client.Context(), shared.COG_REDIS_QUEUE, "0-0", fmt.Sprintf("%d", to)).Result()
	if err != nil {
		log.Error("Error getting pending generation and upscale IDs", "err", err)
		return nil, nil, err
	}

	generationOutputIDs = make([]PendingCogRequestRedis, 0)
	upscaleOutputIDs = make([]PendingCogRequestRedis, 0)

	for _, message := range messages {
		// Get the request ID from the message
		input, ok := message.Values["value"].(string)
		if input == "" || !ok {
			log.Error("Error getting value from message", "message", message)
			continue
		}
		// Deserialize
		var request requests.CogQueueRequest
		err = json.Unmarshal([]byte(input), &request)
		if err != nil {
			log.Error("Error deserializing input", "err", err)
			continue
		}

		if request.Input.ProcessType == shared.UPSCALE {
			parsed, err := uuid.Parse(request.Input.ID)
			if err != nil {
				log.Error("Error parsing upscale output ID", "err", err)
				continue
			}
			upscaleOutputIDs = append(upscaleOutputIDs, PendingCogRequestRedis{
				RedisMsgid: message.ID,
				Type:       request.Input.ProcessType,
				ID:         parsed,
			})
		}
		if request.Input.ProcessType == shared.GENERATE || request.Input.ProcessType == shared.GENERATE_AND_UPSCALE {
			parsed, err := uuid.Parse(request.Input.ID)
			if err != nil {
				log.Error("Error parsing generation output ID", "err", err)
				continue
			}
			generationOutputIDs = append(generationOutputIDs, PendingCogRequestRedis{
				RedisMsgid: message.ID,
				Type:       request.Input.ProcessType,
				ID:         parsed,
			})
		}
	}

	return generationOutputIDs, upscaleOutputIDs, nil
}

type PendingCogRequestRedis struct {
	RedisMsgid string
	Type       shared.ProcessType
	ID         uuid.UUID
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
