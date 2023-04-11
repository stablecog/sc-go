package clip

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/server/responses"
	"github.com/stablecog/sc-go/utils"
	"golang.org/x/exp/slices"
)

type badUrl struct {
	url      string
	markedAt time.Time
}

type ClipService struct {
	redis *database.RedisWrapper
	// Index for round robin
	activeUrl int
	badUrls   []badUrl
	urls      []string
	r         http.RoundTripper
	secret    string
	client    *http.Client
	mu        sync.Mutex
}

func (c *ClipService) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Add("Authorization", c.secret)
	r.Header.Add("Content-Type", "application/json")
	return c.r.RoundTrip(r)
}

// Do a round-robin style request
func (c *ClipService) getActiveUrl() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	active := c.urls[c.activeUrl]
	c.activeUrl++
	if c.activeUrl >= len(c.urls) {
		c.activeUrl = 0
	}
	return active
}

// Mark a URL as not working
func (c *ClipService) markUrlBad(url string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	var badUrlStr string
	for _, u := range c.urls {
		if u == url {
			c.badUrls = append(c.badUrls, badUrl{
				url:      u,
				markedAt: time.Now(),
			})
			badUrlStr = u
		}
	}
	if badUrlStr != "" {
		// Remove from active urls
		urls := make([]string, len(c.urls)-1)
		for _, u := range c.urls {
			if u != badUrlStr {
				urls = append(urls, u)
			}
		}
		c.urls = urls
	}
}

// Unmark all bad URLs that are older than 5 minutes
func (c *ClipService) unmarkBadUrls() {
	if len(c.badUrls) == 0 {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	toDelete := make([]string, 0)
	for _, t := range c.badUrls {
		if time.Since(t.markedAt) > 5*time.Minute {
			c.urls = append(c.urls, t.url)
			toDelete = append(toDelete, t.url)
		}
	}
	var newBadUrls []badUrl
	for _, t := range c.badUrls {
		if !slices.Contains(toDelete, t.url) {
			newBadUrls = append(newBadUrls, t)
		}
	}
	c.badUrls = newBadUrls
}

func NewClipService(redis *database.RedisWrapper) *ClipService {
	svc := &ClipService{
		urls:   strings.Split(os.Getenv("CLIPAPI_URLS"), ","),
		secret: os.Getenv("CLIPAPI_SECRET"),
		r:      http.DefaultTransport,
		redis:  redis,
	}
	svc.client = &http.Client{
		Timeout:   10 * time.Second,
		Transport: svc,
	}
	return svc
}

// GetEmbeddingFromText, retry up to retries times
func (c *ClipService) GetEmbeddingFromText(text string, retries int) (embedding []float32, err error) {
	// Check cache first
	e, err := c.redis.GetEmbeddings(c.redis.Ctx, utils.Sha256(text))
	if err == nil && len(e) > 0 {
		return e, nil
	}
	c.unmarkBadUrls()

	req := []clipApiRequest{{
		Text: text,
	}}

	// Http POST to endpoint with secret
	// Marshal req
	b, err := json.Marshal(req)
	if err != nil {
		log.Errorf("Error marshalling req %v", err)
		return nil, err
	}
	url := c.getActiveUrl()
	request, _ := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	// Do
	resp, err := c.client.Do(request)
	if err != nil {
		if os.IsTimeout(err) || strings.Contains(err.Error(), "connection refused") {
			// Mark this URL as bad
			if len(c.urls) > 1 {
				c.markUrlBad(url)
			}
		}
		log.Errorf("Error getting response from clip api %v", err)
		if retries <= 0 {
			return nil, err
		}
		return c.GetEmbeddingFromText(text, retries-1)
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
