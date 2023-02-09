package sse

// Every server connection is a client instance
type Client struct {
	// Buffered channel of outbound messages.
	Send chan []byte

	// identifier for connected client
	Uid string
}
