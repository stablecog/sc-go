package rest

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/render"
	"github.com/stablecog/sc-go/server/responses"
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
		responses.ErrBadRequest(w, r, "invalid stripe webhook body")
		return
	}

	// Verify signature
	endpointSecret := utils.GetEnv("STRIPE_ENDPOINT_SECRET", "")

	event, err := webhook.ConstructEvent(reqBody, r.Header.Get("Stripe-Signature"), endpointSecret)
	if err != nil {
		klog.Errorf("Unable verifying stripe webhook signature: %v", err)
		responses.ErrBadRequest(w, r, "invalid stripe webhook signature")
		return
	}

	// We can parse the object as an invoice since that's the only thing we care about
	invoice, err := stripeObjectMapToInvoiceObject(event.Data.Object)
	if err != nil || invoice == nil {
		klog.Errorf("Unable parsing stripe invoice object: %v", err)
		responses.ErrInternalServerError(w, r, err.Error())
		return
	}

	// We only care about renewal (cycle), create, and manual
	if invoice.BillingReason != InvoiceBillingReasonSubscriptionCycle && invoice.BillingReason != InvoiceBillingReasonSubscriptionCreate && invoice.BillingReason != InvoiceBillingReasonManual {
		render.Status(r, http.StatusOK)
		render.PlainText(w, r, "OK")
		return
	}

	if invoice.Lines == nil {
		klog.Errorf("Stripe invoice lines is nil %s", invoice.ID)
		responses.ErrInternalServerError(w, r, "Stripe invoice lines is nil")
		return
	}

	for _, line := range invoice.Lines.Data {
		var product string
		if line.Plan == nil && invoice.BillingReason != InvoiceBillingReasonManual {
			klog.Errorf("Stripe plan is nil in line item %s", line.ID)
			responses.ErrInternalServerError(w, r, "Stripe plan is nil in line item")
			return
		}

		if line.Price == nil && invoice.BillingReason == InvoiceBillingReasonManual {
			klog.Errorf("Stripe price is nil in line item %s", line.ID)
			responses.ErrInternalServerError(w, r, "Stripe price is nil in line item")
			return
		}

		if invoice.BillingReason == InvoiceBillingReasonManual {
			product = line.Price.Product
		} else {
			product = line.Plan.Product
		}

		if product == "" {
			klog.Errorf("Stripe product is nil in line item %s", line.ID)
			responses.ErrInternalServerError(w, r, "Stripe product is nil in line item")
			return
		}

		// Get user from customer ID
		user, err := c.Repo.GetUserByStripeCustomerId(invoice.Customer)
		if err != nil {
			klog.Errorf("Unable getting user from stripe customer id: %v", err)
			responses.ErrInternalServerError(w, r, err.Error())
			return
		} else if user == nil {
			klog.Errorf("User does not exist with stripe customer id: %s", invoice.Customer)
			responses.ErrInternalServerError(w, r, "User does not exist with stripe customer id")
			return
		}

		// Get the credit type for this plan
		creditType, err := c.Repo.GetCreditTypeByStripeProductID(product)
		if err != nil {
			klog.Errorf("Unable getting credit type from stripe product id: %v", err)
			responses.ErrInternalServerError(w, r, err.Error())
			return
		} else if creditType == nil {
			klog.Errorf("Credit type does not exist with stripe product id: %s", line.Plan.Product)
			responses.ErrInternalServerError(w, r, "Credit type does not exist with stripe product id")
			return
		}

		if invoice.BillingReason == InvoiceBillingReasonManual {
			// Ad-hoc credit add
			_, err = c.Repo.AddAdhocCreditsIfEligible(creditType, user.ID, line.ID)
			if err != nil {
				klog.Errorf("Unable adding credits to user %s: %v", user.ID.String(), err)
				responses.ErrInternalServerError(w, r, err.Error())
				return
			}
		} else {
			expiresAt := utils.SecondsSinceEpochToTime(line.Period.End)
			// Update user credit
			_, err = c.Repo.AddCreditsIfEligible(creditType, user.ID, expiresAt, line.ID, nil)
			if err != nil {
				klog.Errorf("Unable adding credits to user %s: %v", user.ID.String(), err)
				responses.ErrInternalServerError(w, r, err.Error())
				return
			}
		}
	}

	render.Status(r, http.StatusOK)
	render.PlainText(w, r, "OK")
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

type Price struct {
	Product string `json:"product"`
}

// InvoiceLine is the resource representing a Stripe invoice line item.
// For more details see https://stripe.com/docs/api#invoice_line_item_object.
type InvoiceLine struct {
	ID     string  `json:"id"`
	Period *Period `json:"period"`
	Plan   *Plan   `json:"plan"`
	Price  *Price  `json:"price"`
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
