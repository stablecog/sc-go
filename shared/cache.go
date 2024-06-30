package shared

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
	"golang.org/x/exp/slices"
)

// A singleton that caches the features available to free users
// Avoids having to query the database every time a user requests the features
type Cache struct {
	// Models and options available to free users
	generateModels         []*ent.GenerationModel
	upscaleModels          []*ent.UpscaleModel
	schedulers             []*ent.Scheduler
	voiceoverModels        []*ent.VoiceoverModel
	voiceoverSpeakers      []*ent.VoiceoverSpeaker
	adminIDs               []uuid.UUID
	iPBlacklist            []string
	thumbmarkIDBlacklist   []string
	disposableEmailDomains []string
	usernameBlacklist      []string
	bannedWords            []*ent.BannedWords
	nllbUrls               []string
	clipUrls               []string
	httpClient             *http.Client
	sync.RWMutex
}

var lock = &sync.Mutex{}
var singleCache *Cache

func newCache() *Cache {
	return &Cache{
		httpClient: &http.Client{
			Timeout: time.Second * 30, // Set a timeout for all requests
		},
	}
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
	f.Lock()
	defer f.Unlock()
	f.generateModels = models
}

func (f *Cache) GenerationModels() []*ent.GenerationModel {
	f.RLock()
	defer f.RUnlock()
	return f.generateModels
}

func (f *Cache) UpdateUpscaleModels(models []*ent.UpscaleModel) {
	f.Lock()
	defer f.Unlock()
	f.upscaleModels = models
}

func (f *Cache) UpscaleModels() []*ent.UpscaleModel {
	f.RLock()
	defer f.RUnlock()
	return f.upscaleModels
}

func (f *Cache) UpdateSchedulers(schedulers []*ent.Scheduler) {
	f.Lock()
	defer f.Unlock()
	f.schedulers = schedulers
}

func (f *Cache) Schedulers() []*ent.Scheduler {
	f.RLock()
	defer f.RUnlock()
	return f.schedulers
}

func (f *Cache) UpdateVoiceoverModels(models []*ent.VoiceoverModel) {
	f.Lock()
	defer f.Unlock()
	f.voiceoverModels = models
}

func (f *Cache) VoiceoverModels() []*ent.VoiceoverModel {
	f.RLock()
	defer f.RUnlock()
	return f.voiceoverModels
}

func (f *Cache) UpdateVoiceoverSpeakers(speakers []*ent.VoiceoverSpeaker) {
	f.Lock()
	defer f.Unlock()
	f.voiceoverSpeakers = speakers
}

func (f *Cache) VoiceoverSpeakers() []*ent.VoiceoverSpeaker {
	f.RLock()
	defer f.RUnlock()
	return f.voiceoverSpeakers
}

func (f *Cache) UpdateBannedWords(bannedWords []*ent.BannedWords) {
	f.Lock()
	defer f.Unlock()
	f.bannedWords = bannedWords
}

func (f *Cache) BannedWords() []*ent.BannedWords {
	f.RLock()
	defer f.RUnlock()
	return f.bannedWords
}

func (f *Cache) IsValidGenerationModelID(id uuid.UUID) bool {
	for _, model := range f.GenerationModels() {
		if model.ID == id && model.IsActive {
			return true
		}
	}
	return false
}

func (f *Cache) IsValidUpscaleModelID(id uuid.UUID) bool {
	for _, model := range f.UpscaleModels() {
		if model.ID == id {
			return true
		}
	}
	return false
}

func (f *Cache) IsValidShedulerID(id uuid.UUID) bool {
	for _, scheduler := range f.Schedulers() {
		if scheduler.ID == id {
			return true
		}
	}
	return false
}

func (f *Cache) IsValidVoiceoverModelID(id uuid.UUID) bool {
	for _, model := range f.VoiceoverModels() {
		if model.ID == id {
			return true
		}
	}
	return false
}

func (f *Cache) IsValidVoiceoverSpeakerID(speakerId uuid.UUID, modelId uuid.UUID) bool {
	for _, speaker := range f.VoiceoverSpeakers() {
		if speaker.ID == speakerId && speaker.ModelID == modelId {
			return true
		}
	}
	return false
}

func (f *Cache) GetVoiceoverSpeakersForModel(modelId uuid.UUID) []*ent.VoiceoverSpeaker {
	speakers := []*ent.VoiceoverSpeaker{}
	for _, speaker := range f.VoiceoverSpeakers() {
		if speaker.ModelID == modelId {
			speakers = append(speakers, speaker)
		}
	}
	return speakers
}

func (f *Cache) GetGenerationModelNameFromID(id uuid.UUID) string {
	for _, model := range f.GenerationModels() {
		if model.ID == id {
			return model.NameInWorker
		}
	}
	return ""
}

func (f *Cache) GetUpscaleModelNameFromID(id uuid.UUID) string {
	for _, model := range f.UpscaleModels() {
		if model.ID == id {
			return model.NameInWorker
		}
	}
	return ""
}

func (f *Cache) GetSchedulerNameFromID(id uuid.UUID) string {
	for _, scheduler := range f.Schedulers() {
		if scheduler.ID == id {
			return scheduler.NameInWorker
		}
	}
	return ""
}

func (f *Cache) GetVoiceoverModelNameFromID(id uuid.UUID) string {
	for _, model := range f.VoiceoverModels() {
		if model.ID == id {
			return model.NameInWorker
		}
	}
	return ""
}

func (f *Cache) GetVoiceoverSpeakerNameFromID(speakerId uuid.UUID) string {
	for _, speaker := range f.VoiceoverSpeakers() {
		if speaker.ID == speakerId {
			return speaker.NameInWorker
		}
	}
	return ""
}

func (f *Cache) IsAdmin(id uuid.UUID) bool {
	for _, adminID := range f.AdminIDs() {
		if adminID == id {
			return true
		}
	}
	return false
}

func (f *Cache) SetAdminUUIDs(ids []uuid.UUID) {
	f.Lock()
	defer f.Unlock()
	f.adminIDs = ids
}

func (f *Cache) AdminIDs() []uuid.UUID {
	f.RLock()
	defer f.RUnlock()
	return f.adminIDs
}

func (f *Cache) UpdateDisposableEmailDomains(domains []string) {
	f.Lock()
	defer f.Unlock()
	f.disposableEmailDomains = domains
}

func (f *Cache) DisposableEmailDomains() []string {
	f.RLock()
	defer f.RUnlock()
	return f.disposableEmailDomains
}

func (f *Cache) UpdateIPBlacklist(ips []string) {
	f.Lock()
	defer f.Unlock()
	f.iPBlacklist = ips
}

func (f *Cache) IPBlacklist() []string {
	f.RLock()
	defer f.RUnlock()
	return f.iPBlacklist
}

func (f *Cache) UpdateThumbmarkIDBlacklist(tm []string) {
	f.Lock()
	defer f.Unlock()
	f.thumbmarkIDBlacklist = tm
}

func (f *Cache) ThumbmarkIDBlacklist() []string {
	f.RLock()
	defer f.RUnlock()
	return f.thumbmarkIDBlacklist
}

func (f *Cache) UpdateUsernameBlacklist(blacklist []string) {
	f.Lock()
	defer f.Unlock()
	f.usernameBlacklist = blacklist
}

func (f *Cache) IsUsernameBlacklisted(username string) bool {
	f.RLock()
	defer f.RUnlock()
	for _, blacklistedUsername := range f.usernameBlacklist {
		if username == blacklistedUsername {
			return true
		}
	}
	return false
}

func (f *Cache) IsDisposableEmail(email string) bool {
	disposableEmailDomains := f.DisposableEmailDomains()
	if !strings.Contains(email, "@") {
		for _, disposableDomain := range disposableEmailDomains {
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
	for _, disposableDomain := range disposableEmailDomains {
		if domain == disposableDomain {
			return true
		}
	}
	return false
}

func (f *Cache) IsIPBanned(ip string) bool {
	return slices.Contains(f.IPBlacklist(), ip)
}

func (f *Cache) IsThumbmarkIDBanned(ip string) bool {
	return slices.Contains(f.ThumbmarkIDBlacklist(), ip)
}

func (f *Cache) GetDefaultGenerationModel() *ent.GenerationModel {
	var defaultModel *ent.GenerationModel
	for _, model := range f.GenerationModels() {
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
	for _, model := range f.UpscaleModels() {
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
	for _, model := range f.VoiceoverModels() {
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
	for _, speaker := range f.VoiceoverSpeakers() {
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
	for _, scheduler := range f.Schedulers() {
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
	for _, model := range f.GenerationModels() {
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

func (f *Cache) UpdateWorkerURL(vastAiKey string) error {
	url := fmt.Sprintf("https://console.vast.ai/api/v0/instances?api_key=%s", vastAiKey)

	// HTTP get
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Errorf("Error making creating request %s", err)
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := f.httpClient.Do(req)
	if err != nil {
		log.Error("Error making request:", "err", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Error("Error making request:", "status", resp.Status)
		return fmt.Errorf("Error making request: %s", resp.Status)
	}

	// Try to decode+deserialize
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Error decoding response body %s", err)
		return err
	}

	// Deserialize
	var vastInstances VastAIInstances
	if err := json.Unmarshal(body, &vastInstances); err != nil {
		log.Errorf("Error deserializing response body %s", err)
		return err
	}

	// Update the vast ai worker urls
	nllbUrls := []string{}
	clipUrls := []string{}
	for _, instance := range vastInstances.Instances {
		if len(instance.Ports.One3349TCP) > 0 {
			nllbUrls = append(nllbUrls, fmt.Sprintf("http://%s:%s/predictions", instance.PublicIP, instance.Ports.One3349TCP[0].HostPort))
		}
		if len(instance.Ports.One3339TCP) > 0 {
			clipUrls = append(clipUrls, fmt.Sprintf("http://%s:%s/clip/embed", instance.PublicIP, instance.Ports.One3339TCP[0].HostPort))
		}
	}

	f.Lock()
	defer f.Unlock()
	f.nllbUrls = nllbUrls
	f.clipUrls = clipUrls

	// Log URLs for debugging

	return nil
}

func (f *Cache) GetNLLBUrls() []string {
	f.RLock()
	defer f.RUnlock()
	return f.nllbUrls
}

func (f *Cache) GetClipUrls() []string {
	f.RLock()
	defer f.RUnlock()
	return f.clipUrls
}

// Structs for vast ai
type VastAIInstances struct {
	Instances []VastAIInstance `json:"instances"`
}

type VastAIInstance struct {
	PublicIP string `json:"public_ipaddr"`
	Ports    struct {
		One3339TCP []struct {
			HostIP   string `json:"HostIp"`
			HostPort string `json:"HostPort"`
		} `json:"13339/tcp"`
		One3349TCP []struct {
			HostIP   string `json:"HostIp"`
			HostPort string `json:"HostPort"`
		} `json:"13349/tcp"`
	} `json:"ports"`
}
