package responses

type SSEStatusUpdateResponse struct {
	Status    CogTaskStatus              `json:"status"`
	Id        string                     `json:"id"`
	StreamId  string                     `json:"stream_id"`
	Error     string                     `json:"error,omitempty"`
	NSFWCount int32                      `json:"nsfw_count,omitempty"`
	Outputs   []GenerationOutputResponse `json:"outputs,omitempty"`
}
