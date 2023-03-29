package jobs

import (
	"context"
	"fmt"

	"github.com/meilisearch/meilisearch-go"
	"github.com/stablecog/sc-go/cron/discord"
	"github.com/stablecog/sc-go/database"
	"github.com/stablecog/sc-go/database/repository"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/server/analytics"
)

type JobRunner struct {
	Repo    *repository.Repository
	Redis   *database.RedisWrapper
	Ctx     context.Context
	Meili   *meilisearch.Client
	Discord *discord.DiscordHealthTracker
	Track   *analytics.AnalyticsService
}

// Just wrap logger so we can include the job name without repeating it
type Logger interface {
	Infof(s string, args ...any)
	Errorf(s string, args ...any)
}

type JobLogger struct {
	JobName string
}

func (j *JobLogger) Infof(s string, args ...any) {
	log.Info(fmt.Sprintf("%s -- %v", j.JobName, fmt.Sprintf(s, args...)))
}

func (j *JobLogger) Errorf(s string, args ...any) {
	log.Error(fmt.Sprintf("%s -- %v", j.JobName, fmt.Sprintf(s, args...)))
}

func NewJobLogger(jobName string) *JobLogger {
	return &JobLogger{JobName: jobName}
}
