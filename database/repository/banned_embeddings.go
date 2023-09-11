package repository

import (
	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/log"
)

type MatchBannedPrompts struct {
	ID         uuid.UUID `json:"id" sql:"id"`
	Prompt     string    `json:"prompt" sql:"prompt"`
	Similarity float32   `json:"similarity" sql:"similarity"`
}

func (r *Repository) IsBannedPromptEmbedding(embedding []float32, DB *ent.Client) ([]MatchBannedPrompts, error) {
	if DB == nil {
		DB = r.DB
	}

	rows, err := r.DB.QueryContext(r.Ctx, "SELECT * from match_banned_prompts($1, 0.61, 1)", pgvector.NewVector(embedding))
	if err != nil {
		log.Errorf("Error querying for banned prompt embeddings: %v", err)
	}

	var bannedPrompts []MatchBannedPrompts
	for rows.Next() {
		var bannedPrompt MatchBannedPrompts
		err = rows.Scan(&bannedPrompt.ID, &bannedPrompt.Prompt, &bannedPrompt.Similarity)
		if err != nil {
			log.Errorf("Error scanning banned prompt row: %v", err)
		}
		bannedPrompts = append(bannedPrompts, bannedPrompt)
	}

	return bannedPrompts, nil
}
