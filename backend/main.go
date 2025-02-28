package main

import (
	"context"
	"log"
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

func main() {
	// Create services and components
	dataService := services.NewDataService()
	websocketManager := websocket.NewWebSocketManager()

	// Create handlers
	httpHandler := handlers.NewHTTPHandler(dataService)
	wsHandler := handlers.NewWebSocketHandler(dataService, websocketManager)

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
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")
}
