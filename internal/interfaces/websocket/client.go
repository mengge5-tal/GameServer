package websocket

import (
	"GameServer/internal/domain/valueobject"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// TODO: Implement proper origin checking based on configuration
		return true
	},
}

// Client represents a WebSocket client connection
type Client struct {
	ID       string          // Client unique ID
	UserID   int             // User ID (set after authentication)
	Conn     *websocket.Conn // WebSocket connection
	Send     chan []byte     // Send message channel
	Hub      *Hub            // Owning hub
	IsAuth   bool            // Authentication status
	LastPing time.Time       // Last ping time
}

// NewClient creates a new client instance
func NewClient(conn *websocket.Conn, hub *Hub) *Client {
	return &Client{
		ID:       uuid.New().String(),
		Conn:     conn,
		Send:     make(chan []byte, 256),
		Hub:      hub,
		IsAuth:   false,
		LastPing: time.Now(),
	}
}

// GetID returns client ID
func (c *Client) GetID() string {
	return c.ID
}

// GetUserID returns user ID
func (c *Client) GetUserID() int {
	return c.UserID
}

// SetAuth sets authentication status
func (c *Client) SetAuth(auth bool) {
	c.IsAuth = auth
}

// SetUserID sets user ID
func (c *Client) SetUserID(userID int) {
	c.UserID = userID
}

// IsAuthenticated returns authentication status
func (c *Client) IsAuthenticated() bool {
	return c.IsAuth
}

// SendResponse sends a response to the client
func (c *Client) SendResponse(response *valueobject.Response) {
	data, err := response.ToJSON()
	if err != nil {
		log.Printf("Failed to marshal response: %v", err)
		return
	}

	select {
	case c.Send <- data:
	default:
		log.Printf("Client %s send channel is full", c.ID)
	}
}

// ReadPump handles reading messages from the WebSocket connection
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	// Set read deadline and pong handler
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		c.LastPing = time.Now()
		return nil
	})

	for {
		_, messageData, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Parse message
		message, err := valueobject.ParseMessage(messageData)
		if err != nil {
			log.Printf("Failed to parse message: %v", err)
			response := valueobject.NewErrorResponse("", valueobject.CodeInvalidRequest, "Invalid message format")
			c.SendResponse(response)
			continue
		}

		// Handle message
		c.HandleMessage(message)
	}
}

// WritePump handles writing messages to the WebSocket connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("Failed to write message: %v", err)
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// HandleMessage routes the message to appropriate handler
func (c *Client) HandleMessage(message *valueobject.Message) {
	// Use the message router to handle the message
	response := c.Hub.Router.Handle(c, message)

	// Send response if provided
	if response != nil {
		response.Timestamp = time.Now().Unix()
		c.SendResponse(response)
	}
}