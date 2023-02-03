// * Requests initiated by admins
package requests

import (
	"github.com/google/uuid"
)

type AdminGalleryAction string

const (
	AdminGalleryActionApprove AdminGalleryAction = "approve"
	AdminGalleryActionReject  AdminGalleryAction = "reject"
)

type AdminGalleryRequestBody struct {
	Action              AdminGalleryAction `json:"action"`
	GenerationOutputIDs []uuid.UUID        `json:"generation_output_ids"`
}

type AdminGenerationDeleteRequest struct {
	GenerationIDs []uuid.UUID `json:"generation_ids"`
}
