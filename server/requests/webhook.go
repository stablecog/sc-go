package requests

type WebhookStatus string

const (
	WebhookSucceeded  WebhookStatus = "succeeded"
	WebhookFailed     WebhookStatus = "failed"
	WebhookProcessing WebhookStatus = "processing"
)

type WebhookRequestInput struct {
	Id     string `json:"id"`
	Prompt string `json:"prompt"`
	Model  string `json:"model"`
	Width  string `json:"width"`
	Height string `json:"height"`
}

// Msg from cog to redis
type WebhookRequestData struct {
	Outputs   []string `json:"outputs"`
	NSFWCount int      `json:"nsfw_count"`
}

type WebhookRequest struct {
	Webhook string              `json:"webhook"`
	Input   WebhookRequestInput `json:"input"`
	Status  WebhookStatus       `json:"status"`
	Error   string              `json:"error"`
	Data    WebhookRequestData  `json:"data"`
}
