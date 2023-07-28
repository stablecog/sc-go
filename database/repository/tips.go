package repository

import (
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/ent/credit"
	"github.com/stablecog/sc-go/database/ent/tiplog"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/server/responses"
)

// Handle tips from one user to another
func (r *Repository) TipCreditsToUser(fromUser uuid.UUID, toUser *uuid.UUID, toUserDiscordId string, amount int32) (success bool, err error) {
	// Start transaction
	if err := r.WithTx(func(tx *ent.Tx) error {
		db := tx.Client()

		// Deduct credits from user
		success, err = r.DeductCreditsFromUser(fromUser, amount, true, db)
		if err != nil {
			return err
		}
		if !success {
			return responses.InsufficientCreditsErr
		}

		// Add credits to user, tipped type
		tippedCreditType, err := r.GetOrCreateTippedCreditType(db)
		if err != nil {
			return err
		}

		// Add credits to user (only if registered)
		if toUser != nil {
			// See if they have the tipped type already
			tippedCredits, err := db.Credit.Query().Where(credit.UserIDEQ(*toUser), credit.CreditTypeIDEQ(tippedCreditType.ID), credit.ExpiresAtEQ(NEVER_EXPIRE)).First(r.Ctx)
			if err != nil && !ent.IsNotFound(err) {
				return err
			}
			if err != nil && ent.IsNotFound(err) {
				// Create credit
				_, err := db.Credit.Create().SetCreditTypeID(tippedCreditType.ID).SetUserID(*toUser).SetRemainingAmount(amount).SetExpiresAt(NEVER_EXPIRE).Save(r.Ctx)
				if err != nil {
					return err
				}
			} else {
				_, err = db.Credit.Update().AddRemainingAmount(amount).Where(credit.IDEQ(tippedCredits.ID)).Save(r.Ctx)
				if err != nil {
					return err
				}
			}
		}

		// Log this tip
		tipLogM := db.TipLog.Create()
		if toUser != nil {
			tipLogM.SetTippedTo(*toUser)
		}
		_, err = tipLogM.SetAmount(amount).SetTippedBy(fromUser).SetTippedToDiscordID(toUserDiscordId).Save(r.Ctx)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		log.Errorf("Error in tip transaction %v", err)
		return false, err
	}
	return true, nil
}

// Collect unclaimed tips
func (r *Repository) CollectUnclaimedTips(toUserId uuid.UUID, toUserDiscordId string, DB *ent.Client) (claimted int32, err error) {
	if DB == nil {
		DB = r.DB
	}

	// Get all unclaimed tips
	tips, err := DB.TipLog.Query().Where(tiplog.TippedToDiscordIDEQ(toUserDiscordId), tiplog.TippedToIsNil()).All(r.Ctx)
	if err != nil {
		return 0, err
	}
	var total int32
	for _, tip := range tips {
		total += tip.Amount
	}

	// Update all tips to be claimed
	_, err = DB.TipLog.Update().Where(tiplog.TippedToDiscordIDEQ(toUserDiscordId), tiplog.TippedToIsNil()).SetTippedTo(toUserId).Save(r.Ctx)
	if err != nil {
		return 0, err
	}

	// Add credits to user, tipped type
	tippedCreditType, err := r.GetOrCreateTippedCreditType(DB)
	if err != nil {
		return 0, err
	}

	// See if they have the tipped type already
	tippedCredits, err := DB.Credit.Query().Where(credit.UserIDEQ(toUserId), credit.CreditTypeIDEQ(tippedCreditType.ID), credit.ExpiresAtEQ(NEVER_EXPIRE)).First(r.Ctx)
	if err != nil && !ent.IsNotFound(err) {
		return 0, err
	}
	if err != nil && ent.IsNotFound(err) {
		// Create credit
		_, err := DB.Credit.Create().SetCreditTypeID(tippedCreditType.ID).SetUserID(toUserId).SetRemainingAmount(total).SetExpiresAt(NEVER_EXPIRE).Save(r.Ctx)
		if err != nil {
			return 0, err
		}
	} else {
		_, err = DB.Credit.Update().AddRemainingAmount(total).Where(credit.IDEQ(tippedCredits.ID)).Save(r.Ctx)
		if err != nil {
			return 0, err
		}
	}
	return total, nil
}
