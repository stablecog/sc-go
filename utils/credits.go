package utils

import (
	"math"

	"github.com/stablecog/sc-go/shared"
)

func CalculateVoiceoverCredits(prompt string) int32 {
	return int32(math.Ceil(shared.VOICEOVER_CREDIT_COST_PER_CHARACTER * float64(len(prompt))))
}
