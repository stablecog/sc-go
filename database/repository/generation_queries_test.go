package repository

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent/generation"
	"github.com/stretchr/testify/assert"
)

func TestGetAvgGenerationQueueTime(t *testing.T) {
	// Setup, create two generations, with predictable timestamps
	// Track IDs to clean up test
	var IDs []uuid.UUID
	// First generation, queued for 60s
	// Create two times that are 60s apart
	startTime := time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)
	endTime := startTime.Add(time.Second * 60)
	deviceInfoId, err := MockRepo.GetOrCreateDeviceInfo("browser", "macos", "chrome", nil)
	assert.Nil(t, err)
	insert := MockRepo.DB.Generation.Create().
		SetWidth(512).
		SetHeight(512).
		SetGuidanceScale(1.0).
		SetInferenceSteps(30).
		SetSeed(1234).
		SetModelID(uuid.MustParse(MOCK_GENERATION_MODEL_ID)).
		SetSchedulerID(uuid.MustParse(MOCK_SCHEDULER_ID)).
		SetDeviceInfoID(deviceInfoId).
		SetCountryCode("DE").
		SetUserID(uuid.MustParse(MOCK_ADMIN_UUID)).
		SetWasAutoSubmitted(false).
		SetCreatedAt(startTime).
		SetStartedAt(endTime).
		SetStatus(generation.StatusSucceeded).
		SetNumOutputs(4)
	gen, err := insert.Save(MockRepo.Ctx)
	assert.Nil(t, err)
	IDs = append(IDs, gen.ID)
	// Second generation, queued for 30s
	// Create two times that are 30s apart
	startTime = time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)
	endTime = startTime.Add(time.Second * 30)
	insert = MockRepo.DB.Generation.Create().
		SetWidth(512).
		SetHeight(512).
		SetGuidanceScale(1.0).
		SetInferenceSteps(30).
		SetSeed(1234).
		SetModelID(uuid.MustParse(MOCK_GENERATION_MODEL_ID)).
		SetSchedulerID(uuid.MustParse(MOCK_SCHEDULER_ID)).
		SetDeviceInfoID(deviceInfoId).
		SetCountryCode("DE").
		SetUserID(uuid.MustParse(MOCK_ADMIN_UUID)).
		SetWasAutoSubmitted(false).
		SetCreatedAt(startTime).
		SetStartedAt(endTime).
		SetStatus(generation.StatusSucceeded).
		SetNumOutputs(4)
	gen, err = insert.Save(MockRepo.Ctx)
	assert.Nil(t, err)
	IDs = append(IDs, gen.ID)

	d, err := MockRepo.GetAvgGenerationQueueTime(time.Now().AddDate(0, 0, -1), 2)
	assert.Nil(t, err)
	assert.Equal(t, 45.0, d)

	// Cleanup
	assert.Equal(t, 2, len(IDs))
	for _, id := range IDs {
		MockRepo.DB.Generation.DeleteOneID(id).ExecX(MockRepo.Ctx)
	}
}
