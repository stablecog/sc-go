package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetUpscale(t *testing.T) {
	u, err := MockRepo.CreateMockUpscaleForDeletion(MockRepo.Ctx)
	assert.Nil(t, err)
	assert.NotNil(t, u)

	// Get
	u2, err := MockRepo.GetUpscale(u.ID)
	assert.Nil(t, err)
	assert.NotNil(t, u2)

	// Assert
	assert.Equal(t, u.ID, u2.ID)

	// Delete
	MockRepo.DB.Upscale.DeleteOne(u).ExecX(MockRepo.Ctx)
}
