package responses

import "time"

type GetUserResponse struct {
	TotalRemainingCredits int        `json:"total_remaining_credits"`
	Product               string     `json:"product,omitempty"`
	CancelsAt             *time.Time `json:"cancels_at,omitempty"`
	StripeHadError        bool       `json:"stripe_had_error"`
}
