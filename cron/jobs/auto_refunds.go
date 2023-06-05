package jobs

import (
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/shared"
)

func (j *JobRunner) RefundOldGenerationCredits(log Logger) error {
	log.Infof("Starting refund stale upscale/generation credits job...")
	gens, err := j.Repo.GetGenerationsQueuedOrStarted()
	if err != nil {
		log.Errorf("Error getting queued/started generations %v", err)
		return err
	}
	refunded := 0
	refundedGens := 0
	refundedUpscales := 0
	for _, gen := range gens {
		if err := j.Repo.WithTx(func(tx *ent.Tx) error {
			db := tx.Client()
			j.Repo.SetGenerationFailed(gen.ID.String(), shared.TIMEOUT_ERROR, 0, db)
			// Upscale is always 1 credit
			_, err := j.Repo.RefundCreditsToUser(gen.UserID, gen.NumOutputs, db)
			if err != nil {
				log.Errorf("Error refunding credits for generation %s %s %v", gen.UserID.String(), gen.ID.String(), err)
				return err
			}
			return nil
		}); err != nil {
			return nil
		}
		refunded += int(gen.NumOutputs)
		refundedGens++
	}

	upscales, err := j.Repo.GetUpscalesQueuedOrStarted()
	if err != nil {
		log.Errorf("Error getting queued/started upscales %v", err)
		return err
	}
	for _, us := range upscales {
		if err := j.Repo.WithTx(func(tx *ent.Tx) error {
			db := tx.Client()
			j.Repo.SetUpscaleFailed(us.ID.String(), shared.TIMEOUT_ERROR, db)
			// Upscale is always 1 credit
			_, err := j.Repo.RefundCreditsToUser(us.UserID, 1, db)
			if err != nil {
				log.Errorf("Error refunding credits for upscale %s %s %v", us.UserID.String(), us.ID.String(), err)
				return err
			}
			return nil
		}); err != nil {
			return nil
		}
		refunded += 1
		refundedUpscales++
	}

	log.Infof("Refunded %d credits for %d generations and %d upscales", refunded, refundedGens, refundedUpscales)

	return nil
}
