package rest

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

// HTTP POST - admin approve/reject image in gallery
func (c *RestAPI) HandleGenerationApproveRejectGallery(w http.ResponseWriter, r *http.Request) {
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
	default:
		responses.ErrBadRequest(w, r, fmt.Sprintf("Unsupported action %s", adminGalleryReq.Action))
		return
	}

	render.Status(r, http.StatusOK)
}

// HTTP DELETE - admin delete generation
func (c *RestAPI) HandleDeleteGeneration(w http.ResponseWriter, r *http.Request) {
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

	switch adminGalleryReq.Action {
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
	default:
		responses.ErrBadRequest(w, r, fmt.Sprintf("Unsupported action %s", adminGalleryReq.Action))
		return
	}

	render.Status(r, http.StatusOK)
}
