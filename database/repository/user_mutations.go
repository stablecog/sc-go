package repository

import (
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
)

func (r *Repository) CreateUser(id uuid.UUID, email string, stripeCustomerId string, db *ent.Client) (*ent.User, error) {
	if db == nil {
		db = r.DB
	}
	return db.User.Create().SetID(id).SetStripeCustomerID(stripeCustomerId).SetEmail(email).Save(r.Ctx)
}
