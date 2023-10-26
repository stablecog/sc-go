package database

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/shared"
	"github.com/stablecog/sc-go/utils"
	"github.com/stretchr/testify/assert"
)

func TestGetRedisURL(t *testing.T) {
	// Setup
	origRedisConnectionString := utils.GetEnv().RedisConnectionString
	utils.GetEnv().RedisConnectionString = "wewantthis"
	defer func() {
		utils.GetEnv().RedisConnectionString = origRedisConnectionString
	}()

	// Assert
	assert.Equal(t, "wewantthis", getRedisURL())
}

func TestMockRedis(t *testing.T) {
	origMockRedis := utils.GetEnv().MockRedis
	utils.GetEnv().MockRedis = true
	defer func() {
		utils.GetEnv().MockRedis = origMockRedis
	}()

	// Mock logger
	orgLogInfo := logInfo
	defer func() { logInfo = orgLogInfo }()

	// Write log output to string
	logs := []string{}
	logInfo = func(format interface{}, args ...interface{}) {
		logs = append(logs, format.(string)+fmt.Sprint(args...))
	}

	_, err := NewRedis(context.TODO())
	assert.Nil(t, err)
	assert.Equal(t, "Using mock redis client because MOCK_REDIS=true is set in environment", logs[0])
}

func TestInvalidConnUrlFails(t *testing.T) {
	// Setup
	origRedisConnectionString := utils.GetEnv().RedisConnectionString
	utils.GetEnv().RedisConnectionString = "invalidredisurl"
	defer func() {
		utils.GetEnv().RedisConnectionString = origRedisConnectionString
	}()

	// Mock logger
	orgLogError := logError
	defer func() { logError = orgLogError }()

	// Write log output to string
	logs := []string{}
	logError = func(format interface{}, args ...interface{}) {
		logs = append(logs, format.(string)+fmt.Sprint(args...))
	}

	_, err := NewRedis(context.TODO())
	assert.NotNil(t, err)
	assert.Equal(t, "Error parsing REDIS_CONNECTION_STRINGerrredis: invalid URL scheme: ", logs[0])
}

func TestPingErrorIfCantConnect(t *testing.T) {
	// Setup
	origRedisConnectionString := utils.GetEnv().RedisConnectionString
	utils.GetEnv().RedisConnectionString = "redis://notarealredishost:1234"
	defer func() {
		utils.GetEnv().RedisConnectionString = origRedisConnectionString
	}()

	// Mock logger
	orgLogError := logError
	defer func() { logError = orgLogError }()

	// Write log output to string
	logs := []string{}
	logError = func(format interface{}, args ...interface{}) {
		logs = append(logs, format.(string)+fmt.Sprint(args...))
	}

	_, err := NewRedis(context.TODO())
	assert.NotNil(t, err)
	assert.Equal(t, "Error pinging Rediserrdial tcp: lookup notarealredishos", logs[0][:len("Error pinging Redis: dial tcp: lookup notarealredishost")])
}

func TestGetPendingGenerationAndUpscaleIDs(t *testing.T) {
	// Create redis
	origMockRedis := utils.GetEnv().MockRedis
	utils.GetEnv().MockRedis = true
	defer func() {
		utils.GetEnv().MockRedis = origMockRedis
	}()
	redis, err := NewRedis(context.TODO())
	assert.Nil(t, err)
	// MKStream
	redis.Client.XGroupCreateMkStream(redis.Ctx, shared.COG_REDIS_QUEUE, shared.COG_REDIS_QUEUE, "0-0").Err()

	// Enqueue a few requests
	assert.Nil(t, redis.EnqueueCogRequest(redis.Ctx, shared.COG_REDIS_QUEUE, requests.CogQueueRequest{
		Input: requests.BaseCogRequest{
			ID:          uuid.MustParse("00000000-0000-0000-0000-000000000001"),
			ProcessType: shared.UPSCALE,
		},
	}))

	assert.Nil(t, redis.EnqueueCogRequest(redis.Ctx, shared.COG_REDIS_QUEUE, requests.CogQueueRequest{
		Input: requests.BaseCogRequest{
			ID:          uuid.MustParse("00000000-0000-0000-0000-000000000002"),
			ProcessType: shared.GENERATE,
		},
	}))

	assert.Nil(t, redis.EnqueueCogRequest(redis.Ctx, shared.COG_REDIS_QUEUE, requests.CogQueueRequest{
		Input: requests.BaseCogRequest{
			ID:          uuid.MustParse("00000000-0000-0000-0000-000000000032"),
			ProcessType: shared.GENERATE_AND_UPSCALE,
		},
	}))

	pendingGenerationIDs, pendingUpscaleIDs, err := redis.GetPendingGenerationAndUpscaleIDs(0)
	assert.Nil(t, err)
	assert.Len(t, pendingGenerationIDs, 2)
	assert.Len(t, pendingUpscaleIDs, 1)

	s, err := redis.GetQueueSize()
	assert.Nil(t, err)
	assert.Equal(t, int64(3), s)
}
