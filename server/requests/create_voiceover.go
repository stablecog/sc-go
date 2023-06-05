package requests

import (
	"errors"
	"math"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/stablecog/sc-go/shared"
	"github.com/stablecog/sc-go/utils"
)

type CreateVoiceoverRequest struct {
	Prompt    string     `json:"prompt"`
	ModelId   *uuid.UUID `json:"model_id,omitempty"`
	SpeakerId *uuid.UUID `json:"speaker_id,omitempty"`
	Seed      *int       `json:"seed,omitempty"`
	Temp      *float32   `json:"temp,omitempty"`
	StreamID  string     `json:"stream_id"`
	UIId      string     `json:"ui_id"` // Corresponds to UI identifier
}

func (t *CreateVoiceoverRequest) ApplyDefaults() {
	if t.ModelId == nil {
		t.ModelId = utils.ToPtr(shared.GetCache().GetDefaultVoiceoverModel().ID)
	}

	if t.SpeakerId == nil {
		t.SpeakerId = utils.ToPtr(shared.GetCache().GetDefaultVoiceoverSpeaker().ID)
	}

	if t.Seed == nil || *t.Seed < 0 {
		rand.Seed(time.Now().Unix())
		t.Seed = utils.ToPtr(rand.Intn(math.MaxInt32))
	}

	if t.Temp == nil {
		t.Temp = utils.ToPtr[float32](0.7)
	}

}

func (t *CreateVoiceoverRequest) Validate(api bool) error {
	if !api && !utils.IsSha256Hash(t.StreamID) {
		return errors.New("invalid_stream_id")
	}

	// Apply default settings
	t.ApplyDefaults()

	if !shared.GetCache().IsValidVoiceoverModelID(*t.ModelId) {
		return errors.New("invalid_model_id")
	}

	if !shared.GetCache().IsValidVoiceoverSpeakerID(*t.SpeakerId, *t.ModelId) {
		return errors.New("invalid_speaker_id")
	}

	if *t.Temp < 0.0 || *t.Temp > 1.0 {
		return errors.New("invalid_temp")
	}

	if len(t.Prompt) > shared.VOICEOVER_MAX_TEXT_LENGTH {
		return errors.New("prompt_too_long")
	}

	return nil
}
