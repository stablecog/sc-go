package responses

import "github.com/google/uuid"

type EmbeddingObject struct {
	Embedding      []float32 `json:"embedding"`
	InputText      string    `json:"input_text"`
	TranslatedText string    `json:"translated_text,omitempty"`
	ID             uuid.UUID `json:"id,omitempty"`
	Error          string    `json:"error,omitempty"`
}

type EmbeddingsResponse struct {
	Embeddings []EmbeddingObject `json:"embeddings"`
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
