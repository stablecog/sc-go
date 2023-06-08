package utils

import (
	"math"
	"unicode/utf8"

	"github.com/stablecog/sc-go/shared"
)

func CalculateVoiceoverCredits(prompt string) int32 {
	return int32(math.Ceil(shared.VOICEOVER_CREDIT_COST_PER_CHARACTER * float64(utf8.RuneCountInString(prompt))))
}
