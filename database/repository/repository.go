package repository

import (
	"context"

	"github.com/stablecog/go-apps/database/ent"
)

// Repository is a package that contains all the database access functions

type Repository struct {
	DB  *ent.Client
	Ctx context.Context
}
