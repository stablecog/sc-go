package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/stablecog/go-apps/database/ent"
	"github.com/stablecog/go-apps/database/ent/upscale"
	"github.com/stablecog/go-apps/server/requests"
	"github.com/stablecog/go-apps/shared"
	"k8s.io/klog/v2"
)

// Get upscale by ID
func (r *Repository) GetUpscale(id uuid.UUID) (*ent.Upscale, error) {
	return r.DB.Upscale.Query().Where(upscale.IDEQ(id)).First(r.Ctx)
}

// CreateUpscale creates the initial generation in the database
// Takes in a userID (creator),  device info, countryCode, and a request body
func (r *Repository) CreateUpscale(userID uuid.UUID, width, height int32, deviceType, deviceOs, deviceBrowser, countryCode string, req requests.UpscaleRequestBody, tx *ent.Tx) (*ent.Upscale, error) {
	dbTx := &DBTransaction{TX: tx, DB: r.DB}
	err := dbTx.Start(r.Ctx)
	if err != nil {
		return nil, err
	}
	// Get prompt, negative prompt, device info
	deviceInfoId, err := r.GetOrCreateDeviceInfo(deviceType, deviceOs, deviceBrowser, dbTx.TX)
	if err != nil {
		dbTx.Rollback()
		return nil, err
	}
	u, err := dbTx.TX.Upscale.Create().
		SetStatus(upscale.StatusQueued).
		SetWidth(width).
		SetHeight(height).
		SetModelID(req.ModelId).
		SetDeviceInfoID(deviceInfoId).
		SetCountryCode(countryCode).
		SetScale(shared.DEFAULT_UPSCALE_SCALE).
		SetUserID(userID).Save(r.Ctx)
	if err != nil {
		dbTx.Rollback()
		return nil, err
	}
	err = dbTx.Commit()
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *Repository) SetUpscaleStarted(upscaleID string) error {
	uid, err := uuid.Parse(upscaleID)
	if err != nil {
		klog.Errorf("Error parsing generation id in SetUpscaleStarted %s: %v", upscaleID, err)
		return err
	}
	_, err = r.DB.Upscale.UpdateOneID(uid).SetStatus(upscale.StatusStarted).SetStartedAt(time.Now()).Save(r.Ctx)
	if err != nil {
		// Log error here since this might be happening in a goroutine
		klog.Errorf("Error setting upscale started %s: %v", upscaleID, err)
	}
	return err
}

func (r *Repository) SetUpscaleFailed(upscaleID string, reason string) error {
	uid, err := uuid.Parse(upscaleID)
	if err != nil {
		klog.Errorf("Error parsing generation id in SetUpscaleFailed %s: %v", upscaleID, err)
		return err
	}
	_, err = r.DB.Upscale.UpdateOneID(uid).SetStatus(upscale.StatusFailed).SetFailureReason(reason).Save(r.Ctx)
	if err != nil {
		klog.Errorf("Error setting upscale failed %s: %v", upscaleID, err)
	}
	return err
}

// ! Currently supports 1 output
func (r *Repository) SetUpscaleSucceeded(upscaleID, generationOutputID, output string) (*ent.UpscaleOutput, error) {
	uid, err := uuid.Parse(upscaleID)
	if err != nil {
		klog.Errorf("Error parsing generation id in SetUpscaleSucceeded %s: %v", upscaleID, err)
		return nil, err
	}

	// If output, we also add upscale image output to the corresponding generation_output
	hasGenerationOutput := true
	outputId, err := uuid.Parse(generationOutputID)
	if err != nil {
		hasGenerationOutput = false
	}

	// Start a transaction
	tx, err := r.DB.Tx(r.Ctx)
	if err != nil {
		klog.Errorf("Error starting transaction in SetUpscaleSucceeded %s: %v", upscaleID, err)
		return nil, err
	}

	// Retrieve the upscale
	u, err := r.GetUpscale(uid)
	if err != nil {
		tx.Rollback()
		klog.Errorf("Error retrieving upscale %s: %v", upscaleID, err)
		return nil, err
	}

	// Update the upscale
	_, err = tx.Upscale.UpdateOneID(u.ID).SetStatus(upscale.StatusSucceeded).SetCompletedAt(time.Now()).Save(r.Ctx)
	if err != nil {
		tx.Rollback()
		klog.Errorf("Error setting upscale succeeded %s: %v", upscaleID, err)
		return nil, err
	}

	// Set upscale output
	upscaleOutput, err := tx.UpscaleOutput.Create().SetImageURL(output).SetUpscaleID(uid).Save(r.Ctx)
	if err != nil {
		tx.Rollback()
		klog.Errorf("Error inserting upscale output %s: %v", upscaleID, err)
		return nil, err
	}

	// If necessary add to generation output
	if hasGenerationOutput {
		_, err = tx.GenerationOutput.UpdateOneID(outputId).SetUpscaledImageURL(output).Save(r.Ctx)
		if err != nil {
			tx.Rollback()
			klog.Errorf("Error setting upscaled_image_url %s: %v", upscaleID, err)
			return nil, err
		}
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		klog.Errorf("Error committing transaction in SetUpscaleSucceeded %s: %v", upscaleID, err)
		return nil, err
	}
	return upscaleOutput, nil
}
