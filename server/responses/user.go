// * Responses from user-initiated endpoints
package responses

// API generate simply returns a UUID to track the request to our compute while its in flight
type QueuedResponse struct {
	ID string `json:"id"`
}

// Response for submitting to gallery
type GenerateSubmitToGalleryResponse struct {
	Submitted int `json:"submitted"`
}

// API response for retrieving user generations
// type UserGenerationsResponse struct {
// 	Width          int32                    `json:"width"`
// 	Height         int32                    `json:"height"`
// 	InferenceSteps int32                    `json:"inference_steps"`
// 	GuidanceScale  float32                  `json:"guidance_scale"`
// 	Prompt         string                   `json:"prompt"`
// 	NegativePrompt string                   `json:"negative_prompt,omitempty"`
// 	Model          string                   `json:"model"`
// 	Scheduler      string                   `json:"scheduler"`
// 	Seed           int                      `json:"seed"`
// 	Outputs        []string                 `json:"outputs"`
// 	GalleryStatus  generation.GalleryStatus `json:"gallery_status"`
// 	Status         generation.Status        `json:"status"`
// 	CreatedAt      time.Time                `json:"created_at"`
// 	StartedAt      *time.Time               `json:"started_at,omitempty"`
// 	CompletedAt    *time.Time               `json:"completed_at,omitempty"`
// }
