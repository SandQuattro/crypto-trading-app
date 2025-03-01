package websocket

import (
	"log/slog"
	"net/http"

	"github.com/gorilla/websocket"
)

// Buffer size constants to avoid magic numbers.
const (
	defaultBufferSize = 1024 // 1KB buffer size for WebSocket connections
)

type Manager struct {
	upgrader websocket.Upgrader
}

func NewWebSocketManager() *Manager {
	return &Manager{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  defaultBufferSize,
			WriteBufferSize: defaultBufferSize,
			CheckOrigin: func(_ *http.Request) bool {
				return true // Allow connections from any origin
			},
		},
	}
}

func (m *Manager) Upgrade(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	conn, err := m.upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("Error upgrading to WebSocket", "error", err)
		return nil, err
	}

	// Set handler for connection closure
	conn.SetCloseHandler(func(code int, text string) error {
		slog.Info("WebSocket connection closed", "code", code, "text", text)
		return nil
	})

	return conn, nil
}
