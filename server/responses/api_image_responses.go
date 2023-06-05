package responses

import "github.com/google/uuid"

type ApiOutput struct {
	ID  uuid.UUID `json:"id"`
	URL string    `json:"url"`
}

type ApiSucceededResponse struct {
	Outputs          []ApiOutput `json:"outputs"`
	RemainingCredits float32     `json:"remaining_credits"`
	Settings         interface{} `json:"settings"`
}

type ApiFailedResponse struct {
	Error    string      `json:"error"`
	Settings interface{} `json:"settings"`
}
