package responses

import "github.com/google/uuid"

type ApiGenerateOutput struct {
	ID  uuid.UUID `json:"id"`
	URL string    `json:"url"`
}

type ApiGenerateSucceededResponse struct {
	Outputs          []ApiGenerateOutput `json:"outputs"`
	RemainingCredits int                 `json:"remaining_credits"`
}

type ApiGenerateFailedResponse struct {
	Error string `json:"error"`
}
