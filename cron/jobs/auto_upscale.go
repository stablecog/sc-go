package jobs

import (
	"encoding/json"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/server/scworker"
	"github.com/stablecog/sc-go/shared"
)

func (j *JobRunner) StartAutoUpscaleJob(log Logger) {
	log.Infof("Starting auto upscale job...")
	// Create a SyncMap to track requests
	sMap := shared.NewSyncMap[chan requests.CogWebhookMessage]()
	//Redis subscription for cog messages we should handle
	pubSubInternalMessages := j.Redis.Client.Subscribe(j.Redis.Ctx, shared.REDIS_INTERNAL_COG_CHANNEL)
	defer pubSubInternalMessages.Close()
	// Start in goroutine, this is intended to get info back to the called upscale function
	go func() {
		log.Infof("Listening for internal cog messages %s", shared.REDIS_INTERNAL_COG_CHANNEL)
		for msg := range pubSubInternalMessages.Channel() {
			var cogMessage requests.CogWebhookMessage
			err := json.Unmarshal([]byte(msg.Payload), &cogMessage)
			if err != nil {
				log.Errorf("Error unmarshalling cog internal message %v", err)
				continue
			}
			log.Infof("Received internal cog message %v", cogMessage)

			// See if active channel exists
			activeChannel := sMap.Get(cogMessage.Input.ID.String())
			// Write to channel
			if activeChannel != nil {
				activeChannel <- cogMessage
			}
		}
	}()

	signalChannel := make(chan os.Signal, 2)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
	for {
		select {
		case <-signalChannel:
			log.Infof("Shutting down free credit job...")
			return
		default:
			// Get unscaled outputs
			unscaledOutputs, err := j.Repo.GetNonUpscaledGalleryItems(10)
			refreshedAt := time.Now()
			if err != nil {
				log.Errorf("Error getting unscaled outputs %v", err)
				time.Sleep(5 * time.Second)
				continue
			}
			if len(unscaledOutputs) == 0 {
				log.Infof("No non-unscaled outputs found")
				time.Sleep(shared.AUTO_UPSCALE_RETRY_DURATION)
				continue
			}
			for _, output := range unscaledOutputs {
				// Check if refresh is needed
				if time.Now().Sub(refreshedAt) > 5*time.Minute {
					// Refresh
					break
				}

				// Upscale
				log.Infof("Upscaling output %s", output.ID)
				// Get generation
				g, err := output.QueryGenerations().First(j.Repo.Ctx)
				if err != nil {
					log.Errorf("Error getting generation %v", err)
					time.Sleep(5 * time.Second)
					continue
				}
				err = scworker.CreateUpscaleInternal(j.Track, j.Repo, j.Redis, j.MQClient, sMap, g, output)
				if err != nil {
					log.Errorf("Error creating upscale %v", err)
					time.Sleep(5 * time.Second)
					continue
				}
				log.Infof("Upscale created for output %s", output.ID)
				time.Sleep(5 * time.Second)
			}
		}
	}
}
