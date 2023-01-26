package websocket

import (
	"net/http"

	"github.com/stablecog/go-apps/server/responses"
	"k8s.io/klog/v2"
)

// ServeWS handles new connections to the WS service
func ServeWS(hub *Hub, w http.ResponseWriter, r *http.Request) {
	// Retrieve id from query parameters
	query := r.URL.Query()
	requestId := query.Get("id")
	if requestId == "" {
		responses.ErrBadRequest(w, r, "Missing id query parameter")
		return
	}
	// ! TODO - proper cors check
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		klog.Error(err)
		return
	}

	client := &Client{Hub: hub, Conn: conn, Send: make(chan []byte, 256), Uid: requestId}
	client.Hub.Register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}
