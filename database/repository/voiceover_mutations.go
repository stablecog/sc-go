package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/ent/prompt"
	"github.com/stablecog/sc-go/database/ent/voiceover"
	"github.com/stablecog/sc-go/database/ent/voiceoveroutput"
	"github.com/stablecog/sc-go/database/enttypes"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/utils"
)

// CreateVoiceover creates the initial voiceover in the database
// Takes in a userID (creator),  device info, countryCode, and a request body
func (r *Repository) CreateVoiceover(userID uuid.UUID, deviceType, deviceOs, deviceBrowser, countryCode string, req requests.CreateVoiceoverRequest, productId *string, apiTokenId *uuid.UUID, sourceType enttypes.SourceType, DB *ent.Client) (*ent.Voiceover, error) {
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
		SetWasAutoSubmitted(req.WasAutoSubmitted).
		SetCost(utils.CalculateVoiceoverCredits(req.Prompt)).
		SetSourceType(sourceType)
	if req.DenoiseAudio != nil {
		insert.SetDenoiseAudio(*req.DenoiseAudio)
	}
	if req.RemoveSilence != nil {
		insert.SetRemoveSilence(*req.RemoveSilence)
	}
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
func (r *Repository) SetVoiceoverSucceeded(voiceoverId, promptStr string, submitToGallery bool, output requests.CogWebhookOutput) (*ent.VoiceoverOutput, error) {
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

		// If this voiceover was created with "submit_to_gallery", then submit all outputs to gallery
		var galleryStatus voiceoveroutput.GalleryStatus
		if submitToGallery {
			galleryStatus = voiceoveroutput.GalleryStatusSubmitted
		} else {
			galleryStatus = voiceoveroutput.GalleryStatusNotSubmitted
		}

		// Set audio output
		parsedS3, err := utils.GetPathFromS3URL(output.AudioFiles[0].AudioFile)
		if err != nil {
			log.Error("Error parsing s3 url", "output", output, "err", err)
			parsedS3 = output.AudioFiles[0].AudioFile
		}
		vOutput := tx.VoiceoverOutput.Create().SetAudioPath(parsedS3).SetVoiceoverID(uid).SetGalleryStatus(galleryStatus).SetAudioDuration(output.AudioFiles[0].AudioDuration)
		if output.AudioFiles[0].VideoFile != "" {
			parsedVideoS3, err := utils.GetPathFromS3URL(output.AudioFiles[0].VideoFile)
			if err != nil {
				parsedVideoS3 = output.AudioFiles[0].VideoFile
			}
			vOutput.SetVideoPath(parsedVideoS3)
		}
		if len(output.AudioFiles[0].AudioArray) > 0 {
			vOutput.SetAudioArray(output.AudioFiles[0].AudioArray)
		}
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

// Marks voiceovers for deletions by setting deleted_at
func (r *Repository) MarkVoiceoverOutputsForDeletion(outputIDs []uuid.UUID) (int, error) {
	var deleted int
	var err error
	deletedAt := time.Now()
	// Start transaction
	if err := r.WithTx(func(tx *ent.Tx) error {
		db := tx.Client()
		deleted, err = db.VoiceoverOutput.Update().Where(voiceoveroutput.IDIn(outputIDs...)).SetDeletedAt(deletedAt).Save(r.Ctx)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return 0, err
	}
	return deleted, nil
}

// Marks voiceovers for deletions by setting deleted_at, only if they belong to the user with ID userID
func (r *Repository) MarkVoiceoverOutputsForDeletionForUser(outputIDs []uuid.UUID, userID uuid.UUID) (int, error) {
	// Get outputs belonging to this user
	outputs, err := r.DB.Voiceover.Query().Select().Where(voiceover.UserIDEQ(userID)).QueryVoiceoverOutputs().Select(voiceoveroutput.FieldID).Where(voiceoveroutput.IDIn(outputIDs...)).All(r.Ctx)
	if err != nil {
		return 0, err
	}

	// Filter out outputs that don't belong to the user
	var userVoiceoverOutputIDs []uuid.UUID
	for _, output := range outputs {
		userVoiceoverOutputIDs = append(userVoiceoverOutputIDs, output.ID)
	}

	// Execute delete
	return r.MarkVoiceoverOutputsForDeletion(userVoiceoverOutputIDs)
}
