package utils

import (
	"math/rand"
	"time"
)

// RandInt provides a common interface for random integer generation.
type RandInt interface {
	Intn(n int) int
}

// MathRandInt is a RandInt that uses math/rand.
type MathRandInt struct {
	*rand.Rand
}

func NewMathRandInt(seed int64) *MathRandInt {
	return &MathRandInt{rand.New(rand.NewSource(seed))}
}

func randomString(randInt RandInt, length int) string {
	// Generate a random string of length
	const letterBytes = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		randomIndex := randInt.Intn(len(letterBytes))
		b[i] = letterBytes[randomIndex]
	}
	return string(b)
}

func GenerateUsername(randInt RandInt) string {
	if randInt == nil {
		randInt = NewMathRandInt(time.Now().UnixNano())
	}
	// Generate random 12 character username that is allowed
	username := randomString(randInt, 12)
	for IsValidUsername(username) != nil {
		username = randomString(randInt, 12)
	}
	return username
}
