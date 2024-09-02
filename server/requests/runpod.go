package requests

// Base request data runpod serverless uses to process request
type RunpodInput struct {
	Input BaseCogRequest `json:"input"`
}
