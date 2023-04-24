package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/ent/user"
	"github.com/stablecog/sc-go/shared"
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

// Sync stripe product IDs
func (r *Repository) SyncStripeProductIDs(productCustomerIDMap map[string][]string) error {
	if err := r.WithTx(func(tx *ent.Tx) error {
		allCustomersWithProducts := make([]string, 0)
		for productID, customerIDs := range productCustomerIDMap {
			allCustomersWithProducts = append(allCustomersWithProducts, customerIDs...)
			_, err := tx.User.Update().Where(user.StripeCustomerIDIn(customerIDs...)).SetActiveProductID(productID).Save(r.Ctx)
			if err != nil {
				return err
			}
		}
		err := tx.User.Update().Where(user.StripeCustomerIDNotIn(allCustomersWithProducts...)).ClearActiveProductID().Exec(r.Ctx)
		return err
	}); err != nil {
		return err
	}
	return nil
}

// Ban users
func (r *Repository) BanUsers(userIDs []uuid.UUID) (int, error) {
	return r.DB.User.Update().Where(user.IDIn(userIDs...)).SetBannedAt(time.Now()).SetScheduledForDeletionOn(time.Now().Add(shared.DELETE_BANNED_USER_DATA_AFTER)).Save(r.Ctx)
}

// Unban users
func (r *Repository) UnbanUsers(userIDs []uuid.UUID) (int, error) {
	return r.DB.User.Update().Where(user.IDIn(userIDs...)).ClearBannedAt().ClearScheduledForDeletionOn().Save(r.Ctx)
}
