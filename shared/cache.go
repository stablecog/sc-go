package shared

import (
	"context"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
	"golang.org/x/exp/slices"
)

// A singleton that caches the features available to free users
// Avoids having to query the database every time a user requests the features
type Cache struct {
	// Models and options available to free users
	GenerateModels         []*ent.GenerationModel
	UpscaleModels          []*ent.UpscaleModel
	Schedulers             []*ent.Scheduler
	VoiceoverModels        []*ent.VoiceoverModel
	VoiceoverSpeakers      []*ent.VoiceoverSpeaker
	AdminIDs               []uuid.UUID
	IPBlacklist            []string
	DisposableEmailDomains []string
	BannedWords            []*ent.BannedWords
}

var lock = &sync.Mutex{}

var singleCache *Cache

func newCache() *Cache {
	return &Cache{}
}

func GetCache() *Cache {
	if singleCache == nil {
		lock.Lock()
		defer lock.Unlock()
		if singleCache == nil {
			singleCache = newCache()
		}
	}
	return singleCache
}

func (f *Cache) UpdateGenerationModels(models []*ent.GenerationModel) {
	lock.Lock()
	defer lock.Unlock()
	f.GenerateModels = models
}

func (f *Cache) UpdateUpscaleModels(models []*ent.UpscaleModel) {
	lock.Lock()
	defer lock.Unlock()
	f.UpscaleModels = models
}

func (f *Cache) UpdateSchedulers(schedulers []*ent.Scheduler) {
	lock.Lock()
	defer lock.Unlock()
	f.Schedulers = schedulers
}

func (f *Cache) UpdateVoiceoverModels(models []*ent.VoiceoverModel) {
	lock.Lock()
	defer lock.Unlock()
	f.VoiceoverModels = models
}

func (f *Cache) UpdateVoiceoverSpeakers(speakers []*ent.VoiceoverSpeaker) {
	lock.Lock()
	defer lock.Unlock()
	f.VoiceoverSpeakers = speakers
}

func (f *Cache) UpdateBannedWords(bannedWords []*ent.BannedWords) {
	lock.Lock()
	defer lock.Unlock()
	f.BannedWords = bannedWords
}

func (f *Cache) IsValidGenerationModelID(id uuid.UUID) bool {
	for _, model := range f.GenerateModels {
		if model.ID == id {
			return true
		}
	}
	return false
}

func (f *Cache) IsValidUpscaleModelID(id uuid.UUID) bool {
	for _, model := range f.UpscaleModels {
		if model.ID == id {
			return true
		}
	}
	return false
}

func (f *Cache) IsValidShedulerID(id uuid.UUID) bool {
	for _, scheduler := range f.Schedulers {
		if scheduler.ID == id {
			return true
		}
	}
	return false
}

func (f *Cache) IsValidVoiceoverModelID(id uuid.UUID) bool {
	for _, model := range f.VoiceoverModels {
		if model.ID == id {
			return true
		}
	}
	return false
}

func (f *Cache) IsValidVoiceoverSpeakerID(speakerId uuid.UUID, modelId uuid.UUID) bool {
	for _, speaker := range f.VoiceoverSpeakers {
		if speaker.ID == speakerId && speaker.ModelID == modelId {
			return true
		}
	}
	return false
}

func (f *Cache) GetVoiceoverSpeakersForModel(modelId uuid.UUID) []*ent.VoiceoverSpeaker {
	speakers := []*ent.VoiceoverSpeaker{}
	for _, speaker := range f.VoiceoverSpeakers {
		if speaker.ModelID == modelId {
			speakers = append(speakers, speaker)
		}
	}
	return speakers
}

func (f *Cache) GetGenerationModelNameFromID(id uuid.UUID) string {
	for _, model := range f.GenerateModels {
		if model.ID == id {
			return model.NameInWorker
		}
	}
	return ""
}

func (f *Cache) GetUpscaleModelNameFromID(id uuid.UUID) string {
	for _, model := range f.UpscaleModels {
		if model.ID == id {
			return model.NameInWorker
		}
	}
	return ""
}

func (f *Cache) GetSchedulerNameFromID(id uuid.UUID) string {
	for _, scheduler := range f.Schedulers {
		if scheduler.ID == id {
			return scheduler.NameInWorker
		}
	}
	return ""
}

func (f *Cache) GetVoiceoverModelNameFromID(id uuid.UUID) string {
	for _, model := range f.VoiceoverModels {
		if model.ID == id {
			return model.NameInWorker
		}
	}
	return ""
}

func (f *Cache) GetVoiceoverSpeakerNameFromID(speakerId uuid.UUID) string {
	for _, speaker := range f.VoiceoverSpeakers {
		if speaker.ID == speakerId {
			return speaker.NameInWorker
		}
	}
	return ""
}

func (f *Cache) IsAdmin(id uuid.UUID) bool {
	for _, adminID := range f.AdminIDs {
		if adminID == id {
			return true
		}
	}
	return false
}

func (f *Cache) SetAdminUUIDs(ids []uuid.UUID) {
	lock.Lock()
	defer lock.Unlock()
	f.AdminIDs = ids
}

func (f *Cache) UpdateDisposableEmailDomains(domains []string) {
	lock.Lock()
	defer lock.Unlock()
	f.DisposableEmailDomains = domains
}

func (f *Cache) UpdateIPBlacklist(ips []string) {
	lock.Lock()
	defer lock.Unlock()
	f.IPBlacklist = ips
}

func (f *Cache) IsDisposableEmail(email string) bool {
	if !strings.Contains(email, "@") {
		for _, disposableDomain := range f.DisposableEmailDomains {
			if email == disposableDomain {
				return true
			}
		}
		return false
	}

	segs := strings.Split(email, "@")
	if len(segs) != 2 {
		return false
	}
	domain := strings.ToLower(segs[1])
	for _, disposableDomain := range f.DisposableEmailDomains {
		if domain == disposableDomain {
			return true
		}
	}
	return false
}

func (f *Cache) IsIPBanned(ip string) bool {
	return slices.Contains(f.IPBlacklist, ip)
}

func (f *Cache) GetDefaultGenerationModel() *ent.GenerationModel {
	var defaultModel *ent.GenerationModel
	for _, model := range f.GenerateModels {
		// always set at least 1 as default
		if defaultModel == nil {
			defaultModel = model
		}
		if model.IsDefault {
			defaultModel = model
		}
	}
	return defaultModel
}

func (f *Cache) GetDefaultUpscaleModel() *ent.UpscaleModel {
	var defaultModel *ent.UpscaleModel
	for _, model := range f.UpscaleModels {
		// always set at least 1 as default
		if defaultModel == nil {
			defaultModel = model
		}
		if model.IsDefault {
			defaultModel = model
		}
	}
	return defaultModel
}

func (f *Cache) GetDefaultVoiceoverModel() *ent.VoiceoverModel {
	var defaultModel *ent.VoiceoverModel
	for _, model := range f.VoiceoverModels {
		// always set at least 1 as default
		if defaultModel == nil {
			defaultModel = model
		}
		if model.IsDefault {
			defaultModel = model
		}
	}
	return defaultModel
}

func (f *Cache) GetDefaultVoiceoverSpeaker() *ent.VoiceoverSpeaker {
	var defaultSpeaker *ent.VoiceoverSpeaker
	for _, speaker := range f.VoiceoverSpeakers {
		// always set at least 1 as default
		if defaultSpeaker == nil {
			defaultSpeaker = speaker
		}
		if speaker.IsDefault {
			defaultSpeaker = speaker
		}
	}
	return defaultSpeaker
}

func (f *Cache) GetDefaultScheduler() *ent.Scheduler {
	var defaultScheduler *ent.Scheduler
	for _, scheduler := range f.Schedulers {
		// always set at least 1 as default
		if defaultScheduler == nil {
			defaultScheduler = scheduler
		}
		if scheduler.IsDefault {
			defaultScheduler = scheduler
		}
	}
	return defaultScheduler
}

func (f *Cache) GetGenerationModelByID(id uuid.UUID) *ent.GenerationModel {
	for _, model := range f.GenerateModels {
		if model.ID == id {
			return model
		}
	}
	return nil
}

func (f *Cache) GetCompatibleSchedulerIDsForModel(ctx context.Context, modelId uuid.UUID) []uuid.UUID {
	m := f.GetGenerationModelByID(modelId)
	if m == nil {
		return []uuid.UUID{}
	}
	schedulerIds := make([]uuid.UUID, len(m.Edges.Schedulers))
	for i, scheduler := range m.Edges.Schedulers {
		schedulerIds[i] = scheduler.ID
	}
	return schedulerIds
}

func (f *Cache) GetDefaultSchedulerIDForModel(modelId uuid.UUID) uuid.UUID {
	m := f.GetGenerationModelByID(modelId)
	if m == nil {
		return uuid.Nil
	}
	for _, scheduler := range m.Edges.Schedulers {
		if scheduler.IsDefault {
			return scheduler.ID
		}
	}
	if len(m.Edges.Schedulers) > 0 {
		return m.Edges.Schedulers[0].ID
	}
	return uuid.Nil
}
