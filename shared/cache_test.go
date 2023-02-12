package shared

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stretchr/testify/assert"
)

func TestNewCache(t *testing.T) {
	fc := newCache()
	assert.Equal(t, int32(512), fc.FreeWidths[0])
	assert.Equal(t, int32(512), fc.FreeHeights[0])
	assert.Equal(t, int32(30), fc.FreeInterferenceSteps[0])
}

func resetCache() {
	singleCache = newCache()
}

func TestGetCacheReturnsSameInstance(t *testing.T) {
	resetCache()
	fc1 := GetCache()
	fc1.FreeHeights[0] = 1024
	fc2 := GetCache()
	assert.Equal(t, int32(1024), fc2.FreeHeights[0])
}

func TestUpdateGenerateModels(t *testing.T) {
	resetCache()
	fc := GetCache()
	assert.Len(t, fc.GenerateModels, 0)
	models := []*ent.GenerationModel{
		{NameInWorker: "test"},
	}
	fc.UpdateGenerationModels(models)
	assert.Equal(t, 1, len(fc.GenerateModels))
	assert.Equal(t, "test", fc.GenerateModels[0].NameInWorker)
}

func TestUpdateUpscaleModels(t *testing.T) {
	resetCache()
	fc := GetCache()
	assert.Len(t, fc.UpscaleModels, 0)
	models := []*ent.UpscaleModel{
		{NameInWorker: "test"},
	}
	fc.UpdateUpscaleModels(models)
	assert.Equal(t, 1, len(fc.UpscaleModels))
	assert.Equal(t, "test", fc.UpscaleModels[0].NameInWorker)
}

func TestUpdateSchedulers(t *testing.T) {
	resetCache()
	fc := GetCache()
	assert.Len(t, fc.Schedulers, 0)
	schedulrs := []*ent.Scheduler{
		{NameInWorker: "test"},
	}
	fc.UpdateSchedulers(schedulrs)
	assert.Equal(t, 1, len(fc.Schedulers))
	assert.Equal(t, "test", fc.Schedulers[0].NameInWorker)
}

func TestIsValidGenerationModelID(t *testing.T) {
	resetCache()
	fc := GetCache()
	// Predictable uuid
	uid := uuid.MustParse("00000000-0000-0000-0000-000000000000")
	assert.False(t, fc.IsValidGenerationModelID(uid))
	// Add to models
	fc.UpdateGenerationModels([]*ent.GenerationModel{
		{ID: uid},
	})
	// Assert nil err
	assert.True(t, fc.IsValidGenerationModelID(uid))
}

func TestIsValidUpscaleModelID(t *testing.T) {
	resetCache()
	fc := GetCache()
	// Predictable uuid
	uid := uuid.MustParse("00000000-0000-0000-0000-000000000000")
	assert.False(t, fc.IsValidUpscaleModelID(uid))
	// Add to models
	fc.UpdateUpscaleModels([]*ent.UpscaleModel{
		{ID: uid},
	})
	// Assert nil err
	assert.True(t, fc.IsValidUpscaleModelID(uid))
}

func TestIsValidSchedulerID(t *testing.T) {
	resetCache()
	fc := GetCache()
	// Predictable uuid
	uid := uuid.MustParse("00000000-0000-0000-0000-000000000000")
	assert.False(t, fc.IsValidShedulerID(uid))
	// Add to models
	fc.UpdateSchedulers([]*ent.Scheduler{
		{ID: uid},
	})
	// Assert nil err
	assert.True(t, fc.IsValidShedulerID(uid))
}

func TestIsWidthAvailableForFree(t *testing.T) {
	resetCache()
	fc := GetCache()
	assert.False(t, fc.IsWidthAvailableForFree(1024))
	assert.True(t, fc.IsWidthAvailableForFree(512))
}

func TestIsHeightAvailableForFree(t *testing.T) {
	resetCache()
	fc := GetCache()
	assert.False(t, fc.IsHeightAvailableForFree(1024))
	assert.True(t, fc.IsHeightAvailableForFree(512))
}

func TestIsNumInterferenceStepsAvailableForFree(t *testing.T) {
	resetCache()
	fc := GetCache()
	assert.False(t, fc.IsNumInterferenceStepsAvailableForFree(31))
	assert.True(t, fc.IsNumInterferenceStepsAvailableForFree(30))
}

func TestGetGenerationModelNameFromID(t *testing.T) {
	resetCache()
	fc := GetCache()
	// Predictable uuid
	uid := uuid.MustParse("00000000-0000-0000-0000-000000000000")
	// Add to models
	fc.UpdateGenerationModels([]*ent.GenerationModel{
		{ID: uid, NameInWorker: "test"},
	})
	// Assert
	assert.Equal(t, "test", fc.GetGenerationModelNameFromID(uid))
	// Assert empty if not found
	assert.Equal(t, "", fc.GetGenerationModelNameFromID(uuid.MustParse("00000000-0000-0000-0000-000000000001")))
}

func TestGetUpscaleModelNameFromID(t *testing.T) {
	resetCache()
	fc := GetCache()
	// Predictable uuid
	uid := uuid.MustParse("00000000-0000-0000-0000-000000000000")
	// Add to models
	fc.UpdateUpscaleModels([]*ent.UpscaleModel{
		{ID: uid, NameInWorker: "test"},
	})
	// Assert
	assert.Equal(t, "test", fc.GetUpscaleModelNameFromID(uid))
	// Assert empty if not found
	assert.Equal(t, "", fc.GetUpscaleModelNameFromID(uuid.MustParse("00000000-0000-0000-0000-000000000001")))
}

func TestGetSchedulerNameFromID(t *testing.T) {
	resetCache()
	fc := GetCache()
	// Predictable uuid
	uid := uuid.MustParse("00000000-0000-0000-0000-000000000000")
	// Add to models
	fc.UpdateSchedulers([]*ent.Scheduler{
		{ID: uid, NameInWorker: "test"},
	})
	// Assert
	assert.Equal(t, "test", fc.GetSchedulerNameFromID(uid))
	// Assert empty if not found
	assert.Equal(t, "", fc.GetSchedulerNameFromID(uuid.MustParse("00000000-0000-0000-0000-000000000001")))
}
