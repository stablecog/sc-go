package repository

import (
	"errors"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/ent/credit"
)

// Expiration date for manual invoices (non-recurring)
var EXPIRES_MANUAL_INVOICE = time.Date(2100, 1, 1, 5, 0, 0, 0, time.UTC)

// Add credits of creditType to user if they do not have any un-expired credits of this type
func (r *Repository) AddCreditsIfEligible(creditType *ent.CreditType, userID uuid.UUID, expiresAt time.Time, DB *ent.Client) (added bool, err error) {
	if DB == nil {
		DB = r.DB
	}

	if creditType == nil {
		return false, errors.New("creditType cannot be nil")
	}

	// See if user has any credits of this type
	credits, err := DB.Credit.Query().Where(credit.UserID(userID), credit.CreditTypeID(creditType.ID), credit.ExpiresAtEQ(expiresAt)).First(r.Ctx)
	if err != nil && !ent.IsNotFound(err) {
		return false, err
	}

	if credits != nil {
		// User already has credits of this type
		return false, nil
	}

	// Add credits
	_, err = DB.Credit.Create().SetCreditTypeID(creditType.ID).SetUserID(userID).SetRemainingAmount(creditType.Amount).SetExpiresAt(expiresAt).Save(r.Ctx)
	if err != nil {
		return false, err
	}
	return true, nil
}

// Replenish free credits if eligible
func (r *Repository) ReplenishFreeCreditsIfEligible(userID uuid.UUID, expiresAt time.Time, DB *ent.Client) (added bool, err error) {
	if DB == nil {
		DB = r.DB
	}

	creditType, err := r.GetOrCreateFreeCreditType()
	if err != nil {
		return false, err
	}

	// See if user has any credits of this type
	// ExpiresAt must be greater than or equal to the current time
	credits, err := DB.Credit.Query().Where(credit.UserID(userID), credit.CreditTypeID(creditType.ID), credit.ExpiresAtGTE(time.Now())).Order(ent.Desc(credit.FieldExpiresAt)).First(r.Ctx)
	if err != nil && !ent.IsNotFound(err) {
		return false, err
	}

	if credits != nil {
		// User already has credits of this type
		return false, nil
	}

	// Add credits
	_, err = DB.Credit.Create().SetCreditTypeID(creditType.ID).SetUserID(userID).SetRemainingAmount(creditType.Amount).SetExpiresAt(expiresAt).Save(r.Ctx)
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
	_, err = r.DB.Credit.Create().SetCreditTypeID(creditType.ID).SetUserID(userID).SetRemainingAmount(creditType.Amount).SetStripeLineItemID(lineItemID).SetExpiresAt(EXPIRES_MANUAL_INVOICE).Save(r.Ctx)
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
