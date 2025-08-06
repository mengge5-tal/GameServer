package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"GameServer/internal/infrastructure/config"
	"GameServer/internal/infrastructure/container"
	"GameServer/internal/infrastructure/database"
	"GameServer/internal/interfaces/websocket"
	"GameServer/pkg/logger"
	"GameServer/pkg/metrics"
)

func main() {
	log.Println("Starting Game Server with new architecture...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	log.Printf("Configuration loaded successfully")

	// Initialize logging system
	logger.Init(cfg.Logging.Level, cfg.Logging.Format)
	logger.Info("Logging system initialized", map[string]interface{}{
		"level":  cfg.Logging.Level,
		"format": cfg.Logging.Format,
	})

	// Initialize metrics system
	metrics.Init()
	logger.Info("Metrics system initialized")

	// Connect to database
	dbConnection, err := database.NewConnection(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbConnection.Close()

	// Check database tables
	log.Println("Checking database tables...")
	if err := dbConnection.CreateMissingTables(); err != nil {
		log.Fatalf("Error with database tables: %v", err)
	}

	if err := dbConnection.CheckTableStructure(); err != nil {
		log.Fatalf("Error checking table structure: %v", err)
	}

	// Set all users offline on startup
	if err := dbConnection.SetAllUsersOffline(); err != nil {
		log.Printf("Warning: Failed to set users offline: %v", err)
	}

	log.Println("Database initialization completed successfully!")

	// Initialize dependency injection container
	container, err := container.NewContainer(cfg, dbConnection.GetDB())
	if err != nil {
		log.Fatalf("Failed to initialize container: %v", err)
	}
	defer container.Close()

	logger.Info("Dependency injection container initialized")

	// Create WebSocket hub
	hub := websocket.NewHub(container.GetWebSocketServices())
	go hub.Run()

	logger.Info("WebSocket hub started")

	// Setup HTTP routes
	setupRoutes(hub, container, cfg)

	// Start server
	serverAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Server starting on %s", serverAddr)
	log.Printf("WebSocket endpoint: ws://%s/ws", serverAddr)

	// Graceful shutdown
	go func() {
		if err := http.ListenAndServe(serverAddr, nil); err != nil {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for shutdown signal
	waitForShutdown(hub, dbConnection)
	log.Println("Server shutdown completed")
}

// setupRoutes configures all HTTP routes
func setupRoutes(hub *websocket.Hub, container *container.Container, cfg *config.Config) {
	// WebSocket endpoint
	http.HandleFunc("/ws", hub.HandleWebSocket)

	// Health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := map[string]interface{}{
			"status":    "healthy",
			"timestamp": time.Now().Unix(),
			"version":   "2.0.0", // Updated version for new architecture
			"architecture": "clean-architecture",
		}
		json.NewEncoder(w).Encode(response)
	})

	// Metrics endpoint
	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		currentMetrics := metrics.GetMetrics()
		json.NewEncoder(w).Encode(currentMetrics)
	})

	// API info endpoint
	http.HandleFunc("/info", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		info := map[string]interface{}{
			"service":      "GameServer",
			"version":      "2.0.0",
			"architecture": "Clean Architecture with DDD",
			"endpoints": map[string]string{
				"websocket": "/ws",
				"health":    "/health",
				"metrics":   "/metrics",
				"info":      "/info",
			},
			"features": []string{
				"User Authentication",
				"Player Management", 
				"Equipment System",
				"Friend System",
				"Ranking System",
				"Real-time WebSocket Communication",
			},
		}
		json.NewEncoder(w).Encode(info)
	})

	// Root endpoint
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Write([]byte("Game Server v2.0 with Clean Architecture is running!"))
	})
}

// waitForShutdown waits for interrupt signals for graceful shutdown
func waitForShutdown(hub *websocket.Hub, dbConnection *database.Connection) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	
	log.Println("Shutdown signal received, starting graceful shutdown...")
	
	// Close WebSocket hub to disconnect all clients and set users offline
	if hub != nil {
		log.Println("Closing WebSocket hub...")
		closeHub(hub)
	}
	
	// Set all remaining users offline in database
	if dbConnection != nil {
		log.Println("Setting all users offline...")
		if err := dbConnection.SetAllUsersOffline(); err != nil {
			log.Printf("Warning: Failed to set users offline during shutdown: %v", err)
		}
	}
	
	log.Println("Graceful shutdown completed")
}

// closeHub closes the hub and disconnects all clients
func closeHub(hub *websocket.Hub) {
	// Send close signal to all connected clients
	hub.Mutex.Lock()
	defer hub.Mutex.Unlock()
	
	log.Printf("Disconnecting %d connected clients...", len(hub.Clients))
	
	for client := range hub.Clients {
		if client.UserID > 0 {
			// Set user offline through auth service
			if hub.Services != nil && hub.Services.AuthService != nil {
				if err := hub.Services.AuthService.Logout(client.UserID); err != nil {
					log.Printf("Warning: Failed to logout user %d during shutdown: %v", client.UserID, err)
				} else {
					log.Printf("User %d logged out during shutdown", client.UserID)
				}
			}
		}
		
		// Close client connection
		select {
		case client.Send <- []byte(`{"success":false,"code":1008,"message":"Server shutting down","data":null}`):
		default:
		}
		close(client.Send)
		
		// Close WebSocket connection
		if client.Conn != nil {
			client.Conn.Close()
		}
	}
	
	// Clear all client mappings
	for client := range hub.Clients {
		delete(hub.Clients, client)
	}
	
	for userID := range hub.UserClients {
		delete(hub.UserClients, userID)
	}
}