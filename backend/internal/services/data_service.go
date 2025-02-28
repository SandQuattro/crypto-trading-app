package services

import (
	"crypto/rand"
	"log"
	"math"
	"math/big"
	"time"

	"github.com/gorilla/websocket"

	"github.com/sand/crypto-trading-app/backend/internal/models"
)

type DataService struct {
	TradingPairs map[string]*models.TradingPair
}

func NewDataService() *DataService {
	return &DataService{
		TradingPairs: make(map[string]*models.TradingPair),
	}
}

// NewTradingPair creates a new trading pair
func NewTradingPair(symbol string, initialPrice float64) *models.TradingPair {
	return &models.TradingPair{
		Symbol:      symbol,
		LastPrice:   initialPrice,
		PriceChange: 0,
		CandleData:  make([]models.CandleData, 0),
		Subscribers: make(map[*websocket.Conn]bool),
		StopChan:    make(chan struct{}),
	}
}

// secureFloat64 generates a random number from 0 to 1 using crypto/rand
func secureFloat64() float64 {
	// Generate a random number from 0 to 1<<53
	maxVal := big.NewInt(1 << 53)
	n, err := rand.Int(rand.Reader, maxVal)
	if err != nil {
		// In case of error, return 0.5 as a safe default value
		log.Printf("Error generating secure random number: %v", err)
		return 0.5
	}
	// Convert to float64 from 0 to 1
	return float64(n.Int64()) / float64(maxVal.Int64())
}

// InitializeTradingPairs initializes trading pairs with initial data
func (s *DataService) InitializeTradingPairs() {
	// Create trading pairs with initial prices
	s.TradingPairs["BTCUSDT"] = NewTradingPair("BTCUSDT", 95000.0)
	s.TradingPairs["ETHUSDT"] = NewTradingPair("ETHUSDT", 3500.0)
	s.TradingPairs["SOLUSDT"] = NewTradingPair("SOLUSDT", 180.0)
	s.TradingPairs["BNBUSDT"] = NewTradingPair("BNBUSDT", 600.0)
	s.TradingPairs["XRPUSDT"] = NewTradingPair("XRPUSDT", 0.55)

	// Generate initial candle data
	for _, pair := range s.TradingPairs {
		s.GenerateInitialCandleData(pair)
		// Start simulation in a separate goroutine
		go s.SimulateTradingData(pair)
	}
}

// GenerateInitialCandleData generates initial candle data for a trading pair
func (s *DataService) GenerateInitialCandleData(pair *models.TradingPair) {
	now := time.Now()
	// Round to the beginning of the current 5-minute interval
	currentInterval := time.Date(
		now.Year(), now.Month(), now.Day(),
		now.Hour(), now.Minute()-now.Minute()%5, 0, 0,
		now.Location(),
	)
	startTime := currentInterval.Add(-24 * time.Hour) // 24 hours ago

	// Create slice with required capacity for optimization
	pair.Mutex.Lock()
	defer pair.Mutex.Unlock()

	pair.CandleData = make([]models.CandleData, 0, 288)

	// Base price for the first candle
	basePrice := pair.LastPrice * 0.95

	// Generate candles for the last 24 hours (5-minute candles)
	for i := 0; i < 288; i++ { // 288 candles of 5 minutes each = 24 hours
		candleTime := startTime.Add(time.Duration(i) * 5 * time.Minute)

		// Create a small price change for each candle
		priceChange := basePrice * (secureFloat64()*0.04 - 0.02) // -2% to +2%
		basePrice += priceChange

		// Create candle with random fluctuations
		openPrice := basePrice * (0.995 + secureFloat64()*0.01)
		closePrice := basePrice * (0.995 + secureFloat64()*0.01)
		high := math.Max(openPrice, closePrice) * (1 + secureFloat64()*0.005)
		low := math.Min(openPrice, closePrice) * (0.995 - secureFloat64()*0.005)
		volume := 50 + secureFloat64()*150

		candle := models.CandleData{
			Time:   candleTime.Unix() * 1000, // milliseconds
			Open:   openPrice,
			High:   high,
			Low:    low,
			Close:  closePrice,
			Volume: volume,
		}

		pair.CandleData = append(pair.CandleData, candle)
	}

	// Set last candle
	if len(pair.CandleData) > 0 {
		pair.LastCandle = pair.CandleData[len(pair.CandleData)-1]
		pair.LastPrice = pair.LastCandle.Close
	}

	log.Printf("Generated %d candles for %s", len(pair.CandleData), pair.Symbol)
}

// SimulateTradingData simulates trading data for a pair in real-time
func (s *DataService) SimulateTradingData(pair *models.TradingPair) {
	// Ticker for price updates (every 500ms)
	priceTicker := time.NewTicker(500 * time.Millisecond)
	// Ticker for new candles (every 1 second)
	candleTicker := time.NewTicker(1 * time.Second)

	// Ensure we have candle data
	pair.Mutex.RLock()
	hasData := len(pair.CandleData) > 0
	pair.Mutex.RUnlock()

	if !hasData {
		log.Printf("Generating initial candle data for %s in simulateTradingData", pair.Symbol)
		s.GenerateInitialCandleData(pair)
	}

	// Current candle
	var currentCandle models.CandleData
	pair.Mutex.RLock()
	if len(pair.CandleData) > 0 {
		currentCandle = pair.CandleData[len(pair.CandleData)-1]
	} else {
		now := time.Now()
		// Use a 10-second interval for demonstration
		roundedTime := time.Date(
			now.Year(), now.Month(), now.Day(),
			now.Hour(), now.Minute(), now.Second()/10*10, 0,
			now.Location(),
		)
		currentCandle = models.CandleData{
			Time:   roundedTime.Unix() * 1000,
			Open:   pair.LastPrice,
			High:   pair.LastPrice,
			Low:    pair.LastPrice,
			Close:  pair.LastPrice,
			Volume: 50,
		}
	}
	pair.Mutex.RUnlock()

	for {
		select {
		case <-pair.StopChan:
			priceTicker.Stop()
			candleTicker.Stop()
			return

		case <-priceTicker.C:
			// Update price with small random change
			pair.Mutex.Lock()
			priceChange := pair.LastPrice * (secureFloat64()*0.004 - 0.002) // -0.2% to +0.2%
			pair.LastPrice += priceChange

			// Update current candle
			if pair.LastPrice > currentCandle.High {
				currentCandle.High = pair.LastPrice
			}
			if pair.LastPrice < currentCandle.Low {
				currentCandle.Low = pair.LastPrice
			}
			currentCandle.Close = pair.LastPrice
			currentCandle.Volume += secureFloat64() * 5 // Small increase in volume

			// Update last candle
			pair.LastCandle = currentCandle

			if len(pair.CandleData) > 0 {
				pair.PriceChange = (pair.LastPrice/pair.CandleData[0].Open - 1) * 100 // Calculate % change from first candle
			}
			pair.Mutex.Unlock()

			// Broadcast update to subscribers
			s.BroadcastUpdate(pair)

		case <-candleTicker.C:
			now := time.Now()
			// Use a 10-second interval for demonstration
			roundedTime := time.Date(
				now.Year(), now.Month(), now.Day(),
				now.Hour(), now.Minute(), now.Second()/10*10, 0,
				now.Location(),
			)

			// Check if we need to create a new candle
			if roundedTime.Unix()*1000 > currentCandle.Time {
				pair.Mutex.Lock()

				// Save current candle to history
				if len(pair.CandleData) == 0 || currentCandle.Time > pair.CandleData[len(pair.CandleData)-1].Time {
					pair.CandleData = append(pair.CandleData, currentCandle)
					// Keep only last 288 candles
					if len(pair.CandleData) > 288 {
						pair.CandleData = pair.CandleData[len(pair.CandleData)-288:]
					}
					log.Printf("Created new candle for %s at time %v", pair.Symbol, time.Unix(currentCandle.Time/1000, 0))
				}

				// Create a new current candle
				currentCandle = models.CandleData{
					Time:   roundedTime.Unix() * 1000,
					Open:   pair.LastPrice,
					High:   pair.LastPrice,
					Low:    pair.LastPrice,
					Close:  pair.LastPrice,
					Volume: 50 + secureFloat64()*20,
				}

				// Update last candle
				pair.LastCandle = currentCandle

				pair.Mutex.Unlock()

				// Broadcast update to subscribers
				s.BroadcastUpdate(pair)
			}
		}
	}
}

// BroadcastUpdate sends updates to all subscribers
func (s *DataService) BroadcastUpdate(pair *models.TradingPair) {
	pair.Mutex.RLock()
	defer pair.Mutex.RUnlock()

	// If there are no subscribers, exit
	if len(pair.Subscribers) == 0 {
		return
	}

	// Prepare data for sending
	update := map[string]interface{}{
		"symbol":      pair.Symbol,
		"lastPrice":   pair.LastPrice,
		"priceChange": pair.PriceChange,
		"lastCandle":  pair.LastCandle,
	}

	// Send update to all subscribers
	for conn := range pair.Subscribers {
		err := conn.WriteJSON(update)
		if err != nil {
			log.Printf("Error sending update to subscriber: %v", err)
			conn.Close()
			delete(pair.Subscribers, conn)
		}
	}
}

// GetCandleData returns candle data for a pair
func (s *DataService) GetCandleData(symbol string) ([]models.CandleData, error) {
	pair, ok := s.TradingPairs[symbol]
	if !ok {
		return nil, ErrTradingPairNotFound
	}

	pair.Mutex.RLock()
	defer pair.Mutex.RUnlock()

	if len(pair.CandleData) == 0 {
		// Generate data if it's missing
		s.GenerateInitialCandleData(pair)
	}

	// Create a copy of the data for return
	candles := make([]models.CandleData, len(pair.CandleData))
	copy(candles, pair.CandleData)

	return candles, nil
}

// AddSubscriber adds a subscriber for receiving updates
func (s *DataService) AddSubscriber(symbol string, conn *websocket.Conn) error {
	pair, ok := s.TradingPairs[symbol]
	if !ok {
		return ErrTradingPairNotFound
	}

	pair.Mutex.Lock()
	defer pair.Mutex.Unlock()

	pair.Subscribers[conn] = true
	log.Printf("Added subscriber for %s, total subscribers: %d", symbol, len(pair.Subscribers))
	return nil
}

// RemoveSubscriber removes a subscriber
func (s *DataService) RemoveSubscriber(symbol string, conn *websocket.Conn) error {
	pair, ok := s.TradingPairs[symbol]
	if !ok {
		return ErrTradingPairNotFound
	}

	pair.Mutex.Lock()
	defer pair.Mutex.Unlock()

	delete(pair.Subscribers, conn)
	log.Printf("Removed subscriber for %s, remaining subscribers: %d", symbol, len(pair.Subscribers))
	return nil
}
