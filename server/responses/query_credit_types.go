package responses

import "github.com/google/uuid"

type QueryCreditTypesResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Amount      int32     `json:"amount"`
	Description string    `json:"description,omitempty"`
}
