package utils

import (
	"encoding/hex"
	"net/url"
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

// Validates that a URL is valid with http or https scheme
func IsValidHTTPURL(urlStr string) bool {
	u, err := url.Parse(urlStr)
	return err == nil && (u.Scheme == "https" || u.Scheme == "http") && u.Host != ""
}
