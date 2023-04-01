package rest

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"

	"entgo.io/ent/dialect/sql"
	"github.com/go-chi/render"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"github.com/pgvector/pgvector-go"
	"github.com/stablecog/sc-go/database"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/ent/generationoutput"
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

	res, err := c.Repo.DB.GenerationOutput.Query().Select(generationoutput.FieldImagePath).
		WithGenerations(func(gq *ent.GenerationQuery) {
			gq.WithPrompt()
		}).
		Order(func(s *sql.Selector) {
			s.OrderExpr(sql.Expr("embedding <-> ?", pgvector.NewVector(vector)), sql.Expr("created_at DESC"))
		}).Limit(50).All(r.Context())
	if err != nil {
		log.Errorf("Error searching %v", err)
		responses.ErrBadRequest(w, r, err.Error(), "")
		return
	}

	response := MilvusResponse{
		TranslatedText: clipAPIResponse.Embeddings[0].TranslatedText,
		InputText:      clipAPIResponse.Embeddings[0].InputText,
	}

	for _, v := range res {
		response.Data = append(response.Data, MilvusData{
			Image:  v.ImagePath,
			Prompt: v.Edges.Generations.Edges.Prompt.Text,
		})
	}

	render.Status(r, resp.StatusCode)
	render.JSON(w, r, response)
}
