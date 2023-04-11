package rest

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/render"
	"github.com/google/uuid"
	"github.com/meilisearch/meilisearch-go"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/ent/generationoutput"
	"github.com/stablecog/sc-go/database/qdrant"
	"github.com/stablecog/sc-go/database/repository"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/server/responses"
	"github.com/stablecog/sc-go/utils"
)

const GALLERY_PER_PAGE = 50

func (c *RestAPI) GetGenerationGs(page int, batchSize int, search string, filter string) ([]repository.GalleryData, error) {
	var generationGs = []repository.GalleryData{}
	sortBy := []string{}
	if search == "" {
		sortBy = []string{"created_at:desc"}
	}
	res, err := c.Meili.Index("generation_g").Search(search, &meilisearch.SearchRequest{
		Page:        int64(page),
		HitsPerPage: int64(batchSize),
		Sort:        sortBy,
	})
	if err != nil {
		log.Error("Error searching for generation_g", "err", err)
		return nil, err
	}
	if len(res.Hits) == 0 {
		return generationGs, nil
	}
	for _, hit := range res.Hits {
		j, err := json.Marshal(hit)
		if err != nil {
			log.Error("Error marshalling hit", "err", err)
			return nil, err
		}
		var gen repository.GalleryData
		err = json.Unmarshal(j, &gen)
		if err != nil {
			log.Error("Error unmarshalling hit", "err", err)
			return nil, err
		}
		gen.Seed = 0
		generationGs = append(generationGs, gen)
	}
	return generationGs, nil
}

// Get a specific generation_g by ID
func (c *RestAPI) GetGenerationGByID(outputId uuid.UUID) (*repository.GalleryData, error) {
	res, err := c.Meili.Index("generation_g").Search("", &meilisearch.SearchRequest{
		Page:        int64(1),
		HitsPerPage: int64(1),
		Filter:      []string{fmt.Sprintf("id = %s", outputId.String())},
	})
	if err != nil {
		log.Error("Error searching for generation_g", "err", err)
		return nil, err
	}
	if len(res.Hits) == 0 {
		return nil, nil
	}
	var generationG repository.GalleryData
	for _, hit := range res.Hits {
		j, err := json.Marshal(hit)
		if err != nil {
			log.Error("Error marshalling hit", "err", err)
			return nil, err
		}
		var gen repository.GalleryData
		err = json.Unmarshal(j, &gen)
		if err != nil {
			log.Error("Error unmarshalling hit", "err", err)
			return nil, err
		}
		gen.Seed = 0
		generationG = gen
		break
	}
	return &generationG, nil
}

func (c *RestAPI) HandleQueryGallery(w http.ResponseWriter, r *http.Request) {
	// Get output_id param
	outputId := r.URL.Query().Get("output_id")
	if outputId != "" {
		// Validate output_id
		uid, err := uuid.Parse(outputId)
		if err != nil {
			responses.ErrBadRequest(w, r, "invalid_output_id", "")
			return
		}

		generationG, err := c.GetGenerationGByID(uid)
		if err != nil || generationG == nil {
			log.Error("Error querying generation meili", "err", err)
			responses.ErrInternalServerError(w, r, "Error querying generation")
			return
		}

		// Sanitize
		generationG.UserID = nil

		imageUrl := utils.GetURLFromImagePath(generationG.ImagePath)
		if err != nil {
			log.Error("Error parsing S3 URL", "err", err)
			imageUrl = generationG.ImagePath
		}
		generationG.ImageURL = imageUrl
		generationG.ImagePath = ""
		if generationG.UpscaledImagePath != "" {
			imageUrl := utils.GetURLFromImagePath(generationG.UpscaledImagePath)
			generationG.UpscaledImageURL = imageUrl
			generationG.UpscaledImagePath = ""
		}

		render.Status(r, http.StatusOK)
		render.JSON(w, r, GalleryResponse[int]{
			Page: 1,
			Hits: []repository.GalleryData{*generationG},
		})
		return
	}

	// Get query params
	page, err := strconv.Atoi(r.URL.Query().Get("cursor"))
	if err != nil || page < 1 {
		page = 1
	}

	search := r.URL.Query().Get("search")

	generationGs, err := c.GetGenerationGs(page, GALLERY_PER_PAGE+1, search, "")
	if err != nil {
		log.Error("Error searching meili", "err", err)
		responses.ErrInternalServerError(w, r, "Error querying gallery")
		return
	}
	next := 0
	if len(generationGs) > GALLERY_PER_PAGE {
		next = page + 1
		generationGs = generationGs[:len(generationGs)-1]
	}

	// Shuffle results if no search was specified
	if search == "" {
		// Get seed from query
		seed := r.URL.Query().Get("seed")
		if seed != "" {
			seedInt, err := strconv.Atoi(seed)
			if err != nil {
				log.Error("Error parsing seed", "err", err)
			} else {
				rand.Seed(int64(seedInt))
				rand.Shuffle(
					len(generationGs),
					func(i, j int) { generationGs[i], generationGs[j] = generationGs[j], generationGs[i] },
				)
			}
		}
	}

	// We don't want to leak primary keys, so set to nil
	for i := range generationGs {
		generationGs[i].UserID = nil
	}

	// We want to parse S3 URLs
	for i := range generationGs {
		imageUrl := utils.GetURLFromImagePath(generationGs[i].ImagePath)
		if err != nil {
			log.Error("Error parsing S3 URL", "err", err)
			imageUrl = generationGs[i].ImagePath
		}
		generationGs[i].ImageURL = imageUrl
		generationGs[i].ImagePath = ""
		if generationGs[i].UpscaledImagePath != "" {
			imageUrl := utils.GetURLFromImagePath(generationGs[i].UpscaledImagePath)
			generationGs[i].UpscaledImageURL = imageUrl
			generationGs[i].UpscaledImagePath = ""
		}
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, GalleryResponse[int]{
		Next: next,
		Page: page,
		Hits: generationGs,
	})
}

func (c *RestAPI) HandleSemanticSearchGallery(w http.ResponseWriter, r *http.Request) {
	// Get output_id param
	outputId := r.URL.Query().Get("output_id")
	if outputId != "" {
		// Validate output_id
		uid, err := uuid.Parse(outputId)
		if err != nil {
			responses.ErrBadRequest(w, r, "invalid_output_id", "")
			return
		}

		galleryData, err := c.Repo.RetrieveGalleryDataByID(uid)
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
			Page: 1,
			Hits: []repository.GalleryData{*galleryData},
		})
		return
	}

	search := r.URL.Query().Get("search")
	cursor := r.URL.Query().Get("cursor")
	galleryData := []repository.GalleryData{}
	var nextCursorQdrant *uint
	var nextCursorPostgres *time.Time
	var err error

	// Parse filters
	filters := &requests.QueryGenerationFilters{}
	err = filters.ParseURLQueryParameters(r.URL.Query())
	if err != nil {
		responses.ErrBadRequest(w, r, err.Error(), "")
		return
	}

	// Parse as qdrant filters
	qdrantFilters, scoreThreshold := filters.ToQdrantFilters(true)
	// Append gallery status requirement
	qdrantFilters.Must = append(qdrantFilters.Must, qdrant.SCMatchCondition{
		Key:   "gallery_status",
		Match: &qdrant.SCValue{Value: generationoutput.GalleryStatusApproved},
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
		embeddings, err := c.Clip.GetEmbeddingFromText(search, 3)
		if err != nil {
			log.Error("Error getting embeddings from clip service", "err", err)
			responses.ErrInternalServerError(w, r, "An unknown error occurred")
			return
		}

		res, err := c.Qdrant.QueryGenerations(embeddings, GALLERY_PER_PAGE, offset, scoreThreshold, qdrantFilters, false, false)
		if err != nil {
			log.Error("Error querying qdrant", "err", err)
			responses.ErrInternalServerError(w, r, "An unknown error occurred")
			return
		}

		// Get generation output ids
		var outputIds []uuid.UUID
		var outputIdScoreMap = make(map[uuid.UUID]float32)
		for _, hit := range res.Result {
			outputId, err := uuid.Parse(hit.Id)
			if err != nil {
				log.Error("Error parsing uuid", "err", err)
				continue
			}
			outputIds = append(outputIds, outputId)
			outputIdScoreMap[outputId] = hit.Score
		}

		// Get gallery data
		galleryDataUnsorted, err := c.Repo.RetrieveGalleryDataWithOutputIDs(outputIds)
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
			s := outputIdScoreMap[outputId]
			item.Score = &s
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
		filters.GalleryStatus = []generationoutput.GalleryStatus{generationoutput.GalleryStatusApproved}
		galleryData, nextCursorPostgres, err = c.Repo.RetrieveMostRecentGalleryData(filters, GALLERY_PER_PAGE, qCursor)
		if err != nil {
			log.Error("Error querying gallery data from postgres", "err", err)
			responses.ErrInternalServerError(w, r, "An unknown error occurred")
			return
		}
	}

	// Shuffle results if no search was specified
	if search == "" {
		// Get seed from query
		seed := r.URL.Query().Get("seed")
		if seed != "" {
			seedInt, err := strconv.Atoi(seed)
			if err != nil {
				log.Error("Error parsing seed", "err", err)
			} else {
				rand.Seed(int64(seedInt))
				rand.Shuffle(
					len(galleryData),
					func(i, j int) { galleryData[i], galleryData[j] = galleryData[j], galleryData[i] },
				)
			}
		}
	}

	// We don't want to leak primary keys, so set to nil
	for i := range galleryData {
		galleryData[i].UserID = nil
	}

	if search == "" {
		render.Status(r, http.StatusOK)
		render.JSON(w, r, GalleryResponse[*time.Time]{
			Next: nextCursorPostgres,
			Hits: galleryData,
		})
		return
	}
	render.Status(r, http.StatusOK)
	render.JSON(w, r, GalleryResponse[*uint]{
		Next: nextCursorQdrant,
		Hits: galleryData,
	})
}

type GalleryResponseCursor interface {
	// ! TODO - remove int when meili is gone
	*uint | *time.Time | int
}

type GalleryResponse[T GalleryResponseCursor] struct {
	Next T                        `json:"next,omitempty"`
	Page int                      `json:"page"`
	Hits []repository.GalleryData `json:"hits"`
}

// HTTP PUT submit a generation to gallery - for user
// Only allow submitting user's own gallery items.
func (c *RestAPI) HandleSubmitGenerationToGallery(w http.ResponseWriter, r *http.Request) {
	var user *ent.User
	if user = c.GetUserIfAuthenticated(w, r); user == nil {
		return
	}

	// Parse request body
	reqBody, _ := io.ReadAll(r.Body)
	var submitToGalleryReq requests.SubmitGalleryRequest
	err := json.Unmarshal(reqBody, &submitToGalleryReq)
	if err != nil {
		responses.ErrUnableToParseJson(w, r)
		return
	}

	submitted, err := c.Repo.SubmitGenerationOutputsToGalleryForUser(submitToGalleryReq.GenerationOutputIDs, user.ID)
	if err != nil {
		responses.ErrInternalServerError(w, r, "Error submitting generation outputs to gallery")
		return
	}

	res := responses.SubmitGalleryResponse{
		Submitted: submitted,
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, res)
}
