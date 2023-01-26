package websocket

import (
	"fmt"
	"net/http"

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

	userID := fmt.Sprintf("guest_%d", hub.GetGuestCount()+1)

	client := &Client{Hub: hub, Conn: conn, Send: make(chan []byte, 256), Uid: userID}
	client.Hub.Register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}

type WebsocketConnResponse struct {
	Authenticated bool `json:"authenticated"`
}
