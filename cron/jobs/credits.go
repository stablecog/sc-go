package jobs

import (
	"github.com/stablecog/sc-go/shared"
)

const FREE_JOB_NAME = "FREE_CREDITS_JOB"

func (j *JobRunner) AddFreeCreditsToEligibleUsers(log Logger) error {
	log.Infof("Running free credit job...")
	// Replenish credits
	count, err := j.Repo.ReplenishFreeCreditsToEligibleUsers()
	if err != nil {
		log.Errorf("Error replenishing free credits to eligible users %v", err)
		return err
	}

	if count == 0 {
		log.Infof("No users eligible for free credits")
		return nil
	}

	log.Infof("Added %d credits to %d users", shared.FREE_CREDIT_AMOUNT_DAILY, count)
	return nil
}
