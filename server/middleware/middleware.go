package middleware

import (
	"github.com/stablecog/sc-go/database"
	"github.com/stablecog/sc-go/database/repository"
)

type Middleware struct {
	SupabaseAuth *database.SupabaseAuth
	Repo         *repository.Repository
	Redis        *database.RedisWrapper
}
