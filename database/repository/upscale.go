package repository

import (
	"github.com/google/uuid"
	"github.com/stablecog/go-apps/database/ent"
	"github.com/stablecog/go-apps/database/ent/upscale"
	"github.com/stablecog/go-apps/server/requests"
)

// CreateUpscale creates the initial generation in the database
// Takes in a userID (creator),  device info, countryCode, and a request body
func (r *Repository) CreateUpscale(userID uuid.UUID, width, height int32, deviceType, deviceOs, deviceBrowser, countryCode string, req requests.UpscaleRequestBody) (*ent.Upscale, error) {
	// Get prompt, negative prompt, device info
	deviceInfoId, err := r.GetOrCreateDeviceInfo(deviceType, deviceOs, deviceBrowser)
	if err != nil {
		return nil, err
	}
	return r.DB.Upscale.Create().
		SetStatus(upscale.StatusQueued).
		SetWidth(width).
		SetHeight(height).
		SetModelID(req.ModelId).
		SetDeviceInfoID(deviceInfoId).
		SetCountryCode(countryCode).
		SetUserID(userID).Save(r.Ctx)
}
