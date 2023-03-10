package jobs

import (
	"fmt"
	"time"

	"github.com/meilisearch/meilisearch-go"
)

// General redis key prefix
const redisMeiliKeyPrefix = "meili"
const rTTL = 2 * time.Hour

const maxTotalHits = 5000

var shouldSetSettings = true

var lastSyncedGenUpdatedAtKey = fmt.Sprintf("%s:last_sync_gen_updated_at_v2", redisMeiliKeyPrefix)
var sortableAttributes = []string{"updated_at", "created_at"}
var filterableAttributes = []string{"id"}

func (j *JobRunner) SyncMeili(log Logger) error {
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
		log.Infof("No new generations to sync")
		return nil
	}
	lastGen := galleryItems[len(galleryItems)-1]

	if shouldSetSettings {
		_, err = j.Meili.Index("generation_g").UpdateSortableAttributes(&sortableAttributes)
		if err != nil {
			log.Errorf("Update sortable attributes %v", err)
			return err
		} else {
			log.Infof("Sortable attributes updated")
		}
		_, err = j.Meili.Index("generation_g").UpdateFilterableAttributes(&filterableAttributes)
		if err != nil {
			log.Errorf("Update filterable attributes %v", err)
			return err
		} else {
			log.Infof("Filterable attributes updated")
		}
		_, errMax := j.Meili.Index("generation_g").UpdatePagination(&meilisearch.Pagination{MaxTotalHits: int64(maxTotalHits)})
		if errMax != nil {
			log.Errorf("Update max total hits %v", errMax)
			return errMax
		} else {
			log.Infof("Meili max total hits updated")
		}
		if err == nil && errMax == nil {
			shouldSetSettings = false
		}
	}

	_, errMeili := j.Meili.Index("generation_g").AddDocuments(galleryItems)
	if errMeili != nil {
		log.Errorf("%v", errMeili)
		return errMeili
	} else {
		lastSyncedGenUpdatedAt = lastGen.UpdatedAt
		j.Redis.Client.Set(j.Ctx, lastSyncedGenUpdatedAtKey, lastSyncedGenUpdatedAt.UTC().Format(time.RFC3339), rTTL)
		log.Infof("Successfully indexed at %s", lastSyncedGenUpdatedAt.UTC())
	}

	return nil
}
