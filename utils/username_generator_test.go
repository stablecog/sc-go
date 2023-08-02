package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockRandInt struct{}

func (m *MockRandInt) Intn(n int) int {
	return 1
}

func TestGenerateUsername(t *testing.T) {
	username1 := GenerateUsername(nil)
	assert.Equal(t, 12, len(username1))
	username2 := GenerateUsername(nil)
	assert.Equal(t, 12, len(username2))
	assert.NotEqual(t, username1, username2)

	// Test predictable
	username := GenerateUsername(&MockRandInt{})
	assert.Equal(t, "bbbbbbbbbbbb", username)
}
