package responses

import (
	"time"

	"github.com/stablecog/sc-go/database/ent/userrole"
)

type GetUserResponse struct {
	TotalRemainingCredits int                 `json:"total_remaining_credits"`
	ProductID             string              `json:"product_id,omitempty"`
	CancelsAt             *time.Time          `json:"cancels_at,omitempty"`
	StripeHadError        bool                `json:"stripe_had_error"`
	Roles                 []userrole.RoleName `json:"roles,omitempty"`
}
