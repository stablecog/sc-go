package responses

import "github.com/google/uuid"

type ApiOutput struct {
	ID               uuid.UUID `json:"id"`
	URL              string    `json:"url"`
	ImageURL         *string   `json:"image_url,omitempty"`
	UpscaledImageURL *string   `json:"upscaled_image_url,omitempty"`
	AudioFileURL     *string   `json:"audio_file_url,omitempty"`
	VideoFileURL     *string   `json:"video_file_url,omitempty"`
	AudioDuration    *float32  `json:"audio_duration,omitempty"`
}

type ApiSucceededResponse struct {
	Outputs          []ApiOutput         `json:"outputs"`
	RemainingCredits int                 `json:"remaining_credits"`
	Settings         interface{}         `json:"settings"`
	QueuedResponse   *TaskQueuedResponse `json:"queued_response,omitempty"`
}

type ApiFailedResponse struct {
	Error    string      `json:"error"`
	Settings interface{} `json:"settings,omitempty"`
}
