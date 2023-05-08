package requests

import "github.com/google/uuid"

type DeactiveApiTokenRequest struct {
	ID uuid.UUID `json:"id"`
}
