package middleware

import (
	"github.com/stablecog/go-apps/database"
	"github.com/stablecog/go-apps/database/repository"
)

type Middleware struct {
	SupabaseAuth *database.SupabaseAuth
	Repo         *repository.Repository
}
