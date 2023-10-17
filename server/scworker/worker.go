package scworker

import (
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/stablecog/sc-go/database"
	"github.com/stablecog/sc-go/database/repository"
	"github.com/stablecog/sc-go/server/analytics"
	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/shared"
	"github.com/stablecog/sc-go/shared/queue"
	"github.com/stablecog/sc-go/utils"
)

type SCWorker struct {
	Repo           *repository.Repository
	Redis          *database.RedisWrapper
	SMap           *shared.SyncMap[chan requests.CogWebhookMessage]
	QueueThrottler *shared.UserQueueThrottlerMap
	Track          *analytics.AnalyticsService
	SafetyChecker  *utils.TranslatorSafetyChecker
	S3             *s3.S3
	MQClient       queue.MQClient
}
