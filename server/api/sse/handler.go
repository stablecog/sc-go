package sse

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/stablecog/go-apps/server/responses"
	"github.com/stablecog/go-apps/utils"
)

// Handles client connections to SSE service
func (h *Hub) ServeSSE(w http.ResponseWriter, r *http.Request) {
	// Retrieve id from query parameters
	query := r.URL.Query()
	// They always connect with query param ?stream
	streamID := strings.ToLower(query.Get("stream"))
	if !utils.IsSha256Hash(streamID) {
		responses.ErrBadRequest(w, r, "Invalid ID")
		return
	}

	// Make sure that the writer supports flushing.
	flusher, ok := w.(http.Flusher)

	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	// TODO - Proper cors restrictions
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Register client in the hub
	client := &Client{Send: make(chan []byte, 256), Uid: streamID}
	h.Register <- client

	// Remove this client from the map of connected clients
	// when this handler exits.
	defer func() {
		h.Unregister <- client
	}()

	// Listen to connection close and un-register client
	notify := w.(http.CloseNotifier).CloseNotify()
	for {
		select {
		case <-notify:
			return
		default:
			// Write to the ResponseWriter
			// SSE compatible
			fmt.Fprintf(w, "data: %s\n\n", <-client.Send)

			// Flush the data immediatly instead of buffering it for later.
			flusher.Flush()
		}
	}
}
