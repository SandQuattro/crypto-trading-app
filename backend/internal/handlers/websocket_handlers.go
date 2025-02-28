package handlers

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/sand/crypto-trading-app/backend/internal/services"
	"github.com/sand/crypto-trading-app/backend/internal/websocket"
)

type WebSocketHandler struct {
	dataService      *services.DataService
	websocketManager *websocket.Manager
}

func NewWebSocketHandler(dataService *services.DataService, websocketManager *websocket.Manager) *WebSocketHandler {
	return &WebSocketHandler{
		dataService:      dataService,
		websocketManager: websocketManager,
	}
}

func (h *WebSocketHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/ws/{symbol}", h.HandleConnection)
}

func (h *WebSocketHandler) HandleConnection(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	symbol := vars["symbol"]

	// Check if the trading pair exists
	_, exists := h.dataService.TradingPairs[symbol]
	if !exists {
		http.Error(w, "Trading pair not found", http.StatusNotFound)
		return
	}

	conn, err := h.websocketManager.Upgrade(w, r)
	if err != nil {
		log.Printf("Error upgrading connection: %v", err)
		return
	}

	log.Printf("New WebSocket connection for %s", symbol)

	// Add subscriber
	err = h.dataService.AddSubscriber(symbol, conn)
	if err != nil {
		log.Printf("Error adding subscriber: %v", err)
		conn.Close()
		return
	}

	// Handle messages from client
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Printf("WebSocket connection closed for %s: %v", symbol, err)
			if err := h.dataService.RemoveSubscriber(symbol, conn); err != nil {
				log.Printf("Error removing subscriber: %v", err)
			}
			break
		}
	}
}
