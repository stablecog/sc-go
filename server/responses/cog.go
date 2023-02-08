package responses

// Messages sent from the cog to our application

type CogTaskStatus string

const (
	CogSucceeded  CogTaskStatus = "succeeded"
	CogFailed     CogTaskStatus = "failed"
	CogProcessing CogTaskStatus = "processing"
)

// Should mirror the initial request we made to the cog
type CogInput struct {
	Id                 string `json:"id"`
	Prompt             string `json:"prompt"`
	Model              string `json:"model"`
	Width              string `json:"width"`
	Height             string `json:"height"`
	GenerationOutputID string `json:"generation_output_id,omitempty"`
}

// Msg from cog to redis
type CogStatusUpdate struct {
	Webhook   string        `json:"webhook"`
	Input     CogInput      `json:"input"`
	Status    CogTaskStatus `json:"status"`
	Error     string        `json:"error"`
	Outputs   []string      `json:"output"`
	NSFWCount int           `json:"nsfw_count"`
}
