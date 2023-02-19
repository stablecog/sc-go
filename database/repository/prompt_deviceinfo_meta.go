package repository

import (
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/ent/deviceinfo"
	"github.com/stablecog/sc-go/database/ent/negativeprompt"
	"github.com/stablecog/sc-go/database/ent/prompt"
)

// Create device info and prompts if they don't exist, otherwise get existing
func (r *Repository) GetOrCreateDeviceInfoAndPrompts(promptText, negativePromptText, deviceType, deviceOs, deviceBrowser string, DB *ent.Client) (promptId uuid.UUID, negativePromptId *uuid.UUID, deviceInfoId uuid.UUID, err error) {
	if DB == nil {
		DB = r.DB
	}
	// Check if prompt exists
	var dbPrompt *ent.Prompt
	dbPrompt, err = DB.Prompt.Query().Where(prompt.TextEQ(promptText)).Only(r.Ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			// Create prompt
			dbPrompt, err = DB.Prompt.Create().SetText(promptText).Save(r.Ctx)
			if err != nil {
				return promptId, negativePromptId, deviceInfoId, err
			}
		} else {
			return promptId, negativePromptId, deviceInfoId, err
		}
	}

	// Check if negative prompt exists
	var dbNegativePrompt *ent.NegativePrompt
	if negativePromptText != "" {
		dbNegativePrompt, err = DB.NegativePrompt.Query().Where(negativeprompt.TextEQ(negativePromptText)).Only(r.Ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				// Create negative prompt
				dbNegativePrompt, err = DB.NegativePrompt.Create().SetText(negativePromptText).Save(r.Ctx)
				if err != nil {
					return promptId, negativePromptId, deviceInfoId, err
				}
			} else {
				return promptId, negativePromptId, deviceInfoId, err
			}
		}
	}

	// Check if device info combo exists
	var dbDeviceInfo *ent.DeviceInfo
	dbDeviceInfo, err = DB.DeviceInfo.Query().Where(deviceinfo.Type(deviceType), deviceinfo.Os(deviceOs), deviceinfo.Browser(deviceBrowser)).Only(r.Ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			// Create device info
			dbDeviceInfo, err = DB.DeviceInfo.Create().SetType(deviceType).SetOs(deviceOs).SetBrowser(deviceBrowser).Save(r.Ctx)
			if err != nil {
				return promptId, negativePromptId, deviceInfoId, err
			}
		} else {
			return promptId, negativePromptId, deviceInfoId, err
		}
	}

	// Negative prompt is optional so
	if dbNegativePrompt == nil {
		return dbPrompt.ID, nil, dbDeviceInfo.ID, nil
	}
	return dbPrompt.ID, &dbNegativePrompt.ID, dbDeviceInfo.ID, nil
}

// Get a device_info ID given inputs
func (r *Repository) GetOrCreateDeviceInfo(deviceType, deviceOs, deviceBrowser string, DB *ent.Client) (deviceInfoId uuid.UUID, err error) {
	if DB == nil {
		DB = r.DB
	}
	// Check if device info combo exists
	var dbDeviceInfo *ent.DeviceInfo
	dbDeviceInfo, err = DB.DeviceInfo.Query().Where(deviceinfo.Type(deviceType), deviceinfo.Os(deviceOs), deviceinfo.Browser(deviceBrowser)).Only(r.Ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			// Create device info
			dbDeviceInfo, err = DB.DeviceInfo.Create().SetType(deviceType).SetOs(deviceOs).SetBrowser(deviceBrowser).Save(r.Ctx)
			if err != nil {
				return deviceInfoId, err
			}
		} else {
			return deviceInfoId, err
		}
	}

	return dbDeviceInfo.ID, nil
}
