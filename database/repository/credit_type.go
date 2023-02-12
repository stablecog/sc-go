package repository

import "github.com/stablecog/sc-go/database/ent"

func (r *Repository) CreateCreditType(name string, amount int32, description *string, stripeProductID *string) (*ent.CreditType, error) {
	create := r.DB.CreditType.Create().SetName(name).SetAmount(amount)
	if description != nil {
		create.SetDescription(*description)
	}
	if stripeProductID != nil {
		create.SetStripeProductID(*stripeProductID)
	}
	return create.Save(r.Ctx)
}
