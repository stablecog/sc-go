package jobs

import (
	"github.com/charmbracelet/log"
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/shared"
)

func (j *JobRunner) AddFreeCreditsToEligibleUsers() error {
	users, err := j.Repo.GetUsersThatSignedInSince(shared.FREE_CREDIT_LAST_ACTIVITY_REQUIREMENT)
	if err != nil {
		log.Error("Error getting users eligible for free credits", "err", err)
		return err
	}

	if len(users) == 0 {
		log.Info("No users eligible for free credits")
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
		log.Error("Error replenishing free credits to eligible users", "err", err)
		return err
	}

	log.Info("Replenished free credits to eligible users", "user_count", count, "amount", shared.FREE_CREDIT_AMOUNT_DAILY)
	return nil
}
