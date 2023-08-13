package shared

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stretchr/testify/assert"
)

func resetCache() {
	singleCache = newCache()
}

func TestGetCacheReturnsSameInstance(t *testing.T) {
	resetCache()
	fc1 := GetCache()
	fc1.adminIDs = []uuid.UUID{uuid.New()}
	fc2 := GetCache()
	assert.Len(t, fc2.AdminIDs(), 1)
}

func TestUpdateGenerateModels(t *testing.T) {
	resetCache()
	fc := GetCache()
	assert.Len(t, fc.GenerationModels(), 0)
	models := []*ent.GenerationModel{
		{NameInWorker: "test"},
	}
	fc.UpdateGenerationModels(models)
	assert.Equal(t, 1, len(fc.GenerationModels()))
	assert.Equal(t, "test", fc.GenerationModels()[0].NameInWorker)
}

func TestUpdateUpscaleModels(t *testing.T) {
	resetCache()
	fc := GetCache()
	assert.Len(t, fc.UpscaleModels(), 0)
	models := []*ent.UpscaleModel{
		{NameInWorker: "test"},
	}
	fc.UpdateUpscaleModels(models)
	assert.Equal(t, 1, len(fc.UpscaleModels()))
	assert.Equal(t, "test", fc.UpscaleModels()[0].NameInWorker)
}

func TestUpdateSchedulers(t *testing.T) {
	resetCache()
	fc := GetCache()
	assert.Len(t, fc.Schedulers(), 0)
	schedulrs := []*ent.Scheduler{
		{NameInWorker: "test"},
	}
	fc.UpdateSchedulers(schedulrs)
	assert.Equal(t, 1, len(fc.Schedulers()))
	assert.Equal(t, "test", fc.Schedulers()[0].NameInWorker)
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

func TestIsAdmin(t *testing.T) {
	resetCache()
	fc := GetCache()
	// Predictable uuid
	uid := uuid.MustParse("00000000-0000-0000-0000-000000000000")
	// Assert not admin
	assert.False(t, fc.IsAdmin(uid))
	// Add to models
	fc.SetAdminUUIDs([]uuid.UUID{uid})
	// Assert
	assert.True(t, fc.IsAdmin(uid))
}

func TestUpdateDisposableEmails(t *testing.T) {
	resetCache()
	fc := GetCache()
	assert.Len(t, fc.DisposableEmailDomains(), 0)
	emails := []string{"test"}
	fc.UpdateDisposableEmailDomains(emails)
	assert.Equal(t, 1, len(fc.DisposableEmailDomains()))
	assert.Equal(t, "test", fc.DisposableEmailDomains()[0])
}

func TestIsDisposableEmail(t *testing.T) {
	resetCache()
	fc := GetCache()
	assert.False(t, fc.IsDisposableEmail("test"))
	fc.UpdateDisposableEmailDomains([]string{"test"})
	assert.True(t, fc.IsDisposableEmail("test@test"))
}
