package repository

import (
	"errors"
	"testing"

	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/ent/deviceinfo"
	"github.com/stretchr/testify/assert"
)

// Test that wrapper rolls back transaction when error is thrown
func TestTxWrapper(t *testing.T) {
	err := MockRepo.WithTx(func(tx *ent.Tx) error {
		DB := tx.Client()
		// Change something arbitrary
		_, err := DB.DeviceInfo.Create().SetType("rollback").SetOs("rollback").SetBrowser("rollback").Save(MockRepo.Ctx)
		assert.Nil(t, err)

		// Query to make sure exists
		dinfo := DB.DeviceInfo.Query().Where(deviceinfo.Type("rollback"), deviceinfo.Os("rollback"), deviceinfo.Browser("rollback")).FirstX(MockRepo.Ctx)
		assert.NotNil(t, dinfo)
		assert.Equal(t, "rollback", dinfo.Type)

		// Throw an error to trigger rollback
		return errors.New("rollback")
	})

	assert.NotNil(t, err)
	// Should not be found
	_, err = MockRepo.DB.DeviceInfo.Query().Where(deviceinfo.Type("rollback"), deviceinfo.Os("rollback"), deviceinfo.Browser("rollback")).First(MockRepo.Ctx)
	assert.NotNil(t, err)
	assert.True(t, ent.IsNotFound(err))
}
