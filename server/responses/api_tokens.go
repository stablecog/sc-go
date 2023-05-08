package responses

import (
	"time"

	"github.com/google/uuid"
)

type NewApiTokensResponse struct {
	ID    uuid.UUID `json:"id"`
	Token string    `json:"token"`
}

// For retrieving a list of API tokens
type ApiToken struct {
	ID         uuid.UUID  `json:"id"`
	Uses       int        `json:"uses"`
	IsActive   bool       `json:"is_active"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

type GetApiTokensResponse struct {
	Tokens []ApiToken `json:"tokens"`
}
