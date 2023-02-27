package repository

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent/userrole"
	"github.com/stretchr/testify/assert"
)

func TestGetUserWithRoles(t *testing.T) {
	adminId := uuid.MustParse(MOCK_ADMIN_UUID)

	// Get user with roles
	user, err := MockRepo.GetUserWithRoles(adminId)
	assert.Nil(t, err)
	assert.NotNil(t, user)
	assert.Len(t, user.Roles, 1)
	assert.Equal(t, userrole.RoleNameSUPER_ADMIN, user.Roles[0])
	assert.Equal(t, adminId, user.ID)
	assert.Equal(t, "1", user.StripeCustomerID)

	normalId := uuid.MustParse(MOCK_NORMAL_UUID)
	// No roles
	user, err = MockRepo.GetUserWithRoles(normalId)
	assert.Nil(t, err)
	assert.NotNil(t, user)
	assert.Len(t, user.Roles, 0)
	assert.Equal(t, normalId, user.ID)
	assert.Equal(t, "2", user.StripeCustomerID)

	// Non-existent
	user, err = MockRepo.GetUserWithRoles(uuid.New())
	assert.Nil(t, err)
	assert.Nil(t, user)
}
