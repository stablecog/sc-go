package middleware

import (
	"github.com/stablecog/go-apps/database"
)

type Middleware struct {
	SupabaseAuth *database.SupabaseAuth
}
