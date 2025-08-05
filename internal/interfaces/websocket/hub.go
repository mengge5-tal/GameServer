package websocket

import (
	"log"
	"net/http"
	"sync"
)

// Hub maintains the set of active clients and broadcasts messages to clients
type Hub struct {
	// Registered clients
	Clients map[*Client]bool

	// User ID to client mapping for direct messaging
	UserClients map[int]*Client

	// Register requests from the clients
	Register chan *Client

	// Unregister requests from clients
	Unregister chan *Client

	// Broadcast message to all clients
	Broadcast chan []byte

	// Message router for handling different message types
	Router MessageRouter

	// Mutex for thread-safe operations
	mutex sync.RWMutex

	// Services
	Services *ServiceContainer
}

// ServiceContainer holds all application services
type ServiceContainer struct {
	AuthService    AuthServiceInterface
	PlayerService  PlayerServiceInterface
	FriendService  FriendServiceInterface
	RankingService RankingServiceInterface
}

// NewHub creates a new Hub instance
func NewHub(services *ServiceContainer) *Hub {
	hub := &Hub{
		Clients:     make(map[*Client]bool),
		UserClients: make(map[int]*Client),
		Register:    make(chan *Client),
		Unregister:  make(chan *Client),
		Broadcast:   make(chan []byte),
		Services:    services,
		Router:      NewMessageRouter(services),
	}

	return hub
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.registerClient(client)

		case client := <-h.Unregister:
			h.unregisterClient(client)

		case message := <-h.Broadcast:
			h.broadcastMessage(message)
		}
	}
}

// registerClient registers a new client
func (h *Hub) registerClient(client *Client) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.Clients[client] = true
	log.Printf("Client %s connected", client.ID)
}

// unregisterClient unregisters a client
func (h *Hub) unregisterClient(client *Client) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if _, ok := h.Clients[client]; !ok {
		return
	}

	// Remove from clients map
	delete(h.Clients, client)

	// Remove from user clients map if authenticated
	if client.UserID > 0 {
		delete(h.UserClients, client.UserID)
		
		// Set user offline status
		// This could be handled by a service call
		log.Printf("User %d disconnected", client.UserID)
	}

	// Close send channel
	close(client.Send)
	
	log.Printf("Client %s disconnected", client.ID)
}

// broadcastMessage broadcasts a message to all clients
func (h *Hub) broadcastMessage(message []byte) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for client := range h.Clients {
		select {
		case client.Send <- message:
		default:
			// Client's send channel is full, close it
			close(client.Send)
			delete(h.Clients, client)
			if client.UserID > 0 {
				delete(h.UserClients, client.UserID)
			}
		}
	}
}

// SetUserClient associates a user ID with a client
func (h *Hub) SetUserClient(userID int, client *Client) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.UserClients[userID] = client
}

// GetClientByUserID retrieves a client by user ID
func (h *Hub) GetClientByUserID(userID int) *Client {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return h.UserClients[userID]
}

// RemoveUserClient removes a user client mapping
func (h *Hub) RemoveUserClient(userID int) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	delete(h.UserClients, userID)
}

// SendToUser sends a message to a specific user
func (h *Hub) SendToUser(userID int, message []byte) bool {
	client := h.GetClientByUserID(userID)
	if client == nil {
		return false
	}

	select {
	case client.Send <- message:
		return true
	default:
		return false
	}
}

// HandleWebSocket handles WebSocket upgrade and creates new client
func (h *Hub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	client := NewClient(conn, h)
	h.Register <- client

	// Start client goroutines
	go client.WritePump()
	go client.ReadPump()
}