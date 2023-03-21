package requests

import "github.com/google/uuid"

type CreditAddRequest struct {
	CreditTypeID uuid.UUID `json:"credit_type_id"`
	UserID       uuid.UUID `json:"user_id"`
}
