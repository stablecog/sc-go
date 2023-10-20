package repository

import (
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/ent/credit"
	"github.com/stablecog/sc-go/database/ent/credittype"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/shared"
)

func (r *Repository) GetTippableSumForUser(userID uuid.UUID) (int, error) {
	return r.DB.Credit.Query().
		Where(
			credit.UserID(userID), credit.ExpiresAtGT(time.Now()), credit.CreditTypeIDEQ(uuid.MustParse(TIPPABLE_CREDIT_TYPE_ID)),
		).
		Aggregate(
			ent.Sum(credit.FieldRemainingAmount),
		).
		Int(r.Ctx)
}

// Get credits for user that are not expired
func (r *Repository) GetCreditsForUser(userID uuid.UUID) ([]*UserCreditsQueryResult, error) {
	var res []*UserCreditsQueryResult
	err := r.DB.Credit.Query().Select(credit.FieldID, credit.FieldRemainingAmount, credit.FieldExpiresAt).Where(credit.UserID(userID), credit.ExpiresAtGT(time.Now()), credit.CreditTypeIDNEQ(uuid.MustParse(TIPPABLE_CREDIT_TYPE_ID))).
		Modify(func(s *sql.Selector) {
			ct := sql.Table(credittype.Table)
			s.LeftJoin(ct).On(
				s.C(credit.FieldCreditTypeID), ct.C(credittype.FieldID),
			).AppendSelect(sql.As(ct.C(credittype.FieldID), "credit_type_id"), sql.As(ct.C(credittype.FieldName), "credit_type_name"), sql.As(ct.C(credittype.FieldDescription), "credit_type_description"), sql.As(ct.C(credittype.FieldAmount), "credit_type_amount"))
		}).Scan(r.Ctx, &res)

	return res, err
}

// For mocking
var Now = time.Now

func (r *Repository) GetFreeCreditReplenishesAtForUser(userID uuid.UUID) (*time.Time, *ent.Credit, *ent.CreditType, error) {
	// Free type
	ctype, err := r.GetOrCreateFreeCreditType(nil)
	if err != nil {
		log.Error("Error getting free credit type", "err", err)
		return nil, nil, nil, err
	}
	// get the free credit row
	credit, err := r.DB.Credit.Query().Where(credit.UserID(userID), credit.CreditTypeID(ctype.ID)).Only(r.Ctx)
	if err != nil {
		log.Error("Error getting free credit", "err", err)
		return nil, nil, nil, err
	}
	if credit.RemainingAmount >= ctype.Amount {
		// Already has full amount
		return nil, nil, nil, nil
	}
	// Figure out when it will be replenished
	// It is based on replenished_at and shared.FREE_CREDIT_REPLENISHMENT_INTERVAL
	// It was last replenished at replnished_at and will be replenished every FREE_CREDIT_REPLENISHMENT_INTERVAL
	now := Now()
	// Get time delta between now and replenished_at
	delta := now.Add(credit.ReplenishedAt.Sub(now))
	d := now.Sub(delta)
	diff := shared.FREE_CREDIT_REPLENISHMENT_INTERVAL - d

	replenishesAt := now.Add(diff)

	return &replenishesAt, credit, ctype, nil
}

// Determine if a user has non-free credits or not
func (r *Repository) GetNonFreeCreditSum(userID uuid.UUID) (int, error) {
	var v []struct {
		Sum *int
	}
	err := r.DB.Credit.Query().
		Where(credit.UserIDEQ(userID), credit.CreditTypeIDNEQ(uuid.MustParse(FREE_CREDIT_TYPE_ID)), credit.CreditTypeIDNEQ(uuid.MustParse(TIPPABLE_CREDIT_TYPE_ID))).
		Aggregate(
			ent.Sum(credit.FieldRemainingAmount),
		).
		Scan(r.Ctx, &v)
	if err != nil {
		return 0, err
	}
	if len(v) == 0 || v[0].Sum == nil {
		return 0, nil
	}
	return *v[0].Sum, nil
}

// Determine if user has paid credits, with a non null stripe line item ID
func (r *Repository) GetPaidCreditSum(userID uuid.UUID) (int, error) {
	var v []struct {
		Sum *int
	}
	err := r.DB.Credit.Query().
		Where(credit.UserIDEQ(userID), credit.StripeLineItemIDNotNil(), credit.CreditTypeIDNEQ(uuid.MustParse(FREE_CREDIT_TYPE_ID)), credit.CreditTypeIDNEQ(uuid.MustParse(TIPPABLE_CREDIT_TYPE_ID))).
		Aggregate(
			ent.Sum(credit.FieldRemainingAmount),
		).
		Scan(r.Ctx, &v)
	if err != nil {
		return 0, err
	}
	if len(v) == 0 || v[0].Sum == nil {
		return 0, nil
	}
	return *v[0].Sum, nil
}

func (r *Repository) GetNonExpiredCreditTotalForUser(userID uuid.UUID, DB *ent.Client) (int, error) {
	if DB == nil {
		DB = r.DB
	}
	var total []struct {
		Sum int
	}
	err := DB.Credit.Query().Where(credit.UserID(userID), credit.ExpiresAtGT(time.Now()), credit.CreditTypeIDNEQ(uuid.MustParse(TIPPABLE_CREDIT_TYPE_ID))).Aggregate(ent.Sum(credit.FieldRemainingAmount)).Scan(r.Ctx, &total)
	if err != nil {
		return 0, err
	} else if len(total) == 0 {
		return 0, nil
	}
	return total[0].Sum, err
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
