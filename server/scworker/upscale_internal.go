package scworker

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/favadi/osinfo"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/stablecog/sc-go/database"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/enttypes"
	"github.com/stablecog/sc-go/database/repository"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/server/analytics"
	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/shared"
	"github.com/stablecog/sc-go/shared/queue"
	"github.com/stablecog/sc-go/utils"
)

// Create an Upscale in sc-worker, wait for result
// ! TODO - clean this up and merge with CreateUpscale method
func CreateUpscaleInternal(AsynqClient *asynq.Client, S3 *s3.S3, Track *analytics.AnalyticsService, Repo *repository.Repository, Redis *database.RedisWrapper, MQClient queue.MQClient, sMap *shared.SyncMap[chan requests.CogWebhookMessage], generation *ent.Generation, output *ent.GenerationOutput) error {
	if len(shared.GetCache().UpscaleModels()) == 0 {
		log.Error("No upscale models available")
		return fmt.Errorf("No upscale models available")
	}
	var upscaleModel *ent.UpscaleModel
	for _, model := range shared.GetCache().UpscaleModels() {
		if model.IsActive && model.IsDefault {
			upscaleModel = model
			break
		}
	}
	if upscaleModel == nil {
		log.Error("No active upscale models available")
		return fmt.Errorf("no active upscale model configured")
	}
	// Create req
	upscaleReq := requests.CreateUpscaleRequest{
		Type:    utils.ToPtr(requests.UpscaleRequestTypeOutput),
		Input:   output.ID.String(),
		ModelId: utils.ToPtr(upscaleModel.ID),
	}
	useRunpod := upscaleModel.RunpodEndpoint != nil && upscaleModel.RunpodActive

	var upscale *ent.Upscale
	var requestId uuid.UUID
	var user *ent.User
	var err error

	// Create channel
	activeChl := make(chan requests.CogWebhookMessage)
	// Cleanup
	defer close(activeChl)

	if err := Repo.WithTx(func(tx *ent.Tx) error {
		// Bind transaction to client
		DB := tx.Client()

		user, err = generation.QueryUser().Only(Repo.Ctx)
		if err != nil {
			log.Error("Error getting user", "err", err)
			return err
		}

		// Create upscale
		upscale, err = Repo.CreateUpscale(
			user.ID,
			generation.Width,
			generation.Height,
			"server",
			osinfo.New().String(),
			"",
			"US",
			upscaleReq,
			user.ActiveProductID,
			true,
			nil,
			enttypes.SourceTypeInternal,
			DB)
		if err != nil {
			log.Error("Error creating upscale", "err", err)
			return err
		}

		// Send to the cog
		requestId = upscale.ID
		queueId := utils.Sha256(requestId.String())

		// Live page
		livePageMsg := &shared.LivePageMessage{
			ProcessType:      shared.UPSCALE,
			ID:               utils.Sha256(requestId.String()),
			CountryCode:      "US",
			Status:           shared.LivePageQueued,
			TargetNumOutputs: 1,
			Width:            utils.ToPtr(generation.Width),
			Height:           utils.ToPtr(generation.Height),
			CreatedAt:        upscale.CreatedAt,
			ProductID:        user.ActiveProductID,
			SystemGenerated:  true,
		}

		cogReqBody := requests.CogQueueRequest{
			WebhookEventsFilter: []requests.CogEventFilter{requests.CogEventFilterStart, requests.CogEventFilterStart},
			WebhookUrl:          fmt.Sprintf("%s/v1/worker/webhook", utils.GetEnv().PublicApiUrl),
			Input: requests.BaseCogRequest{
				WebhookPrivateUrl:    fmt.Sprintf("%s/v1/worker/webhook", utils.GetEnv().PrivateApiUrl),
				Internal:             true,
				ID:                   requestId,
				UIId:                 upscaleReq.UIId,
				UserID:               &user.ID,
				StreamID:             upscaleReq.StreamID,
				GenerationOutputID:   output.ID.String(),
				LivePageData:         livePageMsg,
				Image:                utils.GetEnv().GetURLFromImagePath(output.ImagePath),
				ProcessType:          shared.UPSCALE,
				Width:                utils.ToPtr(generation.Width),
				Height:               utils.ToPtr(generation.Height),
				Model:                upscaleModel.NameInWorker,
				UpscaleModel:         upscaleModel.NameInWorker,
				ModelId:              *upscaleReq.ModelId,
				WebhookToken:         upscale.WebhookToken,
				OutputImageExtension: string(shared.DEFAULT_UPSCALE_OUTPUT_EXTENSION),
				OutputImageQuality:   utils.ToPtr(shared.DEFAULT_UPSCALE_OUTPUT_QUALITY),
				Type:                 *upscaleReq.Type,
				RunpodEndpoint:       upscaleModel.RunpodEndpoint,
			},
		}
		if useRunpod {
			cogReqBody.Input.Images = []string{utils.GetEnv().GetURLFromImagePath(output.ImagePath)}
		}

		cogReqBody.Input.SignedUrls = make([]string, 1)
		imgId := fmt.Sprintf("%s.%s", uuid.NewString(), cogReqBody.Input.OutputImageExtension)
		// Sign the URL and append to array
		// If the file does not exist, generate a pre-signed URL
		req, _ := S3.PutObjectRequest(&s3.PutObjectInput{
			Bucket: aws.String(utils.GetEnv().S3BucketName),
			Key:    aws.String(imgId),
		})
		urlStr, err := req.Presign(24 * time.Hour) // URL is valid for 15 minutes
		if err != nil {
			log.Errorf("Failed to sign request: %v\n", err)
			return err
		}

		cogReqBody.Input.SignedUrls[0] = urlStr

		_, err = Repo.AddToQueueLog(queueId, 1, DB)
		if err != nil {
			log.Error("Error adding to queue log", "err", err)
			return err
		}

		if !useRunpod {
			err = MQClient.Publish(queueId, cogReqBody, shared.QUEUE_PRIORITY_1)
			if err != nil {
				log.Error("Failed to write request to queue", "id", queueId, "err", err)
				return err
			}
		} else {
			// use QueCon internal queue
			queueName := shared.QueueByPriority(shared.QUEUE_PRIORITY_1)
			// Enqueue task with priority
			opts := []asynq.Option{
				asynq.MaxRetry(3),
				asynq.TaskID(requestId.String()), // Unique Task ID
				asynq.Queue(queueName),           // Queue name
			}

			// Create payload
			rpInput := requests.RunpodInput{
				Input: cogReqBody.Input,
			}
			payload, err := json.Marshal(rpInput)
			if err != nil {
				log.Error("Error marshalling runpod payload", "err", err)
				return err
			}

			_, err = AsynqClient.Enqueue(asynq.NewTask(
				shared.ASYNQ_TASK_GENERATE,
				payload,
			), opts...)
			if err != nil {
				log.Error("Failed to enqueue task", "err", err)
				return err
			}
		}

		// Analytics
		go Track.UpscaleStarted(user, cogReqBody.Input, enttypes.SourceTypeInternal, "system")

		// Send live page update
		go func() {
			liveResp := repository.TaskStatusUpdateResponse{
				ForLivePage:     true,
				LivePageMessage: livePageMsg,
			}
			respBytes, err := json.Marshal(liveResp)
			if err != nil {
				log.Error("Error marshalling sse live response", "err", err)
				return
			}
			err = Redis.Client.Publish(Redis.Ctx, shared.REDIS_SSE_BROADCAST_CHANNEL, respBytes).Err()
			if err != nil {
				log.Error("Failed to publish live page update", "err", err)
			}
		}()

		return nil
	}); err != nil {
		return err
	}

	// Add channel to sync array (basically a thread-safe map)
	sMap.Put(requestId.String(), activeChl)
	defer sMap.Delete(requestId.String())

	for {
		select {
		case cogMsg := <-activeChl:
			switch cogMsg.Status {
			case requests.CogProcessing:
				err := Repo.SetUpscaleStarted(upscale.ID.String())
				if err != nil {
					log.Error("Failed to set upscale started", "id", upscale.ID, "err", err)
					return err
				}
				// Send live page update
				go func() {
					cogMsg.Input.LivePageData.Status = shared.LivePageProcessing
					now := time.Now()
					cogMsg.Input.LivePageData.StartedAt = &now
					liveResp := repository.TaskStatusUpdateResponse{
						ForLivePage:     true,
						LivePageMessage: cogMsg.Input.LivePageData,
					}
					respBytes, err := json.Marshal(liveResp)
					if err != nil {
						log.Error("Error marshalling sse live response", "err", err)
						return
					}
					err = Redis.Client.Publish(Redis.Ctx, shared.REDIS_SSE_BROADCAST_CHANNEL, respBytes).Err()
					if err != nil {
						log.Error("Failed to publish live page update", "err", err)
					}
				}()
			case requests.CogSucceeded:
				_, err := Repo.SetUpscaleSucceeded(upscale.ID.String(), output.ID.String(), cogMsg.Input.Image, cogMsg.Output)
				if err != nil {
					log.Error("Failed to set upscale succeeded", "id", upscale.ID, "err", err)
				}
				// Send live page update
				go func() {
					cogMsg.Input.LivePageData.Status = shared.LivePageSucceeded
					now := time.Now()
					cogMsg.Input.LivePageData.CompletedAt = &now
					liveResp := repository.TaskStatusUpdateResponse{
						ForLivePage:     true,
						LivePageMessage: cogMsg.Input.LivePageData,
					}
					respBytes, err := json.Marshal(liveResp)
					if err != nil {
						log.Error("Error marshalling sse live response", "err", err)
						return
					}
					err = Redis.Client.Publish(Redis.Ctx, shared.REDIS_SSE_BROADCAST_CHANNEL, respBytes).Err()
					if err != nil {
						log.Error("Failed to publish live page update", "err", err)
					}
				}()
				// Analytics
				upscale, err := Repo.GetUpscale(upscale.ID)
				if err != nil {
					log.Error("Error getting upscale for analytics", "err", err)
					return err
				}
				// Get durations in seconds
				if upscale.StartedAt == nil {
					log.Error("Upscale started at is nil", "id", cogMsg.Input.ID)
					return errors.New("Upscale started at is nil")
				}
				duration := time.Now().Sub(*upscale.StartedAt).Seconds()
				qDuration := (*upscale.StartedAt).Sub(upscale.CreatedAt).Seconds()
				go Track.UpscaleSucceeded(user, cogMsg.Input, duration, qDuration, enttypes.SourceTypeInternal, "system")
				return err
			case requests.CogFailed:
				err := Repo.SetUpscaleFailed(upscale.ID.String(), cogMsg.Error, nil)
				if err != nil {
					log.Error("Failed to set upscale failed", "id", upscale.ID, "err", err)
				}
				// Send live page update
				go func() {
					cogMsg.Input.LivePageData.Status = shared.LivePageFailed
					now := time.Now()
					cogMsg.Input.LivePageData.CompletedAt = &now
					liveResp := repository.TaskStatusUpdateResponse{
						ForLivePage:     true,
						LivePageMessage: cogMsg.Input.LivePageData,
					}
					respBytes, err := json.Marshal(liveResp)
					if err != nil {
						log.Error("Error marshalling sse live response", "err", err)
						return
					}
					err = Redis.Client.Publish(Redis.Ctx, shared.REDIS_SSE_BROADCAST_CHANNEL, respBytes).Err()
					if err != nil {
						log.Error("Failed to publish live page update", "err", err)
					}
				}()
				// Analytics
				duration := time.Now().Sub(cogMsg.Input.LivePageData.CreatedAt).Seconds()
				go Track.UpscaleFailed(user, cogMsg.Input, duration, cogMsg.Error, enttypes.SourceTypeInternal, "system")
				return err
			}
		// Make ~30 minute timeouts, the TTL of MQ messages
		case <-time.After(30 * time.Minute):
			err := Repo.SetUpscaleFailed(upscale.ID.String(), shared.TIMEOUT_ERROR, nil)
			if err != nil {
				log.Error("Failed to set upscale failed", "id", upscale.ID, "err", err)
			}
			return fmt.Errorf(shared.TIMEOUT_ERROR)
		}
	}
}
