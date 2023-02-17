package jobs

import (
	"context"

	"github.com/meilisearch/meilisearch-go"
	"github.com/stablecog/sc-go/database"
	"github.com/stablecog/sc-go/database/repository"
)

type JobRunner struct {
	Repo  *repository.Repository
	Redis *database.RedisWrapper
	Ctx   context.Context
	Meili *meilisearch.Client
}
