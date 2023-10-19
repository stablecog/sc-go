package sse

import (
	"encoding/json"

	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/repository"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/shared"
)

// Broadcasts message from sc-worker to client(s) SSE stream(s)
// It's published by our repository, after we do database-y stuff with out cog message
func (h *Hub) BroadcastStatusUpdate(msg repository.TaskStatusUpdateResponse) {
	// Marshal
	respBytes, err := json.Marshal(msg)
	if err != nil {
		log.Error("Error marshalling sse response", "err", err)
		return
	}

	// Broadcast to all clients subcribed to this stream
	h.Broadcast <- BroadcastPayload{
		ID:      msg.StreamId,
		Message: respBytes,
	}
}

// Broadcast a message for the live page
func (h *Hub) BroadcastLivePageMessage(req shared.LivePageMessage) {
	bytes, err := json.Marshal(req)
	if err != nil {
		log.Error("Error marshalling live page message", "err", err)
		return
	}
	h.Broadcast <- BroadcastPayload{
		ID:      "live",
		Message: bytes,
	}
}

// Broadcast a message to all clients
func (h *Hub) BroadcastQueueUpdate(msg []*ent.MqLog) {
	// Marshal
	respBytes, err := json.Marshal(msg)
	if err != nil {
		log.Error("Error marshalling sse response", "err", err)
		return
	}

	// Broadcast to all clients subcribed to this stream
	h.Broadcast <- BroadcastPayload{
		ID:      ALL_CLIENTS_ID,
		Message: respBytes,
	}
}
