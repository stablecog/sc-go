package requests

import "github.com/google/uuid"

type ClipAPIRequest struct {
	Text string `json:"text"`
}

type ClipAPIImageRequest struct {
	Image   string    `json:"image,omitempty"`
	ImageID string    `json:"image_id,omitempty"`
	ID      uuid.UUID `json:"id"`
}
