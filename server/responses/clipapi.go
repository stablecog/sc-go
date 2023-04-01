package responses

type EmbeddingObject struct {
	Embedding      []float32 `json:"embedding"`
	InputText      string    `json:"input_text"`
	TranslatedText string    `json:"translated_text,omitempty"`
}

type EmbeddingsResponse struct {
	Embeddings []EmbeddingObject `json:"embeddings"`
}
