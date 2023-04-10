package rest

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/go-chi/render"
	"github.com/stablecog/sc-go/database/ent/generationoutput"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/server/responses"
)

func (c *RestAPI) HandleClipQSearch(w http.ResponseWriter, r *http.Request) {
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

	qAPIResponse, err := c.Qdrant.Query(clipAPIResponse.Embeddings[0].Embedding, false)
	if err != nil {
		log.Errorf("Error querying qdrant %v", err)
		responses.ErrBadRequest(w, r, err.Error(), "")
		return
	}

	response := MilvusResponse{
		TranslatedText: clipAPIResponse.Embeddings[0].TranslatedText,
		InputText:      clipAPIResponse.Embeddings[0].InputText,
	}

	response.Data = make([]MilvusData, len(qAPIResponse.Result))
	for i := range qAPIResponse.Result {
		response.Data[i] = MilvusData{
			Image:             qAPIResponse.Result[i].Payload.ImagePath,
			Prompt:            qAPIResponse.Result[i].Payload.Prompt,
			UpscaledImagePath: qAPIResponse.Result[i].Payload.UpscaledImagePath,
			GalleryStatus:     qAPIResponse.Result[i].Payload.GalleryStatus,
			IsFavorited:       qAPIResponse.Result[i].Payload.IsFavorited,
		}
	}

	render.Status(r, resp.StatusCode)
	render.JSON(w, r, response)
}

type MilvusData struct {
	Image             string                         `json:"image"`
	Prompt            string                         `json:"prompt"`
	UpscaledImagePath string                         `json:"upscaled_image_path,omitempty"`
	GalleryStatus     generationoutput.GalleryStatus `json:"gallery_status"`
	IsFavorited       bool                           `json:"is_favorited"`
}

type MilvusResponse struct {
	Data           []MilvusData `json:"data"`
	TranslatedText string       `json:"translated_text,omitempty"`
	InputText      string       `json:"input_text"`
}
