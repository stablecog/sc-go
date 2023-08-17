package requests

import (
	"fmt"
	"math"
	"math/rand"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/stablecog/sc-go/shared"
	"github.com/stablecog/sc-go/utils"
)

type CreateVoiceoverRequest struct {
	Prompt           string     `json:"prompt"`
	ModelId          *uuid.UUID `json:"model_id,omitempty"`
	SpeakerId        *uuid.UUID `json:"speaker_id,omitempty"`
	Seed             *int       `json:"seed,omitempty"`
	Temperature      *float32   `json:"temperature,omitempty"`
	DenoiseAudio     *bool      `json:"denoise_audio,omitempty"`
	RemoveSilence    *bool      `json:"remove_silence,omitempty"`
	StreamID         string     `json:"stream_id"`
	SubmitToGallery  bool       `json:"submit_to_gallery"`
	UIId             string     `json:"ui_id"` // Corresponds to UI identifier
	WasAutoSubmitted bool
}

func (t *CreateVoiceoverRequest) Cost() int32 {
	return utils.CalculateVoiceoverCredits(t.Prompt)
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

	if t.Temperature == nil {
		t.Temperature = utils.ToPtr[float32](shared.DEFAULT_VOICEOVER_TEMPERATURE)
	}

	if t.DenoiseAudio == nil {
		t.DenoiseAudio = utils.ToPtr[bool](shared.DEFAULT_VOICEOVER_DENOISE_AUDIO)
	}

	if t.RemoveSilence == nil {
		t.RemoveSilence = utils.ToPtr[bool](shared.DEFAULT_VOICEOVER_REMOVE_SILENCE)
	}
}

func (t *CreateVoiceoverRequest) Validate(api bool) error {
	if !api && !utils.IsSha256Hash(t.StreamID) {
		return fmt.Errorf("invalid_stream_id")
	}

	// Apply default settings
	t.ApplyDefaults()

	if !shared.GetCache().IsValidVoiceoverModelID(*t.ModelId) {
		return fmt.Errorf("invalid_model_id")
	}

	if !shared.GetCache().IsValidVoiceoverSpeakerID(*t.SpeakerId, *t.ModelId) {
		return fmt.Errorf("invalid_speaker_id")
	}

	if *t.Temperature < 0.0 || *t.Temperature > 1.0 {
		return fmt.Errorf("invalid_temp")
	}

	if utf8.RuneCountInString(t.Prompt) > shared.VOICEOVER_MAX_TEXT_LENGTH {
		return fmt.Errorf("prompt_too_long")
	}

	return nil
}
