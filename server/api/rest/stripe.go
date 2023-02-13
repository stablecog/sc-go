package rest

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/render"
	"github.com/stablecog/sc-go/utils"
	"github.com/stripe/stripe-go/webhook"
	"k8s.io/klog/v2"
)

// invoice.payment_succeeded
func (c *RestAPI) HandleStripeWebhook(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		klog.Errorf("Unable reading stripe webhook body: %v", err)
		render.Status(r, http.StatusServiceUnavailable)
		return
	}

	// Verify signature
	endpointSecret := utils.GetEnv("STRIPE_ENDPOINT_SECRET", "")

	event, err := webhook.ConstructEvent(reqBody, r.Header.Get("Stripe-Signature"), endpointSecret)
	if err != nil {
		klog.Errorf("Unable verifying stripe webhook signature: %v", err)
		render.Status(r, http.StatusBadRequest)
		return
	}

	// We can parse the object as an invoice since that's the only thing we care about
	invoice, err := stripeObjectMapToInvoiceObject(event.Data.Object)
	if err != nil || invoice == nil {
		klog.Errorf("Unable parsing stripe invoice object: %v", err)
		render.Status(r, http.StatusServiceUnavailable)
		return
	}

	// We only care about renewal (cycle) and create
	if invoice.BillingReason != InvoiceBillingReasonSubscriptionCycle && invoice.BillingReason != InvoiceBillingReasonSubscriptionCreate {
		render.Status(r, http.StatusOK)
		return
	}

	if invoice.Lines == nil {
		klog.Errorf("Stripe invoice lines is nil %s", invoice.ID)
		render.Status(r, http.StatusServiceUnavailable)
		return
	}

	for _, line := range invoice.Lines.Data {
		if line.Plan == nil {
			klog.Errorf("Stripe plan is nil in line item %s", line.ID)
			render.Status(r, http.StatusServiceUnavailable)
		}

		// Get user from customer ID
		user, err := c.Repo.GetUserByStripeCustomerId(invoice.Customer)
		if err != nil {
			klog.Errorf("Unable getting user from stripe customer id: %v", err)
			render.Status(r, http.StatusServiceUnavailable)
			return
		} else if user == nil {
			klog.Errorf("User does not exist with stripe customer id: %s", invoice.Customer)
			render.Status(r, http.StatusServiceUnavailable)
			return
		}

		// Get the credit type for this plan
		creditType, err := c.Repo.GetCreditTypeByStripeProductID(line.Plan.Product)
		if err != nil {
			klog.Errorf("Unable getting credit type from stripe product id: %v", err)
			render.Status(r, http.StatusServiceUnavailable)
			return
		} else if creditType == nil {
			klog.Errorf("Credit type does not exist with stripe product id: %s", line.Plan.Product)
			render.Status(r, http.StatusServiceUnavailable)
			return
		}

		expiresAt := utils.SecondsSinceEpochToTime(line.Period.End)

		// Update user credit
		_, err = c.Repo.AddCreditsIfEligible(creditType, user.ID, expiresAt)
		if err != nil {
			klog.Errorf("Unable adding credits to user %s: %v", user.ID.String(), err)
			render.Status(r, http.StatusServiceUnavailable)
			return
		}
	}

	render.Status(r, http.StatusOK)
}

// Parse generic object into stripe invoice struct
func stripeObjectMapToInvoiceObject(obj map[string]interface{}) (*Invoice, error) {
	marshalled, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	var invoice Invoice
	err = json.Unmarshal(marshalled, &invoice)
	if err != nil {
		return nil, err
	}
	return &invoice, nil
}

// ! Stripe types are busted so we modify the ones included in their lib
// InvoiceBillingReason is the reason why a given invoice was created
type InvoiceBillingReason string

// List of values that InvoiceBillingReason can take.
const (
	InvoiceBillingReasonManual                InvoiceBillingReason = "manual"
	InvoiceBillingReasonSubscription          InvoiceBillingReason = "subscription"
	InvoiceBillingReasonSubscriptionCreate    InvoiceBillingReason = "subscription_create"
	InvoiceBillingReasonSubscriptionCycle     InvoiceBillingReason = "subscription_cycle"
	InvoiceBillingReasonSubscriptionThreshold InvoiceBillingReason = "subscription_threshold"
	InvoiceBillingReasonSubscriptionUpdate    InvoiceBillingReason = "subscription_update"
	InvoiceBillingReasonUpcoming              InvoiceBillingReason = "upcoming"
)

// ListMeta is the structure that contains the common properties
// of List iterators. The Count property is only populated if the
// total_count include option is passed in (see tests for example).
type ListMeta struct {
	HasMore    bool   `json:"has_more"`
	TotalCount uint32 `json:"total_count"`
	URL        string `json:"url"`
}

// Period is a structure representing a start and end dates.
type Period struct {
	End   int64 `json:"end"`
	Start int64 `json:"start"`
}

type Plan struct {
	Product string `json:"product"`
}

// InvoiceLine is the resource representing a Stripe invoice line item.
// For more details see https://stripe.com/docs/api#invoice_line_item_object.
type InvoiceLine struct {
	ID     string  `json:"id"`
	Period *Period `json:"period"`
	Plan   *Plan   `json:"plan"`
}

type InvoiceLineList struct {
	ListMeta
	Data []*InvoiceLine `json:"data"`
}

type Invoice struct {
	ID            string               `json:"id"`
	BillingReason InvoiceBillingReason `json:"billing_reason"`
	Lines         *InvoiceLineList     `json:"lines"`
	Customer      string               `json:"customer"`
}
