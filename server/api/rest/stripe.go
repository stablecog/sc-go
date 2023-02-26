package rest

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/render"
	"github.com/stablecog/sc-go/server/responses"
	"github.com/stablecog/sc-go/utils"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/webhook"
	"golang.org/x/exp/slices"
	"k8s.io/klog/v2"
)

type StripeSubscriptionRequest struct {
	TargetPriceID string `json:"target_price_id"`
	Currency      string `json:"currency,omitempty"`
}

var PriceIDs = map[int]string{
	// ultimate
	3: "price_1Mf591ATa0ehBYTA6ggpEEkA",
	// pro
	2: "price_1Mf50bATa0ehBYTAPOcfnOjG",
	// starter
	1: "price_1Mf56NATa0ehBYTAHkCUablG",
}

// For creating a new subscription or upgrading one
// Rejects, if they have a subscription that is at a higher level than the target priceID
func (c *RestAPI) HandleCreateCheckoutSession(w http.ResponseWriter, r *http.Request) {
	userID, _ := c.GetUserIDAndEmailIfAuthenticated(w, r)
	if userID == nil {
		return
	}

	// Parse request body
	reqBody, _ := io.ReadAll(r.Body)
	var generateReq StripeSubscriptionRequest
	err := json.Unmarshal(reqBody, &generateReq)
	if err != nil {
		responses.ErrUnableToParseJson(w, r)
		return
	}

	// Make sure price ID exists in map
	var targetPriceID string
	var targetPriceLevel int
	for level, priceID := range PriceIDs {
		if priceID == generateReq.TargetPriceID {
			targetPriceID = priceID
			targetPriceLevel = level
			break
		}
	}
	if targetPriceID == "" {
		responses.ErrBadRequest(w, r, "invalid_price_id")
		return
	}

	// Get user
	user, err := c.Repo.GetUser(*userID)
	if err != nil {
		klog.Errorf("Error getting user: %v", err)
		responses.ErrInternalServerError(w, r, "An unknown error has occured")
		return
	}

	// Get subscription
	customer, err := c.StripeClient.Customers.Get(user.StripeCustomerID, nil)

	if err != nil {
		klog.Errorf("Error getting customer: %v", err)
		responses.ErrInternalServerError(w, r, "An unknown error has occured")
		return
	}

	var currentPriceID string
	if customer.Subscriptions != nil {
		for _, sub := range customer.Subscriptions.Data {
			if sub.Status == stripe.SubscriptionStatusActive && sub.CancelAt == 0 {
				for _, item := range sub.Items.Data {
					if item.Price.ID == targetPriceID {
						responses.ErrBadRequest(w, r, "already_subscribed")
						return
					}
					// If price ID is in map it's valid
					for _, priceID := range PriceIDs {
						if item.Price.ID == priceID {
							currentPriceID = item.Price.ID
							break
						}
					}
				}
				break
			}
		}
	}

	// If they have a current one, make sure they are upgrading
	if currentPriceID != "" {
		var currentPriceLevel int
		for level, priceID := range PriceIDs {
			if priceID == currentPriceID {
				currentPriceLevel = level
				break
			}
		}

		if currentPriceLevel >= targetPriceLevel {
			responses.ErrBadRequest(w, r, "cannot_downgrade")
			return
		}
	}

	// Create checkout session
	params := &stripe.CheckoutSessionParams{
		Customer: stripe.String(user.StripeCustomerID),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(targetPriceID),
				Quantity: stripe.Int64(1),
			},
		},
		Mode:       stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		SuccessURL: stripe.String(utils.GetPurchaseSucceededURL()),
		CancelURL:  stripe.String(utils.GetPurcahseCancelledURL()),
		Currency:   stripe.String(generateReq.Currency),
	}

	session, err := c.StripeClient.CheckoutSessions.New(params)
	if err != nil {
		klog.Errorf("Error creating checkout session: %v", err)
		responses.ErrInternalServerError(w, r, "An unknown error has occured")
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, session)
}

// HTTP Post - handle stripe subscription downgrade
// Rejects if they don't have a subscription, or if they are not downgrading
func (c *RestAPI) HandleSubscriptionDowngrade(w http.ResponseWriter, r *http.Request) {
	userID, _ := c.GetUserIDAndEmailIfAuthenticated(w, r)
	if userID == nil {
		return
	}

	// Parse request body
	reqBody, _ := io.ReadAll(r.Body)
	var generateReq StripeSubscriptionRequest
	err := json.Unmarshal(reqBody, &generateReq)
	if err != nil {
		responses.ErrUnableToParseJson(w, r)
		return
	}

	// Validate currency
	if !slices.Contains([]string{"usd", "eur"}, generateReq.Currency) {
		responses.ErrBadRequest(w, r, "invalid_currency")
		return
	}

	// Make sure price ID exists in map
	var targetPriceID string
	var targetPriceLevel int
	for level, priceID := range PriceIDs {
		if priceID == generateReq.TargetPriceID {
			targetPriceID = priceID
			targetPriceLevel = level
			break
		}
	}
	if targetPriceID == "" {
		responses.ErrBadRequest(w, r, "invalid_price_id")
		return
	}

	// Get user
	user, err := c.Repo.GetUser(*userID)
	if err != nil {
		klog.Errorf("Error getting user: %v", err)
		responses.ErrInternalServerError(w, r, "An unknown error has occured")
		return
	}

	// Get subscription
	customer, err := c.StripeClient.Customers.Get(user.StripeCustomerID, nil)

	if err != nil {
		klog.Errorf("Error getting customer: %v", err)
		responses.ErrInternalServerError(w, r, "An unknown error has occured")
		return
	}

	if customer.Subscriptions == nil || len(customer.Subscriptions.Data) == 0 || customer.Subscriptions.TotalCount == 0 {
		responses.ErrBadRequest(w, r, "no_active_subscription")
		return
	}

	var currentPriceID string
	var currentSubId string
	for _, sub := range customer.Subscriptions.Data {
		if sub.Status == stripe.SubscriptionStatusActive && sub.CancelAt == 0 {
			for _, item := range sub.Items.Data {
				// If price ID is in map it's valid
				for _, priceID := range PriceIDs {
					if item.Price.ID == priceID {
						currentPriceID = item.Price.ID
						currentSubId = sub.ID
						break
					}
				}
				break
			}
		}
	}

	if currentPriceID == "" {
		responses.ErrBadRequest(w, r, "no_active_subscription")
		return
	}

	if currentPriceID == targetPriceID {
		responses.ErrBadRequest(w, r, "no_downgrade_needed")
		return
	}

	// Make sure this is a downgrade
	for level, priceID := range PriceIDs {
		if priceID == currentPriceID {
			if level <= targetPriceLevel {
				responses.ErrBadRequest(w, r, "no_downgrade_needed")
				return
			}
			break
		}
	}

	// Execute subscription update
	_, err = c.StripeClient.Subscriptions.Update(currentSubId, &stripe.SubscriptionParams{
		ProrationBehavior: stripe.String("none"),
		Items: []*stripe.SubscriptionItemsParams{
			{
				ID:   stripe.String(currentSubId),
				Plan: stripe.String(targetPriceID),
			},
		},
	})

	if err != nil {
		klog.Errorf("Error updating subscription: %v", err)
		responses.ErrInternalServerError(w, r, "An unknown error has occured")
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, map[string]interface{}{
		"success": true,
	})
}

// invoice.payment_succeeded
// customer.subscription.created
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

	switch event.Type {
	// For subscription upgrades, we want to cancel all old subscriptions (schedule to cancel)
	case "customer.subscription.created":
		newSub, err := stripeObjectMapToSubscriptionObject(event.Data.Object)
		if err != nil || newSub == nil {
			klog.Errorf("Unable parsing stripe subscription object: %v", err)
			responses.ErrInternalServerError(w, r, err.Error())
			return
		}
		// We need to see if they have more than one subscription
		subIter := c.StripeClient.Subscriptions.List(&stripe.SubscriptionListParams{
			Customer: stripe.String(newSub.Customer.ID),
		})
		for subIter.Next() {
			sub := subIter.Subscription()
			if sub.ID != newSub.ID {
				// We need to cancel this subscription
				_, err := c.StripeClient.Subscriptions.Update(sub.ID, &stripe.SubscriptionParams{
					CancelAtPeriodEnd: stripe.Bool(true),
				})
				if err != nil {
					klog.Errorf("Unable canceling stripe subscription: %v", err)
					responses.ErrInternalServerError(w, r, err.Error())
					return
				}
			}
		}
	case "invoice.payment_succeeded":
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

// Parse generic object into stripe subscription struct
func stripeObjectMapToSubscriptionObject(obj map[string]interface{}) (*stripe.Subscription, error) {
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
