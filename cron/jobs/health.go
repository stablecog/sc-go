package jobs

import (
	"time"

	"github.com/stablecog/sc-go/cron/discord"
	"github.com/stablecog/sc-go/database/ent/generation"
	"github.com/stablecog/sc-go/shared"
	"k8s.io/klog/v2"
)

// Considered failed if len(failures)/len(generations) > maxGenerationFailWithoutNSFWRate
const maxGenerationFailWithoutNSFWRate = 0.5

// Get this number of generations on each check, sorted by created_at DESC
const generationCountToCheck = 10

// CheckHealth cron job
func (j *JobRunner) CheckHealth() error {
	start := time.Now()
	klog.Infof("Checking health...")

	generations, err := j.Repo.GetGenerations(generationCountToCheck)
	if err != nil || len(generations) == 0 {
		klog.Errorf("Couldn't get generations %v", err)
		return err
	}

	nsfwGenerations := 0
	failedGenerations := 0
	lastGenerationTime := generations[0].CreatedAt

	// Count the number of failed generations distinguishing between NSFW and other failures
	for _, g := range generations {
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
