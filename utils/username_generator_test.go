package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockRandReader struct{}

func (m *MockRandReader) Read(b []byte) (int, error) {
	return 5, nil
}

func TestGenerateUsername(t *testing.T) {
	username1 := GenerateUsername(nil)
	assert.Equal(t, 12, len(username1))
	username2 := GenerateUsername(nil)
	assert.Equal(t, 12, len(username2))
	assert.NotEqual(t, username1, username2)

	// Test predictable
	username := GenerateUsername(&MockRandReader{})
	assert.Equal(t, "aaaaaaaaaaaa", username)
}
