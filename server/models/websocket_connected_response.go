package models

// Sent to a client when they connect to the websocket
type WebsocketConnectedResponse struct {
	Id string `json:"id"`
}
