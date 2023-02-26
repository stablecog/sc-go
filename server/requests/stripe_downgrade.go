package requests

type StripeDowngradeRequest struct {
	TargetPriceID string `json:"target_price_id"`
}
