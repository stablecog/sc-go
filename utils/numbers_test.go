package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMax(t *testing.T) {
	// Test with integers
	assert.Equal(t, 2, Max(int(1), int(2)), "Max of 1 and 2 should be 2")

	// Test with float64
	assert.Equal(t, 3.5, Max(float64(3.5), float64(2.5)), "Max of 3.5 and 2.5 should be 3.5")

	// Test with strings
	assert.Equal(t, "banana", Max(string("apple"), string("banana")), "Max of apple and banana should be banana")
}
