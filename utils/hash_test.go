package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSha256(t *testing.T) {
	toHash := "test"
	assert.Equal(t, "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08", Sha256(toHash))
}
