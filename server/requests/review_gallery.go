// * Requests initiated by admins
package requests

import (
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent/generationoutput"
)

type ReviewGalleryAction string

const (
	GalleryApproveAction ReviewGalleryAction = "approve"
	GalleryRejectAction  ReviewGalleryAction = "reject"
)

type ReviewGalleryRequest struct {
	// ! Deprecated, use gallery_status instead
	Action              ReviewGalleryAction            `json:"action"`
	GenerationOutputIDs []uuid.UUID                    `json:"generation_output_ids"`
	GalleryStatus       generationoutput.GalleryStatus `json:"gallery_status"`
}
