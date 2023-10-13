package responses

import (
	"time"

	"github.com/google/uuid"
)

type GetUserResponse struct {
	UserID                  *uuid.UUID `json:"user_id,omitempty"`
	TotalRemainingCredits   int        `json:"total_remaining_credits"`
	HasNonfreeCredits       bool       `json:"has_nonfree_credits"`
	ProductID               string     `json:"product_id,omitempty"`
	PriceID                 string     `json:"price_id,omitempty"`
	CancelsAt               *time.Time `json:"cancels_at,omitempty"`
	RenewsAt                *time.Time `json:"renews_at,omitempty"`
	RenewsAtAmount          *int       `json:"renews_at_credit_amount,omitempty"`
	MoreCreditsAt           *time.Time `json:"more_credits_at,omitempty"`
	MoreCreditsAtAmount     *int       `json:"more_credits_at_credit_amount,omitempty"`
	MoreFreeCreditsAt       *time.Time `json:"more_free_credits_at,omitempty"`
	MoreFreeCreditsAtAmount *int       `json:"more_free_credits_at_credit_amount,omitempty"`
	WantsEmail              *bool      `json:"wants_email,omitempty"`
	Username                string     `json:"username,omitempty"`
	CreatedAt               time.Time  `json:"created_at"`
	UsernameChangedAt       *time.Time `json:"username_changed_at,omitempty"`
	PurchaseCount           int        `json:"purchase_count"`
	// The current amoount of free credits server offers
	FreeCreditAmount *int     `json:"free_credit_amount,omitempty"`
	StripeHadError   bool     `json:"stripe_had_error"`
	Roles            []string `json:"roles,omitempty"`
}
