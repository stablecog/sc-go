package requests

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/stablecog/sc-go/shared"
	"github.com/stablecog/sc-go/utils"
)

// HTTP Request for creating a new generation
type CreateGenerationRequest struct {
	Prompt               string                `json:"prompt"`
	NegativePrompt       string                `json:"negative_prompt,omitempty"`
	Width                int32                 `json:"width"`
	Height               int32                 `json:"height"`
	InferenceSteps       int32                 `json:"inference_steps"`
	GuidanceScale        float32               `json:"guidance_scale"`
	ModelId              uuid.UUID             `json:"model_id"`
	SchedulerId          uuid.UUID             `json:"scheduler_id"`
	Seed                 int                   `json:"seed"`
	NumOutputs           int32                 `json:"num_outputs,omitempty"`
	StreamID             string                `json:"stream_id"` // Corresponds to SSE stream
	SubmitToGallery      bool                  `json:"submit_to_gallery"`
	ProcessType          shared.ProcessType    `json:"process_type"`
	OutputImageExtension shared.ImageExtension `json:"output_image_extension"`
}

func (t *CreateGenerationRequest) Validate() error {
	if !utils.IsSha256Hash(t.StreamID) {
		return errors.New("invalid_stream_id")
	}

	if t.Height > shared.MAX_GENERATE_HEIGHT {
		return fmt.Errorf("Height is too large, max is: %d", shared.MAX_GENERATE_HEIGHT)
	}

	if t.Width > shared.MAX_GENERATE_WIDTH {
		return fmt.Errorf("Width is too large, max is: %d", shared.MAX_GENERATE_WIDTH)
	}

	if t.Width*t.Height*t.InferenceSteps >= shared.MAX_PRO_PIXEL_STEPS {
		return fmt.Errorf("Pick fewer inference steps or smaller dimensions: %d - %d - %d",
			t.Width,
			t.Height,
			t.InferenceSteps,
		)
	}

	if t.NumOutputs < 0 {
		t.NumOutputs = shared.DEFAULT_GENERATE_NUM_OUTPUTS
	}
	if t.NumOutputs > shared.MAX_GENERATE_NUM_OUTPUTS {
		return fmt.Errorf("Number of outputs can't be more than %d", shared.MAX_GENERATE_NUM_OUTPUTS)
	}

	if !shared.GetCache().IsValidGenerationModelID(t.ModelId) {
		return errors.New("invalid_model_id")
	}

	if !shared.GetCache().IsValidShedulerID(t.SchedulerId) {
		return errors.New("invalid_scheduler_id")
	}

	if t.Seed < 0 {
		rand.Seed(time.Now().Unix())
		t.Seed = rand.Intn(math.MaxInt32)
	}

	return nil
}
