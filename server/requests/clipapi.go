package requests

import "github.com/google/uuid"

type ClipAPIRequest struct {
	Text string `json:"text"`
}

type ClipAPIImageRequest struct {
	Image string    `json:"image"`
	ID    uuid.UUID `json:"id"`
}
