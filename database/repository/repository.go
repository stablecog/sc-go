package repository

import (
	"context"

	"github.com/stablecog/sc-go/database"
	"github.com/stablecog/sc-go/database/ent"
)

// Repository is a package that contains all the database access functions

type Repository struct {
	DB    *ent.Client
	Redis *database.RedisWrapper
	Ctx   context.Context
}
