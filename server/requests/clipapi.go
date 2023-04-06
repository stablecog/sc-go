package requests

type ClipAPIRequest struct {
	Text string `json:"text"`
}

type ClipAPIImageRequest struct {
	Text    string `json:"text,omitempty"`
	Image   string `json:"image,omitempty"`
	ImageID string `json:"image_id,omitempty"`
	ID      string `json:"id"`
}
