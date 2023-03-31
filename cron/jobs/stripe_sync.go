package jobs

import "github.com/stripe/stripe-go/v74"

// Sync stripe subscriptions with active user products
func (j *JobRunner) SyncStripe(log Logger) error {
	iter := j.Stripe.Customers.List(&stripe.CustomerListParams{
		ListParams: stripe.ListParams{
			Expand: []*string{
				stripe.String("subscriptions"),
			},
		},
	})
	// Customers without subscriptions
	customersNoSubscriptions := []string{}
	productCustomerMap := map[string][]string{}
	for iter.Next() {
		customer := iter.Customer()
		found := false
		if customer.Subscriptions.TotalCount == 0 {
			customersNoSubscriptions = append(customersNoSubscriptions, customer.ID)
			continue
		}
		for _, sub := range customer.Subscriptions.Data {
			if sub.Status != stripe.SubscriptionStatusActive && sub.Status != stripe.SubscriptionStatusTrialing {
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
				productCustomerMap[item.Price.Product.ID] = append(productCustomerMap[item.Price.Product.ID], customer.ID)
				found = true
				break
			}
			if found {
				break
			}
		}
		if !found {
			customersNoSubscriptions = append(customersNoSubscriptions, customer.ID)
		}
	}

	log.Infof("Customers without subscriptions: %d", len(customersNoSubscriptions))

	for k, v := range productCustomerMap {
		log.Infof("Product %s has %d customers", k, len(v))
	}
	return nil
}
