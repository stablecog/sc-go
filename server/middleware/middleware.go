package middleware

import (
	"github.com/stablecog/sc-go/database"
	"github.com/stablecog/sc-go/database/repository"
	"github.com/stablecog/sc-go/server/analytics"
)

type Middleware struct {
	SupabaseAuth *database.SupabaseAuth
	Repo         *repository.Repository
	Redis        *database.RedisWrapper
	Track        *analytics.AnalyticsService
}
