package middleware

import (
	"github.com/stablecog/sc-go/database"
	"github.com/stablecog/sc-go/database/repository"
	"github.com/stablecog/sc-go/utils"
)

type Middleware struct {
	SupabaseAuth *database.SupabaseAuth
	Repo         *repository.Repository
	Redis        *database.RedisWrapper
	GeoIP        *utils.GeoIP
}
