package websocket

import (
	"GameServer/internal/domain/valueobject"
	"encoding/json"
	"log"
	"sync"
	"time"
)

// ClientManager 客户端管理器
type ClientManager struct {
	// 在线用户映射 userID -> Client
	clients map[int]*Client
	
	// 用户工会映射 userID -> unionID (用于工会聊天广播)
	userUnions map[int]int
	
	// 工会成员映射 unionID -> set of userIDs (用于快速获取工会成员)
	unionMembers map[int]map[int]bool
	
	// 世界频道用户映射 channelID -> set of userIDs
	worldChannelUsers map[int]map[int]bool
	
	// 用户频道映射 userID -> channelID
	userChannels map[int]int
	
	// 读写锁保护并发访问
	mutex sync.RWMutex
	
	// 服务依赖
	services *ServiceContainer
	
	// 统计信息
	stats *ClientStats
}

// ClientStats 客户端统计信息
type ClientStats struct {
	TotalConnections    int64     `json:"total_connections"`    // 总连接数
	CurrentConnections  int       `json:"current_connections"`  // 当前连接数
	PeakConnections     int       `json:"peak_connections"`     // 峰值连接数
	MessagesSent        int64     `json:"messages_sent"`        // 发送消息数
	MessagesReceived    int64     `json:"messages_received"`    // 接收消息数
	LastPeakTime        time.Time `json:"last_peak_time"`       // 上次峰值时间
	StartTime           time.Time `json:"start_time"`           // 启动时间
}

// NewClientManager 创建客户端管理器
func NewClientManager(services *ServiceContainer) *ClientManager {
	return &ClientManager{
		clients:           make(map[int]*Client),
		userUnions:        make(map[int]int),
		unionMembers:      make(map[int]map[int]bool),
		worldChannelUsers: make(map[int]map[int]bool),
		userChannels:      make(map[int]int),
		services:          services,
		stats: &ClientStats{
			StartTime: time.Now(),
		},
	}
}

// AddClient 添加客户端
func (cm *ClientManager) AddClient(userID int, client *Client) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// 如果用户已经在线，先断开旧连接
	if existingClient, exists := cm.clients[userID]; exists {
		log.Printf("用户 %d 重复登录，断开旧连接", userID)
		go existingClient.Close() // 异步关闭避免死锁
	}

	// 添加新客户端
	cm.clients[userID] = client
	
	// 更新统计信息
	cm.stats.TotalConnections++
	cm.stats.CurrentConnections = len(cm.clients)
	if cm.stats.CurrentConnections > cm.stats.PeakConnections {
		cm.stats.PeakConnections = cm.stats.CurrentConnections
		cm.stats.LastPeakTime = time.Now()
	}

	log.Printf("用户 %d 已连接，当前在线用户数: %d", userID, cm.stats.CurrentConnections)
}

// RemoveClient 移除客户端
func (cm *ClientManager) RemoveClient(userID int) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// 检查客户端是否存在
	if _, exists := cm.clients[userID]; !exists {
		return
	}

	// 移除客户端
	delete(cm.clients, userID)
	
	// 清理世界频道映射
	if channelID, exists := cm.userChannels[userID]; exists {
		if users, exists := cm.worldChannelUsers[channelID]; exists {
			delete(users, userID)
			if len(users) == 0 {
				delete(cm.worldChannelUsers, channelID)
			}
		}
		delete(cm.userChannels, userID)
		
		// 通知世界聊天服务用户断开
		if cm.services.WorldChatService != nil {
			cm.services.WorldChatService.OnUserDisconnect(userID)
		}
	}
	
	// 清理工会映射
	if unionID, exists := cm.userUnions[userID]; exists {
		if members, exists := cm.unionMembers[unionID]; exists {
			delete(members, userID)
			if len(members) == 0 {
				delete(cm.unionMembers, unionID)
			}
		}
		delete(cm.userUnions, userID)
	}
	
	// 更新统计信息
	cm.stats.CurrentConnections = len(cm.clients)

	log.Printf("用户 %d 已断开连接，当前在线用户数: %d", userID, cm.stats.CurrentConnections)
}

// GetClient 获取客户端
func (cm *ClientManager) GetClient(userID int) (*Client, bool) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	client, exists := cm.clients[userID]
	return client, exists
}

// IsOnline 检查用户是否在线
func (cm *ClientManager) IsOnline(userID int) bool {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	_, exists := cm.clients[userID]
	return exists
}

// GetOnlineUsers 获取所有在线用户ID
func (cm *ClientManager) GetOnlineUsers() []int {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	userIDs := make([]int, 0, len(cm.clients))
	for userID := range cm.clients {
		userIDs = append(userIDs, userID)
	}
	return userIDs
}

// GetOnlineCount 获取在线用户数量
func (cm *ClientManager) GetOnlineCount() int {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	return len(cm.clients)
}

// ========== 私聊消息推送 ==========

// SendToUser 向指定用户发送消息
func (cm *ClientManager) SendToUser(userID int, message interface{}) bool {
	client, exists := cm.GetClient(userID)
	if !exists {
		log.Printf("用户 %d 不在线，无法发送消息", userID)
		return false
	}

	// 构建WebSocket响应
	response := valueobject.NewSuccessResponseWithUniqueID(
		valueobject.MessageTypeChat,
		valueobject.ActionSendPrivateMessage,
		message,
	)

	// 序列化消息
	data, err := json.Marshal(response)
	if err != nil {
		log.Printf("序列化消息失败: %v", err)
		return false
	}

	// 发送消息
	select {
	case client.Send <- data:
		cm.mutex.Lock()
		cm.stats.MessagesSent++
		cm.mutex.Unlock()
		return true
	default:
		log.Printf("向用户 %d 发送消息失败：发送通道已满", userID)
		return false
	}
}

// ========== 世界聊天频道管理 ==========

// JoinWorldChannel 用户加入世界聊天频道
func (cm *ClientManager) JoinWorldChannel(userID, channelID int) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// 从旧频道中移除
	if oldChannelID, exists := cm.userChannels[userID]; exists {
		if users, exists := cm.worldChannelUsers[oldChannelID]; exists {
			delete(users, userID)
			if len(users) == 0 {
				delete(cm.worldChannelUsers, oldChannelID)
			}
		}
	}

	// 加入新频道
	if cm.worldChannelUsers[channelID] == nil {
		cm.worldChannelUsers[channelID] = make(map[int]bool)
	}
	cm.worldChannelUsers[channelID][userID] = true
	cm.userChannels[userID] = channelID

	log.Printf("用户 %d 加入世界频道 %d", userID, channelID)
}

// LeaveWorldChannel 用户离开世界聊天频道
func (cm *ClientManager) LeaveWorldChannel(userID int) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	channelID, exists := cm.userChannels[userID]
	if !exists {
		return
	}

	// 从频道中移除用户
	if users, exists := cm.worldChannelUsers[channelID]; exists {
		delete(users, userID)
		if len(users) == 0 {
			delete(cm.worldChannelUsers, channelID)
		}
	}
	delete(cm.userChannels, userID)

	log.Printf("用户 %d 离开世界频道 %d", userID, channelID)
}

// BroadcastToWorldChannel 向世界频道广播消息
func (cm *ClientManager) BroadcastToWorldChannel(channelID int, message interface{}, excludeUserID int) int {
	cm.mutex.RLock()
	channelUsers, exists := cm.worldChannelUsers[channelID]
	if !exists {
		cm.mutex.RUnlock()
		return 0
	}

	// 复制用户列表避免长时间持锁
	userIDs := make([]int, 0, len(channelUsers))
	for userID := range channelUsers {
		if userID != excludeUserID {
			userIDs = append(userIDs, userID)
		}
	}
	cm.mutex.RUnlock()

	// 构建广播消息
	response := valueobject.NewSuccessResponseWithUniqueID(
		valueobject.MessageTypeChat,
		valueobject.ActionSendWorldMessage,
		message,
	)

	data, err := json.Marshal(response)
	if err != nil {
		log.Printf("序列化世界聊天消息失败: %v", err)
		return 0
	}

	// 广播给频道内所有用户
	successCount := 0
	for _, userID := range userIDs {
		if client, exists := cm.GetClient(userID); exists {
			select {
			case client.Send <- data:
				successCount++
			default:
				log.Printf("向用户 %d 广播世界聊天消息失败：发送通道已满", userID)
			}
		}
	}

	cm.mutex.Lock()
	cm.stats.MessagesSent += int64(successCount)
	cm.mutex.Unlock()

	log.Printf("世界频道 %d 广播消息成功发送给 %d/%d 用户", channelID, successCount, len(userIDs))
	return successCount
}

// GetWorldChannelUsers 获取频道内的在线用户
func (cm *ClientManager) GetWorldChannelUsers(channelID int) []int {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	channelUsers, exists := cm.worldChannelUsers[channelID]
	if !exists {
		return nil
	}

	userIDs := make([]int, 0, len(channelUsers))
	for userID := range channelUsers {
		userIDs = append(userIDs, userID)
	}
	return userIDs
}

// ========== 工会聊天管理 ==========

// JoinUnion 用户加入工会
func (cm *ClientManager) JoinUnion(userID, unionID int) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// 从旧工会中移除
	if oldUnionID, exists := cm.userUnions[userID]; exists && oldUnionID != unionID {
		if members, exists := cm.unionMembers[oldUnionID]; exists {
			delete(members, userID)
			if len(members) == 0 {
				delete(cm.unionMembers, oldUnionID)
			}
		}
	}

	// 加入新工会
	if cm.unionMembers[unionID] == nil {
		cm.unionMembers[unionID] = make(map[int]bool)
	}
	cm.unionMembers[unionID][userID] = true
	cm.userUnions[userID] = unionID

	log.Printf("用户 %d 加入工会 %d", userID, unionID)
}

// LeaveUnion 用户离开工会
func (cm *ClientManager) LeaveUnion(userID int) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	unionID, exists := cm.userUnions[userID]
	if !exists {
		return
	}

	// 从工会中移除用户
	if members, exists := cm.unionMembers[unionID]; exists {
		delete(members, userID)
		if len(members) == 0 {
			delete(cm.unionMembers, unionID)
		}
	}
	delete(cm.userUnions, userID)

	log.Printf("用户 %d 离开工会 %d", userID, unionID)
}

// BroadcastToUnion 向工会广播消息
func (cm *ClientManager) BroadcastToUnion(unionID int, message interface{}, excludeUserID int) int {
	cm.mutex.RLock()
	unionMembers, exists := cm.unionMembers[unionID]
	if !exists {
		cm.mutex.RUnlock()
		return 0
	}

	// 复制成员列表避免长时间持锁
	userIDs := make([]int, 0, len(unionMembers))
	for userID := range unionMembers {
		if userID != excludeUserID {
			userIDs = append(userIDs, userID)
		}
	}
	cm.mutex.RUnlock()

	// 构建广播消息
	response := valueobject.NewSuccessResponseWithUniqueID(
		valueobject.MessageTypeChat,
		valueobject.ActionSendUnionMessage,
		message,
	)

	data, err := json.Marshal(response)
	if err != nil {
		log.Printf("序列化工会聊天消息失败: %v", err)
		return 0
	}

	// 广播给工会内所有用户
	successCount := 0
	for _, userID := range userIDs {
		if client, exists := cm.GetClient(userID); exists {
			select {
			case client.Send <- data:
				successCount++
			default:
				log.Printf("向用户 %d 广播工会聊天消息失败：发送通道已满", userID)
			}
		}
	}

	cm.mutex.Lock()
	cm.stats.MessagesSent += int64(successCount)
	cm.mutex.Unlock()

	log.Printf("工会 %d 广播消息成功发送给 %d/%d 用户", unionID, successCount, len(userIDs))
	return successCount
}

// GetUnionMembers 获取工会内的在线成员
func (cm *ClientManager) GetUnionMembers(unionID int) []int {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	unionMembers, exists := cm.unionMembers[unionID]
	if !exists {
		return nil
	}

	userIDs := make([]int, 0, len(unionMembers))
	for userID := range unionMembers {
		userIDs = append(userIDs, userID)
	}
	return userIDs
}

// ========== 系统管理功能 ==========

// GetStats 获取统计信息
func (cm *ClientManager) GetStats() *ClientStats {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	// 复制统计信息
	stats := *cm.stats
	stats.CurrentConnections = len(cm.clients)
	return &stats
}

// GetDetailedStats 获取详细统计信息
func (cm *ClientManager) GetDetailedStats() map[string]interface{} {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	worldChannelCount := 0
	worldUserCount := 0
	for _, users := range cm.worldChannelUsers {
		worldChannelCount++
		worldUserCount += len(users)
	}

	unionCount := 0
	unionUserCount := 0
	for _, members := range cm.unionMembers {
		unionCount++
		unionUserCount += len(members)
	}

	return map[string]interface{}{
		"connections": map[string]interface{}{
			"current": len(cm.clients),
			"total":   cm.stats.TotalConnections,
			"peak":    cm.stats.PeakConnections,
		},
		"messages": map[string]interface{}{
			"sent":     cm.stats.MessagesSent,
			"received": cm.stats.MessagesReceived,
		},
		"world_chat": map[string]interface{}{
			"active_channels": worldChannelCount,
			"total_users":     worldUserCount,
		},
		"union_chat": map[string]interface{}{
			"active_unions": unionCount,
			"total_users":   unionUserCount,
		},
		"uptime": time.Since(cm.stats.StartTime).String(),
	}
}

// BroadcastToAll 向所有在线用户广播系统消息
func (cm *ClientManager) BroadcastToAll(message interface{}) int {
	cm.mutex.RLock()
	userIDs := make([]int, 0, len(cm.clients))
	for userID := range cm.clients {
		userIDs = append(userIDs, userID)
	}
	cm.mutex.RUnlock()

	// 构建系统广播消息
	response := valueobject.NewSuccessResponseWithUniqueID(
		valueobject.MessageTypeHeartbeat, // 使用已有的消息类型
		"broadcast",
		message,
	)

	data, err := json.Marshal(response)
	if err != nil {
		log.Printf("序列化系统广播消息失败: %v", err)
		return 0
	}

	// 广播给所有用户
	successCount := 0
	for _, userID := range userIDs {
		if client, exists := cm.GetClient(userID); exists {
			select {
			case client.Send <- data:
				successCount++
			default:
				log.Printf("向用户 %d 发送系统广播失败：发送通道已满", userID)
			}
		}
	}

	cm.mutex.Lock()
	cm.stats.MessagesSent += int64(successCount)
	cm.mutex.Unlock()

	log.Printf("系统广播消息成功发送给 %d/%d 用户", successCount, len(userIDs))
	return successCount
}

// CleanupInactiveClients 清理非活跃客户端（定时任务）
func (cm *ClientManager) CleanupInactiveClients() {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	now := time.Now()
	inactiveUsers := make([]int, 0)

	for userID, client := range cm.clients {
		// 检查客户端是否超过5分钟未活跃
		if now.Sub(client.LastActivity) > 5*time.Minute {
			inactiveUsers = append(inactiveUsers, userID)
		}
	}

	// 清理非活跃客户端
	for _, userID := range inactiveUsers {
		if client, exists := cm.clients[userID]; exists {
			log.Printf("清理非活跃用户 %d", userID)
			go client.Close() // 异步关闭
		}
	}

	if len(inactiveUsers) > 0 {
		log.Printf("清理了 %d 个非活跃客户端", len(inactiveUsers))
	}
}

// StartHeartbeatChecker 启动心跳检查器
func (cm *ClientManager) StartHeartbeatChecker() {
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				cm.CleanupInactiveClients()
			}
		}
	}()
	log.Println("心跳检查器已启动")
}