package server

import (
	"GameServer/internal/handlers/auth"
	"GameServer/internal/handlers/online"
	"GameServer/internal/models"
	"GameServer/pkg/logger"
	"database/sql"
	"encoding/json"
)

// handleLoginOptimized 优化后的登录处理器，使用缓存
func handleLoginOptimized(c *Client, message *Message, db *sql.DB) *Response {
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
	
	// 查询用户信息
	err = db.QueryRow("SELECT userid, username, passward FROM user WHERE username = ?", 
		loginReq.Username).Scan(&user.UserID, &user.Username, &storedPasswordHash)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return NewErrorResponse(message.RequestID, CodeUserNotFound, "Invalid username or password")
		}
		logger.Error("Database error during login", map[string]interface{}{
			"error":    err.Error(),
			"username": loginReq.Username,
		})
		return NewErrorResponse(message.RequestID, CodeDatabaseError, "Database error")
	}

	// 验证密码
	if !auth.CheckPasswordHash(loginReq.Password, storedPasswordHash) {
		logger.Warn("Failed login attempt", map[string]interface{}{
			"username": loginReq.Username,
			"user_id":  user.UserID,
		})
		return NewErrorResponse(message.RequestID, CodeWrongPassword, "Invalid username or password")
	}

	hub := c.Hub
	
	// 处理重复登录
	if existingClient := hub.GetClientByUserID(user.UserID); existingClient != nil {
		logger.Info("User already logged in, disconnecting previous session", map[string]interface{}{
			"user_id":  user.UserID,
			"username": user.Username,
		})
		
		onlineService := online.NewOnlineService(db)
		onlineService.SetUserOffline(user.UserID)
		existingClient.SetAuth(false)
		existingClient.SetUserID(0)
		
		// 从缓存中清除用户信息，强制重新加载
		if userCache := GetUserCache(); userCache != nil {
			userCache.Delete(user.UserID)
		}
		if playerCache := GetPlayerInfoCache(); playerCache != nil {
			playerCache.Delete(user.UserID)
		}
	}
	
	// 设置客户端状态
	c.SetAuth(true)
	c.SetUserID(user.UserID)
	hub.SetUserClient(user.UserID, c)

	// 设置用户在线状态
	onlineService := online.NewOnlineService(db)
	if err := onlineService.SetUserOnline(user.UserID); err != nil {
		logger.Error("Failed to set user online", map[string]interface{}{
			"user_id": user.UserID,
			"error":   err.Error(),
		})
	}

	// 缓存用户信息
	if userCache := GetUserCache(); userCache != nil {
		userModel := &models.User{
			UserID:   user.UserID,
			Username: user.Username,
		}
		userCache.Set(user.UserID, userModel)
	}

	logger.Info("User logged in successfully", map[string]interface{}{
		"user_id":  user.UserID,
		"username": user.Username,
		"client_id": c.ID,
	})
	
	return NewSuccessResponse(message.RequestID, user)
}

// getUserFromCacheOrDB 从缓存或数据库获取用户信息
func getUserFromCacheOrDB(userID int, db *sql.DB) (*models.User, error) {
	// 先从缓存尝试获取
	if userCache := GetUserCache(); userCache != nil {
		if user, found := userCache.Get(userID); found {
			logger.Debug("User info retrieved from cache", map[string]interface{}{
				"user_id": userID,
			})
			return user, nil
		}
	}
	
	// 缓存未命中，从数据库查询
	var user models.User
	err := db.QueryRow("SELECT userid, username FROM user WHERE userid = ?", userID).
		Scan(&user.UserID, &user.Username)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	
	// 将结果存入缓存
	if userCache := GetUserCache(); userCache != nil {
		userCache.Set(userID, &user)
	}
	
	logger.Debug("User info retrieved from database and cached", map[string]interface{}{
		"user_id": userID,
	})
	
	return &user, nil
}

// handleGetPlayerInfoOptimized 优化后的获取玩家信息处理器
func handleGetPlayerInfoOptimized(c *Client, message *Message, db *sql.DB) *Response {
	if c.UserID == 0 {
		return NewErrorResponse(message.RequestID, CodeUnauthorized, "User not authenticated")
	}
	
	// 先从缓存尝试获取
	if playerCache := GetPlayerInfoCache(); playerCache != nil {
		if cachedInfo, found := playerCache.Get(c.UserID); found {
			logger.Debug("Player info retrieved from cache", map[string]interface{}{
				"user_id": c.UserID,
			})
			
			response := map[string]interface{}{
				"userid":      c.UserID,
				"level":       cachedInfo.Level,
				"experience":  cachedInfo.Experience,
				"gamelevel":   cachedInfo.GameLevel,
				"bloodenergy": cachedInfo.BloodEnergy,
			}
			return NewSuccessResponse(message.RequestID, response)
		}
	}
	
	// 缓存未命中，从数据库查询
	var level, experience, gamelevel, bloodEnergy sql.NullInt64
	err := db.QueryRow("SELECT level, experience, gamelevel, bloodenergy FROM playerinfo WHERE userid = ?",
		c.UserID).Scan(&level, &experience, &gamelevel, &bloodEnergy)

	var playerInfo struct {
		UserID      int `json:"userid"`
		Level       int `json:"level"`
		Experience  int `json:"experience"`
		GameLevel   int `json:"gamelevel"`
		BloodEnergy int `json:"bloodenergy"`
	}
	
	playerInfo.UserID = c.UserID

	if err != nil {
		if err == sql.ErrNoRows {
			// 创建默认玩家信息
			_, err = db.Exec("INSERT INTO playerinfo (userid, level, experience, gamelevel, bloodenergy) VALUES (?, 1, 0, 1, 100)",
				c.UserID)
			if err != nil {
				logger.Error("Failed to create default player info", map[string]interface{}{
					"user_id": c.UserID,
					"error":   err.Error(),
				})
				return NewErrorResponse(message.RequestID, CodeDatabaseError, "Failed to create player info")
			}
			
			// 设置默认值
			playerInfo.Level = 1
			playerInfo.Experience = 0
			playerInfo.GameLevel = 1
			playerInfo.BloodEnergy = 100
		} else {
			logger.Error("Database error getting player info", map[string]interface{}{
				"user_id": c.UserID,
				"error":   err.Error(),
			})
			return NewErrorResponse(message.RequestID, CodeDatabaseError, "Failed to get player info")
		}
	} else {
		// 处理可能的 NULL 值
		playerInfo.Level = int(level.Int64)
		if !level.Valid {
			playerInfo.Level = 1
		}
		
		playerInfo.Experience = int(experience.Int64)
		if !experience.Valid {
			playerInfo.Experience = 0
		}
		
		playerInfo.GameLevel = int(gamelevel.Int64)
		if !gamelevel.Valid {
			playerInfo.GameLevel = 1
		}
		
		playerInfo.BloodEnergy = int(bloodEnergy.Int64)
		if !bloodEnergy.Valid {
			playerInfo.BloodEnergy = 100
		}
	}

	// 存入缓存
	if playerCache := GetPlayerInfoCache(); playerCache != nil {
		playerCache.Set(c.UserID, playerInfo.Level, playerInfo.Experience, 
			playerInfo.GameLevel, playerInfo.BloodEnergy)
	}

	logger.Debug("Player info retrieved from database and cached", map[string]interface{}{
		"user_id": c.UserID,
	})

	return NewSuccessResponse(message.RequestID, playerInfo)
}

// handleUpdatePlayerInfoOptimized 优化后的更新玩家信息处理器
func handleUpdatePlayerInfoOptimized(c *Client, message *Message, db *sql.DB) *Response {
	if c.UserID == 0 {
		return NewErrorResponse(message.RequestID, CodeUnauthorized, "User not authenticated")
	}
	
	var playerInfo struct {
		Level       int `json:"level"`
		Experience  int `json:"experience"`
		GameLevel   int `json:"gamelevel"`
		BloodEnergy int `json:"bloodenergy"`
	}

	dataBytes, err := json.Marshal(message.Data)
	if err != nil {
		return NewErrorResponse(message.RequestID, CodeInvalidParams, "Invalid request data")
	}

	err = json.Unmarshal(dataBytes, &playerInfo)
	if err != nil {
		return NewErrorResponse(message.RequestID, CodeInvalidParams, "Invalid player data")
	}

	// 更新数据库
	_, err = db.Exec("UPDATE playerinfo SET level=?, experience=?, gamelevel=?, bloodenergy=? WHERE userid=?",
		playerInfo.Level, playerInfo.Experience, playerInfo.GameLevel, playerInfo.BloodEnergy, c.UserID)

	if err != nil {
		logger.Error("Database error updating player info", map[string]interface{}{
			"user_id": c.UserID,
			"error":   err.Error(),
		})
		return NewErrorResponse(message.RequestID, CodeDatabaseError, "Failed to update player info")
	}

	// 更新缓存
	if playerCache := GetPlayerInfoCache(); playerCache != nil {
		playerCache.Set(c.UserID, playerInfo.Level, playerInfo.Experience, 
			playerInfo.GameLevel, playerInfo.BloodEnergy)
	}

	// 准备响应
	response := map[string]interface{}{
		"userid":      c.UserID,
		"level":       playerInfo.Level,
		"experience":  playerInfo.Experience,
		"gamelevel":   playerInfo.GameLevel,
		"bloodenergy": playerInfo.BloodEnergy,
	}

	logger.Info("Player info updated successfully", map[string]interface{}{
		"user_id": c.UserID,
		"level":   playerInfo.Level,
	})

	return NewSuccessResponse(message.RequestID, response)
}