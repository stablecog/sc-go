package repository

import (
	"github.com/stablecog/sc-go/database/ent/bannedwords"
	"github.com/stablecog/sc-go/database/ent/disposableemail"
	"github.com/stablecog/sc-go/database/ent/ipblacklist"
	"github.com/stablecog/sc-go/database/ent/thumbmarkidblacklist"
	"github.com/stablecog/sc-go/database/ent/usernameblacklist"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/shared"
)

// Update the cache from the database
func (r *Repository) UpdateCache() error {
	generationModels, err := r.GetAllGenerationModels()
	if err != nil {
		log.Error("Failed to get generation_models", "err", err)
		return err
	}
	shared.GetCache().UpdateGenerationModels(generationModels)

	upscaleModels, err := r.GetAllUpscaleModels()
	if err != nil {
		log.Error("Failed to get upscale_models", "err", err)
		return err
	}
	shared.GetCache().UpdateUpscaleModels(upscaleModels)

	schedulers, err := r.GetAllSchedulers()
	if err != nil {
		log.Error("Failed to get schedulers", "err", err)
		return err
	}
	shared.GetCache().UpdateSchedulers(schedulers)

	voiceoverModels, err := r.GetAllVoiceoverModels()
	if err != nil {
		log.Error("Failed to get voiceover_models", "err", err)
		return err
	}
	shared.GetCache().UpdateVoiceoverModels(voiceoverModels)

	voiceoverSpeakers, err := r.GetAllVoiceoverSpeakers()
	if err != nil {
		log.Error("Failed to get voiceover_speakers", "err", err)
		return err
	}
	shared.GetCache().UpdateVoiceoverSpeakers(voiceoverSpeakers)

	admins, err := r.GetSuperAdminUserIDs()
	if err != nil {
		log.Error("Failed to get super admins", "err", err)
		return err
	}
	shared.GetCache().SetAdminUUIDs(admins)

	disposableEmailDomains, err := r.DB.DisposableEmail.Query().Select(disposableemail.FieldDomain).All(r.Ctx)
	if err != nil {
		log.Error("Failed to get disposable email domains", "err", err)
		return err
	}
	disposableEmailDomainsStr := make([]string, len(disposableEmailDomains))
	for i, domain := range disposableEmailDomains {
		disposableEmailDomainsStr[i] = domain.Domain
	}
	shared.GetCache().UpdateDisposableEmailDomains(disposableEmailDomainsStr)

	ipBlacklist, err := r.DB.IPBlackList.Query().Select(ipblacklist.FieldIP).All(r.Ctx)
	if err != nil {
		log.Error("Failed to get ip blacklist", "err", err)
		return err
	}
	ipBlacklistStr := make([]string, len(ipBlacklist))
	for i, ip := range ipBlacklist {
		ipBlacklistStr[i] = ip.IP
	}
	shared.GetCache().UpdateIPBlacklist(ipBlacklistStr)

	thumbmarkIDBlacklist, err := r.DB.ThumbmarkIdBlackList.Query().Select(thumbmarkidblacklist.FieldThumbmarkID).All(r.Ctx)
	if err != nil {
		log.Error("Failed to get thumbmark id blacklist", "err", err)
		return err
	}
	thumbmarkIDBlacklistStr := make([]string, len(thumbmarkIDBlacklist))
	for i, thumbmarkID := range thumbmarkIDBlacklist {
		thumbmarkIDBlacklistStr[i] = thumbmarkID.ThumbmarkID
	}
	shared.GetCache().UpdateThumbmarkIDBlacklist(thumbmarkIDBlacklistStr)

	bannedWords, err := r.DB.BannedWords.Query().Select(bannedwords.FieldReason, bannedwords.FieldSplitMatch, bannedwords.FieldWords).All(r.Ctx)
	if err != nil {
		log.Error("Failed to get banned words", "err", err)
		return err
	}
	shared.GetCache().UpdateBannedWords(bannedWords)

	// Username
	usernameBlacklist, err := r.DB.UsernameBlacklist.Query().Select(usernameblacklist.FieldUsername).All(r.Ctx)
	if err != nil {
		log.Error("Failed to get username blacklist", "err", err)
		return err
	}
	usernameBlacklistStr := make([]string, len(usernameBlacklist))
	for i, username := range usernameBlacklist {
		usernameBlacklistStr[i] = username.Username
	}
	shared.GetCache().UpdateUsernameBlacklist(usernameBlacklistStr)

	return nil
}
