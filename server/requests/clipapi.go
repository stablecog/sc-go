package requests

import "github.com/google/uuid"

type ClipAPIRequest struct {
	Text string `json:"text"`
}

type ClipAPIImageRequest struct {
	Image   string    `json:"image,omitempty"`
	ImageID string    `json:"image_id,omitempty"`
	ID      uuid.UUID `json:"id"`
}

type QdrantRequest struct {
	Limit       int  `json:"limit"`
	WithPayload bool `json:"with_payload,omitempty"`
	Vector      []float32
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
