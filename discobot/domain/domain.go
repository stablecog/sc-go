package domain

import (
	"github.com/stablecog/sc-go/database"
	"github.com/stablecog/sc-go/database/repository"
)

// Domain logic layer
type DiscoDomain struct {
	Repo         *repository.Repository
	Redis        *database.RedisWrapper
	SupabaseAuth *database.SupabaseAuth
}
