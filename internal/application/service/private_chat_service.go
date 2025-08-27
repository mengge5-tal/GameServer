package service

import (
	"GameServer/internal/application/dto"
	"GameServer/internal/domain/entity"
	"GameServer/internal/domain/repository"
	"GameServer/internal/domain/service"
	"fmt"
	"strconv"
	"time"
)

// PrivateChatService 私聊服务
type PrivateChatService struct {
	privateRepo repository.PrivateMessageRepository
	userRepo    repository.UserRepository
	rateLimiter *service.ChatRateLimiter
}

// NewPrivateChatService 创建私聊服务
func NewPrivateChatService(
	privateRepo repository.PrivateMessageRepository,
	userRepo repository.UserRepository,
	rateLimiter *service.ChatRateLimiter,
) *PrivateChatService {
	return &PrivateChatService{
		privateRepo: privateRepo,
		userRepo:    userRepo,
		rateLimiter: rateLimiter,
	}
}

// SendMessage 发送私聊消息
func (s *PrivateChatService) SendMessage(fromUserID int, request *dto.SendPrivateMessageRequest) (*dto.PrivateMessageResponse, error) {
	// 1. 验证请求参数
	if request.ToUserID <= 0 {
		return nil, fmt.Errorf("接收用户ID无效")
	}
	if request.Content == "" {
		return nil, fmt.Errorf("消息内容不能为空")
	}

	// 2. 检查发送频率限制
	allowed, waitTime, err := s.rateLimiter.CheckPrivateChat(fromUserID)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, fmt.Errorf("发送频率过快，请等待%v后重试", waitTime)
	}

	// 3. 验证发送者和接收者是否存在
	fromUser, err := s.userRepo.GetByID(fromUserID)
	if err != nil {
		return nil, fmt.Errorf("发送者不存在: %v", err)
	}

	toUser, err := s.userRepo.GetByID(request.ToUserID)
	if err != nil {
		return nil, fmt.Errorf("接收者不存在: %v", err)
	}

	// 4. 不能给自己发消息
	if fromUserID == request.ToUserID {
		return nil, fmt.Errorf("不能给自己发送消息")
	}

	// 5. 创建私聊消息实体
	message := &entity.PrivateMessage{
		FromUserID: fromUserID,
		ToUserID:   request.ToUserID,
		Content:    request.Content,
		Status:     0, // 未读状态
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// 6. 验证消息内容
	if err := message.Validate(); err != nil {
		return nil, err
	}

	// 7. 保存到数据库
	if err := s.privateRepo.Create(message); err != nil {
		return nil, fmt.Errorf("保存消息失败: %v", err)
	}

	// 8. 构建响应
	response := &dto.PrivateMessageResponse{
		ID:           message.ID,
		FromUserID:   message.FromUserID,
		FromUsername: fromUser.Username,
		ToUserID:     message.ToUserID,
		ToUsername:   toUser.Username,
		Content:      message.Content,
		Status:       message.Status,
		Timestamp:    message.CreatedAt.Format("2006-01-02 15:04:05"),
		IsFromSelf:   true,
	}

	return response, nil
}

// GetMessages 获取用户的私聊消息
func (s *PrivateChatService) GetMessages(userID int, otherUserID int, page, limit int) (*dto.GetPrivateMessagesResponse, error) {
	// 1. 验证参数
	if userID <= 0 || otherUserID <= 0 {
		return nil, fmt.Errorf("用户ID无效")
	}
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 50 {
		limit = 20 // 默认每页20条，最大50条
	}

	// 2. 验证用户是否存在
	currentUser, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("当前用户不存在: %v", err)
	}

	otherUser, err := s.userRepo.GetByID(otherUserID)
	if err != nil {
		return nil, fmt.Errorf("对话用户不存在: %v", err)
	}

	// 3. 获取对话历史
	messages, total, err := s.privateRepo.GetConversationHistory(userID, otherUserID, page, limit)
	if err != nil {
		return nil, fmt.Errorf("获取对话历史失败: %v", err)
	}

	// 4. 构建响应
	var messageResponses []dto.PrivateMessageResponse
	for _, message := range messages {
		var fromUsername, toUsername string
		var isFromSelf bool

		// 确定发送者和接收者的用户名
		if message.FromUserID == userID {
			fromUsername = currentUser.Username
			toUsername = otherUser.Username
			isFromSelf = true
		} else {
			fromUsername = otherUser.Username
			toUsername = currentUser.Username
			isFromSelf = false
		}

		messageResponse := dto.PrivateMessageResponse{
			ID:           message.ID,
			FromUserID:   message.FromUserID,
			FromUsername: fromUsername,
			ToUserID:     message.ToUserID,
			ToUsername:   toUsername,
			Content:      message.Content,
			Status:       message.Status,
			Timestamp:    message.CreatedAt.Format("2006-01-02 15:04:05"),
			IsFromSelf:   isFromSelf,
		}
		messageResponses = append(messageResponses, messageResponse)
	}

	// 5. 标记对话中的未读消息为已读
	if len(messages) > 0 {
		if err := s.privateRepo.MarkConversationAsRead(userID, otherUserID); err != nil {
			// 标记已读失败不影响主要功能，记录日志即可
			fmt.Printf("标记对话消息已读失败: %v\n", err)
		}
	}

	// 计算是否还有更多消息
	totalPages := (total + limit - 1) / limit
	hasMore := page < totalPages

	response := &dto.GetPrivateMessagesResponse{
		Messages: messageResponses,
		Total:    total,
		HasMore:  hasMore,
	}

	return response, nil
}

// GetUnreadMessages 获取用户所有未读私聊消息
func (s *PrivateChatService) GetUnreadMessages(userID int) (*dto.GetPrivateMessagesResponse, error) {
	// 1. 验证参数
	if userID <= 0 {
		return nil, fmt.Errorf("用户ID无效")
	}

	// 2. 验证用户是否存在
	currentUser, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("用户不存在: %v", err)
	}

	// 3. 获取未读消息
	messages, err := s.privateRepo.GetUnreadMessages(userID)
	if err != nil {
		return nil, fmt.Errorf("获取未读消息失败: %v", err)
	}

	// 4. 构建响应，需要获取发送者的用户名
	var messageResponses []dto.PrivateMessageResponse
	userCache := make(map[int]string) // 缓存用户名

	for _, message := range messages {
		// 获取发送者用户名（使用缓存减少数据库查询）
		var fromUsername string
		if cachedUsername, exists := userCache[message.FromUserID]; exists {
			fromUsername = cachedUsername
		} else {
			fromUser, err := s.userRepo.GetByID(message.FromUserID)
			if err != nil {
				// 如果用户不存在，使用ID作为用户名
				fromUsername = "用户" + strconv.Itoa(message.FromUserID)
			} else {
				fromUsername = fromUser.Username
				userCache[message.FromUserID] = fromUsername
			}
		}

		messageResponse := dto.PrivateMessageResponse{
			ID:           message.ID,
			FromUserID:   message.FromUserID,
			FromUsername: fromUsername,
			ToUserID:     message.ToUserID,
			ToUsername:   currentUser.Username,
			Content:      message.Content,
			Status:       message.Status,
			Timestamp:    message.CreatedAt.Format("2006-01-02 15:04:05"),
			IsFromSelf:   false, // 未读消息都是别人发给自己的
		}
		messageResponses = append(messageResponses, messageResponse)
	}

	response := &dto.GetPrivateMessagesResponse{
		Messages: messageResponses,
		Total:    len(messageResponses),
		HasMore:  false, // 未读消息一次性返回全部
	}

	return response, nil
}

// MarkMessageAsRead 标记消息为已读
func (s *PrivateChatService) MarkMessageAsRead(userID int, messageID int64) error {
	// 1. 验证参数
	if userID <= 0 || messageID <= 0 {
		return fmt.Errorf("参数无效")
	}

	// 2. 获取消息详情
	message, err := s.privateRepo.GetByID(messageID)
	if err != nil {
		return fmt.Errorf("消息不存在: %v", err)
	}

	// 3. 验证用户权限（只有接收者可以标记为已读）
	if message.ToUserID != userID {
		return fmt.Errorf("无权限标记此消息为已读")
	}

	// 4. 如果已经是已读状态，直接返回
	if message.Status == 1 {
		return nil
	}

	// 5. 标记为已读
	if err := s.privateRepo.MarkAsRead(messageID); err != nil {
		return fmt.Errorf("标记消息已读失败: %v", err)
	}

	return nil
}

// MarkAllAsRead 标记用户所有未读消息为已读
func (s *PrivateChatService) MarkAllAsRead(userID int) error {
	// 1. 验证参数
	if userID <= 0 {
		return fmt.Errorf("用户ID无效")
	}

	// 2. 标记所有未读消息为已读
	if err := s.privateRepo.MarkAllAsReadForUser(userID); err != nil {
		return fmt.Errorf("标记所有消息已读失败: %v", err)
	}

	return nil
}

// GetUnreadCount 获取用户未读消息数量
func (s *PrivateChatService) GetUnreadCount(userID int) (int, error) {
	// 1. 验证参数
	if userID <= 0 {
		return 0, fmt.Errorf("用户ID无效")
	}

	// 2. 获取未读消息数量
	count, err := s.privateRepo.GetUnreadCount(userID)
	if err != nil {
		return 0, fmt.Errorf("获取未读消息数量失败: %v", err)
	}

	return count, nil
}

// GetConversationPreview 获取对话预览（最新一条消息）
func (s *PrivateChatService) GetConversationPreview(userID, otherUserID int) (*dto.PrivateMessageResponse, error) {
	// 1. 验证参数
	if userID <= 0 || otherUserID <= 0 {
		return nil, fmt.Errorf("用户ID无效")
	}

	// 2. 获取最近的一条消息
	messages, _, err := s.privateRepo.GetConversationHistory(userID, otherUserID, 1, 1)
	if err != nil {
		return nil, fmt.Errorf("获取对话预览失败: %v", err)
	}

	if len(messages) == 0 {
		return nil, fmt.Errorf("无对话记录")
	}

	message := messages[0]

	// 3. 获取用户信息
	var fromUsername, toUsername string
	var isFromSelf bool

	if message.FromUserID == userID {
		// 当前用户发送的消息
		fromUser, err := s.userRepo.GetByID(userID)
		if err != nil {
			fromUsername = "用户" + strconv.Itoa(userID)
		} else {
			fromUsername = fromUser.Username
		}

		toUser, err := s.userRepo.GetByID(otherUserID)
		if err != nil {
			toUsername = "用户" + strconv.Itoa(otherUserID)
		} else {
			toUsername = toUser.Username
		}
		isFromSelf = true
	} else {
		// 对方发送的消息
		fromUser, err := s.userRepo.GetByID(otherUserID)
		if err != nil {
			fromUsername = "用户" + strconv.Itoa(otherUserID)
		} else {
			fromUsername = fromUser.Username
		}

		toUser, err := s.userRepo.GetByID(userID)
		if err != nil {
			toUsername = "用户" + strconv.Itoa(userID)
		} else {
			toUsername = toUser.Username
		}
		isFromSelf = false
	}

	// 4. 构建响应
	response := &dto.PrivateMessageResponse{
		ID:           message.ID,
		FromUserID:   message.FromUserID,
		FromUsername: fromUsername,
		ToUserID:     message.ToUserID,
		ToUsername:   toUsername,
		Content:      message.Content,
		Status:       message.Status,
		Timestamp:    message.CreatedAt.Format("2006-01-02 15:04:05"),
		IsFromSelf:   isFromSelf,
	}

	return response, nil
}

// DeleteMessage 删除消息（只有发送者可以删除，且只能删除未读消息）
func (s *PrivateChatService) DeleteMessage(userID int, messageID int64) error {
	// 1. 验证参数
	if userID <= 0 || messageID <= 0 {
		return fmt.Errorf("参数无效")
	}

	// 2. 获取消息详情
	message, err := s.privateRepo.GetByID(messageID)
	if err != nil {
		return fmt.Errorf("消息不存在: %v", err)
	}

	// 3. 验证用户权限（只有发送者可以删除）
	if message.FromUserID != userID {
		return fmt.Errorf("无权限删除此消息")
	}

	// 4. 验证消息状态（只能删除未读消息）
	if message.Status == 1 {
		return fmt.Errorf("已读消息无法删除")
	}

	// 5. 检查消息发送时间（只能删除5分钟内的消息）
	if time.Since(message.CreatedAt) > 5*time.Minute {
		return fmt.Errorf("只能删除5分钟内发送的消息")
	}

	// 6. 删除消息
	if err := s.privateRepo.Delete(messageID); err != nil {
		return fmt.Errorf("删除消息失败: %v", err)
	}

	return nil
}