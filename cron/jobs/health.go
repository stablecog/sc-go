package jobs

import (
	"fmt"
	"time"

	"github.com/stablecog/sc-go/cron/discord"
	"github.com/stablecog/sc-go/database/ent/generation"
	"github.com/stablecog/sc-go/shared"
	"k8s.io/klog/v2"
)

// General redis key prefix
const redisHealthKeyPrefix = "health_check"

const maxGenerationFailWithoutNSFWRate = 0.5
const generationCountToCheck = 10
const maxGenerationDuration = 2 * time.Minute

const rTTL = 2 * time.Hour

var lastGenerationKey = fmt.Sprintf("%s:last_generation", redisHealthKeyPrefix)

// CheckHealth cron job
func (j *JobRunner) CheckHealth() error {
	start := time.Now()
	klog.Infof("Checking health...")

	generations, err := j.Repo.GetGenerations(generationCountToCheck)
	if err != nil {
		klog.Errorf("Couldn't get generations %v", err)
		return err
	}

	nsfwGenerations := 0
	failedGenerations := 0
	var lastGenerationTime time.Time

	for i, g := range generations {
		if i == 0 {
			lastGenerationTime = g.CreatedAt
			err := j.Redis.Client.Set(j.Ctx, lastGenerationKey, lastGenerationTime.Format(time.RFC3339), rTTL).Err()
			if err != nil {
				klog.Error("Redis - Error setting last generation key: %v", err)
				return err
			}
		}
		if g.Status == generation.StatusFailed && g.FailureReason != nil && *g.FailureReason == shared.NSFW_ERROR {
			nsfwGenerations++
		} else if g.Status == generation.StatusFailed {
			failedGenerations++
		}
	}

	klog.Infof("Generation fail rate (NSFW): %d/%d", nsfwGenerations, len(generations))
	klog.Infof("Generation fail rate: %d/%d", failedGenerations, len(generations))

	// Figure out if we're healthy
	healthStatus := discord.HEALTHY
	failRate := float64(failedGenerations) / float64(len(generations))
	if failRate > maxGenerationFailWithoutNSFWRate {
		healthStatus = discord.UNHEALTHY
	}

	now := time.Now()
	klog.Infof("Done checking health in: %dms", now.Sub(start).Milliseconds())

	return j.Discord.SendDiscordNotificationIfNeeded(
		healthStatus,
		generations,
		lastGenerationTime,
		now,
	)
}
