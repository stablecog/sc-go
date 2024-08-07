package clip

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"golang.org/x/exp/slices"

	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/server/responses"
	"github.com/stablecog/sc-go/server/translator"
	"github.com/stablecog/sc-go/shared"
	"github.com/stablecog/sc-go/utils"
)

type ClipService struct {
	redis *database.RedisWrapper
	// Index for round robin
	activeUrl     int
	urls          []string
	r             http.RoundTripper
	secret        string
	client        *http.Client
	mu            sync.RWMutex
	SafetyChecker *translator.TranslatorSafetyChecker
	apiUrl        string
	apiAuthToken  string
}

func (c *ClipService) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Add("Authorization", c.secret)
	r.Header.Add("Content-Type", "application/json")
	return c.r.RoundTrip(r)
}

func NewClipService(redis *database.RedisWrapper, safetyChecker *translator.TranslatorSafetyChecker) *ClipService {
	svc := &ClipService{
		secret:        utils.GetEnv().ClipAPISecret,
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

func (c *ClipService) UpdateURLsFromCache() {
	urls := shared.GetCache().GetClipUrls()
	// Compare existing slice
	if len(urls) == 0 {
		return
	}
	if slices.Compare(urls, c.urls) != 0 {
		c.mu.Lock()
		defer c.mu.Unlock()
		c.urls = urls
	}
}

// GetEmbeddingFromText, retry up to retries times
func (c *ClipService) GetEmbeddingFromText(text string, retries int, translate bool) (embedding []float32, err error) {
	c.UpdateURLsFromCache()
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

	req := []clipApiRequest{{
		Text: textTranslated,
	}}

	// Http POST to endpoint with secret
	// Marshal req
	b, err := json.Marshal(req)
	if err != nil {
		log.Errorf("Error marshalling req %v", err)
		return nil, err
	}
	c.mu.RLock()
	if len(c.urls) == 0 {
		c.mu.RUnlock()
		return nil, errors.New("no URLs available")
	}
	// Ensure activeUrl isn't out of range
	if c.activeUrl >= len(c.urls) {
		c.activeUrl = 0
	}
	url := c.urls[c.activeUrl]
	c.mu.RUnlock()
	request, _ := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	// Do
	resp, err := c.client.Do(request)
	if err != nil {
		log.Errorf("Error getting response from clip api %v", err)
		if retries <= 0 {
			return nil, err
		}
		// Set next active index
		c.mu.Lock()
		c.activeUrl = (c.activeUrl + 1) % len(c.urls)
		c.mu.Unlock()
		return c.GetEmbeddingFromText(text, retries-1, translate)
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
	var clipAPIResponse responses.EmbeddingsResponse
	err = json.Unmarshal(readAll, &clipAPIResponse)
	if err != nil {
		log.Errorf("Error unmarshalling resp %v", err)
		return nil, err
	}

	if len(clipAPIResponse.Embeddings) == 0 {
		log.Errorf("No embeddings returned from clip API")
		return nil, fmt.Errorf("no embeddings returned from clip API")
	}

	// Cache
	err = c.redis.CacheEmbeddings(c.redis.Ctx, utils.Sha256(textTranslated), clipAPIResponse.Embeddings[0].Embedding)
	if err != nil {
		log.Errorf("Error caching embeddings %v", err)
	}

	c.mu.Lock()
	c.activeUrl = (c.activeUrl + 1) % len(c.urls)
	c.mu.Unlock()

	return clipAPIResponse.Embeddings[0].Embedding, nil
}

func (c *ClipService) GetEmbeddingsV2(toEmbedObjects []EmbeddingReqObject) (embeddings []EmbeddingResObject, err error) {
	s := time.Now()
	var req []CLIPAPIEmbeddingReqObject = []CLIPAPIEmbeddingReqObject{}
	for _, obj := range toEmbedObjects {
		if obj.Text != "" {
			req = append(req, CLIPAPIEmbeddingReqObject{
				Text: obj.Text,
			})
		} else if obj.Image != "" {
			req = append(req, CLIPAPIEmbeddingReqObject{
				Image:          obj.Image,
				CalculateScore: obj.CalculateScore,
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
	request.Header.Add("Authorization", c.apiAuthToken)

	log.Infof("[] Sending request to URL: %s", url)
	log.Infof("[] Request body: %s", string(b))
	log.Infof("[] Request headers: %v", request.Header)

	resp, err := http.DefaultClient.Do(request)
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

	var clipAPIResponse CLIPAPIEmbeddingRes
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
			Embedding: embedding.Embedding,
			Input:     input,
		})
	}

	duration := time.Since(s)
	log.Info("[] Successfully got %d embedding(s): %dms", len(result), duration.Milliseconds())
	return result, nil
}

// GetEmbeddingFromImagePath, retry up to retries times
func (c *ClipService) GetEmbeddingFromImagePath(imagePath string, noCache bool, retries int) (embedding []float32, err error) {
	c.UpdateURLsFromCache()
	// Check cache first
	if !noCache {
		e, err := c.redis.GetEmbeddings(c.redis.Ctx, utils.Sha256(imagePath))
		if err == nil && len(e) > 0 {
			return e, nil
		}
	}

	req := []clipApiRequest{{
		ImageID: imagePath,
	}}

	// Http POST to endpoint with secret
	// Marshal req
	b, err := json.Marshal(req)
	if err != nil {
		log.Errorf("Error marshalling req %v", err)
		return nil, err
	}
	c.mu.RLock()
	url := c.urls[c.activeUrl]
	c.mu.RUnlock()
	request, _ := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	// Do
	resp, err := c.client.Do(request)
	if err != nil {
		log.Errorf("Error getting response from clip api %v", err)
		if retries <= 0 {
			return nil, err
		}
		// Set next active index
		c.mu.Lock()
		c.activeUrl = (c.activeUrl + 1) % len(c.urls)
		c.mu.Unlock()
		return c.GetEmbeddingFromImagePath(imagePath, noCache, retries-1)
	}
	defer resp.Body.Close()

	readAll, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Error reading resposne body in clip API %v", err)
		return nil, err
	}
	var clipAPIResponse responses.EmbeddingsResponse
	err = json.Unmarshal(readAll, &clipAPIResponse)
	if err != nil {
		log.Errorf("Error unmarshalling resp %v", err)
		return nil, err
	}

	if len(clipAPIResponse.Embeddings) == 0 {
		log.Errorf("No embeddings returned from clip API")
		return nil, fmt.Errorf("no embeddings returned from clip API")
	}

	// Cache
	if !noCache {
		err = c.redis.CacheEmbeddings(c.redis.Ctx, utils.Sha256(imagePath), clipAPIResponse.Embeddings[0].Embedding)
		if err != nil {
			log.Errorf("Error caching embeddings %v", err)
		}
	}

	c.mu.Lock()
	c.activeUrl = (c.activeUrl + 1) % len(c.urls)
	c.mu.Unlock()

	return clipAPIResponse.Embeddings[0].Embedding, nil
}

type clipApiRequest struct {
	Text    string `json:"text,omitempty"`
	Image   string `json:"image,omitempty"`
	ImageID string `json:"image_id,omitempty"`
}

type embeddingObject struct {
	Embedding      []float32 `json:"embedding"`
	InputText      string    `json:"input_text"`
	TranslatedText string    `json:"translated_text,omitempty"`
	ID             uuid.UUID `json:"id,omitempty"`
	Error          string    `json:"error,omitempty"`
}

type embeddingsResponse struct {
	Embeddings []embeddingObject `json:"embeddings"`
}

type CLIPAPIEmbeddingReqObject struct {
	Text           string `json:"text,omitempty"`
	Image          string `json:"image,omitempty"`
	CalculateScore bool   `json:"calculate_score,omitempty"`
}

type CLIPAPIEmbeddingRes struct {
	Embeddings []EmbeddingItem `json:"embeddings"`
}

type EmbeddingItem struct {
	AestheticScore *AestheticScore `json:"aesthetic_score,omitempty"`
	Embedding      []float32       `json:"embedding"`
	InputImage     string          `json:"input_image,omitempty"`
	InputText      string          `json:"input_text,omitempty"`
}

type AestheticScore struct {
	Artifact float32 `json:"artifact"`
	Rating   float32 `json:"rating"`
}

type EmbeddingReqObject struct {
	Text           string `json:"text,omitempty"`
	Image          string `json:"image,omitempty"`
	CalculateScore bool   `json:"calculate_score,omitempty"`
}

type EmbeddingResObject struct {
	Input     EmbeddingReqObject `json:"input"`
	Embedding []float32          `json:"embedding"`
}
