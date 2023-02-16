package rest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/render"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/server/responses"
)

// Admin-related routes, these must be behind admin middleware and auth middleware

// HTTP POST - admin approve/reject image in gallery
func (c *RestAPI) HandleReviewGallerySubmission(w http.ResponseWriter, r *http.Request) {
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

	var updateCount int
	switch adminGalleryReq.Action {
	case requests.AdminGalleryActionApprove, requests.AdminGalleryActionReject:
		updateCount, err = c.Repo.ApproveOrRejectGenerationOutputs(adminGalleryReq.GenerationOutputIDs, adminGalleryReq.Action == requests.AdminGalleryActionApprove)
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

	res := responses.UpdatedResponse{
		Updated: updateCount,
	}
	render.Status(r, http.StatusOK)
	render.JSON(w, r, res)
}

// HTTP DELETE - admin delete generation
func (c *RestAPI) HandleDeleteGenerationOutput(w http.ResponseWriter, r *http.Request) {
	// Get user id (of admin)
	userID := c.GetUserIDIfAuthenticated(w, r)
	if userID == nil {
		return
	}

	// Parse request body
	reqBody, _ := io.ReadAll(r.Body)
	var deleteReq requests.GenerationDeleteRequest
	err := json.Unmarshal(reqBody, &deleteReq)
	if err != nil {
		responses.ErrUnableToParseJson(w, r)
		return
	}

	count, err := c.Repo.MarkGenerationOutputsForDeletion(deleteReq.GenerationOutputIDs)
	if err != nil {
		responses.ErrInternalServerError(w, r, err.Error())
		return
	}

	res := responses.DeletedResponse{
		Deleted: count,
	}
	render.Status(r, http.StatusOK)
	render.JSON(w, r, res)
}
