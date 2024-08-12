package responses

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

type ClipEmbeddingResponse struct {
	Embeddings []ClipEmbeddingItem `json:"embeddings"`
}

type ClipEmbeddingItem struct {
	AestheticScore *ClipAestheticScore `json:"aesthetic_score,omitempty"`
	NsfwScore      *NsfwCheckScore     `json:"nsfw_score,omitempty"`
	Embedding      []float32           `json:"embedding"`
	InputImage     string              `json:"input_image,omitempty"`
	InputText      string              `json:"input_text,omitempty"`
	ID             string              `json:"id,omitempty"`
	Error          string              `json:"error,omitempty"`
}

type ClipAestheticScore struct {
	Artifact float32 `json:"artifact"`
	Rating   float32 `json:"rating"`
}

type NsfwCheckScore struct {
	Nsfw float32 `json:"nsfw"`
}
