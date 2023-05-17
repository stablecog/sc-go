package sse

import (
	"encoding/json"

	"github.com/stablecog/sc-go/database"
	"github.com/stablecog/sc-go/database/repository"
	"github.com/stablecog/sc-go/shared"
)

type BroadcastPayload struct {
	ID      string `json:"id"`
	Message []byte `json:"message"`
}

type Hub struct {
	// Events are pushed to this channel by the main events-gathering routine
	Broadcast chan BroadcastPayload

	// We need to send keepalives to clients so connections stay alive
	KeepAlive chan bool

	// Clients connecting
	Register chan *Client

	// Clients disconnecting
	Unregister chan *Client

	// Client connections registry
	clients map[*Client]bool

	// Database access
	Repo  *repository.Repository
	Redis *database.RedisWrapper
}

func NewHub(redis *database.RedisWrapper, repo *repository.Repository) *Hub {
	return &Hub{
		Broadcast:  make(chan BroadcastPayload, 100),
		KeepAlive:  make(chan bool),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		Repo:       repo,
		Redis:      redis,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.clients[client] = true
		case client := <-h.Unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
			}
		case <-h.KeepAlive:
			keepaliveMsg := map[string]interface{}{
				"keepalive": true,
				"version":   shared.APP_VERSION,
			}
			keepaliveBytes, _ := json.Marshal(keepaliveMsg)
			for client := range h.clients {
				select {
				case client.Send <- keepaliveBytes:
				default:
					close(client.Send)
					delete(h.clients, client)
				}
			}
		case payload := <-h.Broadcast:
			for client := range h.clients {
				if client.Uid == payload.ID {
					select {
					case client.Send <- payload.Message:
					default:
						close(client.Send)
						delete(h.clients, client)
					}
				}
			}
		}
	}
}

func (h *Hub) BraodcastKeepalive() {
	h.KeepAlive <- true
}
