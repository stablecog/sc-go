package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/ent/prompt"
	"github.com/stablecog/sc-go/database/ent/voiceover"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/utils"
)

// CreateVoiceover creates the initial voiceover in the database
// Takes in a userID (creator),  device info, countryCode, and a request body
func (r *Repository) CreateVoiceover(userID uuid.UUID, deviceType, deviceOs, deviceBrowser, countryCode string, req requests.CreateVoiceoverRequest, productId *string, apiTokenId *uuid.UUID, DB *ent.Client) (*ent.Voiceover, error) {
	if DB == nil {
		DB = r.DB
	}
	// Get prompt, negative prompt, device info
	deviceInfoId, err := r.GetOrCreateDeviceInfo(deviceType, deviceOs, deviceBrowser, DB)
	if err != nil {
		return nil, err
	}
	insert := DB.Voiceover.Create().
		SetStatus(voiceover.StatusQueued).
		SetModelID(*req.ModelId).
		SetSpeakerID(*req.SpeakerId).
		SetDeviceInfoID(deviceInfoId).
		SetCountryCode(countryCode).
		SetTemperature(*req.Temperature).
		SetUserID(userID).
		SetSeed(*req.Seed).
		SetWasAutoSubmitted(req.SubmitToGallery)
	if productId != nil {
		insert.SetStripeProductID(*productId)
	}
	if apiTokenId != nil {
		insert.SetAPITokenID(*apiTokenId)
	}
	return insert.Save(r.Ctx)
}

func (r *Repository) SetVoiceoverStarted(voiceoverID string) error {
	uid, err := uuid.Parse(voiceoverID)
	if err != nil {
		log.Error("Error parsing voiceover id in SetVoiceoverStarted", "id", voiceoverID, "err", err)
		return err
	}
	_, err = r.DB.Voiceover.UpdateOneID(uid).SetStatus(voiceover.StatusStarted).SetStartedAt(time.Now()).Save(r.Ctx)
	if err != nil {
		// Log error here since this might be happening in a goroutine
		log.Error("Error setting voiceover started", "id", voiceoverID, "err", err)
	}
	return err
}

func (r *Repository) SetVoiceoverFailed(voiceoverId string, reason string, db *ent.Client) error {
	if db == nil {
		db = r.DB
	}

	uid, err := uuid.Parse(voiceoverId)
	if err != nil {
		log.Error("Error parsing voiceover id in SetVoiceoverFailed", "id", voiceoverId, "err", err)
		return err
	}
	_, err = db.Voiceover.UpdateOneID(uid).SetStatus(voiceover.StatusFailed).SetFailureReason(reason).SetCompletedAt(time.Now()).Save(r.Ctx)
	if err != nil {
		log.Error("Error setting voiceover failed", "id", voiceoverId, "err", err)
	}
	return err
}

// ! Currently supports 1 output
func (r *Repository) SetVoiceoverSucceeded(voiceoverId, promptStr string, output requests.CogWebhookOutput) (*ent.VoiceoverOutput, error) {
	uid, err := uuid.Parse(voiceoverId)
	if err != nil {
		log.Error("Error parsing voiceover id in SetVoiceoverSucceeded", "id", voiceoverId, "err", err)
		return nil, err
	}

	var voiceoverOutput *ent.VoiceoverOutput

	// Start a transaction
	if err := r.WithTx(func(tx *ent.Tx) error {
		db := tx.Client()

		// Retrieve the upscale
		u, err := r.GetVoiceover(uid)
		if err != nil {
			log.Error("Error retrieving voiceover", "id", uid, "err", err)
			return err
		}

		// Get prompt IDs
		promptId, _, err := r.GetOrCreatePrompts(promptStr, "", prompt.TypeImage, db)
		if err != nil || promptId == nil {
			log.Error("Error getting or creating prompts", "id", voiceoverId, "err", err, "prompt", promptStr)
			return err
		}

		// Update the voiceover
		_, err = tx.Voiceover.UpdateOneID(u.ID).SetStatus(voiceover.StatusSucceeded).SetCompletedAt(time.Now()).SetPromptID(*promptId).Save(r.Ctx)
		if err != nil {
			log.Error("Error setting voiceover succeeded", "id", voiceoverId, "err", err)
			return err
		}

		// Set audio output
		parsedS3, err := utils.GetPathFromS3URL(output.AudioFiles[0].AudioFile)
		if err != nil {
			log.Error("Error parsing s3 url", "output", output, "err", err)
			parsedS3 = output.AudioFiles[0].AudioFile
		}
		vOutput := tx.VoiceoverOutput.Create().SetAudioPath(parsedS3).SetVoiceoverID(uid).SetAudioDuration(output.AudioFiles[0].AudioDuration)
		voiceoverOutput, err = vOutput.Save(r.Ctx)
		if err != nil {
			log.Error("Error inserting voiceover output", "id", voiceoverId, "err", err)
			return err
		}

		return nil
	}); err != nil {
		log.Error("Error in SetVoiceoverSucceeded", "id", voiceoverId, "err", err)
		return nil, err
	}

	return voiceoverOutput, nil
}
