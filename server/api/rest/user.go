package rest

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/render"
	"github.com/google/uuid"
	"github.com/stablecog/go-apps/server/responses"
	"github.com/stablecog/go-apps/utils"
	"k8s.io/klog/v2"
)

const DEFAULT_PER_PAGE = 50
const MAX_PER_PAGE = 100

// HTTP Get - generations for user
// Takes query paramers for pagination
// per_page: number of generations to return
// offset: offset for pagination, it is an iso time string in UTC
func (c *RestAPI) HandleQueryGenerations(w http.ResponseWriter, r *http.Request) {
	// See if authenticated
	userIDStr, authenticated := r.Context().Value("user_id").(string)
	// This should always be true because of the auth middleware, but check it anyway
	if !authenticated || userIDStr == "" {
		responses.ErrUnauthorized(w, r)
		return
	}
	// Parse to UUID
	userId, err := uuid.Parse(userIDStr)
	if err != nil {
		responses.ErrUnauthorized(w, r)
		return
	}

	// Validate query parameters
	perPage := DEFAULT_PER_PAGE
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

	var offset *time.Time
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		offsetTime, err := utils.ParseIsoTime(offsetStr)
		if err != nil {
			responses.ErrBadRequest(w, r, "offset must be a valid iso time string")
			return
		}
		offset = &offsetTime
	}

	// Get generaions
	generations, err := c.Repo.GetUserGenerations(userId, perPage, offset)
	if err != nil {
		klog.Errorf("Error getting generations for user: %s", err)
		responses.ErrInternalServerError(w, r, "Error getting generations")
		return
	}

	// Format response
	var generationsResponse []responses.UserGenerationsResponse
	for _, generation := range generations {
		generationResponse := responses.UserGenerationsResponse{
			Width:             generation.Width,
			Height:            generation.Height,
			NumInferenceSteps: generation.InferenceSteps,
			GuidanceScale:     generation.GuidanceScale,
			Seed:              generation.Seed,
			CreatedAt:         generation.CreatedAt,
			StartedAt:         generation.StartedAt,
			CompletedAt:       generation.CompletedAt,
			Status:            generation.Status,
			GalleryStatus:     generation.GalleryStatus,
		}
		// Negative prompt can actually be nil
		if generation.Edges.NegativePrompt != nil {
			generationResponse.NegativePrompt = generation.Edges.NegativePrompt.Text
		}
		// These are nillable because they are joins, but they shouldn't be empty
		// make sure anyway since we don't want to panic
		if generation.Edges.Prompt != nil {
			generationResponse.Prompt = generation.Edges.Prompt.Text
		}
		if generation.Edges.GenerationModel != nil {
			generationResponse.Model = generation.Edges.GenerationModel.Name
		}
		if generation.Edges.Scheduler != nil {
			generationResponse.Scheduler = generation.Edges.Scheduler.Name
		}
		if generation.Edges.GenerationOutputs != nil {
			generationResponse.Outputs = []string{}
			for _, output := range generation.Edges.GenerationOutputs {
				generationResponse.Outputs = append(generationResponse.Outputs, output.ImageURL)
			}
		}
		generationsResponse = append(generationsResponse, generationResponse)
	}

	// Return generations
	render.Status(r, http.StatusOK)
	render.JSON(w, r, generationsResponse)
}
