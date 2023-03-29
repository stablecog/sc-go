package shared

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSyncMap(t *testing.T) {
	sMap := NewSyncMap[string]()
	assert.NotNil(t, sMap)
	iMap := NewSyncMap[int]()
	assert.NotNil(t, iMap)
}

func TestPutAndExistsAndGet(t *testing.T) {
	sMap := NewSyncMap[string]()
	assert.NotNil(t, sMap)
	assert.False(t, sMap.Exists("hello"))
	sMap.Put("hello", "world")
	assert.True(t, sMap.Exists("hello"))
	assert.Equal(t, "world", sMap.Get("hello"))
}

func TestGetReturnsEmptyStringWhenNoValueSet(t *testing.T) {
	sMap := NewSyncMap[string]()
	assert.NotNil(t, sMap)
	assert.False(t, sMap.Exists("hello"))
	assert.Equal(t, "", sMap.Get("hello"))
}

func TestDelete(t *testing.T) {
	sMap := NewSyncMap[string]()
	assert.NotNil(t, sMap)
	assert.False(t, sMap.Exists("hello"))
	sMap.Put("hello", "world")
	assert.True(t, sMap.Exists("hello"))
	assert.Equal(t, "world", sMap.Get("hello"))
	sMap.Delete("hello")
	assert.False(t, sMap.Exists("hello"))
}
