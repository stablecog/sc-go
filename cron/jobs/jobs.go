package jobs

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/meilisearch/meilisearch-go"
	"github.com/stablecog/go-apps/cron/utils"
	"github.com/stablecog/go-apps/database/ent"
)

type JobRunner struct {
	Redis   *redis.Client
	Ctx     context.Context
	Db      *ent.Client
	Discord *utils.DiscordHealthTracker
	Meili   *meilisearch.Client
}
