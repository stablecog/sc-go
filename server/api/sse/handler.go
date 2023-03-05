package sse

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/server/responses"
	"github.com/stablecog/sc-go/utils"
)

// App version
const APP_VERSION = "4.00"

// Special stream ID for live page
const LIVE_STREAM_ID = "live"

// Handles client connections to SSE service
func (h *Hub) ServeSSE(w http.ResponseWriter, r *http.Request) {
	// Retrieve id from query parameters
	query := r.URL.Query()
	// They always connect with query param ?id
	streamID := strings.ToLower(query.Get("id"))
	if !utils.IsSha256Hash(streamID) && streamID != LIVE_STREAM_ID {
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
	w.Header().Set("X-Accel-Buffering", "no")
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

	// Broadcast app version message
	version := AppVersionMessage{Version: APP_VERSION}
	versionBytes, err := json.Marshal(version)
	if err != nil {
		log.Error("Error marshalling app version message", "err", err)
		http.Error(w, "Error marshalling app version message", http.StatusInternalServerError)
		return
	}
	client.Send <- versionBytes

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

type AppVersionMessage struct {
	Version string `json:"version"`
}
