package jobs

import (
	"fmt"
	"time"

	"github.com/meilisearch/meilisearch-go"
	"k8s.io/klog/v2"
)

// General redis key prefix
const redisMeiliKeyPrefix = "meili"
const maxTotalHits = 5000

var shouldSetSettings = true

var lastSyncedGenUpdatedAtKey = fmt.Sprintf("%s:last_sync_gen_updated_at_v2", redisMeiliKeyPrefix)
var sortableAttributes = []string{"updated_at", "created_at"}

func (j *JobRunner) SyncMeili() error {
	var lastSyncedGenUpdatedAt time.Time
	lastSyncedGenUpdatedAtStr := j.Redis.Client.Get(j.Ctx, lastSyncedGenUpdatedAtKey).Val()
	lastSyncedGenUpdatedAt, err := time.Parse(time.RFC3339, lastSyncedGenUpdatedAtStr)
	var lastSyncGenUpdatedAtRef *time.Time
	if err == nil {
		lastSyncGenUpdatedAtRef = &lastSyncedGenUpdatedAt
	}

	galleryItems, err := j.Repo.RetrieveGalleryData(1000, lastSyncGenUpdatedAtRef)

	if err != nil {
		return err
	}

	if len(galleryItems) == 0 {
		klog.Infof("-- MeiliWorker - No new generations to sync")
		return nil
	}
	lastGen := galleryItems[len(galleryItems)-1]

	if shouldSetSettings {
		_, err = j.Meili.Index("generation_g").UpdateSortableAttributes(&sortableAttributes)
		if err != nil {
			klog.Errorf("-- MeiliWorker - Meili update sortable attributes error: %v", err)
			return err
		} else {
			klog.Infof("-- MeiliWorker - Meili sortable attributes updated")
		}
		_, errMax := j.Meili.Index("generation_g").UpdatePagination(&meilisearch.Pagination{MaxTotalHits: int64(maxTotalHits)})
		if errMax != nil {
			klog.Errorf("-- MeiliWorker - Meili update max total hits error: %v", errMax)
			return errMax
		} else {
			klog.Infof("-- MeiliWorker - Meili max total hits updated")
		}
		if err == nil && errMax == nil {
			shouldSetSettings = false
		}
	}

	_, errMeili := j.Meili.Index("generation_g").AddDocuments(galleryItems)
	if errMeili != nil {
		klog.Errorf("-- MeiliWorker - Meili error: %v", errMeili)
		return errMeili
	} else {
		lastSyncedGenUpdatedAt = lastGen.UpdatedAt
		j.Redis.Client.Set(j.Ctx, lastSyncedGenUpdatedAtKey, lastSyncedGenUpdatedAt.UTC().Format(time.RFC3339), rTTL)
		klog.Infof("-- MeiliWorker - Successfully indexed - %s -- ", lastSyncedGenUpdatedAt.UTC())
	}

	return nil
}
