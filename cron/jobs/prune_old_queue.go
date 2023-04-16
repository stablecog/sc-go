package jobs

import (
	"time"
)

// Prune messages older than this time
const PRUNE_OLDER_THAN = 5 * time.Minute

const PRUNE_OLD_QUEUE_JOB_NAME = "PRUNE_OLD_QUEUE_JOB"

// PruneOldQueueItems cron job
func (j *JobRunner) PruneOldQueueItems(log Logger) error {
	start := time.Now()
	log.Infof("Checking for old queue items...")

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
	log.Infof("Finished checking for old queue items in %dms", time.Now().Sub(start).Milliseconds())
	return nil
}
