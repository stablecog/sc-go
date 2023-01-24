package models

type GenerateRequestBody struct {
	// Prompt                string  `json:"prompt"`
	// NegativePrompt        string  `json:"negative_prompt,omitempty"`
	Width             int `json:"width"`
	Height            int `json:"height"`
	NumInferenceSteps int `json:"num_inference_steps"`
	// GuidanceScale         int     `json:"guidance_scale"`
	// ServerUrl             string  `json:"server_url"`
	// ModelId               string  `json:"model_id"`
	// SchedulerId           string  `json:"scheduler_id"`
	Seed int `json:"seed"`
	// OutputImageExt        string  `json:"output_image_ext,omitempty"`
	// InitImage             string  `json:"init_image,omitempty"`
	// Mask                  string  `json:"mask,omitempty"`
	// PromptStrength        float32 `json:"prompt_strength,omitempty"`
	// ShouldSubmitToGallery bool    `json:"should_submit_to_gallery"`
}
