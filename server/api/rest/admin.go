package rest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/render"
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/ent/generationoutput"
	"github.com/stablecog/sc-go/database/ent/userrole"
	"github.com/stablecog/sc-go/database/qdrant"
	"github.com/stablecog/sc-go/database/repository"
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
				responses.ErrBadRequest(w, r, "Generation not found", "")
				return
			}
			responses.ErrInternalServerError(w, r, err.Error())
			return
		}
	default:
		responses.ErrBadRequest(w, r, fmt.Sprintf("Unsupported action %s", adminGalleryReq.Action), "")
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
			responses.ErrBadRequest(w, r, "per_page must be an integer", "")
			return
		} else if perPage < 1 || perPage > MAX_PER_PAGE {
			responses.ErrBadRequest(w, r, fmt.Sprintf("per_page must be between 1 and %d", MAX_PER_PAGE), "")
			return
		}
	}

	cursorStr := r.URL.Query().Get("cursor")
	search := r.URL.Query().Get("search")

	filters := &requests.QueryGenerationFilters{}
	err = filters.ParseURLQueryParameters(r.URL.Query())
	if err != nil {
		responses.ErrBadRequest(w, r, err.Error(), "")
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
			responses.ErrUnauthorized(w, r)
			return
		}
	}

	// For search, use qdrant semantic search
	if search != "" {
		// get embeddings from clip service
		e, err := c.Clip.GetEmbeddingFromText(search, 2)
		if err != nil {
			log.Error("Error getting embedding from clip service", "err", err)
			responses.ErrInternalServerError(w, r, "An unknown error has occured")
			return
		}

		// Parse as qdrant filters
		qdrantFilters, scoreThreshold := filters.ToQdrantFilters(false)
		// Deleted at not empty
		qdrantFilters.Must = append(qdrantFilters.Must, qdrant.SCMatchCondition{
			IsEmpty: &qdrant.SCIsEmpty{Key: "deleted_at"},
		})

		// Get cursor str as uint
		var offset *uint
		var total *uint
		if cursorStr != "" {
			cursoru64, err := strconv.ParseUint(cursorStr, 10, 64)
			if err != nil {
				responses.ErrBadRequest(w, r, "cursor must be a valid uint", "")
				return
			}
			cursoru := uint(cursoru64)
			offset = &cursoru
		} else {
			count, err := c.Qdrant.CountWithFilters(qdrantFilters, false)
			if err != nil {
				log.Error("Error counting qdrant", "err", err)
				responses.ErrInternalServerError(w, r, "An unknown error has occured")
				return
			}
			total = &count
		}

		// Query qdrant
		qdrantRes, err := c.Qdrant.QueryGenerations(e, perPage, offset, scoreThreshold, qdrantFilters, false, false)
		if err != nil {
			log.Error("Error querying qdrant", "err", err)
			responses.ErrInternalServerError(w, r, "An unknown error has occured")
			return
		}

		// Get generation output ids
		var outputIds []uuid.UUID
		for _, hit := range qdrantRes.Result {
			outputId, err := uuid.Parse(hit.Id)
			if err != nil {
				log.Error("Error parsing uuid", "err", err)
				continue
			}
			outputIds = append(outputIds, outputId)
		}

		// Get user generation data in correct format
		generationsUnsorted, err := c.Repo.RetrieveGenerationsWithOutputIDs(outputIds)
		if err != nil {
			log.Error("Error getting generations", "err", err)
			responses.ErrInternalServerError(w, r, "An unknown error has occured")
			return
		}

		// Need to re-sort to preserve qdrant ordering
		gDataMap := make(map[uuid.UUID]repository.GenerationQueryWithOutputsResultFormatted)
		for _, gData := range generationsUnsorted.Outputs {
			gDataMap[gData.ID] = gData
		}

		generations := []repository.GenerationQueryWithOutputsResultFormatted{}
		for _, hit := range qdrantRes.Result {
			outputId, err := uuid.Parse(hit.Id)
			if err != nil {
				log.Error("Error parsing uuid", "err", err)
				continue
			}
			item, ok := gDataMap[outputId]
			if !ok {
				log.Error("Error retrieving gallery data", "output_id", outputId)
				continue
			}
			generations = append(generations, item)
		}
		generationsUnsorted.Outputs = generations

		if total != nil {
			// uint to int
			totalInt := int(*total)
			generationsUnsorted.Total = &totalInt
		}

		// Get next cursor
		generationsUnsorted.Next = qdrantRes.Next

		// Return generations
		render.Status(r, http.StatusOK)
		render.JSON(w, r, generationsUnsorted)
		return
	}

	// Otherwise, query postgres
	var cursor *time.Time
	if cursorStr := r.URL.Query().Get("cursor"); cursorStr != "" {
		cursorTime, err := utils.ParseIsoTime(cursorStr)
		if err != nil {
			responses.ErrBadRequest(w, r, "cursor must be a valid iso time string", "")
			return
		}
		cursor = &cursorTime
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
			responses.ErrBadRequest(w, r, "per_page must be an integer", "")
			return
		} else if perPage < 1 || perPage > MAX_PER_PAGE {
			responses.ErrBadRequest(w, r, fmt.Sprintf("per_page must be between 1 and %d", MAX_PER_PAGE), "")
			return
		}
	}

	var cursor *time.Time
	if cursorStr := r.URL.Query().Get("cursor"); cursorStr != "" {
		cursorTime, err := utils.ParseIsoTime(cursorStr)
		if err != nil {
			responses.ErrBadRequest(w, r, "cursor must be a valid iso time string", "")
			return
		}
		cursor = &cursorTime
	}

	var productIds []string
	if productIdsStr := r.URL.Query().Get("active_product_ids"); productIdsStr != "" {
		productIds = strings.Split(productIdsStr, ",")
	}

	// Get users
	users, err := c.Repo.QueryUsers(r.URL.Query().Get("search"), perPage, cursor, productIds)
	if err != nil {
		log.Error("Error getting users", "err", err)
		responses.ErrInternalServerError(w, r, "Error getting users")
		return
	}

	// Return generations
	render.Status(r, http.StatusOK)
	render.JSON(w, r, users)
}

// Get available credit types admin can gift to user
func (c *RestAPI) HandleQueryCreditTypes(w http.ResponseWriter, r *http.Request) {
	if user, email := c.GetUserIDAndEmailIfAuthenticated(w, r); user == nil || email == "" {
		return
	}

	// Get credit types
	creditTypes, err := c.Repo.GetCreditTypeList()
	if err != nil {
		log.Error("Error getting credit types", "err", err)
		responses.ErrInternalServerError(w, r, "An unknown error has occured")
		return
	}

	resp := make([]responses.QueryCreditTypesResponse, len(creditTypes))
	for i, ct := range creditTypes {
		resp[i].ID = ct.ID
		resp[i].Amount = ct.Amount
		resp[i].Name = ct.Name
		resp[i].Description = ct.Name
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, resp)
}

// Add credits to user
func (c *RestAPI) HandleAddCreditsToUser(w http.ResponseWriter, r *http.Request) {
	if user, email := c.GetUserIDAndEmailIfAuthenticated(w, r); user == nil || email == "" {
		return
	}

	// Parse request body
	reqBody, _ := io.ReadAll(r.Body)
	var addReq requests.CreditAddRequest
	err := json.Unmarshal(reqBody, &addReq)
	if err != nil {
		responses.ErrUnableToParseJson(w, r)
		return
	}

	// Get credit type
	creditType, err := c.Repo.GetCreditTypeByID(addReq.CreditTypeID)
	if err != nil {
		log.Error("Error getting credit type", "err", err)
		responses.ErrInternalServerError(w, r, "An unknown error has occured")
		return
	} else if err == nil && creditType == nil {
		responses.ErrNotFound(w, r, fmt.Sprintf("Invalid credit type %s", addReq.CreditTypeID.String()))
		return
	}

	err = c.Repo.AddCreditsToUser(creditType, addReq.UserID)
	if err != nil {
		log.Error("Error adding credits to user", "err", err)
		responses.ErrInternalServerError(w, r, "An unknown error has occured")
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, map[string]interface{}{
		"added": true,
	})
}
