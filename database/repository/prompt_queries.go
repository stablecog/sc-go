package repository

import (
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent/generation"
)

// Returns UUIDs of prompts that are unique to this user
func (r *Repository) GetUsersUniquePromptIds(promptIds []uuid.UUID, userId uuid.UUID) ([]uuid.UUID, error) {
	generations, err := r.DB.Generation.Query().Where(
		generation.PromptIDIn(promptIds...),
		generation.UserIDNEQ(userId),
	).All(r.Ctx)
	if err != nil {
		return nil, err
	}

	usedPromptIds := make(map[uuid.UUID]bool)
	for _, g := range generations {
		if g.PromptID != nil {
			usedPromptIds[*g.PromptID] = true
		}
	}

	uniquePromptIds := make([]uuid.UUID, 0)
	for _, promptId := range promptIds {
		if _, ok := usedPromptIds[promptId]; !ok {
			uniquePromptIds = append(uniquePromptIds, promptId)
		}
	}
	return uniquePromptIds, nil
}
