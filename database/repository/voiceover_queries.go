package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/ent/voiceover"
)

func (r *Repository) GetAllVoiceoverModels() ([]*ent.VoiceoverModel, error) {
	models, err := r.DB.VoiceoverModel.Query().All(r.Ctx)
	if err != nil {
		return nil, err
	}

	return models, nil
}

func (r *Repository) GetAllVoiceoverSpeakers() ([]*ent.VoiceoverSpeaker, error) {
	speakers, err := r.DB.VoiceoverSpeaker.Query().All(r.Ctx)
	if err != nil {
		return nil, err
	}

	return speakers, nil
}

func (r *Repository) GetVoiceover(id uuid.UUID) (*ent.Voiceover, error) {
	return r.DB.Voiceover.Query().Where(voiceover.ID(id)).Only(r.Ctx)
}

type VoiceoverOutput struct {
	ID               uuid.UUID  `json:"id"`
	AudoFileURL      string     `json:"audio_file_url"`
	CreatedAt        *time.Time `json:"created_at,omitempty"`
	IsFavorited      bool       `json:"is_favorited"`
	WasAutoSubmitted bool       `json:"was_auto_submitted"`
}
