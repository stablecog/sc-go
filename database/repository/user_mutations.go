package repository

import (
	"fmt"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/ent/disposableemail"
	"github.com/stablecog/sc-go/database/ent/user"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/shared"
	"github.com/stablecog/sc-go/utils"
)

func (r *Repository) CreateUser(id uuid.UUID, email string, stripeCustomerId string, lastSignIn *time.Time, db *ent.Client) (*ent.User, error) {
	if db == nil {
		db = r.DB
	}
	cq := db.User.Create().SetID(id).SetStripeCustomerID(stripeCustomerId).SetEmail(email).SetUsername(utils.GenerateUsername(nil))
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

// Ban domains
func (r *Repository) BanDomains(domains []string) (int, error) {
	var bannedUsers int
	if err := r.WithTx(func(tx *ent.Tx) error {
		DB := tx.Client()

		// Insert into disposable emails
		bulk := make([]*ent.DisposableEmailCreate, len(domains))
		for i, domain := range domains {
			bulk[i] = DB.DisposableEmail.Create().SetDomain(domain)
		}
		err := DB.DisposableEmail.CreateBulk(bulk...).OnConflict().DoNothing().Exec(r.Ctx)
		if err != nil {
			return err
		}

		// Update users with domain like this
		for _, domain := range domains {
			updated, err := DB.User.Update().Where(func(s *sql.Selector) {
				s.Where(sql.Like(user.FieldEmail, fmt.Sprintf("%%@%s", domain)))
			}).SetBannedAt(time.Now()).SetScheduledForDeletionOn(time.Now().Add(3 * time.Hour)).Save(r.Ctx)
			if err != nil {
				return err
			}
			bannedUsers += updated
		}

		return nil
	}); err != nil {
		log.Error("Error banning domains", "err", err)
		return 0, err
	}

	return bannedUsers, nil
}

// Ban domains
func (r *Repository) UnbanDomains(domains []string) (int, error) {
	var bannedUsers int
	if err := r.WithTx(func(tx *ent.Tx) error {
		DB := tx.Client()

		// Delete from disposable emails
		_, err := DB.DisposableEmail.Delete().Where(disposableemail.DomainIn(domains...)).Exec(r.Ctx)
		if err != nil {
			return err
		}

		// Update users with domain like this
		for _, domain := range domains {
			updated, err := DB.User.Update().Where(func(s *sql.Selector) {
				s.Where(sql.Like(user.FieldEmail, fmt.Sprintf("%%@%s", domain)))
			}).ClearBannedAt().ClearScheduledForDeletionOn().Save(r.Ctx)
			if err != nil {
				return err
			}
			bannedUsers += updated
		}

		return nil
	}); err != nil {
		log.Error("Error unbanning domains", "err", err)
		return 0, err
	}

	return bannedUsers, nil
}

// Unban users
func (r *Repository) UnbanUsers(userIDs []uuid.UUID) (int, error) {
	return r.DB.User.Update().Where(user.IDIn(userIDs...)).ClearBannedAt().ClearScheduledForDeletionOn().Save(r.Ctx)
}

// Set wants email
func (r *Repository) SetWantsEmail(userId uuid.UUID, wantsEmail bool) error {
	return r.DB.User.Update().Where(user.IDEQ(userId)).SetWantsEmail(wantsEmail).Exec(r.Ctx)
}

// Set discord ID on user
func (r *Repository) SetDiscordID(userId uuid.UUID, discordId string, DB *ent.Client) error {
	if DB == nil {
		DB = r.DB
	}
	return DB.User.Update().Where(user.IDEQ(userId)).SetDiscordID(discordId).Exec(r.Ctx)
}

var UsernameExistsErr = fmt.Errorf("username_exists")

func (r *Repository) SetUsername(userId uuid.UUID, username string) error {
	if err := r.WithTx(func(tx *ent.Tx) error {
		// See if username exists case insenstive
		DB := tx.Client()

		c, err := DB.User.Query().Where(func(s *sql.Selector) {
			s.Where(sql.EQ(sql.Lower(user.FieldUsername), strings.ToLower(username)))
			s.Where(sql.NEQ(user.FieldID, userId))
		}).Count(r.Ctx)

		if err != nil {
			return err
		}

		if c > 0 {
			return UsernameExistsErr
		}

		return DB.User.Update().Where(user.IDEQ(userId)).SetUsername(username).SetUsernameChangedAt(time.Now()).Exec(r.Ctx)
	}); err != nil {
		log.Errorf("Error setting username: %s", err)
		return err
	}
	return nil
}
