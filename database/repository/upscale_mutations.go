package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/ent/upscale"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/shared"
	"github.com/stablecog/sc-go/utils"
)

// CreateUpscale creates the initial generation in the database
// Takes in a userID (creator),  device info, countryCode, and a request body
func (r *Repository) CreateUpscale(userID uuid.UUID, width, height int32, deviceType, deviceOs, deviceBrowser, countryCode string, req requests.CreateUpscaleRequest, productId *string, systemGenerated bool, apiTokenId *uuid.UUID, DB *ent.Client) (*ent.Upscale, error) {
	if DB == nil {
		DB = r.DB
	}
	// Get prompt, negative prompt, device info
	deviceInfoId, err := r.GetOrCreateDeviceInfo(deviceType, deviceOs, deviceBrowser, DB)
	if err != nil {
		return nil, err
	}
	insert := DB.Upscale.Create().
		SetStatus(upscale.StatusQueued).
		SetWidth(width).
		SetHeight(height).
		SetModelID(req.ModelId).
		SetDeviceInfoID(deviceInfoId).
		SetCountryCode(countryCode).
		SetScale(shared.DEFAULT_UPSCALE_SCALE).
		SetUserID(userID).
		SetSystemGenerated(systemGenerated)
	if productId != nil {
		insert.SetStripeProductID(*productId)
	}
	if apiTokenId != nil {
		insert.SetAPITokenID(*apiTokenId)
	}
	return insert.Save(r.Ctx)
}

func (r *Repository) SetUpscaleStarted(upscaleID string) error {
	uid, err := uuid.Parse(upscaleID)
	if err != nil {
		log.Error("Error parsing generation id in SetUpscaleStarted", "id", upscaleID, "err", err)
		return err
	}
	_, err = r.DB.Upscale.UpdateOneID(uid).SetStatus(upscale.StatusStarted).SetStartedAt(time.Now()).Save(r.Ctx)
	if err != nil {
		// Log error here since this might be happening in a goroutine
		log.Error("Error setting upscale started", "id", upscaleID, "err", err)
	}
	return err
}

func (r *Repository) SetUpscaleFailed(upscaleID string, reason string, db *ent.Client) error {
	if db == nil {
		db = r.DB
	}

	uid, err := uuid.Parse(upscaleID)
	if err != nil {
		log.Error("Error parsing generation id in SetUpscaleFailed", "id", upscaleID, "err", err)
		return err
	}
	_, err = db.Upscale.UpdateOneID(uid).SetStatus(upscale.StatusFailed).SetFailureReason(reason).SetCompletedAt(time.Now()).Save(r.Ctx)
	if err != nil {
		log.Error("Error setting upscale failed", "id", upscaleID, "err", err)
	}
	return err
}

// ! Currently supports 1 output
func (r *Repository) SetUpscaleSucceeded(upscaleID, generationOutputID, inputImageUrl string, output requests.CogWebhookOutput) (*ent.UpscaleOutput, error) {
	uid, err := uuid.Parse(upscaleID)
	if err != nil {
		log.Error("Error parsing generation id in SetUpscaleSucceeded", "id", upscaleID, "err", err)
		return nil, err
	}

	// If output, we also add upscale image output to the corresponding generation_output
	hasGenerationOutput := true
	outputId, err := uuid.Parse(generationOutputID)
	if err != nil {
		hasGenerationOutput = false
	}

	var upscaleOutput *ent.UpscaleOutput

	// Start a transaction
	if err := r.WithTx(func(tx *ent.Tx) error {
		// Retrieve the upscale
		u, err := r.GetUpscale(uid)
		if err != nil {
			log.Error("Error retrieving upscale", "id", upscaleID, "err", err)
			return err
		}

		// Update the upscale
		_, err = tx.Upscale.UpdateOneID(u.ID).SetStatus(upscale.StatusSucceeded).SetCompletedAt(time.Now()).Save(r.Ctx)
		if err != nil {
			log.Error("Error setting upscale succeeded", "id", upscaleID, "err", err)
			return err
		}

		// Set upscale output
		parsedS3, err := utils.GetPathFromS3URL(output.Images[0].Image)
		if err != nil {
			log.Error("Error parsing s3 url", "output", output, "err", err)
			parsedS3 = output.Images[0].Image
		}
		uOutput := tx.UpscaleOutput.Create().SetImagePath(parsedS3).SetInputImageURL(inputImageUrl).SetUpscaleID(uid)
		if hasGenerationOutput {
			uOutput.SetGenerationOutputID(outputId)
		}
		upscaleOutput, err = uOutput.Save(r.Ctx)
		if err != nil {
			log.Error("Error inserting upscale output", "id", upscaleID, "err", err)
			return err
		}

		// If necessary add to generation output
		if hasGenerationOutput {
			parsedS3, err := utils.GetPathFromS3URL(output.Images[0].Image)
			if err != nil {
				log.Error("Error parsing s3 url", "output", output, "err", err)
				parsedS3 = output.Images[0].Image
			}
			gOutput, err := tx.GenerationOutput.UpdateOneID(outputId).SetUpscaledImagePath(parsedS3).Save(r.Ctx)
			if err != nil {
				log.Error("Error setting upscaled_image_url", "id", upscaleID, "err", err)
				return err
			}
			if gOutput.HasEmbeddings && r.Qdrant != nil {
				payload := map[string]interface{}{
					"upscaled_image_path": parsedS3,
				}
				err = r.Qdrant.SetPayload(payload, []uuid.UUID{gOutput.ID}, false)
				if err != nil {
					log.Error("Error setting upscaled_image_url in Qdrant", "id", upscaleID, "err", err)
					return err
				}
			}
		}

		return nil
	}); err != nil {
		log.Error("Error in SetUpscaleSucceeded", "id", upscaleID, "err", err)
		return nil, err
	}

	return upscaleOutput, nil
}
