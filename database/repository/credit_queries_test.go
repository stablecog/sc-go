package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent/credit"
	"github.com/stablecog/sc-go/database/ent/credittype"
	"github.com/stablecog/sc-go/shared"
	"github.com/stretchr/testify/assert"
)

func TestCountPaidCreditsForUser(t *testing.T) {
	sum, err := MockRepo.GetNonFreeCreditSum(uuid.MustParse(MOCK_ADMIN_UUID))
	assert.Nil(t, err)
	assert.Equal(t, 100, sum)
}

func TestCreditsForUser(t *testing.T) {
	// Create more credits
	creditType := MockRepo.DB.CreditType.Query().Where(credittype.IDNEQ(uuid.MustParse(TIPPABLE_CREDIT_TYPE_ID))).FirstX(MockRepo.Ctx)
	_, err := MockRepo.DB.Credit.Create().SetCreditTypeID(creditType.ID).SetUserID(uuid.MustParse(MOCK_ADMIN_UUID)).SetRemainingAmount(1234).SetExpiresAt(time.Now().AddDate(1000, 0, 0)).Save(MockRepo.Ctx)
	assert.Nil(t, err)

	// Get credits
	credits, err := MockRepo.GetCreditsForUser(uuid.MustParse(MOCK_ADMIN_UUID))
	assert.Nil(t, err)
	assert.Equal(t, 2, len(credits))
	assert.Equal(t, int32(100), credits[0].RemainingAmount)
	assert.Equal(t, int32(1234), credits[1].RemainingAmount)
}

func TestGetOrCreateFreeCreditType(t *testing.T) {
	ctype, err := MockRepo.GetOrCreateFreeCreditType(nil)
	assert.Nil(t, err)
	assert.Equal(t, "Free", ctype.Name)
}

func TestGetNonExpiredCreditTotalForUser(t *testing.T) {
	u := uuid.MustParse(MOCK_ALT_UUID)
	total, err := MockRepo.GetNonExpiredCreditTotalForUser(u, nil)
	assert.Nil(t, err)
	assert.Equal(t, 1334, total)
}

func TestGetFreeCreditReplenishesAtForUser(t *testing.T) {
	// Mock Now function
	orgNow := Now
	defer func() { Now = orgNow }()
	now := time.Date(2020, 1, 1, 12, 0, 0, 0, time.UTC)
	Now = func() time.Time {
		return now
	}
	// Setup
	u, err := MockRepo.CreateUser(uuid.New(), "TestGetFreeCreditReplenishesAtForUser@stablecog.com", "cus_1234", nil, nil)
	assert.Nil(t, err)
	assert.NotNil(t, u)
	assert.Nil(t, u.ActiveProductID)
	_, err = MockRepo.GiveFreeCredits(u.ID, nil)
	assert.Nil(t, err)
	// Get credits
	ctype, err := MockRepo.GetOrCreateFreeCreditType(nil)
	assert.Nil(t, err)
	credit := MockRepo.DB.Credit.Query().Where(credit.UserID(u.ID), credit.CreditTypeID(ctype.ID)).OnlyX(context.Background())
	assert.NotNil(t, credit)
	// Credits replnished ~6 hours ago
	replenishedAt := time.Date(2020, 1, 1, 6, 0, 0, 0, time.UTC)
	MockRepo.DB.Credit.UpdateOne(credit).SetReplenishedAt(replenishedAt).SetRemainingAmount(50).ExecX(context.Background())

	// Get free credit replenishes at
	replenishesAt, c, ct, err := MockRepo.GetFreeCreditReplenishesAtForUser(u.ID)
	assert.Nil(t, err)
	assert.NotNil(t, c)
	assert.NotNil(t, ct)
	assert.Equal(t, replenishedAt.Add(shared.FREE_CREDIT_REPLENISHMENT_INTERVAL), *replenishesAt)

	// Cleanup
	MockRepo.DB.Credit.DeleteOne(credit).ExecX(context.Background())
	MockRepo.DB.User.DeleteOne(u).ExecX(context.Background())
}
