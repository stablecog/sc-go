package repository

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent/userrole"
	"github.com/stretchr/testify/assert"
)

func TestGetUser(t *testing.T) {
	adminId := uuid.MustParse(MOCK_ADMIN_UUID)

	// Get user
	user, err := MockRepo.GetUser(adminId)
	assert.Nil(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, adminId, user.ID)
	assert.Equal(t, "1", user.StripeCustomerID)

	normalId := uuid.MustParse(MOCK_NORMAL_UUID)
	// Get user
	user, err = MockRepo.GetUser(normalId)
	assert.Nil(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, normalId, user.ID)
	assert.Equal(t, "2", user.StripeCustomerID)

	// Non-existent
	user, err = MockRepo.GetUser(uuid.New())
	assert.Nil(t, err)
	assert.Nil(t, user)
}

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

func TestGetUserByStripeCustomerId(t *testing.T) {
	// Get user
	user, err := MockRepo.GetUserByStripeCustomerId("1")
	assert.Nil(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, uuid.MustParse(MOCK_ADMIN_UUID), user.ID)
	assert.Equal(t, "1", user.StripeCustomerID)

	// Non-existent
	user, err = MockRepo.GetUserByStripeCustomerId("3123213123")
	assert.Nil(t, err)
	assert.Nil(t, user)
}

func TestIsSuperAdmin(t *testing.T) {
	// Super admin
	adminId := uuid.MustParse(MOCK_ADMIN_UUID)
	isSuperAdmin, err := MockRepo.IsSuperAdmin(adminId)
	assert.Nil(t, err)
	assert.True(t, isSuperAdmin)

	// Normal user
	normalId := uuid.MustParse(MOCK_NORMAL_UUID)
	isSuperAdmin, err = MockRepo.IsSuperAdmin(normalId)
	assert.Nil(t, err)
	assert.False(t, isSuperAdmin)

	// Non-existent
	isSuperAdmin, err = MockRepo.IsSuperAdmin(uuid.New())
	assert.Nil(t, err)
	assert.False(t, isSuperAdmin)
}

func TestGetSuperAdminUserIDs(t *testing.T) {
	ids, err := MockRepo.GetSuperAdminUserIDs()
	assert.Nil(t, err)
	assert.Len(t, ids, 1)
	assert.Equal(t, uuid.MustParse(MOCK_ADMIN_UUID), ids[0])
}

func TestGetRoles(t *testing.T) {
	// Get roles
	roles, err := MockRepo.GetRoles(uuid.MustParse(MOCK_ADMIN_UUID))
	assert.Nil(t, err)
	assert.Len(t, roles, 1)

	// No roles
	roles, err = MockRepo.GetRoles(uuid.MustParse(MOCK_NORMAL_UUID))
	assert.Nil(t, err)
	assert.Len(t, roles, 0)
}

func TestQueryUsersCount(t *testing.T) {
	// Query users
	c, cMap, err := MockRepo.QueryUsersCount("")
	assert.Nil(t, err)
	assert.Equal(t, 4, c)
	assert.Len(t, cMap, 1)
	assert.Equal(t, 3, cMap["prod_123"])
}

func TestQueryUsers(t *testing.T) {
	// Query users
	users, err := MockRepo.QueryUsers("mockadmin", 50, nil, []string{}, nil)
	assert.Nil(t, err)
	assert.Equal(t, 1, *users.Total)
	assert.Nil(t, users.Next)
	assert.Equal(t, users.Users[0].ID, uuid.MustParse(MOCK_ADMIN_UUID))

	// Query users
	users, err = MockRepo.QueryUsers("", 1, nil, []string{}, nil)
	assert.Nil(t, err)
	assert.Equal(t, 4, *users.Total)
	assert.NotNil(t, users.Next)
}

func TestGetUsersThatSignedInSince(t *testing.T) {
	// Get users
	users, err := MockRepo.GetUsersThatSignedInSince(1 * time.Hour)
	assert.Nil(t, err)
	assert.Len(t, users, 1)
}

func TestGetNSubscribers(t *testing.T) {
	// Get users
	users, err := MockRepo.GetNSubscribers()
	assert.Nil(t, err)
	assert.Equal(t, 3, users)
}
