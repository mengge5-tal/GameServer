package server

// 临时处理器方法，用于兼容旧的处理逻辑
// TODO: 将这些方法重构为独立的处理器包

import (
	"GameServer/internal/handlers/auth"
	"GameServer/internal/handlers/online"
	"database/sql"
	"encoding/json"
	"log"
)

// 认证相关处理器
func handleLogin(c *Client, message *Message, db *sql.DB) *Response {
	var loginReq struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	
	dataBytes, err := json.Marshal(message.Data)
	if err != nil {
		return NewErrorResponse(message.RequestID, CodeInvalidParams, "Invalid request data")
	}
	
	err = json.Unmarshal(dataBytes, &loginReq)
	if err != nil {
		return NewErrorResponse(message.RequestID, CodeInvalidParams, "Invalid login data")
	}

	if loginReq.Username == "" || loginReq.Password == "" {
		return NewErrorResponse(message.RequestID, CodeInvalidParams, "Username and password required")
	}

	var user struct {
		UserID   int    `json:"userid"`
		Username string `json:"username"`
	}
	var storedPasswordHash string
	err = db.QueryRow("SELECT userid, username, passward FROM user WHERE username = ?", 
		loginReq.Username).Scan(&user.UserID, &user.Username, &storedPasswordHash)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return NewErrorResponse(message.RequestID, CodeUserNotFound, "Invalid username or password")
		}
		log.Printf("Database error during login: %v", err)
		return NewErrorResponse(message.RequestID, CodeDatabaseError, "Database error")
	}

	if !auth.CheckPasswordHash(loginReq.Password, storedPasswordHash) {
		return NewErrorResponse(message.RequestID, CodeWrongPassword, "Invalid username or password")
	}

	hub := c.Hub
	if existingClient := hub.GetClientByUserID(user.UserID); existingClient != nil {
		// 设置旧连接离线
		onlineService := online.NewOnlineService(db)
		onlineService.SetUserOffline(user.UserID)
		existingClient.SetAuth(false)
		existingClient.SetUserID(0)
	}
	
	c.SetAuth(true)
	c.SetUserID(user.UserID)
	hub.SetUserClient(user.UserID, c)

	// 设置用户在线状态
	onlineService := online.NewOnlineService(db)
	if err := onlineService.SetUserOnline(user.UserID); err != nil {
		log.Printf("Failed to set user %d online: %v", user.UserID, err)
		// 不影响登录流程，只记录错误
	}

	log.Printf("User %s (ID: %d) logged in successfully", user.Username, user.UserID)
	
	return NewSuccessResponse(message.RequestID, user)
}

func handleRegister(c *Client, message *Message, db *sql.DB) *Response {
	var registerReq struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	
	dataBytes, err := json.Marshal(message.Data)
	if err != nil {
		return NewErrorResponse(message.RequestID, CodeInvalidParams, "Invalid request data")
	}
	
	err = json.Unmarshal(dataBytes, &registerReq)
	if err != nil {
		return NewErrorResponse(message.RequestID, CodeInvalidParams, "Invalid register data")
	}

	if registerReq.Username == "" || registerReq.Password == "" {
		return NewErrorResponse(message.RequestID, CodeInvalidParams, "Username and password required")
	}

	if err := auth.ValidateUsername(registerReq.Username); err != nil {
		return NewErrorResponse(message.RequestID, CodeInvalidParams, err.Error())
	}

	if err := auth.ValidatePassword(registerReq.Password); err != nil {
		return NewErrorResponse(message.RequestID, CodeInvalidParams, err.Error())
	}

	var existingUser string
	err = db.QueryRow("SELECT username FROM user WHERE username = ?", registerReq.Username).Scan(&existingUser)
	if err != sql.ErrNoRows {
		if err == nil {
			return NewErrorResponse(message.RequestID, CodeUserExists, "Username already exists")
		}
		log.Printf("Database error during registration check: %v", err)
		return NewErrorResponse(message.RequestID, CodeDatabaseError, "Database error")
	}

	hashedPassword, err := auth.HashPassword(registerReq.Password)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		return NewErrorResponse(message.RequestID, CodeServerError, "Internal server error")
	}

	result, err := db.Exec("INSERT INTO user (username, passward) VALUES (?, ?)", 
		registerReq.Username, hashedPassword)
	if err != nil {
		log.Printf("Database error during registration: %v", err)
		return NewErrorResponse(message.RequestID, CodeDatabaseError, "Failed to create user")
	}

	userID, err := result.LastInsertId()
	if err != nil {
		log.Printf("Failed to get user ID: %v", err)
		return NewErrorResponse(message.RequestID, CodeDatabaseError, "Failed to get user ID")
	}

	_, err = db.Exec("INSERT INTO playerinfo (userid, level, experience, gamelevel, bloodenergy) VALUES (?, 1, 0, 1, 100)", userID)
	if err != nil {
		log.Printf("Failed to create player info: %v", err)
	}

	user := struct {
		UserID   int    `json:"userid"`
		Username string `json:"username"`
	}{
		UserID:   int(userID),
		Username: registerReq.Username,
	}

	log.Printf("User %s registered successfully with ID: %d", registerReq.Username, userID)
	
	return NewSuccessResponse(message.RequestID, user)
}

func handleLogout(c *Client, message *Message, db *sql.DB) *Response {
	if c.GetUserID() > 0 {
		userID := c.GetUserID()
		hub := c.Hub
		hub.RemoveUserClient(userID)
		
		// 设置用户离线状态
		onlineService := online.NewOnlineService(db)
		if err := onlineService.SetUserOffline(userID); err != nil {
			log.Printf("Failed to set user %d offline: %v", userID, err)
			// 不影响登出流程，只记录错误
		}
		
		log.Printf("User ID %d logged out", userID)
		c.SetUserID(0)
		c.SetAuth(false)
	}
	
	return NewSuccessResponse(message.RequestID, "Logged out successfully")
}

func handlePing(c *Client, message *Message) *Response {
	return NewSuccessResponse(message.RequestID, map[string]string{
		"action": "pong",
		"time":   "ok",
	})
}


