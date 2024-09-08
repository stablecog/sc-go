package scworker

import (
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hibiken/asynq"
	"github.com/stablecog/sc-go/database"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/repository"
	"github.com/stablecog/sc-go/server/analytics"
	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/server/translator"
	"github.com/stablecog/sc-go/shared"
	"github.com/stablecog/sc-go/shared/queue"
)

type SCWorker struct {
	Repo           *repository.Repository
	Redis          *database.RedisWrapper
	SMap           *shared.SyncMap[chan requests.CogWebhookMessage]
	QueueThrottler *shared.UserQueueThrottlerMap
	Track          *analytics.AnalyticsService
	SafetyChecker  *translator.TranslatorSafetyChecker
	S3Img          *s3.S3
	S3             *s3.S3
	MQClient       queue.MQClient
	AsynqClient    *asynq.Client
}

const USE_RUNPOD_FALLBACK = true

func ShouldUseRunpodGenerate(model *ent.GenerationModel, redis *database.RedisWrapper) bool {
	if USE_RUNPOD_FALLBACK == false {
		return model.RunpodEndpoint != nil && model.RunpodActive
	}

	health, err := redis.GetWorkerHealth()
	if err != nil {
		return model.RunpodEndpoint != nil && model.RunpodActive
	}

	return model.RunpodEndpoint != nil && health != shared.HEALTHY
}

func ShouldUseRunpodUpscale(model *ent.UpscaleModel, redis *database.RedisWrapper) bool {
	if USE_RUNPOD_FALLBACK == false {
		return model.RunpodEndpoint != nil && model.RunpodActive
	}

	health, err := redis.GetWorkerHealth()
	if err != nil {
		return model.RunpodEndpoint != nil && model.RunpodActive
	}

	return model.RunpodEndpoint != nil && health != shared.HEALTHY
}
