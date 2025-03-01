package models

import (
	"sync"

	"github.com/gorilla/websocket"
)

// CandleData represents candle data for the chart
type CandleData struct {
	Time   int64   `json:"time"`   // Time in milliseconds
	Open   float64 `json:"open"`   // Opening price
	High   float64 `json:"high"`   // Highest price
	Low    float64 `json:"low"`    // Lowest price
	Close  float64 `json:"close"`  // Closing price
	Volume float64 `json:"volume"` // Trading volume
}

// TradingPair represents a trading pair
type TradingPair struct {
	Symbol      string                   `json:"symbol"`      // Pair symbol (e.g., BTCUSDT)
	LastPrice   float64                  `json:"lastPrice"`   // Last price
	PriceChange float64                  `json:"priceChange"` // Price change percentage
	CandleData  []CandleData             `json:"-"`           // Historical candle data
	LastCandle  CandleData               `json:"-"`           // Last candle
	Subscribers map[*websocket.Conn]bool `json:"-"`           // WebSocket update subscribers
	Mutex       sync.RWMutex             `json:"-"`           // Mutex for safe data access
	StopChan    chan struct{}            `json:"-"`           // Channel for stopping goroutines
}
