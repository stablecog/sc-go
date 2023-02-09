package repository

import (
	"context"

	"github.com/stablecog/go-apps/database"
	"github.com/stablecog/go-apps/database/ent"
)

// Repository is a package that contains all the database access functions

type Repository struct {
	DB    *ent.Client
	Redis *database.RedisWrapper
	Ctx   context.Context
}
