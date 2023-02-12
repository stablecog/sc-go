package rest

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/render"
	"github.com/stablecog/sc-go/utils"
	"github.com/stripe/stripe-go"
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
	if invoice.BillingReason != stripe.InvoiceBillingReasonSubscriptionCycle && invoice.BillingReason != stripe.InvoiceBillingReasonSubscriptionCreate {
		render.Status(r, http.StatusOK)
		return
	}

	subscription := invoice.Subscription
	if subscription == nil {
		klog.Errorf("Stripe subscription is nil %s", invoice.ID)
		render.Status(r, http.StatusServiceUnavailable)
		return
	}

	if subscription.Plan == nil && subscription.Plan.Product == nil {
		klog.Errorf("Stripe plan or subscription is nil %s", subscription.ID)
		render.Status(r, http.StatusServiceUnavailable)
		return
	}

	if subscription.Customer == nil {
		klog.Errorf("Stripe customer is nil %s", subscription.ID)
		render.Status(r, http.StatusServiceUnavailable)
		return
	}

	// Get user from customer ID
	user, err := c.Repo.GetUserByStripeCustomerId(subscription.Customer.ID)
	if err != nil {
		klog.Errorf("Unable getting user from stripe customer id: %v", err)
		render.Status(r, http.StatusServiceUnavailable)
		return
	} else if user == nil {
		klog.Errorf("User does not exist with stripe customer id: %s", subscription.Customer.ID)
		render.Status(r, http.StatusServiceUnavailable)
		return
	}

	// Get the credit type for this plan
	creditType, err := c.Repo.GetCreditTypeByStripeProductID(subscription.Plan.Product.ID)
	if err != nil {
		klog.Errorf("Unable getting credit type from stripe product id: %v", err)
		render.Status(r, http.StatusServiceUnavailable)
		return
	} else if creditType == nil {
		klog.Errorf("Credit type does not exist with stripe product id: %s", subscription.Plan.Product.ID)
		render.Status(r, http.StatusServiceUnavailable)
		return
	}

	expiresAt := utils.SecondsSinceEpochToTime(subscription.CurrentPeriodEnd)

	// Update user credit
	_, err = c.Repo.AddCreditsIfEligible(creditType, user.ID, expiresAt)
	if err != nil {
		klog.Errorf("Unable adding credits to user %s: %v", user.ID.String(), err)
	}
}

// Parse generic object into stripe invoice struct
func stripeObjectMapToInvoiceObject(obj map[string]interface{}) (*stripe.Invoice, error) {
	marshalled, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	var invoice stripe.Invoice
	err = json.Unmarshal(marshalled, &invoice)
	if err != nil {
		return nil, err
	}
	return &invoice, nil
}
