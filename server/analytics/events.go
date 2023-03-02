package analytics

import (
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
)

// Sign Up
func (a *AnalyticsService) SignUp(userId uuid.UUID, email, ipAddress string) error {
	return a.Dispatch(Event{
		DistinctId: userId.String(),
		EventName:  "Sign Up",
		Properties: map[string]interface{}{
			"email":      email,
			"SC - Email": email,
			"$ip":        ipAddress,
		},
	})
}

// New Subscription
func (a *AnalyticsService) Subscription(user *ent.User, productId string) error {
	return a.Dispatch(Event{
		DistinctId: user.ID.String(),
		EventName:  "Subscription",
		Properties: map[string]interface{}{
			"SC - Stripe Product Id": productId,
			"SC - Email":             user.Email,
			"SC - Stripe ID":         user.StripeCustomerID,
			"$geoip_disable":         true,
		},
	})
}

// Renewed Subscription
func (a *AnalyticsService) SubscriptionRenewal(user *ent.User, productId string) error {
	return a.Dispatch(Event{
		DistinctId: user.ID.String(),
		EventName:  "Subscription | Renewed",
		Properties: map[string]interface{}{
			"SC - Stripe Product Id": productId,
			"SC - Email":             user.Email,
			"SC - Stripe ID":         user.StripeCustomerID,
			"$geoip_disable":         true,
		},
	})
}

// Cancelled Subscription
func (a *AnalyticsService) SubscriptionCancelled(user *ent.User, productId string) error {
	return a.Dispatch(Event{
		DistinctId: user.ID.String(),
		EventName:  "Subscription | Cancelled",
		Properties: map[string]interface{}{
			"SC - Stripe Product Id": productId,
			"SC - Email":             user.Email,
			"SC - Stripe ID":         user.StripeCustomerID,
			"$geoip_disable":         true,
		},
	})
}

// Upgraded subscription
func (a *AnalyticsService) SubscriptionUpgraded(user *ent.User, oldProductId string, productId string) error {
	return a.Dispatch(Event{
		DistinctId: user.ID.String(),
		EventName:  "Subscription | Upgraded",
		Properties: map[string]interface{}{
			"SC - Old Product Id":    oldProductId,
			"SC - Stripe Product Id": productId,
			"SC - Email":             user.Email,
			"SC - Stripe ID":         user.StripeCustomerID,
			"$geoip_disable":         true,
		},
	})
}

// Credit purchase
func (a *AnalyticsService) CreditPurchase(user *ent.User, productId string, amount int) error {
	return a.Dispatch(Event{
		DistinctId: user.ID.String(),
		EventName:  "Credit | Purchase",
		Properties: map[string]interface{}{
			"SC - Stripe Product Id": productId,
			"SC - Email":             user.Email,
			"SC - Stripe ID":         user.StripeCustomerID,
			"SC - Amount":            amount,
			"$geoip_disable":         true,
		},
	})
}

// Free credits replenished
func (a *AnalyticsService) FreeCreditsReplenished(userId uuid.UUID, email string, amount int) error {
	return a.Dispatch(Event{
		DistinctId: userId.String(),
		EventName:  "Credit | Free Replenished",
		Properties: map[string]interface{}{
			"SC - Email":     email,
			"SC - Amount":    amount,
			"$geoip_disable": true,
		},
	})
}
