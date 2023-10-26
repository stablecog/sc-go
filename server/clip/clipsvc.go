package clip

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/server/responses"
	"github.com/stablecog/sc-go/utils"
)

type badUrl struct {
	url      string
	markedAt time.Time
}

type ClipService struct {
	redis *database.RedisWrapper
	// Index for round robin
	activeUrl     int
	urls          []string
	r             http.RoundTripper
	secret        string
	client        *http.Client
	mu            sync.RWMutex
	SafetyChecker *utils.TranslatorSafetyChecker
}

func (c *ClipService) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Add("Authorization", c.secret)
	r.Header.Add("Content-Type", "application/json")
	return c.r.RoundTrip(r)
}

func NewClipService(redis *database.RedisWrapper, safetyChecker *utils.TranslatorSafetyChecker) *ClipService {
	svc := &ClipService{
		urls:          utils.GetEnv().ClipAPIURLs,
		secret:        utils.GetEnv().ClipAPISecret,
		r:             http.DefaultTransport,
		redis:         redis,
		SafetyChecker: safetyChecker,
	}
	svc.client = &http.Client{
		Timeout:   10 * time.Second,
		Transport: svc,
	}
	return svc
}

// GetEmbeddingFromText, retry up to retries times
func (c *ClipService) GetEmbeddingFromText(text string, retries int, translate bool) (embedding []float32, err error) {
	// Check cache first
	e, err := c.redis.GetEmbeddings(c.redis.Ctx, utils.Sha256(text))
	if err == nil && len(e) > 0 {
		return e, nil
	}

	// Translate text
	textTranslated := text
	if translate {
		textTranslated, _, err = c.SafetyChecker.TranslatePrompt(text, "")
		if err != nil {
			log.Errorf("Error translating text %v", err)
			return nil, err
		}
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
	err = c.redis.CacheEmbeddings(c.redis.Ctx, utils.Sha256(text), clipAPIResponse.Embeddings[0].Embedding)
	if err != nil {
		log.Errorf("Error caching embeddings %v", err)
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
