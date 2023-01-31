package jobs

import (
	"context"

	"github.com/meilisearch/meilisearch-go"
	"github.com/stablecog/go-apps/cron/utils"
	"github.com/stablecog/go-apps/database"
	"github.com/stablecog/go-apps/database/ent"
)

type JobRunner struct {
	Redis            *database.RedisWrapper
	Ctx              context.Context
	Db               *ent.Client
	Discord          *utils.DiscordHealthTracker
	Meili            *meilisearch.Client
	LastHealthStatus string
}
