package repository

import (
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent/generation"
)

// Returns UUIDs of prompts that are unique to this user
func (r *Repository) GetUsersUniqueNegativePromptIds(negativePromptIds []uuid.UUID, userId uuid.UUID) ([]uuid.UUID, error) {
	generations, err := r.DB.Generation.Query().Where(
		generation.NegativePromptIDIn(negativePromptIds...),
		generation.UserIDNEQ(userId),
	).All(r.Ctx)
	if err != nil {
		return nil, err
	}

	usedPromptIds := make(map[uuid.UUID]bool)
	for _, g := range generations {
		if g.NegativePromptID != nil {
			usedPromptIds[*g.NegativePromptID] = true
		}
	}

	uniquePromptIds := make([]uuid.UUID, 0)
	for _, promptId := range negativePromptIds {
		if _, ok := usedPromptIds[promptId]; !ok {
			uniquePromptIds = append(uniquePromptIds, promptId)
		}
	}
	return uniquePromptIds, nil
}
