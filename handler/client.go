package handler

import (
	"net/http"
	"sync"

	"github.com/coder/websocket"
)

type ClientHandler struct {
	lock sync.Mutex
	clients []*websocket.Conn
}

func NewClientHandler() ClientHandler {
	return ClientHandler{ lock: sync.Mutex{}, clients: []*websocket.Conn{} }
}

func (c *ClientHandler) Connection(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hi mom\n"))
}
