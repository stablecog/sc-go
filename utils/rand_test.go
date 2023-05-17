package utils

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateRandHex(t *testing.T) {
	seed, _ := GenerateRandomHex(nil, 32)
	assert.Len(t, seed, 64)

	// Test predictable
	seed, _ = GenerateRandomHex(strings.NewReader("9f729340e07eee69abac049c2fdd4a3c4b50e4672a2fabdf1ae295f2b4f3040b"), 32)
	assert.Len(t, seed, 64)
	assert.Equal(t, "3966373239333430653037656565363961626163303439633266646434613363", seed)

	// Test different nBytes
	seed, _ = GenerateRandomHex(strings.NewReader("9f729340e07eee69abac049c2fdd4a3c4b50e4672a2fabdf1ae295f2b4f3040b"), 24)
	assert.Len(t, seed, 48)
	assert.Equal(t, "396637323933343065303765656536396162616330343963", seed)
}
