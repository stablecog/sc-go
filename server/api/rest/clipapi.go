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
	// Get Authorization header
	auth := r.Header.Get("Authorization")
	if auth != os.Getenv("CLIPAPI_SECRET") {
		responses.ErrUnauthorized(w, r)
		return
	}

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

	sp, _ := entity.NewIndexHNSWSearchParam(128)
	vec2search := []entity.Vector{
		entity.FloatVector(clipAPIResponse.Embeddings[0].Embedding),
	}
	res, err := c.Milvus.Client.Search(c.Milvus.Ctx, database.MILVUS_COLLECTION_NAME, []string{}, "", []string{"prompt_text", "image_path"}, vec2search, "image_embedding", entity.IP, 50, sp, client.WithSearchQueryConsistencyLevel(entity.ClSession), client.WithLimit(50), client.WithOffset(0))
	if err != nil {
		log.Errorf("Error searching %v", err)
		responses.ErrBadRequest(w, r, err.Error(), "")
		return
	}

	response := MilvusResponse{
		TranslatedText: clipAPIResponse.Embeddings[0].TranslatedText,
		InputText:      clipAPIResponse.Embeddings[0].InputText,
	}

	var promptData []string
	var imageData []string
	for _, v := range res {
		for _, v2 := range v.Fields {
			if v2.Name() == "prompt_text" {
				promptData = v2.FieldData().GetScalars().GetStringData().Data
			}
			if v2.Name() == "image_path" {
				imageData = v2.FieldData().GetScalars().GetStringData().Data
			}
		}
	}

	// Combine prompt and image data as if they correlate to each other
	response.Data = make([]MilvusData, len(promptData))
	for i := range promptData {
		response.Data[i] = MilvusData{
			Image:  imageData[i],
			Prompt: promptData[i],
		}
	}

	render.Status(r, resp.StatusCode)
	render.JSON(w, r, response)
}

type MilvusData struct {
	Image  string `json:"image"`
	Prompt string `json:"prompt"`
}

type MilvusResponse struct {
	Data           []MilvusData `json:"data"`
	TranslatedText string       `json:"translated_text,omitempty"`
	InputText      string       `json:"input_text"`
}

func (c *RestAPI) HandleClipSearchPGVector(w http.ResponseWriter, r *http.Request) {
	// Get Authorization header
	auth := r.Header.Get("Authorization")
	if auth != os.Getenv("CLIPAPI_SECRET") {
		responses.ErrUnauthorized(w, r)
		return
	}
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

	vector := clipAPIResponse.Embeddings[0].Embedding

	res, err := c.Weaviate.SearchNearVector(vector)
	if err != nil {
		log.Errorf("Error searching %v", err)
		responses.ErrBadRequest(w, r, err.Error(), "")
		return
	}

	jObj, ok := res["Get"]
	if !ok {
		log.Error("Error getting Get from response")
		render.Status(r, http.StatusOK)
		render.JSON(w, r, []float32{})
		return
	}
	jObjInf, ok := jObj.(map[string]interface{})
	if !ok {
		log.Error("Error converting Get to map[string]interface{}")
		render.Status(r, http.StatusOK)
		render.JSON(w, r, []float32{})
		return
	}
	items, ok := jObjInf["Test"]
	if !ok {
		log.Error("Error getting Test from response")
		render.Status(r, http.StatusOK)
		render.JSON(w, r, []float32{})
		return
	}
	itmsList, ok := items.([]interface{})
	if !ok {
		log.Error("Error converting items to map[string]string")
		render.Status(r, http.StatusOK)
		render.JSON(w, r, []float32{})
		return
	}

	response := MilvusResponse{
		TranslatedText: clipAPIResponse.Embeddings[0].TranslatedText,
		InputText:      clipAPIResponse.Embeddings[0].InputText,
	}

	for _, v := range itmsList {
		response.Data = append(response.Data, MilvusData{
			Image:  v.(map[string]interface{})["image_path"].(string),
			Prompt: v.(map[string]interface{})["prompt"].(string),
		})
	}

	render.Status(r, resp.StatusCode)
	render.JSON(w, r, res)
}
