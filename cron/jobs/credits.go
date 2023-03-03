package jobs

import (
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/shared"
)

const FREE_JOB_NAME = "FREE_CREDITS_JOB"

func (j *JobRunner) AddFreeCreditsToEligibleUsers(log Logger) error {
	log.Infof("Running free credit job...")
	users, err := j.Repo.GetUsersThatSignedInSince(shared.FREE_CREDIT_LAST_ACTIVITY_REQUIREMENT)
	if err != nil {
		log.Errorf("Error getting users eligible for free credits %v", err)
		return err
	}

	if len(users) == 0 {
		log.Infof("No users eligible for free credits")
		return nil
	}

	// Get the uuids as array
	uuids := make([]uuid.UUID, len(users))
	for i, user := range users {
		uuids[i] = user.ID
	}

	// Replenish credits
	count, err := j.Repo.ReplenishFreeCreditsToEligibleUsers(uuids)
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
