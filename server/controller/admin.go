package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/render"
	"github.com/stablecog/go-apps/database/ent"
	"github.com/stablecog/go-apps/server/requests"
	"github.com/stablecog/go-apps/server/responses"
)

// Admin-related routes, these must be behind admin middleware and auth middleware

// HTTP POST/DELETE - admin approve/reject image in gallery. Also allows deleting generations
// POST - approve/reject
// DELETE - delete
func (c *HttpController) HandleGenerationDeleteAndApproveRejectGallery(w http.ResponseWriter, r *http.Request) {
	// Get user id (of admin)
	userID := c.GetUserIDIfAuthenticated(w, r)
	if userID == nil {
		return
	}

	// Parse request body
	reqBody, _ := io.ReadAll(r.Body)
	var adminGalleryReq requests.AdminGalleryRequestBody
	err := json.Unmarshal(reqBody, &adminGalleryReq)
	if err != nil {
		responses.ErrUnableToParseJson(w, r)
		return
	}

	// Check request method
	// Post only for updates
	if r.Method == http.MethodPost && adminGalleryReq.Action == requests.AdminGalleryActionDelete {
		responses.ErrMethodNotAllowed(w, r, "Cannot use POST to delete image")
		return
	}
	// Delete only for deletes
	if r.Method == http.MethodDelete && adminGalleryReq.Action != requests.AdminGalleryActionDelete {
		responses.ErrMethodNotAllowed(w, r, "Cannot use DELETE to approve/reject image")
		return
	}

	// Ensure action is supported
	if adminGalleryReq.Action != requests.AdminGalleryActionApprove && adminGalleryReq.Action != requests.AdminGalleryActionReject && adminGalleryReq.Action != requests.AdminGalleryActionDelete {
		responses.ErrBadRequest(w, r, fmt.Sprintf("Unsupported action %s", adminGalleryReq.Action))
		return
	}

	switch adminGalleryReq.Action {
	case requests.AdminGalleryActionApprove, requests.AdminGalleryActionReject:
		err = c.Repo.ApproveOrRejectGeneration(adminGalleryReq.GenerationID, adminGalleryReq.Action == requests.AdminGalleryActionApprove)
		if err != nil {
			if ent.IsNotFound(err) {
				responses.ErrBadRequest(w, r, "Generation not found")
				return
			}
			responses.ErrInternalServerError(w, r, err.Error())
			return
		}
	case requests.AdminGalleryActionDelete:
		err = c.Repo.DeleteGeneration(adminGalleryReq.GenerationID)
		if err != nil {
			if ent.IsNotFound(err) {
				responses.ErrBadRequest(w, r, "Generation not found")
				return
			}
			responses.ErrInternalServerError(w, r, err.Error())
			return
		}
	}

	render.Status(r, http.StatusOK)
}
