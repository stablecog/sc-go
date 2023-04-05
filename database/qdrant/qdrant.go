package qdrant

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/utils"
)

type QDrantClient struct {
	ActiveUrl string
	URLs      []string
	token     string
	r         http.RoundTripper
	Client    *ClientWithResponses
	Ctx       context.Context
}

func (q QDrantClient) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Add("Authorization", "Basic "+q.token)
	return q.r.RoundTrip(r)
}

func NewQdrantClient(ctx context.Context) (*QDrantClient, error) {
	// Get URLs from env, comma separated
	urlEnv := os.Getenv("QDRANT_URLS")
	if urlEnv == "" {
		log.Errorf("QDRANT_URLS not set")
		return nil, errors.New("QDRANT_URLS not set")
	}
	// Split by comma
	urls := strings.Split(urlEnv, ",")
	// Token
	auth := os.Getenv("QDRANT_USERNAME") + ":" + os.Getenv("QDRANT_PASSWORD")
	// Create client
	qClient := &QDrantClient{
		URLs:      urls,
		ActiveUrl: urls[0],
		token:     base64.StdEncoding.EncodeToString([]byte(auth)),
		r:         http.DefaultTransport,
		Ctx:       ctx,
	}

	c, err := NewClientWithResponses(qClient.ActiveUrl, WithHTTPClient(&http.Client{
		Timeout:   10 * time.Second,
		Transport: qClient,
	}))
	if err != nil {
		log.Errorf("Error creating qdrant client %v", err)
		return nil, err
	}
	qClient.Client = c

	return qClient, nil
}

// Update the client if the active url is not responding
func (q *QDrantClient) UpdateActiveClient() error {
	var targetUrl string
	for _, url := range q.URLs {
		if url != q.ActiveUrl {
			targetUrl = url
			break
		}
	}
	if targetUrl == "" {
		log.Errorf("No other urls to try")
		return errors.New("No other urls to try")
	}

	q.ActiveUrl = targetUrl

	c, err := NewClientWithResponses(q.ActiveUrl, WithHTTPClient(&http.Client{
		Timeout:   10 * time.Second,
		Transport: q,
	}))
	if err != nil {
		log.Errorf("Error creating qdrant client %v", err)
		return err
	}

	q.Client = c
	return nil
}

func (q *QDrantClient) GetCollections(noRetry bool) (*CollectionsResponse, error) {
	resp, err := q.Client.GetCollectionsWithResponse(q.Ctx)
	if err != nil {
		if !noRetry && (os.IsTimeout(err) || strings.Contains(err.Error(), "connection refused")) {
			err = q.UpdateActiveClient()
			if err == nil {
				return q.GetCollections(true)
			}
		}
		log.Errorf("Error getting collections %v", err)
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		log.Errorf("Error getting collections %v", resp.StatusCode())
		return nil, errors.New("Error getting collections " + string(resp.Body))
	}
	return resp.JSON200.Result, nil
}

func (q *QDrantClient) CreateCollection(name string, noRetry bool) error {
	// create optimizers config
	optimizersConfig := &CreateCollection_OptimizersConfig{}
	err := optimizersConfig.FromOptimizersConfigDiff(OptimizersConfigDiff{
		MemmapThreshold: utils.ToPtr[uint](20000),
	})
	if err != nil {
		log.Errorf("Error creating optimizers config %v", err)
		return err
	}

	// create quantization config
	quantizationConfig := QuantizationConfig{}
	err = quantizationConfig.FromScalarQuantization(ScalarQuantization{
		Scalar: ScalarQuantizationConfig{
			AlwaysRam: utils.ToPtr(false),
			Quantile:  utils.ToPtr[float32](0.99),
			Type:      ScalarType("int8"),
		},
	})
	if err != nil {
		log.Errorf("Error creating quantization config %v", err)
		return err
	}
	createCollectionQuantizationConfig := &CreateCollection_QuantizationConfig{}
	err = createCollectionQuantizationConfig.FromQuantizationConfig(quantizationConfig)
	if err != nil {
		log.Errorf("Error creating create collection quantization config %v", err)
		return err
	}

	// Create vectors config
	vectorsConfig := VectorsConfig{}
	err = vectorsConfig.FromVectorParams(VectorParams{
		Size:     uint64(1024),
		Distance: "Cosine",
	})
	if err != nil {
		log.Errorf("Error creating vectors config %v", err)
		return err
	}

	test := CreateCollection{
		OptimizersConfig:   optimizersConfig,
		QuantizationConfig: createCollectionQuantizationConfig,
		Vectors:            vectorsConfig,
	}
	// Marshal and print as json
	json, err := json.Marshal(test)
	if err != nil {
		log.Errorf("Error marshalling json %v", err)
		return err
	}
	log.Infof(string(json))

	resp, err := q.Client.CreateCollectionWithResponse(q.Ctx, name, &CreateCollectionParams{}, CreateCollection{
		OptimizersConfig:   optimizersConfig,
		QuantizationConfig: createCollectionQuantizationConfig,
		Vectors:            vectorsConfig,
	})

	if err != nil {
		if !noRetry && (os.IsTimeout(err) || strings.Contains(err.Error(), "connection refused")) {
			err = q.UpdateActiveClient()
			if err == nil {
				return q.CreateCollection(name, true)
			}
		}
		log.Errorf("Error getting collections %v", err)
		return err
	}
	if resp.StatusCode() != http.StatusOK {
		log.Errorf("Error getting collections %v", resp.StatusCode())
		return errors.New("Error getting collections " + string(resp.Body))
	}

	return nil
}
