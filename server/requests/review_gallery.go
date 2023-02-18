// * Requests initiated by admins
package requests

import (
	"github.com/google/uuid"
)

type ReviewGalleryAction string

const (
	GalleryApproveAction ReviewGalleryAction = "approve"
	GalleryRejectAction  ReviewGalleryAction = "reject"
)

type ReviewGalleryRequest struct {
	Action              ReviewGalleryAction `json:"action"`
	GenerationOutputIDs []uuid.UUID         `json:"generation_output_ids"`
}
