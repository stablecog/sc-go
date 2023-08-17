package requests

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/stablecog/sc-go/shared"
	"github.com/stablecog/sc-go/utils"
	"golang.org/x/exp/slices"
)

// HTTP Request for creating a new generation
type CreateGenerationRequest struct {
	Prompt               string                `json:"prompt"`
	NegativePrompt       string                `json:"negative_prompt,omitempty"`
	Width                *int32                `json:"width,omitempty"`
	Height               *int32                `json:"height,omitempty"`
	InferenceSteps       *int32                `json:"inference_steps,omitempty"`
	GuidanceScale        *float32              `json:"guidance_scale,omitempty"`
	ModelId              *uuid.UUID            `json:"model_id,omitempty"`
	SchedulerId          *uuid.UUID            `json:"scheduler_id,omitempty"`
	Seed                 *int                  `json:"seed,omitempty"`
	NumOutputs           *int32                `json:"num_outputs,omitempty"`
	StreamID             string                `json:"stream_id"` // Corresponds to SSE stream
	UIId                 string                `json:"ui_id"`     // Corresponds to UI identifier
	InitImageUrl         string                `json:"init_image_url,omitempty"`
	MaskImageUrl         string                `json:"mask_image_url,omitempty"`
	PromptStrength       *float32              `json:"prompt_strength,omitempty"`
	SubmitToGallery      bool                  `json:"submit_to_gallery"`
	ProcessType          shared.ProcessType    `json:"process_type"`
	OutputImageExtension shared.ImageExtension `json:"output_image_extension"`
	WasAutoSubmitted     bool
}

func (t *CreateGenerationRequest) Cost() int32 {
	return *t.NumOutputs
}

// Apply defaults for missing parameters
func (t *CreateGenerationRequest) ApplyDefaults() {
	if t.InferenceSteps == nil {
		t.InferenceSteps = utils.ToPtr(shared.DEFAULT_GENERATE_INFERENCE_STEPS)
	}
	if t.GuidanceScale == nil {
		t.GuidanceScale = utils.ToPtr(shared.DEFAULT_GENERATE_GUIDANCE_SCALE)
	}
	if t.NumOutputs == nil {
		t.NumOutputs = utils.ToPtr(shared.DEFAULT_GENERATE_NUM_OUTPUTS)
	}
	if t.ModelId == nil {
		t.ModelId = utils.ToPtr(shared.GetCache().GetDefaultGenerationModel().ID)
	}
	if t.SchedulerId == nil {
		t.SchedulerId = utils.ToPtr(shared.GetCache().GetDefaultSchedulerIDForModel(*t.ModelId))
	}
	if t.Width == nil {
		t.Width = utils.ToPtr(shared.GetCache().GetGenerationModelByID(*t.ModelId).DefaultWidth)
	}
	if t.Height == nil {
		t.Height = utils.ToPtr(shared.GetCache().GetGenerationModelByID(*t.ModelId).DefaultHeight)
	}
	if t.InitImageUrl != "" && t.PromptStrength == nil {
		t.PromptStrength = utils.ToPtr(shared.DEFAULT_GENERATE_PROMPT_STRENGTH)
	}
	if t.Seed == nil || *t.Seed < 0 {
		rand.Seed(time.Now().Unix())
		t.Seed = utils.ToPtr(rand.Intn(math.MaxInt32))
	}
}

func (t *CreateGenerationRequest) Validate(api bool) error {
	if !api && !utils.IsSha256Hash(t.StreamID) {
		return errors.New("invalid_stream_id")
	}

	t.ApplyDefaults()

	// Only apply scheduler check to API for now
	if api {
		compatibleSchedulerIds := shared.GetCache().GetCompatibleSchedulerIDsForModel(context.TODO(), *t.ModelId)
		if !slices.Contains(compatibleSchedulerIds, *t.SchedulerId) {
			fmt.Printf("MOdel ID %s", (*t.ModelId).String())
			fmt.Printf("Scheduler ID %s", (*t.SchedulerId).String())
			return errors.New("invalid_scheduler_id")
		}
	}

	if *t.Height > shared.MAX_GENERATE_HEIGHT {
		return fmt.Errorf("Height is too large, max is: %d", shared.MAX_GENERATE_HEIGHT)
	}

	if *t.Height < shared.MIN_GENERATE_HEIGHT {
		return fmt.Errorf("Height is too small, min is: %d", shared.MIN_GENERATE_HEIGHT)
	}

	if *t.Width > shared.MAX_GENERATE_WIDTH {
		return fmt.Errorf("Width is too large, max is: %d", shared.MAX_GENERATE_WIDTH)
	}

	if *t.Width < shared.MIN_GENERATE_WIDTH {
		return fmt.Errorf("Width is too small, min is: %d", shared.MIN_GENERATE_WIDTH)
	}

	if *t.GuidanceScale < shared.MIN_GUIDANCE_SCALE {
		return fmt.Errorf("Guidance scale is too small, min is: %f", shared.MIN_GUIDANCE_SCALE)
	}

	if *t.GuidanceScale > shared.MAX_GUIDANCE_SCALE {
		return fmt.Errorf("Guidance scale is too large, max is: %f", shared.MAX_GUIDANCE_SCALE)
	}

	if *t.InferenceSteps < shared.MIN_INFERENCE_STEPS {
		return fmt.Errorf("Inference steps is too small, min is: %d", shared.MIN_INFERENCE_STEPS)
	}

	if (*t.Width)*(*t.Height)*(*t.InferenceSteps) > shared.MAX_PRO_PIXEL_STEPS {
		return fmt.Errorf("Pick fewer inference steps or smaller dimensions: %d - %d - %d",
			*t.Width,
			*t.Height,
			*t.InferenceSteps,
		)
	}

	if *t.NumOutputs < 0 {
		t.NumOutputs = utils.ToPtr(shared.DEFAULT_GENERATE_NUM_OUTPUTS)
	}
	if *t.NumOutputs > shared.MAX_GENERATE_NUM_OUTPUTS {
		return fmt.Errorf("Number of outputs can't be more than %d", shared.MAX_GENERATE_NUM_OUTPUTS)
	}

	if !shared.GetCache().IsValidGenerationModelID(*t.ModelId) {
		return errors.New("invalid_model_id")
	}

	if !shared.GetCache().IsValidShedulerID(*t.SchedulerId) {
		return errors.New("invalid_scheduler_id")
	}

	// Ensure http, https, or s3
	if t.InitImageUrl != "" && !strings.HasPrefix(t.InitImageUrl, "s3://") && !strings.HasPrefix(t.InitImageUrl, "http://") && !strings.HasPrefix(t.InitImageUrl, "https://") {
		return errors.New("invalid_init_image_url")
	}

	if t.Seed == nil || *t.Seed < 0 {
		rand.Seed(time.Now().Unix())
		t.Seed = utils.ToPtr(rand.Intn(math.MaxInt32))
	}

	if t.PromptStrength != nil {
		if *t.PromptStrength < shared.MIN_PROMPT_STRENGTH {
			*t.PromptStrength = shared.MIN_PROMPT_STRENGTH
		}
		if *t.PromptStrength > shared.MAX_PROMPT_STRENGTH {
			*t.PromptStrength = shared.MAX_PROMPT_STRENGTH
		}
	}

	return nil
}
