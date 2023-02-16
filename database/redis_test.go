package database

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/shared"
	"github.com/stretchr/testify/assert"
)

func TestGetRedisURL(t *testing.T) {
	// Setup
	os.Setenv("RENDER", "true")
	defer os.Unsetenv("RENDER")
	os.Setenv("REDIS_CONNECTION_STRING_RENDER", "wewantthis")
	defer os.Unsetenv("REDIS_CONNECTION_STRING_RENDER")
	os.Setenv("REDIS_CONNECTION_STRING", "weDoNotwantthisIfRenderIsSet")
	defer os.Unsetenv("REDIS_CONNECTION_STRING")

	// Assert
	assert.Equal(t, "wewantthis", getRedisURL())
	os.Unsetenv("RENDER")
	assert.Equal(t, "weDoNotwantthisIfRenderIsSet", getRedisURL())
}

func TestMockRedis(t *testing.T) {
	os.Setenv("MOCK_REDIS", "true")
	defer os.Unsetenv("MOCK_REDIS")

	// Mock logger
	orgKlogInfof := klogInfof
	defer func() { klogInfof = orgKlogInfof }()

	// Write log output to string
	logs := []string{}
	klogInfof = func(format string, args ...interface{}) {
		logs = append(logs, fmt.Sprintf(format, args...))
	}

	_, err := NewRedis(context.TODO())
	assert.Nil(t, err)
	assert.Equal(t, "Using mock redis client because MOCK_REDIS=true is set in environment", logs[0])
}

func TestInvalidConnUrlFails(t *testing.T) {
	// Setup
	os.Setenv("REDIS_CONNECTION_STRING", "invalidredisurl")

	// Mock logger
	orgKlogErrorf := klogErrorf
	defer func() { klogErrorf = orgKlogErrorf }()

	// Write log output to string
	logs := []string{}
	klogErrorf = func(format string, args ...interface{}) {
		logs = append(logs, fmt.Sprintf(format, args...))
	}

	_, err := NewRedis(context.TODO())
	assert.NotNil(t, err)
	assert.Equal(t, "Error parsing REDIS_CONNECTION_STRING: redis: invalid URL scheme: ", logs[0])
}

func TestPingErrorIfCantConnect(t *testing.T) {
	// Setup
	os.Setenv("REDIS_CONNECTION_STRING", "redis://notarealredishost:1234")

	// Mock logger
	orgKlogErrorf := klogErrorf
	defer func() { klogErrorf = orgKlogErrorf }()

	// Write log output to string
	logs := []string{}
	klogErrorf = func(format string, args ...interface{}) {
		logs = append(logs, fmt.Sprintf(format, args...))
	}

	_, err := NewRedis(context.TODO())
	assert.NotNil(t, err)
	assert.Equal(t, "Error pinging Redis: dial tcp: lookup notarealredishost", logs[0][:len("Error pinging Redis: dial tcp: lookup notarealredishost")])
}

func TestGetPendingGenerationAndUpscaleIDs(t *testing.T) {
	// Create redis
	os.Setenv("MOCK_REDIS", "true")
	defer os.Unsetenv("MOCK_REDIS")
	redis, err := NewRedis(context.TODO())
	assert.Nil(t, err)
	// MKStream
	redis.Client.XGroupCreateMkStream(redis.Client.Context(), shared.COG_REDIS_QUEUE, shared.COG_REDIS_QUEUE, "0-0").Err()

	// Enqueue a few requests
	assert.Nil(t, redis.EnqueueCogRequest(redis.Client.Context(), requests.CogQueueRequest{
		Input: requests.BaseCogRequest{
			ID:          uuid.MustParse("00000000-0000-0000-0000-000000000001").String(),
			ProcessType: shared.UPSCALE,
		},
	}))

	assert.Nil(t, redis.EnqueueCogRequest(redis.Client.Context(), requests.CogQueueRequest{
		Input: requests.BaseCogRequest{
			ID:          uuid.MustParse("00000000-0000-0000-0000-000000000002").String(),
			ProcessType: shared.GENERATE,
		},
	}))

	assert.Nil(t, redis.EnqueueCogRequest(redis.Client.Context(), requests.CogQueueRequest{
		Input: requests.BaseCogRequest{
			ID:          uuid.MustParse("00000000-0000-0000-0000-000000000032").String(),
			ProcessType: shared.GENERATE_AND_UPSCALE,
		},
	}))

	pendingGenerationIDs, pendingUpscaleIDs, err := redis.GetPendingGenerationAndUpscaleIDs(0)
	assert.Nil(t, err)
	assert.Len(t, pendingGenerationIDs, 2)
	assert.Len(t, pendingUpscaleIDs, 1)
}
