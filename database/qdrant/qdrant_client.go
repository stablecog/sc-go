package qdrant

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/utils"
)

type QDrantClient struct {
	ActiveUrl      string
	URLs           []string
	token          string
	r              http.RoundTripper
	Client         *ClientWithResponses
	Doer           HttpRequestDoer
	Ctx            context.Context
	CollectionName string
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
		URLs:           urls,
		ActiveUrl:      urls[0],
		token:          base64.StdEncoding.EncodeToString([]byte(auth)),
		r:              http.DefaultTransport,
		Ctx:            ctx,
		CollectionName: utils.GetEnv("QDRANT_COLLECTION_NAME", "stablecog"),
	}

	c, doer, err := NewClientWithResponses(qClient.ActiveUrl, WithHTTPClient(&http.Client{
		Timeout:   10 * time.Second,
		Transport: qClient,
	}))
	if err != nil {
		log.Errorf("Error creating qdrant client %v", err)
		return nil, err
	}
	qClient.Client = c
	qClient.Doer = doer

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

	c, doer, err := NewClientWithResponses(q.ActiveUrl, WithHTTPClient(&http.Client{
		Timeout:   10 * time.Second,
		Transport: q,
	}))
	if err != nil {
		log.Errorf("Error creating qdrant client %v", err)
		return err
	}

	q.Client = c
	q.Doer = doer
	return nil
}

// Get all collections in qdrant
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

// Creates our app collection if it doesnt exist
func (q *QDrantClient) CreateCollectionIfNotExists(noRetry bool) error {
	// Check if collection exists
	collections, err := q.GetCollections(false)
	if err != nil {
		return err
	}
	for _, collection := range collections.Collections {
		if collection.Name == q.CollectionName {
			return nil
		}
	}

	// create optimizers config
	optimizersConfig := &CreateCollection_OptimizersConfig{}
	err = optimizersConfig.FromOptimizersConfigDiff(OptimizersConfigDiff{
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

	resp, err := q.Client.CreateCollectionWithResponse(q.Ctx, q.CollectionName, &CreateCollectionParams{}, CreateCollection{
		OptimizersConfig:   optimizersConfig,
		QuantizationConfig: createCollectionQuantizationConfig,
		Vectors:            vectorsConfig,
	})

	if err != nil {
		if !noRetry && (os.IsTimeout(err) || strings.Contains(err.Error(), "connection refused")) {
			err = q.UpdateActiveClient()
			if err == nil {
				return q.CreateCollectionIfNotExists(true)
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

// Upsert
func (q *QDrantClient) Upsert(id uuid.UUID, payload map[string]interface{}, embedding []float32, noRetry bool) error {
	// id
	rId := ExtendedPointId{}
	err := rId.FromExtendedPointId1(id)
	if err != nil {
		log.Errorf("Error creating id %v", err)
		return err
	}
	// payload
	rPayload := PointStruct_Payload{}
	err = rPayload.FromPayload(payload)
	if err != nil {
		log.Errorf("Error creating payload %v", err)
		return err
	}
	// vector
	v := VectorStruct{}
	err = v.FromVectorStruct0(embedding)
	if err != nil {
		log.Errorf("Error creating vector %v", err)
		return err
	}

	// request
	b := UpsertPointsJSONRequestBody{}
	b.FromPointsList(PointsList{
		[]PointStruct{
			{
				Id:      rId,
				Payload: &rPayload,
				Vector:  v,
			},
		},
	})
	resp, err := q.Client.UpsertPoints(q.Ctx, q.CollectionName, &UpsertPointsParams{}, b)
	if err != nil {
		if !noRetry && (os.IsTimeout(err) || strings.Contains(err.Error(), "connection refused")) {
			err = q.UpdateActiveClient()
			if err == nil {
				return q.Upsert(id, payload, embedding, true)
			}
		}
		log.Errorf("Error upserting to collection %v", err)
		return err
	}
	if resp.StatusCode != http.StatusOK {
		log.Errorf("Error getting collections %v", resp.StatusCode)
		return fmt.Errorf("Error upserting to collection %v", resp.StatusCode)
	}

	return nil
}

// Query
func (q *QDrantClient) Query(embedding []float32, noRetry bool) (*QResponse, error) {
	qReq := QdrantRequest{
		Limit:       50,
		WithPayload: true,
		Vector:      embedding,
		Params: QdrantRequestParams{
			HNSWEf: 128,
			Exact:  false,
			Quantization: QdrantRequestParamsQuantization{
				Ignore:  false,
				Rescore: true,
			},
		},
	}
	// Http POST to endpoint with secret
	// Marshal req
	b, err := json.Marshal(qReq)
	if err != nil {
		log.Errorf("Error marshalling req %v", err)
		return nil, err
	}

	qRequest, _ := http.NewRequest(http.MethodPost, q.ActiveUrl, bytes.NewReader(b))
	qRequest.Header.Set("Content-Type", "application/json")
	// Do
	qResp, err := q.Doer.Do(qRequest)
	if err != nil {
		log.Errorf("Error making request %v", err)
		return nil, err
	}
	defer qResp.Body.Close()
	if err != nil {
		if !noRetry && (os.IsTimeout(err) || strings.Contains(err.Error(), "connection refused")) {
			err = q.UpdateActiveClient()
			if err == nil {
				return q.Query(embedding, true)
			}
		}
		log.Errorf("Error getting collections %v", err)
		return nil, err
	}
	if qResp.StatusCode != http.StatusOK {
		log.Errorf("Error querying collection %v", qResp.StatusCode)
		return nil, fmt.Errorf("Error querying collection %v", qResp.StatusCode)
	}

	qReadAll, qErr := io.ReadAll(qResp.Body)
	if qErr != nil {
		log.Error(qErr)
		return nil, qErr
	}

	var qAPIResponse QResponse
	err = json.Unmarshal(qReadAll, &qAPIResponse)
	if err != nil {
		log.Errorf("Error unmarshalling resp %v", err)
		return nil, err
	}

	return &qAPIResponse, nil
}

type QResponse struct {
	Result []QResponseResult `json:"result"`
	Status string            `json:"status"`
	Time   float32           `json:"time"`
}

type QResponseResult struct {
	Id      string                 `json:"id"`
	Version int                    `json:"version"`
	Score   float32                `json:"score"`
	Payload QResponseResultPayload `json:"payload"`
}

type QResponseResultPayload struct {
	CreatedAt string `json:"created_at"`
	ImagePath string `json:"image_path"`
	Prompt    string `json:"prompt"`
}

type QdrantRequest struct {
	Limit       int                 `json:"limit"`
	WithPayload bool                `json:"with_payload,omitempty"`
	Vector      []float32           `json:"vector"`
	Params      QdrantRequestParams `json:"params,omitempty"`
}

type QdrantRequestParams struct {
	HNSWEf       int                             `json:"hnsw_ef"`
	Exact        bool                            `json:"exact"`
	Quantization QdrantRequestParamsQuantization `json:"quantization,omitempty"`
}

type QdrantRequestParamsQuantization struct {
	Ignore  bool `json:"ignore"`
	Rescore bool `json:"rescore"`
}
