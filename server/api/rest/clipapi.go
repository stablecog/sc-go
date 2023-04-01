package rest

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/go-chi/render"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"github.com/stablecog/sc-go/database"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/server/responses"
)

func (c *RestAPI) HandleClipSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")

	if query == "" {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, []float32{})
		return
	}

	req := []requests.ClipAPIRequest{{
		Text: query,
	}}

	secret := os.Getenv("CLIPAPI_SECRET")
	endpoint := os.Getenv("CLIPAPI_ENDPOINT")

	// Http POST to endpoint with secret
	// Marshal req
	b, err := json.Marshal(req)
	if err != nil {
		log.Errorf("Error marshalling req %v", err)
		responses.ErrBadRequest(w, r, err.Error(), "")
		return
	}
	request, _ := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(b))
	request.Header.Set("Authorization", secret)
	request.Header.Set("Content-Type", "application/json")
	// Do
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Errorf("Error making request %v", err)
		responses.ErrBadRequest(w, r, err.Error(), "")
		return
	}
	defer resp.Body.Close()

	readAll, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error(err)
		responses.ErrBadRequest(w, r, err.Error(), "")
		return
	}
	var clipAPIResponse responses.EmbeddingsResponse
	err = json.Unmarshal(readAll, &clipAPIResponse)
	if err != nil {
		log.Errorf("Error unmarshalling resp %v", err)
		responses.ErrBadRequest(w, r, err.Error(), "")
		return
	}

	if len(clipAPIResponse.Embeddings) == 0 {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, []float32{})
		return
	}

	sp, _ := entity.NewIndexHNSWSearchParam(256)
	vec2search := []entity.Vector{
		entity.FloatVector(clipAPIResponse.Embeddings[0].Embedding),
	}
	res, err := c.Milvus.Client.Search(c.Milvus.Ctx, database.MILVUS_COLLECTION_NAME, []string{}, "", []string{"image_path"}, vec2search, "image_embedding", entity.IP, 50, sp, client.WithSearchQueryConsistencyLevel(entity.ClSession), client.WithLimit(50), client.WithOffset(0))
	if err != nil {
		log.Errorf("Error searching %v", err)
		responses.ErrBadRequest(w, r, err.Error(), "")
		return
	}

	var response []string
	for _, v := range res {
		for _, v2 := range v.Fields {
			response = v2.FieldData().GetScalars().GetStringData().Data
		}
	}

	// Return as-is
	render.Status(r, resp.StatusCode)
	render.JSON(w, r, response)
}

type MilvusResponse struct {
	Data []string `json:"data"`
}
