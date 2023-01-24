package responses

// API generate simply returns a UUID to track the request to our compute while its in flight
type GenerateResponse struct {
	ID string `json:"id"`
}
