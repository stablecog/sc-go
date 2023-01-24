package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCache(t *testing.T) {
	fc := newCache()
	assert.Equal(t, 512, fc.FreeWidths[0])
	assert.Equal(t, 512, fc.FreeHeights[0])
	assert.Equal(t, 30, fc.FreeInterferenceSteps[0])
}

func TestGetCacheReturnsSameInstance(t *testing.T) {
	fc1 := GetCache()
	fc1.FreeHeights[0] = 1024
	fc2 := GetCache()
	assert.Equal(t, 1024, fc2.FreeHeights[0])
}

func TestUpdateFreeModelsAndSchedulers(t *testing.T) {
	fc := GetCache()
	fc.UpdateFreeModelsAndSchedulers([]string{"model1", "model2"}, []string{"scheduler1", "scheduler2"})
	assert.Equal(t, 2, len(fc.FreeModelIDs))
	assert.Equal(t, "model1", fc.FreeModelIDs[0])
	assert.Equal(t, "model2", fc.FreeModelIDs[1])
	assert.Equal(t, 2, len(fc.FreeSchedulerIDs))
	assert.Equal(t, "scheduler1", fc.FreeSchedulerIDs[0])
	assert.Equal(t, "scheduler2", fc.FreeSchedulerIDs[1])
}
