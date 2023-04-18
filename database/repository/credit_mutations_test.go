package repository

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/ent/credit"
	"github.com/stablecog/sc-go/database/ent/credittype"
	"github.com/stablecog/sc-go/database/ent/user"
	"github.com/stretchr/testify/assert"
)

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

func TestDeductCreditsFromUserMultipart(t *testing.T) {
	// Test with 1+1+1 credits deducting 3, they should have 0 left
	// Create a mock user
	u, err := MockRepo.CreateUser(uuid.MustParse("c66b47be-aa0b-4840-965b-f19c8847904e"), "testmultipart@stablecog.com", "1234", nil, nil)
	assert.Nil(t, err)
	// Create a few credit types
	c1, err := MockRepo.CreateCreditType("TestDeductCreditsFromUserMultipart1", 100, nil, nil, credittype.TypeOneTime)
	assert.Nil(t, err)
	c2, err := MockRepo.CreateCreditType("TestDeductCreditsFromUserMultipart2", 100, nil, nil, credittype.TypeOneTime)
	assert.Nil(t, err)
	c3, err := MockRepo.CreateCreditType("TestDeductCreditsFromUserMultipart3", 100, nil, nil, credittype.TypeOneTime)
	assert.Nil(t, err)

	// Give user 1 credit of each type
	_, err = MockRepo.DB.Credit.Create().SetCreditTypeID(c1.ID).SetUserID(u.ID).SetRemainingAmount(1).SetExpiresAt(time.Now().Add(30 * time.Minute)).Save(MockRepo.Ctx)
	assert.Nil(t, err)
	_, err = MockRepo.DB.Credit.Create().SetCreditTypeID(c2.ID).SetUserID(u.ID).SetRemainingAmount(1).SetExpiresAt(time.Now().Add(30 * time.Minute)).Save(MockRepo.Ctx)
	assert.Nil(t, err)
	_, err = MockRepo.DB.Credit.Create().SetCreditTypeID(c3.ID).SetUserID(u.ID).SetRemainingAmount(1).SetExpiresAt(time.Now().Add(30 * time.Minute)).Save(MockRepo.Ctx)
	assert.Nil(t, err)

	// Test deduct too many
	success, err := MockRepo.DeductCreditsFromUser(uuid.MustParse("c66b47be-aa0b-4840-965b-f19c8847904e"), 4, nil)
	assert.Nil(t, err)
	assert.Equal(t, false, success)

	// user should have 3 credits
	userCredit, err := MockRepo.GetNonExpiredCreditTotalForUser(u.ID, nil)
	assert.Nil(t, err)
	assert.Equal(t, 3, userCredit)

	// Test deduct
	success, err = MockRepo.DeductCreditsFromUser(uuid.MustParse("c66b47be-aa0b-4840-965b-f19c8847904e"), 3, nil)
	assert.Nil(t, err)
	assert.Equal(t, true, success)

	// User should have 0 credits
	userCredit, err = MockRepo.GetNonExpiredCreditTotalForUser(u.ID, nil)
	assert.Nil(t, err)
	assert.Equal(t, 0, userCredit)

	// Cleanup
	_, err = MockRepo.DB.Credit.Delete().Where(credit.UserIDEQ(u.ID)).Exec(MockRepo.Ctx)
	assert.Nil(t, err)
	_, err = MockRepo.DB.CreditType.Delete().Where(credittype.IDEQ(c1.ID)).Exec(MockRepo.Ctx)
	assert.Nil(t, err)
	_, err = MockRepo.DB.CreditType.Delete().Where(credittype.IDEQ(c2.ID)).Exec(MockRepo.Ctx)
	assert.Nil(t, err)
	_, err = MockRepo.DB.CreditType.Delete().Where(credittype.IDEQ(c3.ID)).Exec(MockRepo.Ctx)
	assert.Nil(t, err)
	_, err = MockRepo.DB.User.Delete().Where(user.IDEQ(u.ID)).Exec(MockRepo.Ctx)
	assert.Nil(t, err)
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

	creditsStarted := MockRepo.DB.Credit.Query().Where(credit.UserIDEQ(uuid.MustParse(MOCK_NORMAL_UUID))).Order(ent.Desc(credit.FieldExpiresAt)).FirstX(MockRepo.Ctx)
	assert.Equal(t, int32(1234-50), creditsStarted.RemainingAmount)

	// Refund
	success, err = MockRepo.RefundCreditsToUser(uuid.MustParse(MOCK_NORMAL_UUID), 50, nil)
	assert.Nil(t, err)
	assert.Equal(t, true, success)

	creditsStarted = MockRepo.DB.Credit.Query().Where(credit.IDEQ(creditsStarted.ID)).FirstX(MockRepo.Ctx)
	assert.Equal(t, int32(1234-50), creditsStarted.RemainingAmount)

	// Get refund credit type
	refundCreditType, err := MockRepo.GetOrCreateRefundCreditType(nil)
	assert.Nil(t, err)
	// Get credits of this type for user
	creditsRefund := MockRepo.DB.Credit.Query().Where(credit.UserIDEQ(uuid.MustParse(MOCK_NORMAL_UUID))).Where(credit.CreditTypeIDEQ(refundCreditType.ID)).FirstX(MockRepo.Ctx)
	assert.Equal(t, int32(50), creditsRefund.RemainingAmount)

	// Cleanup
	MockRepo.DB.Credit.Delete().Where(credit.IDEQ(creditsRefund.ID)).ExecX(MockRepo.Ctx)
	MockRepo.DB.Credit.UpdateOne(creditsStarted).SetRemainingAmount(1234).ExecX(MockRepo.Ctx)
}
