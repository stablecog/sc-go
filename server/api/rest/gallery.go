package rest

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/render"
	"github.com/meilisearch/meilisearch-go"
	"github.com/stablecog/sc-go/database/repository"
	"github.com/stablecog/sc-go/server/responses"
	"github.com/stablecog/sc-go/utils"
	"k8s.io/klog/v2"
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
		klog.Errorf("Error searching for generation_g: %v", err)
		return nil, err
	}
	if len(res.Hits) == 0 {
		return generationGs, nil
	}
	for _, hit := range res.Hits {
		j, err := json.Marshal(hit)
		if err != nil {
			klog.Errorf("Error marshalling hit: %v", err)
			return nil, err
		}
		var gen repository.GalleryData
		err = json.Unmarshal(j, &gen)
		if err != nil {
			klog.Errorf("Error unmarshalling hit: %v", err)
			return nil, err
		}
		generationGs = append(generationGs, gen)
	}
	return generationGs, nil
}

func (c *RestAPI) HandleQueryGallery(w http.ResponseWriter, r *http.Request) {
	// Get query params
	page, err := strconv.Atoi(r.URL.Query().Get("cursor"))
	if err != nil || page < 1 {
		page = 1
	}

	search := r.URL.Query().Get("search")

	generationGs, err := c.GetGenerationGs(page, GALLERY_PER_PAGE+1, search, "")
	if err != nil {
		klog.Errorf("Error searching meili: %v", err)
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
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(
			len(generationGs),
			func(i, j int) { generationGs[i], generationGs[j] = generationGs[j], generationGs[i] },
		)
	}

	// We don't want to leak primary keys, so set to nil
	for i := range generationGs {
		generationGs[i].UserID = nil
	}

	// We want to parse S3 URLs
	for i := range generationGs {
		imageUrl := utils.GetURLFromImagePath(generationGs[i].ImagePath)
		if err != nil {
			klog.Errorf("Error parsing S3 URL: %v", err)
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
	render.JSON(w, r, GalleryResponse{
		Next: next,
		Page: page,
		Hits: generationGs,
	})
}

type GalleryResponse struct {
	Next int                      `json:"next"`
	Page int                      `json:"page"`
	Hits []repository.GalleryData `json:"hits"`
}
