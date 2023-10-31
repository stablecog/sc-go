package rest

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/qdrant"
	"github.com/stablecog/sc-go/database/repository"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/server/responses"
	"github.com/stablecog/sc-go/utils"
)

type UserProfileMetadata struct {
	CreatedAt       time.Time `json:"created_at"`
	ActiveProductID *string   `json:"active_product_id,omitempty"`
	Username        string    `json:"username"`
}

// For v1/profile/{username}/metadata
func (c *RestAPI) HandleGetUserProfileMetadata(w http.ResponseWriter, r *http.Request) {
	// Get username
	username := chi.URLParam(r, "username")
	user, err := c.Repo.GetUserByUsername(username)
	if err != nil {
		if ent.IsNotFound(err) {
			responses.ErrNotFound(w, r, "user_not_found")
			return
		}
		log.Error("Error retrieving user", "err", err)
		responses.ErrInternalServerError(w, r, "An unknown error has occurred")
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, UserProfileMetadata{
		CreatedAt:       user.CreatedAt,
		ActiveProductID: user.ActiveProductID,
		Username:        user.Username,
	})
}

// For v1/profile/{username}/outputs
func (c *RestAPI) HandleUserProfileSemanticSearch(w http.ResponseWriter, r *http.Request) {
	// Get user for like data, if authenticated
	callingUser, err := c.GetUserIfAuthenticatedOnly(w, r)
	if err != nil {
		log.Error("Error getting user", "err", err)
		responses.ErrInternalServerError(w, r, "An unknown error has occurred")
		return
	}
	var callingUserId *uuid.UUID
	if callingUser != nil {
		callingUserId = utils.ToPtr(callingUser.ID)
	}

	// Get username
	username := chi.URLParam(r, "username")
	user, err := c.Repo.GetUserByUsername(username)
	if err != nil {
		if ent.IsNotFound(err) {
			responses.ErrNotFound(w, r, "user_not_found")
			return
		}
		log.Error("Error retrieving user", "err", err)
		responses.ErrInternalServerError(w, r, "An unknown error has occurred")
		return
	}

	metadata := &UserProfileMetadata{
		CreatedAt:       user.CreatedAt,
		ActiveProductID: user.ActiveProductID,
		Username:        user.Username,
	}

	if user.BannedAt != nil {
		render.Status(r, http.StatusOK)
		render.JSON(w, r, GalleryResponse[*uint]{
			Next:     nil,
			Hits:     []repository.GalleryData{},
			Metadata: metadata,
		})
		return
	}

	// Get output_id param
	outputId := r.URL.Query().Get("output_id")
	if outputId != "" {
		// Validate output_id
		uid, err := uuid.Parse(outputId)
		if err != nil {
			responses.ErrBadRequest(w, r, "invalid_output_id", "")
			return
		}

		galleryData, err := c.Repo.RetrieveGalleryDataByID(uid, utils.ToPtr(user.ID), callingUserId, true)
		if err != nil {
			if ent.IsNotFound(err) {
				responses.ErrNotFound(w, r, "generation_not_found")
				return
			}
			log.Error("Error retrieving gallery data", "err", err)
			responses.ErrInternalServerError(w, r, "Error retrieving gallery data")
			return
		}

		// Sanitize
		galleryData.UserID = nil

		render.Status(r, http.StatusOK)
		render.JSON(w, r, GalleryResponse[int]{
			Page:     1,
			Hits:     []repository.GalleryData{*galleryData},
			Metadata: metadata,
		})
		return
	}

	search := r.URL.Query().Get("search")
	cursor := r.URL.Query().Get("cursor")
	galleryData := []repository.GalleryData{}
	var nextCursorQdrant *uint
	var nextCursorPostgres *time.Time

	// Parse filters
	filters := &requests.QueryGenerationFilters{}
	err = filters.ParseURLQueryParameters(r.URL.Query())
	if err != nil {
		responses.ErrBadRequest(w, r, err.Error(), "")
		return
	}
	filters.UserID = utils.ToPtr(user.ID)

	// Validate query parameters
	perPage := GALLERY_PER_PAGE
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

	// Parse as qdrant filters
	qdrantFilters, scoreThreshold := filters.ToQdrantFilters(true)
	// Append gallery status requirement
	qdrantFilters.Must = append(qdrantFilters.Must, qdrant.SCMatchCondition{
		Key:   "is_public",
		Match: &qdrant.SCValue{Value: true},
	})
	qdrantFilters.Must = append(qdrantFilters.Must, qdrant.SCMatchCondition{
		IsEmpty: &qdrant.SCIsEmpty{Key: "deleted_at"},
	})

	// Leverage qdrant for semantic search
	if search != "" {
		var offset *uint
		if cursor != "" {
			cursoru64, err := strconv.ParseUint(cursor, 10, 64)
			if err != nil {
				responses.ErrBadRequest(w, r, "cursor must be a valid uint", "")
				return
			}
			cursorU := uint(cursoru64)
			offset = &cursorU
		}
		// See if search is a uuid
		uid, err := uuid.Parse(search)
		var embeddings []float32
		if err == nil {
			// Get embeddings from qdrant
			getPointRes, err := c.Qdrant.GetPoint(uid, false)
			if err != nil {
				log.Error("Error getting point from qdrant", "err", err)
				if strings.Contains(err.Error(), "Error querying collection 404") {
					responses.ErrNotFound(w, r, "generation_not_found")
					return
				}
				responses.ErrInternalServerError(w, r, "An unknown error occurred")
				return
			}
			embeddings = getPointRes.Result.Vector.Image
		} else {
			embeddings, err = c.Clip.GetEmbeddingFromText(search, 3, true)
			if err != nil {
				log.Error("Error getting embeddings from clip service", "err", err)
				responses.ErrInternalServerError(w, r, "An unknown error occurred")
				return
			}
		}

		res, err := c.Qdrant.QueryGenerations(embeddings, perPage, offset, scoreThreshold, qdrantFilters, false, false)
		if err != nil {
			log.Error("Error querying qdrant", "err", err)
			responses.ErrInternalServerError(w, r, "An unknown error occurred")
			return
		}

		// Get generation output ids
		var outputIds []uuid.UUID
		for _, hit := range res.Result {
			outputId, err := uuid.Parse(hit.Id)
			if err != nil {
				log.Error("Error parsing uuid", "err", err)
				continue
			}
			outputIds = append(outputIds, outputId)
		}

		// Get gallery data
		galleryDataUnsorted, err := c.Repo.RetrieveGalleryDataWithOutputIDs(outputIds, callingUserId, true)
		if err != nil {
			log.Error("Error querying gallery data", "err", err)
			responses.ErrInternalServerError(w, r, "An unknown error occurred")
			return
		}
		gDataMap := make(map[uuid.UUID]repository.GalleryData)
		for _, gData := range galleryDataUnsorted {
			gDataMap[gData.ID] = gData
		}

		for _, hit := range res.Result {
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
			galleryData = append(galleryData, item)
		}

		// Set next cursor
		nextCursorQdrant = res.Next
	} else {
		// Get most recent gallery data
		var qCursor *time.Time
		if cursor != "" {
			cursorTime, err := utils.ParseIsoTime(cursor)
			if err != nil {
				responses.ErrBadRequest(w, r, "cursor must be a valid iso time string", "")
				return
			}
			qCursor = &cursorTime
		}

		// Retrieve from postgres
		filters.IsPublic = utils.ToPtr(true)
		galleryData, nextCursorPostgres, err = c.Repo.RetrieveMostRecentGalleryDataV2(filters, callingUserId, perPage, qCursor)
		if err != nil {
			log.Error("Error querying gallery data from postgres", "err", err)
			responses.ErrInternalServerError(w, r, "An unknown error occurred")
			return
		}
	}

	// We don't want to leak primary keys, so set to nil
	for i := range galleryData {
		galleryData[i].UserID = nil
	}

	if search == "" {
		render.Status(r, http.StatusOK)
		render.JSON(w, r, GalleryResponse[*time.Time]{
			Next:     nextCursorPostgres,
			Hits:     galleryData,
			Metadata: metadata,
		})
		return
	}
	render.Status(r, http.StatusOK)
	render.JSON(w, r, GalleryResponse[*uint]{
		Next:     nextCursorQdrant,
		Hits:     galleryData,
		Metadata: metadata,
	})
}
