package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/ent/upscale"
	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/shared"
	"k8s.io/klog/v2"
)

// Get upscale by ID
func (r *Repository) GetUpscale(id uuid.UUID) (*ent.Upscale, error) {
	return r.DB.Upscale.Query().Where(upscale.IDEQ(id)).First(r.Ctx)
}

// CreateUpscale creates the initial generation in the database
// Takes in a userID (creator),  device info, countryCode, and a request body
func (r *Repository) CreateUpscale(userID uuid.UUID, width, height int32, deviceType, deviceOs, deviceBrowser, countryCode string, req requests.CreateUpscaleRequest, DB *ent.Client) (*ent.Upscale, error) {
	if DB == nil {
		DB = r.DB
	}
	// Get prompt, negative prompt, device info
	deviceInfoId, err := r.GetOrCreateDeviceInfo(deviceType, deviceOs, deviceBrowser, DB)
	if err != nil {
		return nil, err
	}
	return DB.Upscale.Create().
		SetStatus(upscale.StatusQueued).
		SetWidth(width).
		SetHeight(height).
		SetModelID(req.ModelId).
		SetDeviceInfoID(deviceInfoId).
		SetCountryCode(countryCode).
		SetScale(shared.DEFAULT_UPSCALE_SCALE).
		SetUserID(userID).Save(r.Ctx)
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

func (r *Repository) SetUpscaleFailed(upscaleID string, reason string, db *ent.Client) error {
	if db == nil {
		db = r.DB
	}

	uid, err := uuid.Parse(upscaleID)
	if err != nil {
		klog.Errorf("Error parsing generation id in SetUpscaleFailed %s: %v", upscaleID, err)
		return err
	}
	_, err = db.Upscale.UpdateOneID(uid).SetStatus(upscale.StatusFailed).SetFailureReason(reason).SetCompletedAt(time.Now()).Save(r.Ctx)
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

	var upscaleOutput *ent.UpscaleOutput

	// Start a transaction
	if err := r.WithTx(func(tx *ent.Tx) error {
		// Retrieve the upscale
		u, err := r.GetUpscale(uid)
		if err != nil {
			klog.Errorf("Error retrieving upscale %s: %v", upscaleID, err)
			return err
		}

		// Update the upscale
		_, err = tx.Upscale.UpdateOneID(u.ID).SetStatus(upscale.StatusSucceeded).SetCompletedAt(time.Now()).Save(r.Ctx)
		if err != nil {
			klog.Errorf("Error setting upscale succeeded %s: %v", upscaleID, err)
			return err
		}

		// Set upscale output
		upscaleOutput, err = tx.UpscaleOutput.Create().SetImagePath(output).SetUpscaleID(uid).Save(r.Ctx)
		if err != nil {
			klog.Errorf("Error inserting upscale output %s: %v", upscaleID, err)
			return err
		}

		// If necessary add to generation output
		if hasGenerationOutput {
			_, err = tx.GenerationOutput.UpdateOneID(outputId).SetUpscaledImagePath(output).Save(r.Ctx)
			if err != nil {
				klog.Errorf("Error setting upscaled_image_url %s: %v", upscaleID, err)
				return err
			}
		}

		return nil
	}); err != nil {
		klog.Errorf("Error in SetUpscaleSucceeded %s: %v", upscaleID, err)
		return nil, err
	}

	return upscaleOutput, nil
}
