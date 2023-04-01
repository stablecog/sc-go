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
