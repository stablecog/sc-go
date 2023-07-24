package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemovePlusFromEmail(t *testing.T) {
	assert.Equal(t, "hello@stablecog.com", RemovePlusFromEmail("hello@stablecog.com"))
	assert.Equal(t, "hello@stablecog.com", RemovePlusFromEmail("hello+123@stablecog.com"))
	assert.Equal(t, "hello@stablecog.com", RemovePlusFromEmail("hello+1@stablecog.com"))
	assert.Equal(t, "hello@stablecog.com", RemovePlusFromEmail("hello+abcdef@stablecog.com"))
	assert.Equal(t, "hello", RemovePlusFromEmail("hello"))
}
