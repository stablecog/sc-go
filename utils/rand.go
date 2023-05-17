package utils

import (
	cryptorand "crypto/rand"
	"encoding/hex"
	"io"
)

// Generates random N bytes and returns them as a hex string
func GenerateRandomHex(rand io.Reader, nBytes int) (string, error) {
	if rand == nil {
		rand = cryptorand.Reader
	}
	bytes := make([]byte, nBytes)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
