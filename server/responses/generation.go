package responses

import (
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent/generationoutput"
)

type GenerationOutputResponse struct {
	ID               uuid.UUID                      `json:"id"`
	ImageUrl         string                         `json:"image_url"`
	UpscaledImageUrl string                         `json:"upscaled_image_url,omitempty"`
	GalleryStatus    generationoutput.GalleryStatus `json:"gallery_status,omitempty"`
}
