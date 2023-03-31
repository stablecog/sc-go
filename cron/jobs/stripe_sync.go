package jobs

import (
	"time"

	"github.com/stripe/stripe-go/v74"
)

// Sync stripe subscriptions with active user products
func (j *JobRunner) SyncStripe(log Logger) error {
	log.Infof("Starting stripe customer sync job...")
	start := time.Now()
	iter := j.Stripe.Subscriptions.List(&stripe.SubscriptionListParams{
		Status: stripe.String(string(stripe.SubscriptionStatusActive)),
	})
	// Customers without subscriptions
	productCustomerMap := map[string][]string{}
	for iter.Next() {
		sub := iter.Subscription()
		if sub.Customer == nil {
			log.Errorf("Customer is nil for subscription %s", sub.ID)
			continue
		}
		if sub.Items == nil || sub.Items.TotalCount == 0 {
			log.Errorf("Items is nil for subscription %s", sub.ID)
			continue
		}
		for _, item := range sub.Items.Data {
			if item.Price == nil {
				continue
			}
			if item.Price.Product == nil {
				log.Errorf("Product is nil for price %s", item.Price.ID)
				continue
			}
			productCustomerMap[item.Price.Product.ID] = append(productCustomerMap[item.Price.Product.ID], sub.Customer.ID)
			break
		}
	}

	err := j.Repo.SyncStripeProductIDs(productCustomerMap)
	if err != nil {
		log.Errorf("Error syncing stripe product ids: %v", err)
		return err
	}
	end := time.Now()
	log.Infof("Finished stripe customer sync job in %v", end.Sub(start))
	return nil
}
