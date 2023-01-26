package utils

import (
	"encoding/hex"
)

func IsSha256Hash(str string) bool {
	if len(str) != 64 {
		return false
	}

	// Validate string is hex
	_, err := hex.DecodeString(str)
	if err != nil {
		return false
	}

	return true
}
