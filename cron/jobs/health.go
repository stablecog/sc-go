package jobs

import (
	"fmt"
	"time"

	"github.com/stablecog/go-apps/database/ent"
	dbgeneration "github.com/stablecog/go-apps/database/ent/generation"
)

// General redis key prefix
const redisHealthKeyPrefix = "health_check"

const maxGenerationFailWithoutNSFWRate = 0.5
const generationCountToCheck = 10
const maxGenerationDuration = 2 * time.Minute

const rTTL = 2 * time.Hour

var lastGenerationKey = fmt.Sprintf("%s:last_generation", redisHealthKeyPrefix)

// Query last generations from database
func (j *JobRunner) GetLastGenerations(limit int) ([]*ent.Generation, error) {
	return j.Db.Generation.Query().
		Select(dbgeneration.FieldStatus, dbgeneration.FieldCreatedAt, dbgeneration.FieldFailureReason).
		Order(ent.Desc(dbgeneration.FieldCreatedAt)).
		Limit(limit).
		All(j.Ctx)
}

// CheckHealth cron job
func (j *JobRunner) CheckHealth() error {
	return nil
	// start := time.Now()
	// klog.Infof("Checking health...")

	// generations, err := j.GetLastGenerations(generationCountToCheck)
	// if err != nil {
	// 	klog.Errorf("Couldn't get generations %v", err)
	// 	return err
	// }

	// if j.LastHealthStatus == "" {
	// 	j.LastHealthStatus = "unknown"
	// }

	// var generationsFailed int
	// var generationsFailedWithoutNSFW int
	// var lastGenerationTime time.Time
	// lastGenerationTimeStr := j.Redis.Get(j.Ctx, lastGenerationKey).Val()
	// lastGenerationTime, _ = time.Parse(time.RFC3339, lastGenerationTimeStr)
	// for i, generation := range generations {
	// 	if i == 0 {
	// 		lastGenerationTime = generation.CreatedAt
	// 		err := j.Redis.Set(j.Ctx, lastGenerationKey, lastGenerationTime.Format(time.RFC3339), rTTL).Err()
	// 		if err != nil {
	// 			klog.Error("Redis - Error setting last generation key: %v", err)
	// 			return err
	// 		}
	// 	}
	// 	if generation.Status == nil {
	// 		continue
	// 	} else if *generation.Status == dbgeneration.StatusFailed {
	// 		generationsFailed++
	// 		if generation.FailureReason == nil || *generation.FailureReason != "NSFW" {
	// 			generationsFailedWithoutNSFW++
	// 		}
	// 	} else if *generation.Status == dbgeneration.StatusStarted && time.Now().Sub(generation.CreatedAt) > maxGenerationDuration {
	// 		generationsFailed++
	// 		generationsFailedWithoutNSFW++
	// 	}
	// }
	// /* generationFailRate := float64(generationsFailed) / float64(len(generations)) */
	// generationFailWithoutNSFWRate := float64(generationsFailedWithoutNSFW) / float64(len(generations))
	// klog.Infof("Generation fail rate: %d/%d", generationsFailed, len(generations))
	// klog.Infof("Generation fail rate without NSFW: %d/%d", generationsFailedWithoutNSFW, len(generations))

	// lastStatusPrev := j.LastHealthStatus
	// if generationFailWithoutNSFWRate > maxGenerationFailWithoutNSFWRate {
	// 	j.LastHealthStatus = "unhealthy"
	// } else {
	// 	j.LastHealthStatus = "healthy"
	// }
	// now := time.Now()
	// klog.Infof("Done checking health in: %dms", now.Sub(start).Milliseconds())

	// return j.Discord.SendDiscordNotificationIfNeeded(
	// 	j.LastHealthStatus,
	// 	lastStatusPrev,
	// 	generations,
	// 	lastGenerationTime,
	// 	now,
	// )
}
