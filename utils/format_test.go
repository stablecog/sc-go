package utils

import (
	"testing"

	"github.com/stablecog/sc-go/shared"
	"github.com/stretchr/testify/assert"
)

func TestRemoveRedundantSpaces(t *testing.T) {
	assert.Equal(t, "hello world", RemoveRedundantSpaces("  hello   world  "))
}

func TestRemoveLineBreaks(t *testing.T) {
	assert.Equal(t, "hello world", RemoveLineBreaks("hello\nworld"))
}

func TestFormatPrompt(t *testing.T) {
	assert.Equal(t, "hello world", FormatPrompt("  hello   world  "))
	assert.Equal(t, "hello world", FormatPrompt("hello\nworld"))
	assert.Equal(t, "hello world", FormatPrompt("hello\nworld\n"))
	assert.Equal(t, "", FormatPrompt(""))
	// Create a string longer than the max prompt length
	var longStr string
	for i := 0; i < shared.MAX_PROMPT_LENGTH+1; i++ {
		longStr += "a"
	}
	assert.Equal(t, shared.MAX_PROMPT_LENGTH+1, len(longStr))
	assert.Equal(t, shared.MAX_PROMPT_LENGTH, len(FormatPrompt(longStr)))
}

func TestEnsureTrailingSlash(t *testing.T) {
	assert.Equal(t, "hello/", EnsureTrailingSlash("hello"))
	assert.Equal(t, "hello/", EnsureTrailingSlash("hello/"))
}
