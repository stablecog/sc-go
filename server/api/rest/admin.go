package rest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/render"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/ent/generationoutput"
	"github.com/stablecog/sc-go/database/ent/userrole"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/server/responses"
	"github.com/stablecog/sc-go/utils"
	"golang.org/x/exp/slices"
)

// Admin-related routes, these must be behind admin middleware and auth middleware

// HTTP POST - admin approve/reject image in gallery
func (c *RestAPI) HandleReviewGallerySubmission(w http.ResponseWriter, r *http.Request) {
	if user, email := c.GetUserIDAndEmailIfAuthenticated(w, r); user == nil || email == "" {
		return
	}

	// Parse request body
	reqBody, _ := io.ReadAll(r.Body)
	var adminGalleryReq requests.ReviewGalleryRequest
	err := json.Unmarshal(reqBody, &adminGalleryReq)
	if err != nil {
		responses.ErrUnableToParseJson(w, r)
		return
	}

	var updateCount int
	switch adminGalleryReq.Action {
	case requests.GalleryApproveAction, requests.GalleryRejectAction:
		updateCount, err = c.Repo.ApproveOrRejectGenerationOutputs(adminGalleryReq.GenerationOutputIDs, adminGalleryReq.Action == requests.GalleryApproveAction)
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
	// Get user
	if user, email := c.GetUserIDAndEmailIfAuthenticated(w, r); user == nil || email == "" {
		return
	}

	// Get user_role from context
	userRole, ok := r.Context().Value("user_role").(string)
	if !ok || userRole != userrole.RoleNameSUPER_ADMIN.String() {
		responses.ErrUnauthorized(w, r)
		return
	}

	// Parse request body
	reqBody, _ := io.ReadAll(r.Body)
	var deleteReq requests.DeleteGenerationRequest
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

// HTTP Get - generations for admin
func (c *RestAPI) HandleQueryGenerationsForAdmin(w http.ResponseWriter, r *http.Request) {
	if user, email := c.GetUserIDAndEmailIfAuthenticated(w, r); user == nil || email == "" {
		return
	}

	// Get user_role from context
	userRole, ok := r.Context().Value("user_role").(string)
	if !ok {
		responses.ErrUnauthorized(w, r)
		return
	}
	superAdmin := userRole == userrole.RoleNameSUPER_ADMIN.String()

	// Validate query parameters
	perPage := DEFAULT_PER_PAGE
	var err error
	if perPageStr := r.URL.Query().Get("per_page"); perPageStr != "" {
		perPage, err = strconv.Atoi(perPageStr)
		if err != nil {
			responses.ErrBadRequest(w, r, "per_page must be an integer")
			return
		} else if perPage < 1 || perPage > MAX_PER_PAGE {
			responses.ErrBadRequest(w, r, fmt.Sprintf("per_page must be between 1 and %d", MAX_PER_PAGE))
			return
		}
	}

	var cursor *time.Time
	if cursorStr := r.URL.Query().Get("cursor"); cursorStr != "" {
		cursorTime, err := utils.ParseIsoTime(cursorStr)
		if err != nil {
			responses.ErrBadRequest(w, r, "cursor must be a valid iso time string")
			return
		}
		cursor = &cursorTime
	}

	filters := &requests.QueryGenerationFilters{}
	err = filters.ParseURLQueryParameters(r.URL.Query())
	if err != nil {
		responses.ErrBadRequest(w, r, err.Error())
		return
	}
	// Make sure non-super admin can't get private generations
	if !superAdmin {
		if len(filters.GalleryStatus) == 0 {
			filters.GalleryStatus = []generationoutput.GalleryStatus{
				generationoutput.GalleryStatusApproved,
				generationoutput.GalleryStatusRejected,
				generationoutput.GalleryStatusSubmitted,
			}
		} else if slices.Contains(filters.GalleryStatus, generationoutput.GalleryStatusNotSubmitted) {
			responses.ErrBadRequest(w, r, "Only super admins can query for not submitted generations")
			return
		}
	}

	// Get generaions
	generations, err := c.Repo.QueryGenerationsAdmin(perPage, cursor, filters)
	if err != nil {
		log.Error("Error getting generations for admin", "err", err)
		responses.ErrInternalServerError(w, r, "Error getting generations")
		return
	}

	// Return generations
	render.Status(r, http.StatusOK)
	render.JSON(w, r, generations)
}

// HTTP Get - users for admin
func (c *RestAPI) HandleQueryUsers(w http.ResponseWriter, r *http.Request) {
	if user, email := c.GetUserIDAndEmailIfAuthenticated(w, r); user == nil || email == "" {
		return
	}

	// Validate query parameters
	perPage := DEFAULT_PER_PAGE
	var err error
	if perPageStr := r.URL.Query().Get("per_page"); perPageStr != "" {
		perPage, err = strconv.Atoi(perPageStr)
		if err != nil {
			responses.ErrBadRequest(w, r, "per_page must be an integer")
			return
		} else if perPage < 1 || perPage > MAX_PER_PAGE {
			responses.ErrBadRequest(w, r, fmt.Sprintf("per_page must be between 1 and %d", MAX_PER_PAGE))
			return
		}
	}

	var cursor *time.Time
	if cursorStr := r.URL.Query().Get("cursor"); cursorStr != "" {
		cursorTime, err := utils.ParseIsoTime(cursorStr)
		if err != nil {
			responses.ErrBadRequest(w, r, "cursor must be a valid iso time string")
			return
		}
		cursor = &cursorTime
	}

	// Get users
	users, err := c.Repo.QueryUsers(r.URL.Query().Get("search"), perPage, cursor)
	if err != nil {
		log.Error("Error getting users", "err", err)
		responses.ErrInternalServerError(w, r, "Error getting users")
		return
	}

	// Return generations
	render.Status(r, http.StatusOK)
	render.JSON(w, r, users)
}
