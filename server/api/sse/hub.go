package sse

import (
	"sync"
)

type Hub struct {
	// Events are pushed to this channel by the main events-gathering routine
	Broadcast chan []byte

	// Clients connecting
	Register chan *Client

	// Clients disconnecting
	Unregister chan *Client

	// Client connections registry
	clients map[*Client]bool

	// We need a mutex to protect the clients map
	mu sync.Mutex
}

// Braodcast a message to all clients that match the given uid
func (h *Hub) BroadcastToClientsWithUid(uid string, message []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for client := range h.clients {
		if client.Uid == uid {
			select {
			case client.Send <- message:
			default:
				close(client.Send)
				delete(h.clients, client)
			}
		}
	}
}

func (h *Hub) GetClientByUid(uid string) *Client {
	h.mu.Lock()
	defer h.mu.Unlock()
	for client := range h.clients {
		if client.Uid == uid {
			return client
		}
	}
	return nil
}

func NewHub() *Hub {
	return &Hub{
		Broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			func() {
				h.mu.Lock()
				defer h.mu.Unlock()
				h.clients[client] = true
			}()
		case client := <-h.Unregister:
			func() {
				h.mu.Lock()
				defer h.mu.Unlock()
				if _, ok := h.clients[client]; ok {
					delete(h.clients, client)
					close(client.Send)
				}
			}()
		case message := <-h.Broadcast:
			func() {
				h.mu.Lock()
				defer h.mu.Unlock()
				for client := range h.clients {
					select {
					case client.Send <- message:
					default:
						close(client.Send)
						delete(h.clients, client)
					}
				}
			}()
		}
	}

}
