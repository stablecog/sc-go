package jobs

import (
	"fmt"
	"math/big"
	"time"

	"github.com/meilisearch/meilisearch-go"
	"github.com/stablecog/go-apps/database/ent"
	"github.com/stablecog/go-apps/database/ent/generationg"
	"github.com/stablecog/go-apps/database/ent/model"
	"github.com/stablecog/go-apps/database/ent/negativeprompt"
	"github.com/stablecog/go-apps/database/ent/prompt"
	"github.com/stablecog/go-apps/database/ent/scheduler"
	"k8s.io/klog/v2"
)

// General redis key prefix
const redisMeiliKeyPrefix = "meili"
const maxTotalHits = 5000

var shouldSetSettings = true

var lastSyncedGenUpdatedAtKey = fmt.Sprintf("%s:last_sync_gen_updated_at", redisMeiliKeyPrefix)
var sortableAttributes = []string{"updated_at", "created_at"}

func (j *JobRunner) SyncMeili() error {
	var lastSyncedGenUpdatedAt time.Time
	lastSyncedGenUpdatedAtStr := j.Redis.Get(j.Ctx, lastSyncedGenUpdatedAtKey).Val()
	lastSyncedGenUpdatedAt, _ = time.Parse(time.RFC3339, lastSyncedGenUpdatedAtStr)

	generations, err := j.Db.GenerationG.Query().Select(generationg.FieldCreatedAt, generationg.FieldUpdatedAt, generationg.FieldGuidanceScale, generationg.FieldHidden,
		generationg.FieldHeight, generationg.FieldWidth, generationg.FieldSchedulerID, generationg.FieldModelID, generationg.FieldImageID,
		generationg.FieldSeed, generationg.FieldNegativePromptID, generationg.FieldNumInferenceSteps, generationg.FieldPromptID, generationg.FieldUserID,
		generationg.FieldUserTier).WithPrompt(
		func(query *ent.PromptQuery) {
			query.Select(prompt.FieldText)
		}).WithNegativePrompt(
		func(query *ent.NegativePromptQuery) {
			query.Select(negativeprompt.FieldText)
		}).WithModel(
		func(query *ent.ModelQuery) {
			query.Select(model.FieldName)
		}).WithScheduler(
		func(query *ent.SchedulerQuery) {
			query.Select(scheduler.FieldName)
		}).
		Where(generationg.Hidden(false), generationg.UpdatedAtGT(lastSyncedGenUpdatedAt)).
		Order(ent.Asc(generationg.FieldUpdatedAt)).
		Limit(1000).All(j.Ctx)

	if err != nil {
		return err
	}

	if len(generations) == 0 {
		klog.Infof("-- MeiliWorker - No new generations to sync")
		return nil
	}
	lastGen := generations[len(generations)-1]

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

	var generationsMeili []MeiliGenerationG

	for _, gen := range generations {
		var uid *string
		if gen.UserID != nil {
			asStr := gen.UserID.String()
			uid = &asStr
		}
		newG := MeiliGenerationG{
			Id:                gen.ID.String(),
			ImageId:           gen.ImageID,
			Width:             gen.Width,
			Height:            gen.Height,
			Hidden:            gen.Hidden,
			Prompt:            MeiliPrompt{Id: gen.Edges.Prompt.ID.String(), Text: gen.Edges.Prompt.Text},
			Model:             MeiliModel{Id: gen.Edges.Model.ID.String(), Name: gen.Edges.Model.Name},
			Scheduler:         MeiliScheduler{Id: gen.Edges.Scheduler.ID.String(), Name: gen.Edges.Scheduler.Name},
			NumInferenceSteps: gen.NumInferenceSteps,
			GuidanceScale:     gen.GuidanceScale,
			Seed:              gen.Seed.Int,
			UserId:            uid,
			UserTier:          gen.UserTier.String(),
			CreatedAt:         gen.CreatedAt.Format(time.RFC3339),
			UpdatedAt:         gen.UpdatedAt.Format(time.RFC3339),
		}
		if gen.Edges.NegativePrompt != nil {
			newG.NegativePrompt = &MeiliNegativePrompt{Id: gen.Edges.NegativePrompt.ID.String(), Text: gen.Edges.NegativePrompt.Text}
		}
		generationsMeili = append(generationsMeili, newG)
	}

	_, errMeili := j.Meili.Index("generation_g").AddDocuments(generationsMeili)
	if errMeili != nil {
		klog.Errorf("-- MeiliWorker - Meili error: %v", errMeili)
		return errMeili
	} else {
		lastSyncedGenUpdatedAt = lastGen.UpdatedAt
		j.Redis.Set(j.Ctx, lastSyncedGenUpdatedAtKey, lastSyncedGenUpdatedAt.UTC().Format(time.RFC3339), rTTL)
		klog.Infof("-- MeiliWorker - Successfully indexed - %s -- ", lastSyncedGenUpdatedAt.UTC())
	}

	return nil
}

type MeiliGenerationG struct {
	Id                string               `json:"id"`
	ImageId           string               `json:"image_id"`
	Width             int                  `json:"width"`
	Height            int                  `json:"height"`
	Hidden            bool                 `json:"hidden"`
	Prompt            MeiliPrompt          `json:"prompt"`
	NegativePrompt    *MeiliNegativePrompt `json:"negative_prompt,omitempty"`
	Model             MeiliModel           `json:"model"`
	Scheduler         MeiliScheduler       `json:"scheduler"`
	NumInferenceSteps *int                 `json:"num_inference_steps"`
	GuidanceScale     float64              `json:"guidance_scale"`
	Seed              *big.Int             `json:"seed"`
	UserId            *string              `json:"user_id,omitempty"`
	UserTier          string               `json:"user_tier"`
	CreatedAt         string               `json:"created_at"`
	UpdatedAt         string               `json:"updated_at"`
}

type MeiliPrompt struct {
	Id   string `json:"id"`
	Text string `json:"text"`
}

type MeiliNegativePrompt struct {
	Id   string `json:"id,omitempty"`
	Text string `json:"text,omitempty"`
}

type MeiliModel struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type MeiliScheduler struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}
