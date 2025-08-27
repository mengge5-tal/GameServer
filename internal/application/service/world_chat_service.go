package service

import (
	"GameServer/internal/application/dto"
	"GameServer/internal/domain/entity"
	"GameServer/internal/domain/repository"
	"GameServer/internal/domain/service"
	"GameServer/internal/interfaces/websocket"
	"fmt"
	"sync"
	"time"
)

// WorldChatService 世界聊天服务
type WorldChatService struct {
	worldRepo      repository.WorldChatRepository
	userRepo       repository.UserRepository
	rateLimiter    *service.ChatRateLimiter
	channelManager *ChannelManager
	clientManager  websocket.ClientManagerInterface // 用于消息广播
	mutex          sync.RWMutex
}

// ChannelManager 频道管理器
type ChannelManager struct {
	// 频道ID -> 在线用户映射
	channelUsers map[int]map[int]*ChannelUser
	// 用户ID -> 频道ID映射
	userChannels map[int]int
	mutex        sync.RWMutex
}

// ChannelUser 频道用户信息
type ChannelUser struct {
	UserID   int       `json:"user_id"`
	Username string    `json:"username"`
	JoinTime time.Time `json:"join_time"`
}

// NewWorldChatService 创建世界聊天服务
func NewWorldChatService(
	worldRepo repository.WorldChatRepository,
	userRepo repository.UserRepository,
	rateLimiter *service.ChatRateLimiter,
	clientManager websocket.ClientManagerInterface,
) *WorldChatService {
	return &WorldChatService{
		worldRepo:     worldRepo,
		userRepo:      userRepo,
		rateLimiter:   rateLimiter,
		clientManager: clientManager,
		channelManager: &ChannelManager{
			channelUsers: make(map[int]map[int]*ChannelUser),
			userChannels: make(map[int]int),
		},
	}
}

// SendMessage 发送世界聊天消息
func (s *WorldChatService) SendMessage(userID int, request *dto.SendWorldMessageRequest) (*dto.WorldMessageResponse, error) {
	// 1. 验证请求参数
	if request.Content == "" {
		return nil, fmt.Errorf("消息内容不能为空")
	}

	// 2. 检查发送频率限制
	allowed, waitTime, err := s.rateLimiter.CheckWorldChat(userID)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, fmt.Errorf("世界聊天发送过于频繁，请等待%v后重试", waitTime)
	}

	// 3. 验证用户是否存在
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("用户不存在: %v", err)
	}

	// 4. 获取用户所在的频道
	channelID, err := s.getUserChannel(userID)
	if err != nil {
		return nil, fmt.Errorf("获取用户频道失败: %v", err)
	}

	// 5. 验证消息内容长度
	if len([]rune(request.Content)) > 30 {
		return nil, fmt.Errorf("消息内容不能超过30个字符")
	}

	// 6. 创建世界聊天消息
	message := &entity.WorldChatMessage{
		ChannelID: channelID,
		UserID:    userID,
		Username:  user.Username,
		Content:   request.Content,
		Timestamp: time.Now(),
		MessageID: s.generateMessageID(userID, channelID),
	}

	// 7. 验证消息
	if err := message.Validate(); err != nil {
		return nil, err
	}

	// 8. 构建响应
	response := &dto.WorldMessageResponse{
		ChannelID: message.ChannelID,
		UserID:    message.UserID,
		Username:  message.Username,
		Content:   message.Content,
		Timestamp: message.Timestamp.Format("2006-01-02 15:04:05"),
		MessageID: message.MessageID,
	}

	return response, nil
}

// JoinChannel 加入世界聊天频道
func (s *WorldChatService) JoinChannel(userID int, request *dto.JoinWorldChannelRequest) (*dto.WorldChannelResponse, error) {
	// 1. 验证用户是否存在
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("用户不存在: %v", err)
	}

	var targetChannelID int

	// 2. 如果没有指定频道ID，自动分配
	if request.ChannelID == 0 {
		channel, err := s.worldRepo.GetLeastUsersChannel()
		if err != nil {
			return nil, fmt.Errorf("获取可用频道失败: %v", err)
		}
		targetChannelID = channel.ChannelID
	} else {
		targetChannelID = request.ChannelID
	}

	// 3. 获取目标频道信息
	channel, err := s.worldRepo.GetChannelByID(targetChannelID)
	if err != nil {
		return nil, fmt.Errorf("频道不存在: %v", err)
	}

	// 4. 检查频道是否可以加入
	if !channel.CanJoin() {
		return nil, fmt.Errorf("频道已满或不可用")
	}

	// 5. 先从当前频道离开
	if err := s.leaveCurrentChannel(userID); err != nil {
		return nil, fmt.Errorf("离开当前频道失败: %v", err)
	}

	// 6. 加入新频道
	if err := s.joinChannelInternal(userID, user.Username, targetChannelID); err != nil {
		return nil, fmt.Errorf("加入频道失败: %v", err)
	}

	// 7. 更新数据库中的频道用户数
	if err := s.worldRepo.IncrementChannelUsers(targetChannelID); err != nil {
		// 如果数据库更新失败，回滚内存状态
		s.channelManager.removeUser(userID)
		return nil, fmt.Errorf("更新频道用户数失败: %v", err)
	}

	// 8. 重新获取频道信息
	updatedChannel, err := s.worldRepo.GetChannelByID(targetChannelID)
	if err != nil {
		return nil, fmt.Errorf("获取更新后频道信息失败: %v", err)
	}

	// 9. 构建响应
	response := &dto.WorldChannelResponse{
		ChannelID:    updatedChannel.ChannelID,
		ChannelName:  updatedChannel.ChannelName,
		MaxUsers:     updatedChannel.MaxUsers,
		CurrentUsers: updatedChannel.CurrentUsers,
		IsActive:     updatedChannel.IsActive,
	}

	return response, nil
}

// LeaveChannel 离开世界聊天频道
func (s *WorldChatService) LeaveChannel(userID int) error {
	return s.leaveCurrentChannel(userID)
}

// GetChannels 获取所有世界聊天频道
func (s *WorldChatService) GetChannels(userID int) (*dto.GetWorldChannelsResponse, error) {
	// 1. 获取所有活跃频道
	channels, err := s.worldRepo.GetActiveChannels()
	if err != nil {
		return nil, fmt.Errorf("获取频道列表失败: %v", err)
	}

	// 2. 构建频道响应列表
	var channelResponses []dto.WorldChannelResponse
	for _, channel := range channels {
		channelResponse := dto.WorldChannelResponse{
			ChannelID:    channel.ChannelID,
			ChannelName:  channel.ChannelName,
			MaxUsers:     channel.MaxUsers,
			CurrentUsers: channel.CurrentUsers,
			IsActive:     channel.IsActive,
		}
		channelResponses = append(channelResponses, channelResponse)
	}

	// 3. 获取用户当前频道信息
	var currentUserInfo *dto.UserChannelInfo
	if currentChannelID, exists := s.channelManager.getUserChannel(userID); exists {
		if channelUser, exists := s.channelManager.getChannelUser(currentChannelID, userID); exists {
			currentUserInfo = &dto.UserChannelInfo{
				UserID:    userID,
				ChannelID: currentChannelID,
				JoinTime:  channelUser.JoinTime.Format("2006-01-02 15:04:05"),
			}
		}
	}

	// 4. 构建响应
	response := &dto.GetWorldChannelsResponse{
		Channels:      channelResponses,
		CurrentUser:   currentUserInfo,
		TotalChannels: len(channelResponses),
	}

	return response, nil
}

// GetChannelUsers 获取频道内的用户列表
func (s *WorldChatService) GetChannelUsers(channelID int) ([]*ChannelUser, error) {
	// 1. 验证频道是否存在
	_, err := s.worldRepo.GetChannelByID(channelID)
	if err != nil {
		return nil, fmt.Errorf("频道不存在: %v", err)
	}

	// 2. 获取频道用户
	users := s.channelManager.getChannelUsers(channelID)
	return users, nil
}

// BroadcastToChannel 向频道广播消息
func (s *WorldChatService) BroadcastToChannel(channelID int, message *dto.WorldMessageResponse) error {
	// 使用ClientManager进行真实的消息广播
	if s.clientManager != nil {
		successCount := s.clientManager.BroadcastToWorldChannel(channelID, message, message.UserID)
		fmt.Printf("世界频道 %d 广播消息成功发送给 %d 个用户\n", channelID, successCount)
		return nil
	}
	
	// 如果没有ClientManager，使用本地频道管理器记录日志
	users := s.channelManager.getChannelUsers(channelID)
	for _, user := range users {
		fmt.Printf("向用户 %d (%s) 广播世界聊天消息: %s\n", user.UserID, user.Username, message.Content)
	}

	fmt.Printf("世界频道 %d 广播消息: %s (来自用户 %d)\n", channelID, message.Content, message.UserID)
	return nil
}

// 内部方法

// getUserChannel 获取用户所在的频道ID
func (s *WorldChatService) getUserChannel(userID int) (int, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// 检查用户是否已在某个频道
	if channelID, exists := s.channelManager.getUserChannel(userID); exists {
		return channelID, nil
	}

	// 用户不在任何频道，自动分配到用户最少的频道
	channel, err := s.worldRepo.GetLeastUsersChannel()
	if err != nil {
		return 0, fmt.Errorf("获取可用频道失败: %v", err)
	}

	// 获取用户信息
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return 0, fmt.Errorf("获取用户信息失败: %v", err)
	}

	// 自动加入频道
	if err := s.joinChannelInternal(userID, user.Username, channel.ChannelID); err != nil {
		return 0, fmt.Errorf("自动加入频道失败: %v", err)
	}

	// 更新数据库
	if err := s.worldRepo.IncrementChannelUsers(channel.ChannelID); err != nil {
		// 回滚内存状态
		s.channelManager.removeUser(userID)
		return 0, fmt.Errorf("更新频道用户数失败: %v", err)
	}

	return channel.ChannelID, nil
}

// joinChannelInternal 内部加入频道方法
func (s *WorldChatService) joinChannelInternal(userID int, username string, channelID int) error {
	channelUser := &ChannelUser{
		UserID:   userID,
		Username: username,
		JoinTime: time.Now(),
	}

	return s.channelManager.addUser(channelID, channelUser)
}

// leaveCurrentChannel 离开当前频道
func (s *WorldChatService) leaveCurrentChannel(userID int) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 获取用户当前频道
	currentChannelID, exists := s.channelManager.getUserChannel(userID)
	if !exists {
		// 用户不在任何频道
		return nil
	}

	// 从内存中移除用户
	s.channelManager.removeUser(userID)

	// 更新数据库中的频道用户数
	if err := s.worldRepo.DecrementChannelUsers(currentChannelID); err != nil {
		return fmt.Errorf("更新频道用户数失败: %v", err)
	}

	return nil
}

// generateMessageID 生成消息ID
func (s *WorldChatService) generateMessageID(userID, channelID int) string {
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("world_%d_%d_%d", channelID, userID, timestamp)
}

// OnUserDisconnect 用户断开连接时调用
func (s *WorldChatService) OnUserDisconnect(userID int) {
	if err := s.leaveCurrentChannel(userID); err != nil {
		fmt.Printf("用户 %d 断开连接时离开频道失败: %v\n", userID, err)
	}
}

// GetUserChannelInfo 获取用户频道信息
func (s *WorldChatService) GetUserChannelInfo(userID int) (*dto.UserChannelInfo, error) {
	if currentChannelID, exists := s.channelManager.getUserChannel(userID); exists {
		if channelUser, exists := s.channelManager.getChannelUser(currentChannelID, userID); exists {
			return &dto.UserChannelInfo{
				UserID:    userID,
				ChannelID: currentChannelID,
				JoinTime:  channelUser.JoinTime.Format("2006-01-02 15:04:05"),
			}, nil
		}
	}

	return nil, fmt.Errorf("用户不在任何频道中")
}

// ChannelManager 方法实现

// addUser 添加用户到频道
func (cm *ChannelManager) addUser(channelID int, user *ChannelUser) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// 初始化频道用户映射
	if cm.channelUsers[channelID] == nil {
		cm.channelUsers[channelID] = make(map[int]*ChannelUser)
	}

	// 检查用户是否已在频道中
	if _, exists := cm.channelUsers[channelID][user.UserID]; exists {
		return fmt.Errorf("用户已在频道中")
	}

	// 添加用户到频道
	cm.channelUsers[channelID][user.UserID] = user
	cm.userChannels[user.UserID] = channelID

	return nil
}

// removeUser 从频道中移除用户
func (cm *ChannelManager) removeUser(userID int) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// 获取用户当前频道
	channelID, exists := cm.userChannels[userID]
	if !exists {
		return
	}

	// 从频道中移除用户
	if cm.channelUsers[channelID] != nil {
		delete(cm.channelUsers[channelID], userID)
		
		// 如果频道没有用户了，清理频道映射
		if len(cm.channelUsers[channelID]) == 0 {
			delete(cm.channelUsers, channelID)
		}
	}

	// 清理用户频道映射
	delete(cm.userChannels, userID)
}

// getUserChannel 获取用户所在频道
func (cm *ChannelManager) getUserChannel(userID int) (int, bool) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	channelID, exists := cm.userChannels[userID]
	return channelID, exists
}

// getChannelUsers 获取频道内的所有用户
func (cm *ChannelManager) getChannelUsers(channelID int) []*ChannelUser {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	users := make([]*ChannelUser, 0)
	if channelUsers, exists := cm.channelUsers[channelID]; exists {
		for _, user := range channelUsers {
			users = append(users, user)
		}
	}

	return users
}

// getChannelUser 获取频道中的特定用户
func (cm *ChannelManager) getChannelUser(channelID, userID int) (*ChannelUser, bool) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	if channelUsers, exists := cm.channelUsers[channelID]; exists {
		if user, exists := channelUsers[userID]; exists {
			return user, true
		}
	}

	return nil, false
}

// GetChannelUserCount 获取频道用户数量
func (cm *ChannelManager) GetChannelUserCount(channelID int) int {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	if channelUsers, exists := cm.channelUsers[channelID]; exists {
		return len(channelUsers)
	}

	return 0
}

// SetClientManager 设置客户端管理器
func (s *WorldChatService) SetClientManager(clientManager websocket.ClientManagerInterface) {
	s.clientManager = clientManager
}