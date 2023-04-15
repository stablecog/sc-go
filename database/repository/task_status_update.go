package repository

import (
	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/shared"
)

// Represents an update to a generation/upscale in our database
type TaskStatusUpdateResponse struct {
	Status           requests.CogTaskStatus    `json:"status"`
	ProcessType      shared.ProcessType        `json:"process_type"`
	Id               string                    `json:"id"`
	UIId             string                    `json:"ui_id,omitempty"`
	StreamId         string                    `json:"stream_id"`
	Error            string                    `json:"error,omitempty"`
	NSFWCount        int32                     `json:"nsfw_count,omitempty"`
	Outputs          []GenerationUpscaleOutput `json:"outputs,omitempty"`
	ForLivePage      bool                      `json:"for_live_page,omitempty"`
	LivePageMessage  *shared.LivePageMessage   `json:"live_page_message,omitempty"`
	RemainingCredits int                       `json:"total_remaining_credits,omitempty"`
}
