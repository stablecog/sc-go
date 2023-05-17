package requests

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/stablecog/sc-go/shared"
	"github.com/stablecog/sc-go/utils"
)

// Request for initiationg an upscale
type UpscaleRequestType string

const (
	UpscaleRequestTypeImage  UpscaleRequestType = "from_image"
	UpscaleRequestTypeOutput UpscaleRequestType = "from_output"
)

// Can be initiated with either an image_url or a generation_output_id
type CreateUpscaleRequest struct {
	Type     UpscaleRequestType `json:"type"`
	Input    string             `json:"input"`
	ModelId  uuid.UUID          `json:"model_id"`
	StreamID string             `json:"stream_id"`
	UIId     string             `json:"ui_id"` // Corresponds to UI identifier
	OutputID uuid.UUID
}

func (t *CreateUpscaleRequest) Validate(api bool) error {
	if !api && !utils.IsSha256Hash(t.StreamID) {
		return errors.New("invalid_stream_id")
	}

	if t.Type != UpscaleRequestTypeImage && t.Type != UpscaleRequestTypeOutput {
		return fmt.Errorf("Invalid upscale type, should be %s or %s", UpscaleRequestTypeImage, UpscaleRequestTypeOutput)
	}

	if t.Type == UpscaleRequestTypeImage && !utils.IsValidHTTPURL(t.Input) {
		return errors.New("invalid_image_url")
	} else if t.Type == UpscaleRequestTypeOutput {
		outputID, err := uuid.Parse(t.Input)
		if err != nil {
			return errors.New("invalid_output_id")
		}
		t.OutputID = outputID
	}

	if !shared.GetCache().IsValidUpscaleModelID(t.ModelId) {
		return errors.New("invalid_model_id")
	}

	return nil
}
