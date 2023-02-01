package requests

import (
	"github.com/google/uuid"
)

type AdminGalleryAction string

const (
	AdminGalleryActionApprove AdminGalleryAction = "approve"
	AdminGalleryActionReject  AdminGalleryAction = "reject"
	AdminGalleryActionDelete  AdminGalleryAction = "delete"
)

type AdminGalleryRequestBody struct {
	Action       AdminGalleryAction `json:"action"`
	GenerationID uuid.UUID          `json:"generation_id"`
}
