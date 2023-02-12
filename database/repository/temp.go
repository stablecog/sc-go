package repository

import (
	"github.com/google/uuid"
	"github.com/stablecog/go-apps/database/ent"
	"github.com/stablecog/go-apps/database/ent/deviceinfo"
	"github.com/stablecog/go-apps/database/ent/negativeprompt"
	"github.com/stablecog/go-apps/database/ent/prompt"
)

// ! Temporary repo methods since these should go away

// ! TODO - move this to a postgres stored procedure
func (r *Repository) GetOrCreateDeviceInfoAndPrompts(promptText, negativePromptText, deviceType, deviceOs, deviceBrowser string, tx *ent.Tx) (promptId uuid.UUID, negativePromptId *uuid.UUID, deviceInfoId uuid.UUID, err error) {
	dbTx := &DBTransaction{TX: tx, DB: r.DB}
	err = dbTx.Start(r.Ctx)
	if err != nil {
		return promptId, negativePromptId, deviceInfoId, err
	}
	// Check if prompt exists
	var dbPrompt *ent.Prompt
	dbPrompt, err = dbTx.TX.Prompt.Query().Where(prompt.TextEQ(promptText)).Only(r.Ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			// Create prompt
			dbPrompt, err = dbTx.TX.Prompt.Create().SetText(promptText).Save(r.Ctx)
			if err != nil {
				dbTx.Rollback()
				return promptId, negativePromptId, deviceInfoId, err
			}
		} else {
			dbTx.Rollback()
			return promptId, negativePromptId, deviceInfoId, err
		}
	}

	// Check if negative prompt exists
	var dbNegativePrompt *ent.NegativePrompt
	if negativePromptText != "" {
		dbNegativePrompt, err = dbTx.TX.NegativePrompt.Query().Where(negativeprompt.TextEQ(negativePromptText)).Only(r.Ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				// Create negative prompt
				dbNegativePrompt, err = dbTx.TX.NegativePrompt.Create().SetText(negativePromptText).Save(r.Ctx)
				if err != nil {
					dbTx.Rollback()
					return promptId, negativePromptId, deviceInfoId, err
				}
			} else {
				dbTx.Rollback()
				return promptId, negativePromptId, deviceInfoId, err
			}
		}
	}

	// Check if device info combo exists
	var dbDeviceInfo *ent.DeviceInfo
	dbDeviceInfo, err = dbTx.TX.DeviceInfo.Query().Where(deviceinfo.Type(deviceType), deviceinfo.Os(deviceOs), deviceinfo.Browser(deviceBrowser)).Only(r.Ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			// Create device info
			dbDeviceInfo, err = dbTx.TX.DeviceInfo.Create().SetType(deviceType).SetOs(deviceOs).SetBrowser(deviceBrowser).Save(r.Ctx)
			if err != nil {
				dbTx.Rollback()
				return promptId, negativePromptId, deviceInfoId, err
			}
		} else {
			dbTx.Rollback()
			return promptId, negativePromptId, deviceInfoId, err
		}
	}

	err = dbTx.Commit()
	if err != nil {
		return promptId, negativePromptId, deviceInfoId, err
	}

	// Negative prompt is optional so
	if dbNegativePrompt == nil {
		return dbPrompt.ID, nil, dbDeviceInfo.ID, nil
	}
	return dbPrompt.ID, &dbNegativePrompt.ID, dbDeviceInfo.ID, nil
}

// Get a device_info ID given inputs
func (r *Repository) GetOrCreateDeviceInfo(deviceType, deviceOs, deviceBrowser string, tx *ent.Tx) (deviceInfoId uuid.UUID, err error) {
	dbTx := &DBTransaction{TX: tx, DB: r.DB}
	err = dbTx.Start(r.Ctx)
	if err != nil {
		return deviceInfoId, err
	}

	// Check if device info combo exists
	var dbDeviceInfo *ent.DeviceInfo
	dbDeviceInfo, err = dbTx.TX.DeviceInfo.Query().Where(deviceinfo.Type(deviceType), deviceinfo.Os(deviceOs), deviceinfo.Browser(deviceBrowser)).Only(r.Ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			// Create device info
			dbDeviceInfo, err = dbTx.TX.DeviceInfo.Create().SetType(deviceType).SetOs(deviceOs).SetBrowser(deviceBrowser).Save(r.Ctx)
			if err != nil {
				dbTx.Rollback()
				return deviceInfoId, err
			}
		} else {
			dbTx.Rollback()
			return deviceInfoId, err
		}
	}

	err = dbTx.Commit()
	if err != nil {
		return deviceInfoId, err
	}

	return dbDeviceInfo.ID, nil
}
