# Crypto Trading

A real-time cryptocurrency trading chart built with Go backend and React frontend.

## Features

- Real-time candlestick chart visualization
- Multiple cryptocurrency pair support (BTC/USDT, ETH/USDT, SOL/USDT, BNB/USDT, XRP/USDT)
- Simulated trading data generation
- Responsive design with dark theme
- Consistent trading pair order

## Tech Stack

- **Backend**: Go with Gorilla WebSocket for real-time data streaming
- **Frontend**: React and TradingView Lightweight Charts
- **Communication**: WebSockets for real-time updates, REST API for historical data

## Project Structure

```
crypto-trading-app/
├── backend/           # Go backend
├── frontend/          # React frontend
│   └── src/
│       ├── components/ # React components
│       └── services/   # API services
├── Dockerfile         # Multi-stage Docker build
└── docker-compose.yml # Docker Compose configuration
```

## Getting Started

### Using Docker (Recommended)

The easiest way to run the application is using Docker:

```bash
# Build and run with Docker
make docker

# Stop the Docker container
make docker-stop

# View Docker logs
make docker-logs
```

### Manual Setup

If you prefer to run the application without Docker:

1. Install dependencies:
   ```bash
   make install
   ```

2. Run the backend and frontend in separate terminals:
   ```bash
   # Terminal 1: Run backend
   make run-backend

   # Terminal 2: Run frontend
   make run-frontend
   ```

3. Open your browser and navigate to:
   - Frontend: http://localhost:3000
   - Backend API: http://localhost:8080

## Development

### Backend

The backend is written in Go and provides:
- REST API for trading pairs and historical data
- WebSocket endpoint for real-time updates
- Simulated candlestick data generation

#### Code Quality

The project uses golangci-lint for static code analysis. To run the linter:

```bash
# Install golangci-lint
make lint-install

# Run linter
make lint

# Fix issues automatically
make lint-fix
```

### Frontend

The frontend is built with React and includes:
- TradingView Lightweight Charts for candlestick visualization
- WebSocket connection for real-time updates
- Trading pair selection
- Price display with change indicators

## API Documentation

### Overview

The Crypto Trading App API provides access to cryptocurrency trading pair data, historical candle data, and real-time updates via WebSocket.

### Base URL

```
http://localhost:8080
```

### REST API

#### Get Trading Pairs List

Returns a list of all available trading pairs with current prices and changes.

**URL**: `/api/pairs`

**Method**: `GET`

**Request Example**:
```bash
curl -X GET http://localhost:8080/api/pairs
```

**Successful Response**:

```json
[
  {
    "symbol": "BTCUSDT",
    "lastPrice": 65000.0,
    "priceChange": 2.5
  },
  {
    "symbol": "ETHUSDT",
    "lastPrice": 3500.0,
    "priceChange": 1.2
  },
  {
    "symbol": "SOLUSDT",
    "lastPrice": 180.0,
    "priceChange": 3.7
  },
  {
    "symbol": "BNBUSDT",
    "lastPrice": 600.0,
    "priceChange": -0.5
  },
  {
    "symbol": "XRPUSDT",
    "lastPrice": 0.55,
    "priceChange": 0.8
  }
]
```

**Response Codes**:

- `200 OK`: Successful request
- `500 Internal Server Error`: Server error

#### Get Candle Data

Returns historical candle data for the specified trading pair.

**URL**: `/api/candles/{symbol}`

**Method**: `GET`

**URL Parameters**:

- `{symbol}`: Trading pair symbol (e.g., BTCUSDT)

**Request Example**:
```bash
curl -X GET http://localhost:8080/api/candles/BTCUSDT
```

**Successful Response**:

```json
[
  {
    "time": 1677676800000,
    "open": 64500.0,
    "high": 65100.0,
    "low": 64400.0,
    "close": 65000.0,
    "volume": 100.5
  },
  {
    "time": 1677677100000,
    "open": 65000.0,
    "high": 65200.0,
    "low": 64900.0,
    "close": 65100.0,
    "volume": 90.2
  }
]
```

**Response Codes**:

- `200 OK`: Successful request
- `404 Not Found`: Trading pair not found
- `500 Internal Server Error`: Server error

### WebSocket API

#### WebSocket Connection

To receive real-time updates, the client must establish a WebSocket connection.

**URL**: `ws://localhost:8080/ws/{symbol}`

**URL Parameters**:

- `{symbol}`: Trading pair symbol (e.g., BTCUSDT)

**Connection Example**:

```javascript
const socket = new WebSocket('ws://localhost:8080/ws/BTCUSDT');

socket.onopen = () => {
  console.log('WebSocket connection established');
};

socket.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('Data received:', data);
};

socket.onclose = () => {
  console.log('WebSocket connection closed');
};

socket.onerror = (error) => {
  console.error('WebSocket error:', error);
};
```

#### Message Format

The server sends updates in JSON format:

```json
{
  "time": 1677677400000,
  "open": 65100.0,
  "high": 65300.0,
  "low": 65000.0,
  "close": 65200.0,
  "volume": 110.7
}
```

#### Error Handling

If an error occurs, the server may close the connection. The client should handle such situations and reconnect if necessary.

## Technical Documentation

### Application Architecture

The application is built on the principle of separation of concerns and consists of the following components:

#### Backend (Go)

The backend is divided into the following layers:

1. **Models** (`internal/models/`) - define data structures
2. **Services** (`internal/services/`) - contain business logic
3. **Handlers** (`internal/handlers/`) - process HTTP requests and WebSocket connections
4. **WebSocket** (`internal/websocket/`) - manage WebSocket connections

#### Frontend (React)

The frontend is organized as follows:

1. **Components** (`src/components/`) - UI components
2. **Services** (`src/services/`) - API interaction

### Detailed Backend Description

#### Data Models (`internal/models/`)

##### TradingPair

```go
type TradingPair struct {
    Symbol       string                   // Pair symbol (e.g., BTCUSDT)
    LastPrice    float64                  // Last price
    PriceChange  float64                  // Price change percentage
    CandleData   []CandleData             // Historical candle data
    LastCandle   CandleData               // Last candle
    Subscribers  map[*websocket.Conn]bool // WebSocket update subscribers
    Mutex        sync.RWMutex             // Mutex for safe data access
    StopChan     chan struct{}            // Channel for stopping goroutines
}
```

##### CandleData

```go
type CandleData struct {
    Time   int64   // Time in milliseconds
    Open   float64 // Opening price
    High   float64 // Highest price
    Low    float64 // Lowest price
    Close  float64 // Closing price
    Volume float64 // Trading volume
}
```

#### Services (`internal/services/`)

##### DataService

Responsible for generating and managing trading pair data:

- `InitializeTradingPairs()` - initializes trading pairs with initial data
- `GenerateInitialCandleData()` - generates historical candle data
- `SimulateTradingData()` - simulates real-time trading data
- `BroadcastUpdate()` - sends updates to all subscribers
- `GetCandleData()` - returns candle data for a pair
- `AddSubscriber()` / `RemoveSubscriber()` - manages subscribers

#### WebSocket (`internal/websocket/`)

##### WebSocketManager

Manages WebSocket connections:

- `Upgrade()` - upgrades HTTP connection to WebSocket

#### Handlers (`internal/handlers/`)

##### HTTPHandler

Processes HTTP requests:

- `GetTradingPairsHandler()` - returns list of trading pairs
- `GetCandlesHandler()` - returns candle data for a trading pair

##### WebSocketHandler

Processes WebSocket connections:

- `HandleConnection()` - handles WebSocket connections for the specified trading pair

### WebSocket: Implementation Details

#### Connection Establishment

1. Client initiates WebSocket connection via URL `ws://localhost:8080/ws/{symbol}`
2. Server upgrades HTTP connection to WebSocket using `websocketManager.Upgrade()`
3. Client is added to the subscriber list for the specified trading pair

#### Message Format

##### From Server to Client

The server sends updates in JSON format representing a new candle:

```json
{
  "time": 1677677400000,
  "open": 65100.0,
  "high": 65300.0,
  "low": 65000.0,
  "close": 65200.0,
  "volume": 110.7
}
```

### Concurrency and Synchronization

#### Concurrent Access

- `sync.RWMutex` is used for safe access to trading pair data
- Data reading is protected using `RLock()` / `RUnlock()`
- Data writing is protected using `Lock()` / `Unlock()`

## License

MIT
