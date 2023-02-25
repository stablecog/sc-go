package repository

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/ent/credit"
	"github.com/stretchr/testify/assert"
)

func TestCreditsForUser(t *testing.T) {
	// Create more credits
	creditType := MockRepo.DB.CreditType.Query().FirstX(MockRepo.Ctx)
	_, err := MockRepo.DB.Credit.Create().SetCreditTypeID(creditType.ID).SetUserID(uuid.MustParse(MOCK_ADMIN_UUID)).SetRemainingAmount(1234).SetExpiresAt(time.Now().AddDate(1000, 0, 0)).Save(MockRepo.Ctx)
	assert.Nil(t, err)

	// Get credits
	credits, err := MockRepo.GetCreditsForUser(uuid.MustParse(MOCK_ADMIN_UUID))
	assert.Nil(t, err)
	assert.Equal(t, 2, len(credits))
	assert.Equal(t, int32(100), credits[0].RemainingAmount)
	assert.Equal(t, int32(1234), credits[1].RemainingAmount)
}

func TestDeductCreditsFromUser(t *testing.T) {
	// User should have 100 credits
	success, err := MockRepo.DeductCreditsFromUser(uuid.MustParse(MOCK_NORMAL_UUID), 50, nil)
	assert.Nil(t, err)
	assert.Equal(t, true, success)

	// User should have 50 credits
	userCredit := MockRepo.DB.Credit.Query().Where(credit.UserIDEQ(uuid.MustParse(MOCK_NORMAL_UUID))).FirstX(MockRepo.Ctx)
	assert.Equal(t, int32(50), userCredit.RemainingAmount)

	// Create another row of non-expiring credits for user
	creditType := MockRepo.DB.CreditType.Query().FirstX(MockRepo.Ctx)
	_, err = MockRepo.DB.Credit.Create().SetCreditTypeID(creditType.ID).SetUserID(uuid.MustParse(MOCK_NORMAL_UUID)).SetRemainingAmount(1234).SetExpiresAt(time.Now().AddDate(1000, 0, 0)).Save(MockRepo.Ctx)
	assert.Nil(t, err)

	// Deduct 50 credits
	success, err = MockRepo.DeductCreditsFromUser(uuid.MustParse(MOCK_NORMAL_UUID), 50, nil)
	assert.Nil(t, err)
	assert.Equal(t, true, success)

	// Ensure most recent expiritng credits were used
	credits := MockRepo.DB.Credit.Query().Where(credit.UserIDEQ(uuid.MustParse(MOCK_NORMAL_UUID))).Order(ent.Desc(credit.FieldExpiresAt)).AllX(MockRepo.Ctx)
	assert.Equal(t, int32(1234), credits[0].RemainingAmount)
	assert.Equal(t, int32(0), credits[1].RemainingAmount)

	// Take again to ensure we get the next expiring credits
	success, err = MockRepo.DeductCreditsFromUser(uuid.MustParse(MOCK_NORMAL_UUID), 50, nil)
	assert.Nil(t, err)
	assert.Equal(t, true, success)

	// Ensure most recent expiritng credits were used
	credits = MockRepo.DB.Credit.Query().Where(credit.UserIDEQ(uuid.MustParse(MOCK_NORMAL_UUID))).Order(ent.Desc(credit.FieldExpiresAt)).AllX(MockRepo.Ctx)
	assert.Equal(t, int32(1234-50), credits[0].RemainingAmount)
	assert.Equal(t, int32(0), credits[1].RemainingAmount)

	// Expire all credits and make sure we get 0 for deduction
	_, err = MockRepo.DB.Credit.Delete().Where(credit.UserIDEQ(uuid.MustParse(MOCK_NORMAL_UUID))).Exec(MockRepo.Ctx)
	assert.Nil(t, err)
	// Create new credits that are expired
	_, err = MockRepo.DB.Credit.Create().SetCreditTypeID(creditType.ID).SetUserID(uuid.MustParse(MOCK_NORMAL_UUID)).SetRemainingAmount(1234).SetExpiresAt(time.Now().AddDate(-1, 0, 0)).Save(MockRepo.Ctx)

	// Try to deduct
	success, err = MockRepo.DeductCreditsFromUser(uuid.MustParse(MOCK_NORMAL_UUID), 50, nil)
	assert.Nil(t, err)
	assert.Equal(t, false, success)
}

func TestRefundCreditsToUser(t *testing.T) {
	// Delete all credits
	_, err := MockRepo.DB.Credit.Delete().Where(credit.UserIDEQ(uuid.MustParse(MOCK_NORMAL_UUID))).Exec(MockRepo.Ctx)
	assert.Nil(t, err)
	// Create new credits
	creditType := MockRepo.DB.CreditType.Query().FirstX(MockRepo.Ctx)
	_, err = MockRepo.DB.Credit.Create().SetCreditTypeID(creditType.ID).SetUserID(uuid.MustParse(MOCK_NORMAL_UUID)).SetRemainingAmount(1234).SetExpiresAt(time.Now().AddDate(1, 0, 0)).Save(MockRepo.Ctx)

	// Deduct
	success, err := MockRepo.DeductCreditsFromUser(uuid.MustParse(MOCK_NORMAL_UUID), 50, nil)
	assert.Nil(t, err)
	assert.Equal(t, true, success)

	credits := MockRepo.DB.Credit.Query().Where(credit.UserIDEQ(uuid.MustParse(MOCK_NORMAL_UUID))).Order(ent.Desc(credit.FieldExpiresAt)).FirstX(MockRepo.Ctx)
	assert.Equal(t, int32(1234-50), credits.RemainingAmount)

	// Refund
	success, err = MockRepo.RefundCreditsToUser(uuid.MustParse(MOCK_NORMAL_UUID), 50, nil)
	assert.Nil(t, err)
	assert.Equal(t, true, success)

	credits = MockRepo.DB.Credit.Query().Where(credit.UserIDEQ(uuid.MustParse(MOCK_NORMAL_UUID))).Order(ent.Desc(credit.FieldExpiresAt)).FirstX(MockRepo.Ctx)
	assert.Equal(t, int32(1234), credits.RemainingAmount)
}

func TestGetFreeCreditType(t *testing.T) {
	ctype, err := MockRepo.GetFreeCreditType()
	assert.Nil(t, err)
	assert.Equal(t, "Free", ctype.Name)
}
