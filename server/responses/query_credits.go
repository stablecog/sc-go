package responses

import (
	"time"

	"github.com/google/uuid"
)

// Response for retrieving user credits
type CreditType struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Amount      int32     `json:"amount"`
	Description string    `json:"description"`
}

type Credit struct {
	ID              uuid.UUID  `json:"id"`
	RemainingAmount int32      `json:"remaining_amount"`
	ExpiresAt       time.Time  `json:"expires_at"`
	Type            CreditType `json:"type"`
}

type QueryCreditsResponse struct {
	TotalRemainingCredits int32    `json:"total_remaining_credits"`
	Credits               []Credit `json:"credits"`
}
