package requests

type StripeCheckoutRequest struct {
	TargetPriceID   string `json:"target_price_id"`
	Currency        string `json:"currency,omitempty"`
	SuccessUrl      string `json:"success_url"`
	CancelUrl       string `json:"cancel_url"`
	PromotionCodeID string `json:"promotion_code_id,omitempty"`
}
