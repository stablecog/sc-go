package repository

import (
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/ent/credit"
	"github.com/stablecog/sc-go/database/ent/credittype"
)

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

func (r *Repository) GetNonExpiredCreditTotalForUser(userID uuid.UUID, DB *ent.Client) (int, error) {
	if DB == nil {
		DB = r.DB
	}
	var total []struct {
		Sum int
	}
	err := DB.Credit.Query().Where(credit.UserID(userID), credit.ExpiresAtGT(time.Now())).Aggregate(ent.Sum(credit.FieldRemainingAmount)).Scan(r.Ctx, &total)
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
