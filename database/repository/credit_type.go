package repository

import (
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/ent/credittype"
)

const FREE_CREDIT_TYPE_ID = "3b12b23e-478b-4c18-8e34-70b3f0af1ee6"
const REFUND_CREDIT_TYPE_ID = "7ca94fd6-c201-4ca6-a9bf-4473c83e30b4"

func (r *Repository) CreateCreditType(name string, amount int32, description *string, stripeProductID *string, ctype credittype.Type) (*ent.CreditType, error) {
	create := r.DB.CreditType.Create().SetName(name).SetAmount(amount).SetType(ctype)
	if description != nil {
		create.SetDescription(*description)
	}
	if stripeProductID != nil {
		create.SetStripeProductID(*stripeProductID)
	}
	return create.Save(r.Ctx)
}

func (r *Repository) GetCreditTypeByID(id uuid.UUID) (*ent.CreditType, error) {
	creditType, err := r.DB.CreditType.Query().Where(credittype.IDEQ(id)).Only(r.Ctx)
	if err != nil && ent.IsNotFound(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return creditType, nil
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

func (r *Repository) GetOrCreateFreeCreditType() (*ent.CreditType, error) {
	freeId := uuid.MustParse(FREE_CREDIT_TYPE_ID)
	creditType, err := r.DB.CreditType.Query().Where(credittype.IDEQ(freeId)).Only(r.Ctx)
	if err != nil && ent.IsNotFound(err) {
		// Create it
		creditType, err := r.DB.CreditType.Create().SetID(freeId).SetName("Free").SetAmount(100).SetType(credittype.TypeFree).Save(r.Ctx)
		if err != nil {
			return nil, err
		}
		return creditType, nil
	} else if err != nil {
		return nil, err
	}
	return creditType, nil
}

// Refund credit types are used to refund credits in case of generation failure, they do not expire
func (r *Repository) GetOrCreateRefundCreditType(db *ent.Client) (*ent.CreditType, error) {
	if db == nil {
		db = r.DB
	}

	freeId := uuid.MustParse(REFUND_CREDIT_TYPE_ID)
	creditType, err := db.CreditType.Query().Where(credittype.IDEQ(freeId)).Only(r.Ctx)
	if err != nil && ent.IsNotFound(err) {
		// Create it
		creditType, err := db.CreditType.Create().SetID(freeId).SetName("Refund").SetDescription("For generate/upscale failure refunds").SetAmount(0).SetType(credittype.TypeOneTime).Save(r.Ctx)
		if err != nil {
			return nil, err
		}
		return creditType, nil
	} else if err != nil {
		return nil, err
	}
	return creditType, nil
}

func (r *Repository) GetCreditTypeList() ([]*ent.CreditType, error) {
	return r.DB.CreditType.Query().Where(credittype.TypeEQ(credittype.TypeOneTime)).All(r.Ctx)
}
