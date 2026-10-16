package handler

import (
	"net/http"
	"sync"

	ws "github.com/coder/websocket"
	"github.com/yukiisen/ohana/logger"
)

type ClientHandler struct {
	Mu sync.Mutex
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
		logger.Warnf("Failed to upgrade websocket connection")
		return
	}

	defer func() {
		h.Mu.Lock()
		defer h.Mu.Unlock()

		for i, c := range h.Clients {
			if c == conn {
				h.Clients = append(h.Clients[:i], h.Clients[i+1:]...)
				break
			}
		}

		conn.CloseNow()
	}()

	h.Mu.Lock()
	h.Clients = append(h.Clients, conn)
	h.Mu.Unlock()

	logger.Infof("Client Connected.")

	ctx := conn.CloseRead(r.Context())

	// Wait until the client disconnects.
	<-ctx.Done()

}
