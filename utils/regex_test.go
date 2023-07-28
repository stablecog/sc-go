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

func TestExtractAmountsFromString(t *testing.T) {
	amt, err := ExtractAmountsFromString("hello 123 45")
	assert.Equal(t, AmountAmbiguousError, err)
	amt, err = ExtractAmountsFromString("hello 123.45")
	assert.Equal(t, AmountNotIntegerError, err)
	amt, err = ExtractAmountsFromString("hello world")
	assert.Equal(t, AmountMissingError, err)
	amt, err = ExtractAmountsFromString("!hello 123 here is for gas")
	assert.NoError(t, err)
	assert.Equal(t, 123, amt)
}
