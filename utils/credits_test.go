package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateVoiceoverCredits(t *testing.T) {
	assert.Equal(t, int32(1), CalculateVoiceoverCredits("1"))
	// Up to 57 chars also 1
	assert.Equal(t, int32(1), CalculateVoiceoverCredits("123456789012345678901234567890123456789012345678901234567"))
	// 58 is 2
	assert.Equal(t, int32(2), CalculateVoiceoverCredits("1234567890123456789012345678901234567890123456789012345678"))
}
