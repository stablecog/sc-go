package sse

import (
	"encoding/json"

	"github.com/stablecog/sc-go/database/repository"
	"github.com/stablecog/sc-go/shared"
	"k8s.io/klog/v2"
)

// Broadcasts message from sc-worker to client(s) SSE stream(s)
// It's published by our repository, after we do database-y stuff with out cog message
func (h *Hub) BroadcastStatusUpdate(msg repository.TaskStatusUpdateResponse) {
	// If the stream isn't connected to us, do nothing
	if h.GetClientByUid(msg.StreamId) == nil {
		return
	}

	// Marshal
	respBytes, err := json.Marshal(msg)
	if err != nil {
		klog.Errorf("--- Error marshalling sse response: %v", err)
		return
	}

	// Broadcast to all clients subcribed to this stream
	h.BroadcastToClientsWithUid(msg.StreamId, respBytes)
}

// Broadcast a message for the live page
func (h *Hub) BroadcastLivePageMessage(req shared.LivePageMessage) {
	bytes, err := json.Marshal(req)
	if err != nil {
		klog.Errorf("Error marshalling live page message: %v", err)
		return
	}
	h.BroadcastToClientsWithUid("live", bytes)
}
