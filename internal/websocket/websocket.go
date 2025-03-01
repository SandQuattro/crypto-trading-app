package websocket

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type Manager struct {
	upgrader websocket.Upgrader
}

func NewWebSocketManager() *Manager {
	return &Manager{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(_ *http.Request) bool {
				return true // Allow connections from any origin
			},
		},
	}
}

func (m *Manager) Upgrade(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	conn, err := m.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading to WebSocket: %v", err)
		return nil, err
	}

	// Set handler for connection closure
	conn.SetCloseHandler(func(code int, text string) error {
		log.Printf("WebSocket connection closed with code %d: %s", code, text)
		return nil
	})

	return conn, nil
}
