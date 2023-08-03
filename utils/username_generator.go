package utils

import (
	cryptorand "crypto/rand"
	"math/big"
)

type RandReader interface {
	Read([]byte) (int, error)
}

type CryptoRandReader struct{}

func (crr *CryptoRandReader) Read(b []byte) (int, error) {
	return cryptorand.Read(b)
}

func randInt(randReader RandReader, n int) (int, error) {
	val, err := cryptorand.Int(randReader, big.NewInt(int64(n)))
	if err != nil {
		return 0, err
	}
	return int(val.Int64()), nil
}

func randomString(randReader RandReader, length int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		randomIndex, _ := randInt(randReader, len(letterBytes))
		b[i] = letterBytes[randomIndex]
	}
	return string(b)
}

func GenerateUsername(randReader RandReader) string {
	if randReader == nil {
		randReader = &CryptoRandReader{}
	}
	username := randomString(randReader, 12)
	for IsValidUsername(username) != nil {
		username = randomString(randReader, 12)
	}
	return username
}
