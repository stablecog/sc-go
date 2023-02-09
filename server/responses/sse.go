package responses

import (
	"github.com/google/uuid"
	"github.com/stablecog/go-apps/database/ent/generationoutput"
)

type WebhookStatusUpdateOutputs struct {
	ID               uuid.UUID                      `json:"id"`
	ImageUrl         string                         `json:"image_url"`
	UpscaledImageUrl *string                        `json:"upscaled_image_url,omitempty"`
	GalleryStatus    generationoutput.GalleryStatus `json:"gallery_status,omitempty"`
}

type SSEStatusUpdateResponse struct {
	Status    CogTaskStatus                `json:"status"`
	Id        string                       `json:"id"`
	Error     string                       `json:"error,omitempty"`
	NSFWCount int                          `json:"nsfw_count,omitempty"`
	Outputs   []WebhookStatusUpdateOutputs `json:"outputs,omitempty"`
}
