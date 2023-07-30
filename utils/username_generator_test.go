package utils

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateUsername(t *testing.T) {
	username := GenerateUsername(nil)
	assert.Equal(t, 3, len(strings.Split(username, "-")))

	// Test predictable
	fixedReader := strings.NewReader("9f729340e07eee69abac049c2fdd4a3c4b50e4672a2fabdf1ae295f2b4f3040b")
	username = GenerateUsername(fixedReader)
	assert.Equal(t, "dull-emergence-3LLL", username)
}
