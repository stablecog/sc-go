package qdrant

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/go-multierror"
	"github.com/stablecog/sc-go/database/ent/generationoutput"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/utils"
	"golang.org/x/exp/slices"
)

type qdrantIndexField struct {
	Name string            `json:"name"`
	Type PayloadSchemaType `json:"type"`
}

// The fields we create indexes for on app startup
var fieldsToIndex = []qdrantIndexField{
	{
		Name: "gallery_status",
		Type: PayloadSchemaTypeKeyword,
	},
	{
		Name: "user_id",
		Type: PayloadSchemaTypeKeyword,
	},
	{
		Name: "width",
		Type: PayloadSchemaTypeInteger,
	},
	{
		Name: "height",
		Type: PayloadSchemaTypeInteger,
	},
	{
		Name: "inference_steps",
		Type: PayloadSchemaTypeInteger,
	},
	{
		Name: "guidance_scale",
		Type: PayloadSchemaTypeFloat,
	},
	{
		Name: "created_at",
		Type: PayloadSchemaTypeInteger,
	},
	{
		Name: "deleted_at",
		Type: PayloadSchemaTypeInteger,
	},
	{
		Name: "model",
		Type: PayloadSchemaTypeKeyword,
	},
	{
		Name: "scheduler",
		Type: PayloadSchemaTypeKeyword,
	},
	{
		Name: "is_favorited",
		Type: PayloadSchemaTypeBool,
	},
	{
		Name: "was_auto_submitted",
		Type: PayloadSchemaTypeBool,
	},
	{
		Name: "is_public",
		Type: PayloadSchemaTypeBool,
	},
	{
		Name: "prompt",
		Type: PayloadSchemaTypeKeyword,
	},
	{
		Name: "prompt_id",
		Type: PayloadSchemaTypeKeyword,
	},
}

type QdrantClient struct {
	ActiveUrl      string
	token          string
	r              http.RoundTripper
	Client         *ClientWithResponses
	Doer           HttpRequestDoer
	Ctx            context.Context
	CollectionName string
}

func (q QdrantClient) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Add("Authorization", "Basic "+q.token)
	return q.r.RoundTrip(r)
}

func NewQdrantClient(ctx context.Context) (*QdrantClient, error) {
	// Get URLs from env, comma separated
	urlEnv := os.Getenv("QDRANT_URL")
	if urlEnv == "" {
		log.Errorf("QDRANT_URL not set")
		return nil, errors.New("QDRANT_URL not set")
	}
	var auth string
	if os.Getenv("QDRANT_USERNAME") != "" && os.Getenv("QDRANT_PASSWORD") != "" {
		auth = base64.StdEncoding.EncodeToString([]byte(os.Getenv("QDRANT_USERNAME") + ":" + os.Getenv("QDRANT_PASSWORD")))
	}
	// Create client
	qClient := &QdrantClient{
		ActiveUrl:      urlEnv,
		Ctx:            ctx,
		token:          auth,
		r:              http.DefaultTransport,
		CollectionName: utils.GetEnv("QDRANT_COLLECTION_NAME", "stablecog"),
	}

	transport := http.DefaultTransport
	if auth != "" {
		transport = qClient
	}

	c, doer, err := NewClientWithResponses(qClient.ActiveUrl, WithHTTPClient(&http.Client{
		Timeout:   10 * time.Second,
		Transport: transport,
	}))
	if err != nil {
		log.Errorf("Error creating qdrant client %v", err)
		return nil, err
	}
	qClient.Client = c
	qClient.Doer = doer

	return qClient, nil
}

// Get all collections in qdrant
func (q *QdrantClient) GetCollections(noRetry bool) (*CollectionsResponse, error) {
	resp, err := q.Client.GetCollectionsWithResponse(q.Ctx)
	if err != nil {
		if !noRetry && (os.IsTimeout(err) || strings.Contains(err.Error(), "connection refused")) {
			return q.GetCollections(true)
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

// Create indexes
const (
	PayloadTypeKeyword = "keyword"
	PayloadTypeFloat   = "float"
	PayloadTypeInt     = "integer"
	PayloadTypeGeo     = "geo"
	PayloadTypeText    = "text"
)

func (q *QdrantClient) DeleteIndex(fieldName string, noRetry bool) error {
	_, err := q.Client.DeleteFieldIndex(q.Ctx, q.CollectionName, fieldName, &DeleteFieldIndexParams{})
	if err != nil {
		return err
	}
	return nil
}

func (q *QdrantClient) CreateIndex(fieldName string, schemaType PayloadSchemaType, noRetry bool) error {
	schema := &CreateFieldIndex_FieldSchema{}
	plSchema := PayloadFieldSchema{}
	plSchema.FromPayloadSchemaType(schemaType)
	schema.FromPayloadFieldSchema(plSchema)
	// Create indexes
	res, err := q.Client.CreateFieldIndexWithResponse(q.Ctx, q.CollectionName, &CreateFieldIndexParams{}, CreateFieldIndex{
		FieldName:   fieldName,
		FieldSchema: schema,
	})
	if err != nil {
		if !noRetry && (os.IsTimeout(err) || strings.Contains(err.Error(), "connection refused")) {
			return q.CreateIndex(fieldName, schemaType, true)
		}
		log.Errorf("Error creating index %v", err)
		return err
	}
	if res.StatusCode() != http.StatusOK {
		log.Errorf("Error creating index %v", res.StatusCode())
		return errors.New("Error creating index " + string(res.Body))
	}
	return nil
}

// Creates our app collection if it doesnt exist
func (q *QdrantClient) CreateCollectionIfNotExists(noRetry bool) error {
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
	vectorsConfigMulti := VectorsConfig1{}
	vectorsConfigMulti["image"] = VectorParams{
		Size:     uint64(1024),
		Distance: "Dot",
	}
	vectorsConfigMulti["text"] = VectorParams{
		Size:     uint64(1024),
		Distance: "Cosine",
	}
	vectorsConfig.FromVectorsConfig1(vectorsConfigMulti)
	if err != nil {
		log.Errorf("Error creating vectors config %v", err)
		return err
	}

	test := CreateCollection{
		OptimizersConfig:   optimizersConfig,
		QuantizationConfig: createCollectionQuantizationConfig,
		Vectors:            vectorsConfig,
		ShardNumber:        utils.ToPtr[uint32](2),
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
		ShardNumber:        utils.ToPtr[uint32](2),
	})

	if err != nil {
		if !noRetry && (os.IsTimeout(err) || strings.Contains(err.Error(), "connection refused")) {
			return q.CreateCollectionIfNotExists(true)
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
func (q *QdrantClient) Upsert(id uuid.UUID, payload map[string]interface{}, imageEmbedding []float32, promptEmbedding []float32, noRetry bool) error {
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
	err = v.FromVectorStruct1(VectorStruct1{
		"image": imageEmbedding,
		"text":  promptEmbedding,
	})
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
	resp, err := q.Client.UpsertPointsWithResponse(q.Ctx, q.CollectionName, &UpsertPointsParams{}, b)
	if err != nil {
		if !noRetry && (os.IsTimeout(err) || strings.Contains(err.Error(), "connection refused")) {
			return q.Upsert(id, payload, imageEmbedding, promptEmbedding, true)
		}
		log.Errorf("Error upserting to collection %v", err)
		return err
	}
	if resp.StatusCode() != http.StatusOK {
		log.Errorf("Error getting collections %v", resp.StatusCode())
		return fmt.Errorf("Error upserting to collection %v", resp.StatusCode())
	}

	return nil
}

// Set payload on points
func (q *QdrantClient) SetPayload(payload map[string]interface{}, ids []uuid.UUID, noRetry bool) error {
	var points []ExtendedPointId
	for _, id := range ids {
		rId := ExtendedPointId{}
		err := rId.FromExtendedPointId1(id)
		if err != nil {
			log.Errorf("Error creating id %v", err)
			return err
		}
		points = append(points, rId)
	}
	res, err := q.Client.SetPayloadWithResponse(q.Ctx, q.CollectionName, &SetPayloadParams{}, SetPayload{
		Points:  &points,
		Payload: payload,
	})
	if err != nil {
		if !noRetry && (os.IsTimeout(err) || strings.Contains(err.Error(), "connection refused")) {
			return q.SetPayload(payload, ids, true)
		}
		log.Errorf("Error setting payload %v", err)
		return err
	}
	if res.StatusCode() != http.StatusOK {
		if res.JSON4XX != nil {
			marshalled, err := json.Marshal(*res.JSON4XX)
			if err != nil {
				log.Errorf("Error marshalling response %v", err)
			} else {
				log.Errorf("Error setting payload res %v", string(marshalled))
			}
		}
		log.Errorf("Error setting payload %v", res.StatusCode())
		return fmt.Errorf("Error setting payload %v", res.StatusCode())
	}
	return nil
}

// Count with filters
func (q *QdrantClient) CountWithFilters(filters *SearchRequest_Filter, noRetry bool) (uint, error) {
	resp, err := q.Client.CountPointsWithResponse(q.Ctx, q.CollectionName, CountPointsJSONRequestBody{
		Filter: filters,
		Exact:  utils.ToPtr(false),
	})
	if err != nil {
		if !noRetry && (os.IsTimeout(err) || strings.Contains(err.Error(), "connection refused")) {
			q.CountWithFilters(filters, true)
		}
		log.Errorf("Error counting points %v", err)
		return 0, err
	}
	if resp.StatusCode() != http.StatusOK {
		log.Errorf("Error counting points %v", resp.StatusCode())
		return 0, fmt.Errorf("Error counting points %v", resp.StatusCode())
	}
	return resp.JSON200.Result.Count, nil
}

// Count
func (q *QdrantClient) Count(noRetry bool) (uint, error) {
	resp, err := q.Client.CountPointsWithResponse(q.Ctx, q.CollectionName, CountPointsJSONRequestBody{})
	if err != nil {
		if !noRetry && (os.IsTimeout(err) || strings.Contains(err.Error(), "connection refused")) {
			q.Count(true)
		}
		log.Errorf("Error counting points %v", err)
		return 0, err
	}
	if resp.StatusCode() != http.StatusOK {
		log.Errorf("Error counting points %v", resp.StatusCode())
		return 0, fmt.Errorf("Error counting points %v", resp.StatusCode())
	}
	return resp.JSON200.Result.Count, nil
}

// Upsert
func (q *QdrantClient) BatchUpsert(payload []map[string]interface{}, noRetry bool) error {
	payloadCopy := make([]map[string]interface{}, len(payload))
	copy(payloadCopy, payload)
	var points []PointStruct
	for _, p := range payload {
		// See if ID in payload and remove it
		idStr, hasId := p["id"].(string)
		if !hasId {
			log.Errorf("Error upserting point: no id")
			return fmt.Errorf("Error upserting point: no id")
		}
		id, err := uuid.Parse(idStr)
		delete(p, "id")
		// Get embedding from payload and remove it
		embedding, hasEmbedding := p["embedding"].([]float32)
		if hasEmbedding {
			delete(p, "embedding")
		}
		textEmbedding, hasTextEmbedding := p["text_embedding"].([]float32)
		if hasTextEmbedding {
			delete(p, "text_embedding")
		}

		rId := ExtendedPointId{}
		err = rId.FromExtendedPointId1(id)
		if err != nil {
			log.Errorf("Error creating id %v", err)
			return err
		}

		// payload
		rPayload := PointStruct_Payload{}
		err = rPayload.FromPayload(p)
		if err != nil {
			log.Errorf("Error creating payload %v", err)
			return err
		}

		// vector
		v := VectorStruct{}
		vMulti := VectorStruct1{}
		vMulti["image"] = embedding
		vMulti["text"] = textEmbedding
		err = v.FromVectorStruct1(vMulti)
		if err != nil {
			log.Errorf("Error creating vector %v", err)
			return err
		}

		points = append(points, PointStruct{
			Id:      rId,
			Payload: &rPayload,
			Vector:  v,
		})
	}
	// request
	b := UpsertPointsJSONRequestBody{}
	b.FromPointsList(PointsList{
		points,
	})
	resp, err := q.Client.UpsertPointsWithResponse(q.Ctx, q.CollectionName, &UpsertPointsParams{}, b)
	if err != nil {
		if !noRetry && (os.IsTimeout(err) || strings.Contains(err.Error(), "connection refused")) {
			return q.BatchUpsert(payloadCopy, true)
		}
		log.Errorf("Error upserting to collection %v", err)
		return err
	}
	if resp.StatusCode() != http.StatusOK {
		log.Errorf("Error getting collections %v", resp.StatusCode())
		return fmt.Errorf("Error upserting to collection %v", resp.StatusCode())
	}

	return nil
}

// Type for vector result
type GetPointVector struct {
	Image []float32 `json:"image"`
	Text  []float32 `json:"text"`
}

// Type for GetPoint response
type GetPointResult struct {
	ID      uuid.UUID              `json:"id"`
	Payload map[string]interface{} `json:"payload"`
	Vector  GetPointVector         `json:"vector"`
}

type GetPointResponseSC struct {
	Result GetPointResult `json:"result"`
}

// Get vectors for an ID
func (q *QdrantClient) GetPoint(id uuid.UUID, noRetry bool) (*GetPointResponseSC, error) {
	rId := ExtendedPointId{}
	rId.FromExtendedPointId1(id)
	resp, err := q.Client.GetPointWithResponse(q.Ctx, q.CollectionName, rId, &GetPointParams{})
	if err != nil {
		if !noRetry && (os.IsTimeout(err) || strings.Contains(err.Error(), "connection refused")) {
			return q.GetPoint(id, true)
		}
		log.Errorf("Error getting point %v", err)
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		log.Errorf("Error querying collection %v", resp.StatusCode())
		return nil, fmt.Errorf("Error querying collection %v", resp.StatusCode())
	}
	if resp.StatusCode() == http.StatusNotFound {
		return nil, fmt.Errorf("not_found")
	}

	var unmarshalled GetPointResponseSC
	err = json.Unmarshal(resp.Body, &unmarshalled)
	if err != nil {
		log.Errorf("Error unmarshalling response %v", err)
		return nil, err
	}
	return &unmarshalled, nil
}

// Query
func (q *QdrantClient) Query(embedding []float32, noRetry bool) (*QResponse, error) {
	qParams := &SearchParams_Quantization{}
	qParams.FromQuantizationSearchParams(QuantizationSearchParams{
		Ignore:  utils.ToPtr(false),
		Rescore: utils.ToPtr(false),
	})
	params := &SearchRequest_Params{}
	params.FromSearchParams(SearchParams{
		HnswEf:       utils.ToPtr[uint](128),
		Exact:        utils.ToPtr(false),
		Quantization: &SearchParams_Quantization{},
	})
	namedVectorParams := NamedVectorStruct{}
	err := namedVectorParams.FromNamedVector(NamedVector{
		Name:   "image",
		Vector: embedding,
	})
	if err != nil {
		log.Errorf("Error creating vector search param %v", err)
		return nil, err
	}
	resp, err := q.Client.SearchPointsWithResponse(q.Ctx, q.CollectionName, &SearchPointsParams{}, SearchPointsJSONRequestBody{
		Limit:       50,
		WithPayload: true,
		Vector:      namedVectorParams,
		Params:      params,
	})

	if err != nil {
		if !noRetry && (os.IsTimeout(err) || strings.Contains(err.Error(), "connection refused")) {
			return q.Query(embedding, true)
		}
		log.Errorf("Error getting collections %v", err)
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		log.Errorf("Error querying collection %v", resp.StatusCode())
		return nil, fmt.Errorf("Error querying collection %v", resp.StatusCode())
	}

	var qAPIResponse QResponse
	err = json.Unmarshal(resp.Body, &qAPIResponse)
	if err != nil {
		log.Errorf("Error unmarshalling resp %v", err)
		return nil, err
	}

	return &qAPIResponse, nil
}

func (q *QdrantClient) DeleteById(id uuid.UUID, noRetry bool) error {
	rId := ExtendedPointId{}
	rId.FromExtendedPointId1(id)
	body := DeletePointsJSONRequestBody{}
	body.FromPointIdsList(PointIdsList{
		Points: []ExtendedPointId{
			rId,
		},
	})
	resp, err := q.Client.DeletePointsWithResponse(q.Ctx, q.CollectionName, &DeletePointsParams{}, body)
	if err != nil {
		if !noRetry && (os.IsTimeout(err) || strings.Contains(err.Error(), "connection refused")) {
			return q.DeleteById(id, true)
		}
		log.Errorf("Error deleting from collection %v", err)
		return err
	}
	if resp.StatusCode() != http.StatusOK {
		log.Errorf("Error getting collections %v", resp.StatusCode())
		return fmt.Errorf("Error upserting to collection %v", resp.StatusCode())
	}

	return nil
}

// Public gallery search
func (q *QdrantClient) QueryGenerations(embedding []float32, per_page int, offset *uint, scoreThreshold *float32, filters *SearchRequest_Filter, withPayload bool, noRetry bool) (*QResponse, error) {
	qParams := &SearchParams_Quantization{}
	qParams.FromQuantizationSearchParams(QuantizationSearchParams{
		Ignore:  utils.ToPtr(false),
		Rescore: utils.ToPtr(false),
	})
	params := &SearchRequest_Params{}
	params.FromSearchParams(SearchParams{
		HnswEf:       utils.ToPtr[uint](128),
		Exact:        utils.ToPtr(false),
		Quantization: &SearchParams_Quantization{},
	})
	namedVectorParams := NamedVectorStruct{}
	err := namedVectorParams.FromNamedVector(NamedVector{
		Name:   "image",
		Vector: embedding,
	})
	if err != nil {
		log.Errorf("Error creating vector search param %v", err)
		return nil, err
	}

	resp, err := q.Client.SearchPointsWithResponse(q.Ctx, q.CollectionName, &SearchPointsParams{}, SearchPointsJSONRequestBody{
		Limit:          uint(per_page + 1),
		WithPayload:    withPayload,
		Vector:         namedVectorParams,
		Offset:         offset,
		Filter:         filters,
		Params:         params,
		ScoreThreshold: scoreThreshold,
	})

	if err != nil {
		if !noRetry && (os.IsTimeout(err) || strings.Contains(err.Error(), "connection refused")) {
			return q.QueryGenerations(embedding, per_page, offset, scoreThreshold, filters, withPayload, true)
		}
		log.Errorf("Error getting collections %v", err)
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		log.Errorf("Error querying collection %v", resp.StatusCode())
		return nil, fmt.Errorf("Error querying collection %v", resp.StatusCode())
	}

	var qAPIResponse QResponse
	err = json.Unmarshal(resp.Body, &qAPIResponse)
	if err != nil {
		log.Errorf("Error unmarshalling resp %v", err)
		return nil, err
	}

	if len(qAPIResponse.Result) > per_page {
		// Remove last result and get next offset
		if offset != nil {
			qAPIResponse.Next = utils.ToPtr(uint(*offset) + uint(per_page))
		} else {
			qAPIResponse.Next = utils.ToPtr(uint(per_page))
		}
		// Remove last item
		qAPIResponse.Result = qAPIResponse.Result[:len(qAPIResponse.Result)-1]
	}
	return &qAPIResponse, nil
}

// Get list of fields with index
func (q *QdrantClient) GetIndexedPayloadFields(noRetry bool) ([]string, error) {
	resp, err := q.Client.GetCollectionWithResponse(q.Ctx, q.CollectionName)
	if err != nil {
		log.Errorf("Error getting collections %v", err)
		return nil, err
	}
	if err != nil {
		if !noRetry && (os.IsTimeout(err) || strings.Contains(err.Error(), "connection refused")) {
			return q.GetIndexedPayloadFields(true)
		}
		log.Errorf("Error getting collections %v", err)
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		log.Errorf("Error getting collection %v", resp.StatusCode())
		return nil, fmt.Errorf("Error getting collection %v", resp.StatusCode())
	}
	res := make([]string, len(resp.JSON200.Result.PayloadSchema))
	i := 0
	for fieldName := range resp.JSON200.Result.PayloadSchema {
		res[i] = fieldName
		i++
	}
	return res, nil
}

func (q *QdrantClient) CreateAllIndexes() error {
	// Get indexed fields
	indexFields, err := q.GetIndexedPayloadFields(false)
	if err != nil {
		return err
	}
	var mErr *multierror.Error
	for _, field := range fieldsToIndex {
		if !slices.Contains(indexFields, field.Name) {
			mErr = multierror.Append(q.CreateIndex(field.Name, field.Type, false))
		}
	}
	return mErr.ErrorOrNil()
}

func (q *QdrantClient) DeleteAllIDs(ids []uuid.UUID, noRetry bool) error {
	p := &DeletePointsParams{
		Wait: utils.ToPtr(false),
	}
	body := DeletePointsJSONRequestBody{}
	extPointId := make([]ExtendedPointId, len(ids))
	for i, id := range ids {
		rId := ExtendedPointId{}
		err := rId.FromExtendedPointId1(id)
		if err != nil {
			return err
		}
		extPointId[i] = rId
	}
	ls := PointIdsList{
		Points: extPointId,
	}
	body.FromPointIdsList(ls)
	resp, err := q.Client.DeletePointsWithResponse(q.Ctx, q.CollectionName, p, body)
	if err != nil {
		if !noRetry && (os.IsTimeout(err) || strings.Contains(err.Error(), "connection refused")) {
			return q.DeleteAllIDs(ids, true)
		}
		log.Errorf("Error deleting multi points %v", err)
		return err
	}
	if resp.StatusCode() != http.StatusOK {
		log.Errorf("Error deleting multi points %v", resp.StatusCode())
		return fmt.Errorf("Error deleting multi points %v", resp.StatusCode())
	}

	return nil

}

type QResponse struct {
	Result []QResponseResult `json:"result"`
	Status string            `json:"status"`
	Time   float32           `json:"time"`
	Next   *uint             `json:"next,omitempty"`
}

type QResponseResult struct {
	Id      string                 `json:"id"`
	Version int                    `json:"version"`
	Score   float32                `json:"score"`
	Payload QResponseResultPayload `json:"payload,omitempty"`
}

type QResponseResultPayload struct {
	CreatedAt         int64                          `json:"created_at"`
	ImagePath         string                         `json:"image_path"`
	UpscaledImagePath string                         `json:"upscaled_image_path,omitempty"`
	GalleryStatus     generationoutput.GalleryStatus `json:"gallery_status"`
	IsFavorited       bool                           `json:"is_favorited"`
	Prompt            string                         `json:"prompt"`
	GenerationID      string                         `json:"generation_id"`
}
