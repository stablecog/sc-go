// * Responses from user-initiated endpoints
package responses

import (
	"time"

	"github.com/google/uuid"
)

// API generate simply returns a UUID to track the request to our compute while its in flight
type QueuedResponse struct {
	ID string `json:"id"`
}

// Response for submitting to gallery
type GenerateSubmitToGalleryResponse struct {
	Submitted int `json:"submitted"`
}

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

type UserCreditsResponse struct {
	TotalRemainingCredits int32    `json:"total_remaining_credits"`
	Credits               []Credit `json:"credits"`
}
