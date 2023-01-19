package websocket

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/stablecog/go-apps/server/models"
	"k8s.io/klog/v2"
)

// ServeWS handles new connections to the WS service
func ServeWS(hub *Hub, w http.ResponseWriter, r *http.Request) {
	// ! TODO - proper cors check
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		klog.Error(err)
		return
	}

	// Generate a unique ID for this client
	uid := uuid.New()
	wsResp := models.WebsocketConnectedResponse{
		Id: uid.String(),
	}
	// Struct to byte
	uidByte, err := json.Marshal(wsResp)
	if err != nil {
		klog.Error(err)
		return
	}

	client := &Client{Hub: hub, Conn: conn, Send: make(chan []byte, 256), Uid: uid}
	client.Hub.Register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()

	client.Send <- uidByte
}
