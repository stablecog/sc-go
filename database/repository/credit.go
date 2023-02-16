package repository

import (
	"errors"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/ent/credit"
	"github.com/stablecog/sc-go/database/ent/credittype"
)

// Expiration date for manual invoices (non-recurring)
var EXPIRES_MANUAL_INVOICE = time.Date(2100, 1, 1, 5, 0, 0, 0, time.UTC)

// Add credits of creditType to user if they do not have any un-expired credits of this type
func (r *Repository) AddCreditsIfEligible(creditType *ent.CreditType, userID uuid.UUID, expiresAt time.Time) (added bool, err error) {
	if creditType == nil {
		return false, errors.New("creditType cannot be nil")
	}
	// See if user has any credits of this type
	credits, err := r.DB.Credit.Query().Where(credit.UserID(userID), credit.CreditTypeID(creditType.ID), credit.ExpiresAtEQ(expiresAt)).First(r.Ctx)
	if err != nil && !ent.IsNotFound(err) {
		return false, err
	}

	if credits != nil {
		// User already has credits of this type
		return false, nil
	}

	// Add credits
	_, err = r.DB.Credit.Create().SetCreditTypeID(creditType.ID).SetUserID(userID).SetRemainingAmount(creditType.Amount).SetExpiresAt(expiresAt).Save(r.Ctx)
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

// Get credits for user that are not expired
func (r *Repository) GetCreditsForUser(userID uuid.UUID) ([]*UserCreditsQueryResult, error) {
	var res []*UserCreditsQueryResult
	err := r.DB.Credit.Query().Select(credit.FieldID, credit.FieldRemainingAmount, credit.FieldExpiresAt).Where(credit.UserID(userID), credit.ExpiresAtGT(time.Now())).
		Modify(func(s *sql.Selector) {
			ct := sql.Table(credittype.Table)
			s.LeftJoin(ct).On(
				s.C(credit.FieldCreditTypeID), ct.C(credittype.FieldID),
			).AppendSelect(sql.As(ct.C(credittype.FieldID), "credit_type_id"), sql.As(ct.C(credittype.FieldName), "credit_type_name"), sql.As(ct.C(credittype.FieldDescription), "credit_type_description"), sql.As(ct.C(credittype.FieldAmount), "credit_type_amount"))
		}).Scan(r.Ctx, &res)

	return res, err
}

type UserCreditsQueryResult struct {
	ID                    uuid.UUID `json:"id" sql:"id"`
	RemainingAmount       int32     `json:"remaining_amount" sql:"remaining_amount"`
	ExpiresAt             time.Time `json:"expires_at" sql:"expires_at"`
	CreditTypeID          uuid.UUID `json:"credit_type_id" sql:"credit_type_id"`
	CreditTypeName        string    `json:"credit_type_name" sql:"credit_type_name"`
	CreditTypeDescription string    `json:"credit_type_description" sql:"credit_type_description"`
	CreditTypeAmount      int32     `json:"credit_type_amount" sql:"credit_type_amount"`
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
