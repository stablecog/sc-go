package repository

import (
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/ent/credittype"
)

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

func (r *Repository) GetCreditTypeByStripeProductID(stripeProductID string) (*ent.CreditType, error) {
	creditType, err := r.DB.CreditType.Query().Where(credittype.StripeProductIDEQ(stripeProductID)).Only(r.Ctx)
	if err != nil && ent.IsNotFound(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return creditType, nil
}

func (r *Repository) GetFreeCreditType() (*ent.CreditType, error) {
	creditType, err := r.DB.CreditType.Query().Where(credittype.NameEQ("free")).Only(r.Ctx)
	if err != nil && ent.IsNotFound(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return creditType, nil
}
