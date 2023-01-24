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
func (r *Repository) GetOrCreateDeviceInfoAndPrompts(promptText, negativePromptText, deviceType, deviceOs, deviceBrowser string) (promptId uuid.UUID, negativePromptId *uuid.UUID, deviceInfoId uuid.UUID, err error) {
	// Start tx
	tx, err := r.DB.Tx(r.Ctx)
	if err != nil {
		return uuid.UUID{}, nil, uuid.UUID{}, err
	}
	// Check if prompt exists
	var dbPrompt *ent.Prompt
	dbPrompt, err = tx.Prompt.Query().Where(prompt.TextEQ(promptText)).Only(r.Ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			// Create prompt
			dbPrompt, err = tx.Prompt.Create().SetText(promptText).Save(r.Ctx)
			if err != nil {
				tx.Rollback()
				return uuid.UUID{}, nil, uuid.UUID{}, err
			}
		} else {
			tx.Rollback()
			return uuid.UUID{}, nil, uuid.UUID{}, err
		}
	}

	// Check if negative prompt exists
	var dbNegativePrompt *ent.NegativePrompt
	if negativePromptText != "" {
		dbNegativePrompt, err = tx.NegativePrompt.Query().Where(negativeprompt.TextEQ(negativePromptText)).Only(r.Ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				// Create negative prompt
				dbNegativePrompt, err = tx.NegativePrompt.Create().SetText(negativePromptText).Save(r.Ctx)
				if err != nil {
					tx.Rollback()
					return uuid.UUID{}, nil, uuid.UUID{}, err
				}
			} else {
				tx.Rollback()
				return uuid.UUID{}, nil, uuid.UUID{}, err
			}
		}
	}

	// Check if device info combo exists
	var dbDeviceInfo *ent.DeviceInfo
	dbDeviceInfo, err = tx.DeviceInfo.Query().Where(deviceinfo.Type(deviceType), deviceinfo.Os(deviceOs), deviceinfo.Browser(deviceBrowser)).Only(r.Ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			// Create device info
			dbDeviceInfo, err = tx.DeviceInfo.Create().SetType(deviceType).SetOs(deviceOs).SetBrowser(deviceBrowser).Save(r.Ctx)
			if err != nil {
				tx.Rollback()
				return uuid.UUID{}, nil, uuid.UUID{}, err
			}
		} else {
			tx.Rollback()
			return uuid.UUID{}, nil, uuid.UUID{}, err
		}
	}

	err = tx.Commit()
	if err != nil {
		return uuid.UUID{}, nil, uuid.UUID{}, err
	}

	// Negative prompt is optional so
	if dbNegativePrompt == nil {
		return dbPrompt.ID, nil, dbDeviceInfo.ID, nil
	}
	return dbPrompt.ID, &dbNegativePrompt.ID, dbDeviceInfo.ID, nil
}
