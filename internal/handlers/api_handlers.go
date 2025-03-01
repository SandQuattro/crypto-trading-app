package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/sand/crypto-trading-app/backend/internal/services"
)

type HTTPHandler struct {
	logger      *slog.Logger
	dataService *services.DataService
}

func NewHTTPHandler(logger *slog.Logger, dataService *services.DataService) *HTTPHandler {
	return &HTTPHandler{
		logger:      logger,
		dataService: dataService,
	}
}

func (h *HTTPHandler) RegisterRoutes(router *mux.Router) {
	// API endpoints.
	router.HandleFunc("/api/pairs", h.GetTradingPairsHandler).Methods("GET")
	router.HandleFunc("/api/candles/{symbol}", h.GetCandlesHandler).Methods("GET")

	// Static files - register last to avoid intercepting other routes.
	fs := http.FileServer(http.Dir("./static"))
	router.PathPrefix("/").Handler(http.StripPrefix("/", fs))
}

// GetTradingPairsHandler returns a list of trading pairs.
func (h *HTTPHandler) GetTradingPairsHandler(w http.ResponseWriter, _ *http.Request) {
	pairs := make([]map[string]any, 0, len(h.dataService.TradingPairs))

	for _, pair := range h.dataService.TradingPairs {
		pair.Mutex.RLock()
		pairData := map[string]any{
			"symbol":      pair.Symbol,
			"lastPrice":   pair.LastPrice,
			"priceChange": pair.PriceChange,
		}
		pair.Mutex.RUnlock()

		pairs = append(pairs, pairData)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(pairs); err != nil {
		h.logger.Error("Error encoding trading pairs", "error", err)
	}
}

// GetCandlesHandler returns candle data for a trading pair.
func (h *HTTPHandler) GetCandlesHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	symbol := vars["symbol"]

	candles, err := h.dataService.GetCandleData(symbol)
	if err != nil {
		if errors.Is(err, services.ErrTradingPairNotFound) {
			http.Error(w, "Trading pair not found", http.StatusNotFound)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	h.logger.Info("Sending candles", "count", len(candles), "symbol", symbol)
	w.Header().Set("Content-Type", "application/json")
	encodeErr := json.NewEncoder(w).Encode(candles)
	if encodeErr != nil {
		h.logger.Error("Error encoding candles", "error", encodeErr)
	}
}
