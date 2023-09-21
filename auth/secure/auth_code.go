package secure

import (
	"crypto/rand"
	"math/big"
)

// Use crypto/random to generate a secure auth code
func GenerateAuthCode(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)

	for i := range result {
		randomInt, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		result[i] = charset[randomInt.Int64()]
	}

	return string(result), nil
}
