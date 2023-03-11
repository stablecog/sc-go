package repository

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCreateUser(t *testing.T) {
	u, err := MockRepo.CreateUser(uuid.New(), "testcreateuser@stablecog.com", "cus_1234", nil)
	assert.Nil(t, err)
	assert.NotNil(t, u)

	// Delete
	MockRepo.DB.User.DeleteOne(u).ExecX(MockRepo.Ctx)
}

func TestSetActiveProductID(t *testing.T) {
	u, err := MockRepo.CreateUser(uuid.New(), "testsetactiveproductid@stablecog.com", "cus_1234", nil)
	assert.Nil(t, err)
	assert.NotNil(t, u)
	assert.Nil(t, u.ActiveProductID)

	// Set
	err = MockRepo.SetActiveProductID(u.ID, "prod_1234", nil)
	assert.Nil(t, err)

	// Get user
	u, err = MockRepo.GetUser(u.ID)
	assert.Nil(t, err)
	assert.NotNil(t, u)
	assert.Equal(t, "prod_1234", *u.ActiveProductID)

	// Delete
	MockRepo.DB.User.DeleteOne(u).ExecX(MockRepo.Ctx)
}

func TestUnsetActiveProductID(t *testing.T) {
	u, err := MockRepo.CreateUser(uuid.New(), "testunsetactiveproductid@stablecog.com", "cus_1234", nil)
	assert.Nil(t, err)
	assert.NotNil(t, u)
	assert.Nil(t, u.ActiveProductID)

	// Set
	err = MockRepo.SetActiveProductID(u.ID, "prod_1234", nil)
	assert.Nil(t, err)

	// Get user
	u, err = MockRepo.GetUser(u.ID)
	assert.Nil(t, err)
	assert.NotNil(t, u)
	assert.Equal(t, "prod_1234", *u.ActiveProductID)

	// Unset
	changed, err := MockRepo.UnsetActiveProductID(u.ID, "prod_12345", nil)
	assert.Nil(t, err)
	assert.Equal(t, 0, changed)
	changed, err = MockRepo.UnsetActiveProductID(u.ID, "prod_1234", nil)
	assert.Nil(t, err)
	assert.Equal(t, 1, changed)

	// Get user
	u, err = MockRepo.GetUser(u.ID)
	assert.Nil(t, err)
	assert.NotNil(t, u)
	assert.Nil(t, u.ActiveProductID)

	// Delete
	MockRepo.DB.User.DeleteOne(u).ExecX(MockRepo.Ctx)
}
