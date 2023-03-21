package repository

import (
	"errors"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/ent/credit"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/shared"
)

// Expiration date for manual invoices (non-recurring)
var NEVER_EXPIRE = time.Date(2100, 1, 1, 5, 0, 0, 0, time.UTC)

func (r *Repository) DeleteCreditsWithLineItemID(lineItemID string) error {
	_, err := r.DB.Credit.Delete().Where(credit.StripeLineItemIDEQ(lineItemID)).Exec(r.Ctx)
	if err != nil {
		return err
	}
	return nil
}

// Give credits to user
func (r *Repository) AddCreditsToUser(creditType *ent.CreditType, userID uuid.UUID) error {
	if creditType == nil {
		return errors.New("creditType cannot be nil")
	}

	_, err := r.DB.Credit.Create().SetCreditTypeID(creditType.ID).SetUserID(userID).SetRemainingAmount(creditType.Amount).SetExpiresAt(NEVER_EXPIRE).Save(r.Ctx)
	return err
}

// Add credits of creditType to user if they do not have any un-expired credits of this type
func (r *Repository) AddCreditsIfEligible(creditType *ent.CreditType, userID uuid.UUID, expiresAt time.Time, lineItemId string, DB *ent.Client) (added bool, err error) {
	if DB == nil {
		DB = r.DB
	}

	if creditType == nil {
		return false, errors.New("creditType cannot be nil")
	}

	// See if user has any credits of this type
	credits, err := DB.Credit.Query().Where(credit.UserID(userID), credit.CreditTypeID(creditType.ID), credit.StripeLineItemIDEQ(lineItemId)).First(r.Ctx)
	if err != nil && !ent.IsNotFound(err) {
		return false, err
	}

	if credits != nil {
		// User already has credits of this type
		return false, nil
	}

	// Add credits
	// Add an extra day to expiresAt
	expiresAtBuffer := expiresAt.AddDate(0, 0, 1)
	_, err = DB.Credit.Create().SetCreditTypeID(creditType.ID).SetUserID(userID).SetRemainingAmount(creditType.Amount).SetExpiresAt(expiresAtBuffer).SetStripeLineItemID(lineItemId).Save(r.Ctx)
	if err != nil {
		return false, err
	}
	return true, nil
}

// Give free credits if eligible
func (r *Repository) GiveFreeCredits(userID uuid.UUID, DB *ent.Client) (added bool, err error) {
	if DB == nil {
		DB = r.DB
	}

	creditType, err := r.GetOrCreateFreeCreditType()
	if err != nil {
		return false, err
	}

	// See if user has any credits of this type
	credits, err := DB.Credit.Query().Where(credit.UserID(userID), credit.CreditTypeID(creditType.ID)).First(r.Ctx)
	if err != nil && !ent.IsNotFound(err) {
		return false, err
	}

	if credits != nil {
		// User already has credits of this type
		return false, nil
	}

	// Add credits
	credits, err = DB.Credit.Create().SetCreditTypeID(creditType.ID).SetUserID(userID).SetRemainingAmount(creditType.Amount).SetExpiresAt(NEVER_EXPIRE).Save(r.Ctx)
	if err != nil {
		return false, err
	}
	return true, nil
}

// Adds credits of creditType to user if they do not already have any belonging to stripe invoice line item
func (r *Repository) AddAdhocCreditsIfEligible(creditType *ent.CreditType, userID uuid.UUID, lineItemID string) (added bool, err error) {
	if creditType == nil {
		return false, errors.New("creditType cannot be nil")
	}
	// See if user has any credits of this type
	credits, err := r.DB.Credit.Query().Where(credit.UserID(userID), credit.CreditTypeID(creditType.ID), credit.StripeLineItemIDEQ(lineItemID)).First(r.Ctx)
	if err != nil && !ent.IsNotFound(err) {
		return false, err
	}

	if credits != nil {
		// User already has credits of this type
		return false, nil
	}

	// Add credits
	_, err = r.DB.Credit.Create().SetCreditTypeID(creditType.ID).SetUserID(userID).SetRemainingAmount(creditType.Amount).SetStripeLineItemID(lineItemID).SetExpiresAt(NEVER_EXPIRE).Save(r.Ctx)
	if err != nil {
		return false, err
	}
	return true, nil
}

// Deduct credits from user, starting with credits that expire soonest. Return true if deduction was successful
func (r *Repository) DeductCreditsFromUser(userID uuid.UUID, amount int32, DB *ent.Client) (success bool, err error) {
	if DB == nil {
		DB = r.DB
	}
	rowsAffected, err := DB.Credit.Update().
		Where(func(s *sql.Selector) {
			t := sql.Table(credit.Table)
			s.Where(
				sql.EQ(t.C(credit.FieldID),
					sql.Select(credit.FieldID).From(t).Where(
						sql.And(
							// Not expired
							sql.GT(t.C(credit.FieldExpiresAt), time.Now()),
							// Our user
							sql.EQ(t.C(credit.FieldUserID), userID),
							// Has remaining amount
							sql.GTE(t.C(credit.FieldRemainingAmount), amount),
						),
					).OrderBy(sql.Asc(t.C(credit.FieldExpiresAt))).Limit(1),
				),
			)
		}).AddRemainingAmount(-1 * amount).Save(r.Ctx)
	if err != nil {
		return false, err
	}
	return rowsAffected > 0, nil
}

// Refund credits for user, starting with credits that expire soonest. Return true if refund was successful
func (r *Repository) RefundCreditsToUser(userID uuid.UUID, amount int32, db *ent.Client) (success bool, err error) {
	if db == nil {
		db = r.DB
	}
	rowsAffected, err := db.Credit.Update().
		Where(func(s *sql.Selector) {
			t := sql.Table(credit.Table)
			s.Where(
				sql.EQ(t.C(credit.FieldID),
					sql.Select(credit.FieldID).From(t).Where(
						sql.And(
							// Not expired
							sql.GT(t.C(credit.FieldExpiresAt), time.Now()),
							// Our user
							sql.EQ(t.C(credit.FieldUserID), userID),
							// Has remaining amount
							sql.GTE(t.C(credit.FieldRemainingAmount), amount),
						),
					).OrderBy(sql.Asc(t.C(credit.FieldExpiresAt))).Limit(1),
				),
			)
		}).AddRemainingAmount(amount).Save(r.Ctx)
	if err != nil {
		return false, err
	}
	return rowsAffected > 0, nil
}

// Replenish free credits where eligible
func (r *Repository) ReplenishFreeCreditsToEligibleUsers(userIDs []uuid.UUID) (int, error) {
	// Get free credit type
	creditType, err := r.GetOrCreateFreeCreditType()
	if err != nil {
		log.Error("Error getting free credit type", "err", err)
		return 0, err
	}

	// Add where
	// - user is in userIDs
	// - credit type is free credit type
	// - remaining amount is less than amount (cap)
	// - item was last updated more than FREE_CREDIT_REPLENISHMENT_INTERVAL ago
	updatedAtSince := time.Now().Add(-shared.FREE_CREDIT_REPLENISHMENT_INTERVAL)
	var updated int
	if err := r.WithTx(func(tx *ent.Tx) error {
		updated, err = tx.Credit.Update().
			Where(
				credit.UserIDIn(userIDs...),
				credit.CreditTypeID(creditType.ID),
				credit.RemainingAmountLT(creditType.Amount),
				credit.ReplenishedAtLT(updatedAtSince),
			).
			SetReplenishedAt(time.Now()).
			AddRemainingAmount(shared.FREE_CREDIT_AMOUNT_DAILY).Save(r.Ctx)
		if err != nil {
			return err
		}
		// Ensure nothing is higher than the cap
		_, err = tx.Credit.Update().
			Where(
				credit.UserIDIn(userIDs...),
				credit.CreditTypeID(creditType.ID),
				credit.RemainingAmountGT(creditType.Amount),
			).
			SetRemainingAmount(creditType.Amount).Save(r.Ctx)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return 0, err
	}
	return updated, nil
}
