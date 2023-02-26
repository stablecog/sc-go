package responses

type StripeSessionResponse struct {
	CheckoutURL       string `json:"checkout_url,omitempty"`
	CustomerPortalURL string `json:"customer_portal_url,omitempty"`
}
