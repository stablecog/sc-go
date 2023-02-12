package jobs

import (
	"context"

	"github.com/meilisearch/meilisearch-go"
	"github.com/stablecog/sc-go/cron/utils"
	"github.com/stablecog/sc-go/database"
	"github.com/stablecog/sc-go/database/ent"
)

type JobRunner struct {
	Redis            *database.RedisWrapper
	Ctx              context.Context
	Db               *ent.Client
	Discord          *utils.DiscordHealthTracker
	Meili            *meilisearch.Client
	LastHealthStatus string
}
