package rest

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/render"
	"github.com/stablecog/go-apps/utils"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/webhook"
	"k8s.io/klog/v2"
)

// customer.subscription.created
// customer.subscription.deleted
// customer.subscription.updated
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

	// We can parse the object as a subscription since that's the only thing we care about
	subscription, err := stripeObjectMapToSubscription(event.Data.Object)
	if err != nil || subscription == nil {
		klog.Errorf("Unable parsing stripe subscription object: %v", err)
		render.Status(r, http.StatusServiceUnavailable)
		return
	}

	if subscription.Plan == nil && subscription.Plan.Product == nil {
		klog.Errorf("Stripe plan or subscription is nil %s", subscription.ID)
		render.Status(r, http.StatusServiceUnavailable)
		return
	}

	switch event.Type {
	case "customer.subscription.created":
		klog.Infof("CREATED %s", subscription.Customer.ID)
	case "customer.subscription.deleted":
		klog.Infof("DELETED %s", subscription.Customer.ID)
	case "customer.subscription.updated":
		klog.Infof("UPDATED %s", subscription.Customer.ID)
	}

}

// Parse generic object into stripe subscription struct
func stripeObjectMapToSubscription(obj map[string]interface{}) (*stripe.Subscription, error) {
	marshalled, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	var subscription stripe.Subscription
	err = json.Unmarshal(marshalled, &subscription)
	if err != nil {
		return nil, err
	}
	return &subscription, nil
}
