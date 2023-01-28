package shared

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stablecog/go-apps/database/ent"
	"github.com/stretchr/testify/assert"
)

func TestNewCache(t *testing.T) {
	fc := newCache()
	assert.Equal(t, int32(512), fc.FreeWidths[0])
	assert.Equal(t, int32(512), fc.FreeHeights[0])
	assert.Equal(t, int32(30), fc.FreeInterferenceSteps[0])
}

func TestGetCacheReturnsSameInstance(t *testing.T) {
	fc1 := GetCache()
	fc1.FreeHeights[0] = 1024
	fc2 := GetCache()
	assert.Equal(t, int32(1024), fc2.FreeHeights[0])
}

func TestUpdateGenerateModels(t *testing.T) {
	fc := GetCache()
	assert.Len(t, fc.GenerateModels, 0)
	models := []*ent.GenerationModel{
		{Name: "test"},
	}
	fc.UpdateGenerationModels(models)
	assert.Equal(t, 1, len(fc.GenerateModels))
	assert.Equal(t, "test", fc.GenerateModels[0].Name)
}

func TestUpdateSchedulers(t *testing.T) {
	fc := GetCache()
	assert.Len(t, fc.Schedulers, 0)
	schedulrs := []*ent.Scheduler{
		{Name: "test"},
	}
	fc.UpdateSchedulers(schedulrs)
	assert.Equal(t, 1, len(fc.Schedulers))
	assert.Equal(t, "test", fc.Schedulers[0].Name)
}

func TestIsValidGenerationModelID(t *testing.T) {
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
func TestIsValidSchedulerID(t *testing.T) {
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

func TestIsGenerationModelAvailableForFree(t *testing.T) {
	fc := GetCache()
	// Predictable uuid
	uid := uuid.MustParse("00000000-0000-0000-0000-000000000000")
	assert.False(t, fc.IsGenerationModelAvailableForFree(uid))
	// Add to models
	fc.UpdateGenerationModels([]*ent.GenerationModel{
		{ID: uid, IsFree: false},
	})
	// Assert not available
	assert.False(t, fc.IsGenerationModelAvailableForFree(uid))
	// Update to free
	fc.UpdateGenerationModels([]*ent.GenerationModel{
		{ID: uid, IsFree: true},
	})
	// Assert available
	assert.True(t, fc.IsGenerationModelAvailableForFree(uid))
}

func TestIsSchedulerAvailableForFree(t *testing.T) {
	fc := GetCache()
	// Predictable uuid
	uid := uuid.MustParse("00000000-0000-0000-0000-000000000000")
	assert.False(t, fc.IsSchedulerAvailableForFree(uid))
	// Add to models
	fc.UpdateSchedulers([]*ent.Scheduler{
		{ID: uid, IsFree: false},
	})
	// Assert not available
	assert.False(t, fc.IsSchedulerAvailableForFree(uid))
	// Update to free
	fc.UpdateSchedulers([]*ent.Scheduler{
		{ID: uid, IsFree: true},
	})
	// Assert available
	assert.True(t, fc.IsSchedulerAvailableForFree(uid))
}

func TestIsWidthAvailableForFree(t *testing.T) {
	fc := GetCache()
	assert.False(t, fc.IsWidthAvailableForFree(1024))
	assert.True(t, fc.IsWidthAvailableForFree(512))
}

func TestIsHeightAvailableForFree(t *testing.T) {
	fc := GetCache()
	assert.False(t, fc.IsHeightAvailableForFree(1024))
	assert.True(t, fc.IsHeightAvailableForFree(512))
}

func TestIsNumInterferenceStepsAvailableForFree(t *testing.T) {
	fc := GetCache()
	assert.False(t, fc.IsNumInterferenceStepsAvailableForFree(31))
	assert.True(t, fc.IsNumInterferenceStepsAvailableForFree(30))
}

func TestGetGenerationModelNameFromID(t *testing.T) {
	fc := GetCache()
	// Predictable uuid
	uid := uuid.MustParse("00000000-0000-0000-0000-000000000000")
	// Add to models
	fc.UpdateGenerationModels([]*ent.GenerationModel{
		{ID: uid, Name: "test"},
	})
	// Assert
	assert.Equal(t, "test", fc.GetGenerationModelNameFromID(uid))
	// Assert empty if not found
	assert.Equal(t, "", fc.GetGenerationModelNameFromID(uuid.MustParse("00000000-0000-0000-0000-000000000001")))
}

func TestGetSchedulerNameFromID(t *testing.T) {
	fc := GetCache()
	// Predictable uuid
	uid := uuid.MustParse("00000000-0000-0000-0000-000000000000")
	// Add to models
	fc.UpdateSchedulers([]*ent.Scheduler{
		{ID: uid, Name: "test"},
	})
	// Assert
	assert.Equal(t, "test", fc.GetSchedulerNameFromID(uid))
	// Assert empty if not found
	assert.Equal(t, "", fc.GetSchedulerNameFromID(uuid.MustParse("00000000-0000-0000-0000-000000000001")))
}
