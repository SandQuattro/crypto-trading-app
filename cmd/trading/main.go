package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"

	"github.com/sand/crypto-trading-app/backend/internal/handlers"
	"github.com/sand/crypto-trading-app/backend/internal/services"
	"github.com/sand/crypto-trading-app/backend/internal/websocket"
)

// Server timeout constants.
const (
	readTimeoutSeconds     = 15
	writeTimeoutSeconds    = 15
	idleTimeoutSeconds     = 60
	shutdownTimeoutSeconds = 5
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	// Create services and components
	dataService := services.NewDataService(logger)
	websocketManager := websocket.NewWebSocketManager(logger)

	// Create handlers
	httpHandler := handlers.NewHTTPHandler(logger, dataService)
	wsHandler := handlers.NewWebSocketHandler(logger, dataService, websocketManager)

	// Initialize trading pairs
	dataService.InitializeTradingPairs()

	// Create router
	router := mux.NewRouter()

	// Register WebSocket routes before HTTP routes
	wsHandler.RegisterRoutes(router)
	httpHandler.RegisterRoutes(router)

	// Configure CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	// Wrap router in CORS middleware
	handler := c.Handler(router)

	// Create HTTP server with timeouts
	port := ":8080"
	srv := &http.Server{
		Addr:         port,
		Handler:      handler,
		ReadTimeout:  readTimeoutSeconds * time.Second,
		WriteTimeout: writeTimeoutSeconds * time.Second,
		IdleTimeout:  idleTimeoutSeconds * time.Second,
	}

	// Start server in a separate goroutine
	go func() {
		log.Printf("Starting server on %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error starting server: %v", err)
		}
	}()

	// Set up graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Give 5 seconds to complete current requests
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeoutSeconds*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
		return
	}

	log.Println("Server exited properly")
}
