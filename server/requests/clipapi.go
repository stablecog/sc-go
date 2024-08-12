package requests

type ClipApiEmbeddingRequest struct {
	Text           string `json:"text,omitempty"`
	Image          string `json:"image,omitempty"`
	CalculateScore bool   `json:"calculate_score,omitempty"`
	CheckNsfw      bool   `json:"check_nsfw,omitempty"`
	ID             string `json:"id,omitempty"`
}
