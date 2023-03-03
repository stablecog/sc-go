package sse

import (
	"github.com/stablecog/sc-go/database"
	"github.com/stablecog/sc-go/database/repository"
	"github.com/stablecog/sc-go/log"
)

type BroadcastPayload struct {
	ID      string `json:"id"`
	Message []byte `json:"message"`
}

type Hub struct {
	// Events are pushed to this channel by the main events-gathering routine
	Broadcast chan BroadcastPayload

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
		case payload := <-h.Broadcast:
			for client := range h.clients {
				if client.Uid == payload.ID {
					select {
					case client.Send <- payload.Message:
						log.Infof("Sent message to client %s", client.Uid)
					default:
						close(client.Send)
						delete(h.clients, client)
					}
				}
			}
		}
	}

}
