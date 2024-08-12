package clip

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/stablecog/sc-go/database"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/server/responses"
	"github.com/stablecog/sc-go/server/translator"
	"github.com/stablecog/sc-go/utils"
)

type ClipService struct {
	redis *database.RedisWrapper
	// Index for round robin
	r             http.RoundTripper
	client        *http.Client
	SafetyChecker *translator.TranslatorSafetyChecker
	apiUrl        string
	apiAuthToken  string
}

func (c *ClipService) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Add("Authorization", c.apiAuthToken)
	r.Header.Add("Content-Type", "application/json")
	return c.r.RoundTrip(r)
}

func NewClipService(redis *database.RedisWrapper, safetyChecker *translator.TranslatorSafetyChecker) *ClipService {
	svc := &ClipService{
		r:             http.DefaultTransport,
		redis:         redis,
		SafetyChecker: safetyChecker,
		apiUrl:        utils.GetEnv().ClipApiUrl,
		apiAuthToken:  utils.GetEnv().ClipApiAuthToken,
	}
	svc.client = &http.Client{
		Timeout:   10 * time.Second,
		Transport: svc,
	}
	return svc
}

// GetEmbeddingFromText, retry up to retries times
func (c *ClipService) GetEmbeddingFromText(text string, translate bool) (embedding []float32, err error) {
	// Translate text
	textTranslated := text
	if translate {
		textTranslated, _, err = c.SafetyChecker.TranslatePrompt(text, "")
		if err != nil {
			log.Errorf("Error translating text %v", err)
			return nil, err
		}
	}

	// Check cache first
	e, err := c.redis.GetEmbeddings(c.redis.Ctx, utils.Sha256(textTranslated))
	if err == nil && len(e) > 0 {
		return e, nil
	}

	req := []requests.ClipApiEmbeddingRequest{{
		Text: textTranslated,
	}}

	// Http POST to endpoint with secret
	// Marshal req
	b, err := json.Marshal(req)
	if err != nil {
		log.Errorf("Error marshalling req %v", err)
		return nil, err
	}
	url := c.apiUrl + "/embed"
	request, _ := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	// Do
	resp, err := c.client.Do(request)
	if err != nil {
		log.Errorf("Error getting response from clip api %v", err)
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Error("Error sending clip request", "status", resp.StatusCode, "url", url, "response", resp.Body)
		return nil, errors.New("clip request failed")
	}

	readAll, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Error reading resposne body in clip API %v", err)
		return nil, err
	}
	var clipAPIResponse responses.ClipEmbeddingResponse
	err = json.Unmarshal(readAll, &clipAPIResponse)
	if err != nil {
		log.Errorf("Error unmarshalling resp %v", err)
		return nil, err
	}

	if len(clipAPIResponse.Embeddings) == 0 {
		log.Errorf("No embeddings returned from clip API")
		return nil, fmt.Errorf("no embeddings returned from clip API")
	}

	embed := clipAPIResponse.Embeddings[0].Embedding

	// Cache
	err = c.redis.CacheEmbeddings(c.redis.Ctx, utils.Sha256(textTranslated), embed)
	if err != nil {
		log.Errorf("Error caching embeddings %v", err)
	}

	return embed, nil
}

func (c *ClipService) GetEmbeddings(toEmbedObjects []EmbeddingReqObject) (embeddings []EmbeddingResObject, err error) {
	s := time.Now()
	var req []requests.ClipApiEmbeddingRequest = []requests.ClipApiEmbeddingRequest{}
	for _, obj := range toEmbedObjects {
		if obj.Text != "" {
			req = append(req, requests.ClipApiEmbeddingRequest{
				Text: obj.Text,
			})
		} else if obj.Image != "" {
			req = append(req, requests.ClipApiEmbeddingRequest{
				Image:          obj.Image,
				CalculateScore: obj.CalculateScore,
				CheckNsfw:      obj.CheckNsfw,
			})
		}
	}

	b, err := json.Marshal(req)
	if err != nil {
		log.Errorf("[] Error marshalling req: %v", err)
		return nil, err
	}

	url := c.apiUrl + "/embed"
	request, _ := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))

	resp, err := c.client.Do(request)
	if err != nil {
		log.Errorf("[] Error sending request: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("[] Error reading response body: %v", err)
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		log.Error(
			"[] Error from CLIP API",
			"status", resp.StatusCode,
			"url", url,
			"response", string(bodyBytes),
		)
		return nil, fmt.Errorf("CLIP API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var clipAPIResponse responses.ClipEmbeddingResponse
	err = json.Unmarshal(bodyBytes, &clipAPIResponse)
	if err != nil {
		log.Errorf("[] Error unmarshalling response: %v", err)
		return nil, err
	}

	if len(clipAPIResponse.Embeddings) != len(toEmbedObjects) {
		log.Errorf("[] Mismatch in number of embeddings returned: %d vs %d", len(clipAPIResponse.Embeddings), len(toEmbedObjects))
		return nil, fmt.Errorf("mismatch in number of embeddings returned from CLIP API")
	}

	var result []EmbeddingResObject
	for i, embedding := range clipAPIResponse.Embeddings {
		input := toEmbedObjects[i]
		result = append(result, EmbeddingResObject{
			Embedding:      embedding.Embedding,
			Input:          input,
			AestheticScore: embedding.AestheticScore,
			NsfwScore:      embedding.NsfwScore,
		})
	}

	duration := time.Since(s)
	log.Infof("[] Successfully got %d embedding(s): %dms", len(result), duration.Milliseconds())
	return result, nil
}

func (c *ClipService) GetNsfwScores(imageUrls []string) (scores []float32, err error) {
	s := time.Now()
	var req []string
	for _, url := range imageUrls {
		req = append(req, url)
	}

	b, err := json.Marshal(req)
	if err != nil {
		log.Errorf("👙 Error marshalling req: %v", err)
		return nil, err
	}

	url := c.apiUrl + "/nsfw-check"
	request, _ := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))

	resp, err := c.client.Do(request)
	if err != nil {
		log.Errorf("👙 Error sending request: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("👙 Error reading response body: %v", err)
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		log.Error(
			"👙 Error from CLIP API",
			"status", resp.StatusCode,
			"url", url,
			"response", string(bodyBytes),
		)
		return nil, fmt.Errorf("CLIP API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var clipAPIResponse responses.ClipNsfwCheckResponse
	err = json.Unmarshal(bodyBytes, &clipAPIResponse)
	if err != nil {
		log.Errorf("👙 Error unmarshalling response: %v", err)
		return nil, err
	}

	if len(clipAPIResponse.Data) != len(imageUrls) {
		log.Errorf("👙 Mismatch in number of NSFW scores returned: %d vs %d", len(clipAPIResponse.Data), len(imageUrls))
		return nil, fmt.Errorf("mismatch in number of NSFW scores returned from CLIP API")
	}

	var result []float32
	for _, item := range clipAPIResponse.Data {
		result = append(result, item.NsfwScore.Nsfw)
	}

	duration := time.Since(s)
	log.Infof("👙 Successfully got %d NSFW score(s): %dms", len(result), duration.Milliseconds())
	return result, nil
}

type EmbeddingReqObject struct {
	Text           string `json:"text,omitempty"`
	Image          string `json:"image,omitempty"`
	CalculateScore bool   `json:"calculate_score,omitempty"`
	CheckNsfw      bool   `json:"check_nsfw,omitempty"`
}

type EmbeddingResObject struct {
	Input          EmbeddingReqObject            `json:"input"`
	Embedding      []float32                     `json:"embedding"`
	AestheticScore *responses.ClipAestheticScore `json:"aesthetic_score,omitempty"`
	NsfwScore      *responses.NsfwCheckScore     `json:"nsfw_score,omitempty"`
}
