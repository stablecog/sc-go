package shared

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func MockRedis(ctx context.Context) (*redis.Client, error) {
	mr, _ := miniredis.Run()
	opts := &redis.Options{
		Addr: mr.Addr(),
	}
	redis := redis.NewClient(opts)
	_, err := redis.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}
	return redis, nil
}

func TestIncrement(t *testing.T) {
	ctx := context.Background()
	redis, err := MockRedis(ctx)
	assert.Nil(t, err)
	m := NewQueueThrottler(ctx, redis, time.Minute)
	assert.Nil(t, m.IncrementBy(1, "test"))
	nq, err := m.NumQueued("test")
	assert.Nil(t, err)
	assert.Equal(t, 1, nq)
	assert.Nil(t, m.IncrementBy(3, "test"))
	assert.Nil(t, m.IncrementBy(2, "test"))
	nq, err = m.NumQueued("test")
	assert.Equal(t, 6, nq)
}

func TestDecrement(t *testing.T) {
	ctx := context.Background()
	redis, err := MockRedis(ctx)
	assert.Nil(t, err)
	m := NewQueueThrottler(ctx, redis, time.Minute)
	assert.Nil(t, m.IncrementBy(1, "test"))
	assert.Nil(t, m.IncrementBy(1, "test"))
	assert.Nil(t, m.IncrementBy(1, "test"))
	assert.Nil(t, m.DecrementBy(1, "test"))
	nq, err := m.NumQueued("test")
	assert.Equal(t, 2, nq)
	assert.Nil(t, m.DecrementBy(1, "test"))
	assert.Nil(t, m.DecrementBy(1, "test"))
	assert.Nil(t, m.DecrementBy(1, "test"))
	nq, err = m.NumQueued("test")
	assert.Equal(t, 0, nq)

	// It shouldn't go below 0
	assert.Nil(t, m.DecrementBy(1, "test"))
	nq, err = m.NumQueued("test")
	assert.Nil(t, err)
	assert.Equal(t, 0, nq)

	// New key should be 0
	nq, err = m.NumQueued("test2")
	assert.Nil(t, err)
	assert.Equal(t, 0, nq)
}

func TestNumQueued(t *testing.T) {
	ctx := context.Background()
	redis, err := MockRedis(ctx)
	assert.Nil(t, err)
	m := NewQueueThrottler(ctx, redis, time.Minute)
	assert.Nil(t, m.IncrementBy(1, "test"))
	assert.Nil(t, m.IncrementBy(1, "test"))
	assert.Nil(t, m.IncrementBy(1, "test"))
	nq, err := m.NumQueued("test")
	assert.Equal(t, 3, nq)
	assert.Nil(t, m.DecrementBy(1, "test"))
	nq, err = m.NumQueued("test")
	assert.Equal(t, 2, nq)
	assert.Nil(t, m.DecrementBy(1, "test"))
	assert.Nil(t, m.DecrementBy(2, "test"))
	nq, err = m.NumQueued("test")
	assert.Equal(t, 0, nq)
}

func TestNumQueuedTimeout(t *testing.T) {
	ctx := context.Background()
	redis, err := MockRedis(ctx)
	assert.Nil(t, err)
	m := NewQueueThrottler(ctx, redis, time.Millisecond)
	assert.Nil(t, m.IncrementBy(1, "test"))
	assert.Nil(t, m.IncrementBy(1, "test"))
	assert.Nil(t, m.IncrementBy(1, "test"))
	nq, err := m.NumQueued("test")
	assert.Nil(t, err)
	assert.Equal(t, 3, nq)
	time.Sleep(time.Millisecond * 2)
	nq, err = m.NumQueued("test")
	assert.Nil(t, err)
	assert.Equal(t, 0, nq)
}
