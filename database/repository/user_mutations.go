package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/ent/user"
)

func (r *Repository) CreateUser(id uuid.UUID, email string, stripeCustomerId string, lastSignIn *time.Time, db *ent.Client) (*ent.User, error) {
	if db == nil {
		db = r.DB
	}
	cq := db.User.Create().SetID(id).SetStripeCustomerID(stripeCustomerId).SetEmail(email)
	if lastSignIn != nil {
		cq.SetLastSignInAt(*lastSignIn)
	}
	return cq.Save(r.Ctx)
}

func (r *Repository) SetActiveProductID(id uuid.UUID, stripeProductID string, db *ent.Client) error {
	if db == nil {
		db = r.DB
	}
	return db.User.UpdateOneID(id).SetActiveProductID(stripeProductID).Exec(r.Ctx)
}

// Only unset if the active product ID matches the stripe product ID given
func (r *Repository) UnsetActiveProductID(id uuid.UUID, stripeProductId string, db *ent.Client) (int, error) {
	if db == nil {
		db = r.DB
	}
	return db.User.Update().Where(user.IDEQ(id), user.ActiveProductIDEQ(stripeProductId)).ClearActiveProductID().Save(r.Ctx)
}

// Update last_seen_at
func (r *Repository) UpdateLastSeenAt(id uuid.UUID) error {
	return r.DB.User.UpdateOneID(id).SetLastSeenAt(time.Now()).Exec(r.Ctx)
}
