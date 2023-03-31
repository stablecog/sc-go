package responses

type EmbeddingObject struct {
	Embedding []float32 `json:"embedding"`
	InputText string    `json:"input_text"`
}

type EmbeddingsResponse struct {
	Embeddings []EmbeddingObject `json:"embeddings"`
}
