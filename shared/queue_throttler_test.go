package shared

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIncrement(t *testing.T) {
	m := NewQueueThrottler(time.Minute)
	m.Increment("1", "test")
	assert.Equal(t, 1, m.NumQueued("test"))
	m.Increment("1", "test")
	m.Increment("1", "test")
	assert.Equal(t, 3, m.NumQueued("test"))
}

func TestDecrement(t *testing.T) {
	m := NewQueueThrottler(time.Microsecond)
	m.Increment("1", "test")
	m.Increment("1", "test")
	m.Increment("1", "test")
	m.Decrement("1", "test")
	assert.Equal(t, 2, m.NumQueued("test"))
	m.Decrement("1", "test")
	m.Decrement("1", "test")
	m.Decrement("1", "test")
	assert.Equal(t, 0, m.NumQueued("test"))

	// It shouldn't go below 0
	m.Decrement("1", "test")
	assert.Equal(t, 0, m.NumQueued("test"))

	// New key should be 0
	assert.Equal(t, 0, m.NumQueued("test2"))
}

func TestNumQueued(t *testing.T) {
	m := NewQueueThrottler(time.Minute)
	m.Increment("1", "test")
	m.Increment("1", "test")
	m.Increment("1", "test")
	assert.Equal(t, 3, m.NumQueued("test"))
	m.Decrement("1", "test")
	assert.Equal(t, 2, m.NumQueued("test"))
	m.Decrement("1", "test")
	m.Decrement("1", "test")
	m.Decrement("1", "test")
	assert.Equal(t, 0, m.NumQueued("test"))
}

func TestNumQueuedTimeout(t *testing.T) {
	m := NewQueueThrottler(time.Millisecond)
	m.Increment("1", "test")
	m.Increment("1", "test")
	m.Increment("1", "test")
	assert.Equal(t, 3, m.NumQueued("test"))
	time.Sleep(time.Millisecond * 2)
	assert.Equal(t, 0, m.NumQueued("test"))
}
