package handler

import (
	"net/http"
	"sync"

	ws "github.com/coder/websocket"
	"github.com/yukiisen/hana/logger"
)

// clientLog scopes client messages so they are tagged [client].
var clientLog = logger.New(logger.LevelInfo).WithScope("client")

type ClientHandler struct {
	Mu      sync.Mutex
	Clients []*ws.Conn
}

// Count returns the number of connected clients.
func (h *ClientHandler) Count() int {
	h.Mu.Lock()
	defer h.Mu.Unlock()

	return len(h.Clients)
}

func (h *ClientHandler) Connection(w http.ResponseWriter, r *http.Request) {
	conn, err := ws.Accept(w, r, nil)
	if err != nil {
		clientLog.Warnf("Failed to upgrade websocket connection: %v", err)
		return
	}

	defer func() {
		h.Mu.Lock()

		for i, c := range h.Clients {
			if c == conn {
				h.Clients = append(h.Clients[:i], h.Clients[i+1:]...)
				break
			}
		}
		remaining := len(h.Clients)
		h.Mu.Unlock()

		conn.CloseNow()
		clientLog.Infof("Client disconnected (%d connected)", remaining)
	}()

	h.Mu.Lock()
	h.Clients = append(h.Clients, conn)
	h.Mu.Unlock()

	clientLog.Infof("Client connected (%d connected)", h.Count())

	ctx := conn.CloseRead(r.Context())

	// Wait until the client disconnects.
	<-ctx.Done()

}
