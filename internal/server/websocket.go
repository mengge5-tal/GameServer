package server

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"GameServer/internal/handlers/online"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return isOriginAllowed(r.Header.Get("Origin"))
	},
}

// Check if the origin is in the allowed list
func isOriginAllowed(origin string) bool {
	if origin == "" {
		// Allow requests with no origin (e.g., same-origin requests)
		return true
	}

	// Get allowed origins from global configuration
	// Note: This requires passing config to the function or making it global
	// For now, we'll keep the environment variable approach
	allowedOriginsEnv := os.Getenv("WS_ALLOWED_ORIGINS")
	if allowedOriginsEnv == "" {
		allowedOriginsEnv = "http://localhost:3000,http://localhost:8080"
	}
	allowedOrigins := strings.Split(allowedOriginsEnv, ",")

	// Trim whitespace from each origin
	for i, o := range allowedOrigins {
		allowedOrigins[i] = strings.TrimSpace(o)
	}

	// Check if the request origin is in the allowed list
	for _, allowedOrigin := range allowedOrigins {
		if strings.EqualFold(origin, allowedOrigin) {
			return true
		}
	}

	log.Printf("WebSocket connection rejected: origin %s not allowed", origin)
	return false
}

// 客户端连接结构
type Client struct {
	ID       string          // 客户端唯一ID
	UserID   int             // 用户ID（登录后设置）
	Conn     *websocket.Conn // WebSocket连接
	Send     chan []byte     // 发送消息通道
	Hub      *Hub            // 所属的Hub
	IsAuth   bool            // 是否已认证
	LastPing time.Time       // 最后ping时间
}

// 实现Client接口方法
func (c *Client) GetID() string {
	return c.ID
}

func (c *Client) GetUserID() int {
	return c.UserID
}

func (c *Client) SetAuth(auth bool) {
	c.IsAuth = auth
}

func (c *Client) SetUserID(userID int) {
	c.UserID = userID
}

func (c *Client) GetHub() *Hub {
	return c.Hub
}

// 连接管理器
type Hub struct {
	Clients     map[*Client]bool // 所有连接的客户端
	UserClients map[int]*Client  // 用户ID到客户端的映射
	Register    chan *Client     // 注册新客户端
	Unregister  chan *Client     // 注销客户端
	Broadcast   chan []byte      // 广播消息
	DB          *sql.DB          // 数据库连接
	Router      *Router          // 消息路由器
	mutex       sync.RWMutex     // 读写锁
}

// 实现Hub接口方法
func (h *Hub) SetUserClient(userID int, client *Client) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.UserClients[userID] = client
}

func (h *Hub) GetClientByUserID(userID int) *Client {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return h.UserClients[userID]
}

func (h *Hub) RemoveUserClient(userID int) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	delete(h.UserClients, userID)
}

// 创建新的Hub
func NewHub(db *sql.DB, router *Router) *Hub {
	return &Hub{
		Clients:     make(map[*Client]bool),
		UserClients: make(map[int]*Client),
		Register:    make(chan *Client),
		Unregister:  make(chan *Client),
		Broadcast:   make(chan []byte),
		DB:          db,
		Router:      router,
	}
}

// Hub运行主循环
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mutex.Lock()
			h.Clients[client] = true
			h.mutex.Unlock()
			log.Printf("Client %s connected", client.ID)

		case client := <-h.Unregister:
			h.mutex.Lock()
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				if client.UserID > 0 {
					// 设置用户离线状态
					onlineService := online.NewOnlineService(h.DB)
					if err := onlineService.SetUserOffline(client.UserID); err != nil {
						log.Printf("Failed to set user %d offline on disconnect: %v", client.UserID, err)
					}
					delete(h.UserClients, client.UserID)
				}
				
				// 从限流器中移除客户端
				if rateLimiter := GetRateLimiter(); rateLimiter != nil {
					rateLimiter.RemoveClient(client.ID)
				}
				
				close(client.Send)
			}
			h.mutex.Unlock()
			log.Printf("Client %s disconnected", client.ID)

		case message := <-h.Broadcast:
			h.mutex.RLock()
			for client := range h.Clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.Clients, client)
					if client.UserID > 0 {
						// 设置用户离线状态
						onlineService := online.NewOnlineService(h.DB)
						if err := onlineService.SetUserOffline(client.UserID); err != nil {
							log.Printf("Failed to set user %d offline on broadcast failure: %v", client.UserID, err)
						}
						delete(h.UserClients, client.UserID)
					}
				}
			}
			h.mutex.RUnlock()
		}
	}
}

// 向指定用户发送消息
func (h *Hub) SendToUser(userID int, message []byte) bool {
	client := h.GetClientByUserID(userID)
	if client != nil {
		select {
		case client.Send <- message:
			return true
		default:
			return false
		}
	}
	return false
}

// 客户端读取消息
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	// 设置读取超时
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

		// 解析消息
		message, err := ParseMessage(messageData)
		if err != nil {
			log.Printf("Failed to parse message: %v", err)
			response := NewErrorResponse("", CodeInvalidRequest, "Invalid message format")
			c.SendResponse(response)
			continue
		}

		// 处理消息
		c.HandleMessage(message)
	}
}

// 客户端写入消息
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

// 发送响应
func (c *Client) SendResponse(response *Response) {
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

// 处理消息路由
func (c *Client) HandleMessage(message *Message) {
	// 使用路由器处理消息
	response := c.Hub.Router.Handle(c, message)

	// 如果响应为nil，说明处理器直接发送了响应（兼容旧处理器）
	if response != nil {
		// 设置时间戳
		response.Timestamp = time.Now().Unix()

		// 发送响应
		c.SendResponse(response)
	}
}

// WebSocket处理器
func (h *Hub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	client := &Client{
		ID:       uuid.New().String(),
		Conn:     conn,
		Send:     make(chan []byte, 256),
		Hub:      h,
		IsAuth:   false,
		LastPing: time.Now(),
	}

	client.Hub.Register <- client

	// 启动读写协程
	go client.WritePump()
	go client.ReadPump()
}
