package jobs

import (
	"time"

	"github.com/stablecog/sc-go/database/ent/mqlog"
)

// Prune messages older than this time
const PRUNE_OLDER_THAN = 5 * time.Minute

const PRUNE_OLD_QUEUE_JOB_NAME = "PRUNE_OLD_QUEUE_JOB"

// PruneOldQueueItems cron job
func (j *JobRunner) PruneOldQueueItems(log Logger) error {
	start := time.Now()
	log.Infof("Checking for old queue items...")

	// For new mq_log, delete stuff older than 30 minutes
	deletedPg, err := j.Repo.DB.MqLog.Delete().Where(mqlog.CreatedAtLT(time.Now().Add(-30 * time.Minute))).Exec(j.Repo.Ctx)
	if err != nil {
		log.Errorf("Couldn't delete old queue items from mq_log %v", err)
		return err
	}

	generations, upscales, err := j.Redis.GetPendingGenerationAndUpscaleIDs(PRUNE_OLDER_THAN)
	if err != nil {
		log.Errorf("Couldn't get xrange from redis %v", err)
		return err
	}

	if len(generations) == 0 && len(upscales) == 0 {
		log.Infof("No old queue items to delete")
		return nil
	}

	idsToRemove := make([]string, len(generations)+len(upscales))

	lastI := 0
	for i, g := range generations {
		idsToRemove[i] = g.RedisMsgid
		lastI = i + 1
	}
	for i, u := range upscales {
		idsToRemove[lastI+i] = u.RedisMsgid
	}

	deleted, err := j.Redis.XDelListOfIDs(idsToRemove)
	if err != nil {
		log.Errorf("Couldn't delete old queue items %v", err)
		return err
	}

	log.Infof("Deleted %d old queue items", deleted)
	log.Infof("Deleted %d old queue items from mq_log", deletedPg)
	log.Infof("Finished checking for old queue items in %dms", time.Now().Sub(start).Milliseconds())
	return nil
}
