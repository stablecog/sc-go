package responses

import "time"

type GetUserResponse struct {
	TotalRemainingCredits int32      `json:"total_remaining_credits"`
	Product               string     `json:"product,omitempty"`
	CancelsAt             *time.Time `json:"cancels_at,omitempty"`
	CustomerNotFound      bool       `json:"customer_not_found"`
	StripeHadError        bool       `json:"stripe_had_error"`
}
