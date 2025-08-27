package websocket

import (
	"GameServer/internal/application/dto"
	"GameServer/internal/domain/valueobject"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// ChatHandler 聊天处理器
type ChatHandler struct {
	privateChatService PrivateChatServiceInterface
	worldChatService   WorldChatServiceInterface
	unionChatService   UnionChatServiceInterface
	clientManager      *ClientManager
	services           *ServiceContainer
}

// NewChatHandler 创建聊天处理器
func NewChatHandler(
	privateChatService PrivateChatServiceInterface,
	worldChatService WorldChatServiceInterface,
	unionChatService UnionChatServiceInterface,
	clientManager *ClientManager,
	services *ServiceContainer,
) *ChatHandler {
	return &ChatHandler{
		privateChatService: privateChatService,
		worldChatService:   worldChatService,
		unionChatService:   unionChatService,
		clientManager:      clientManager,
		services:           services,
	}
}

// Handle 实现MessageHandler接口
func (h *ChatHandler) Handle(client *Client, message *valueobject.Message) *valueobject.Response {
	switch message.Action {
	// 私聊相关
	case valueobject.ActionSendPrivateMessage:
		return h.HandleSendPrivateMessage(client, message.RequestID, message.Data)
	case valueobject.ActionGetPrivateMessages:
		return h.HandleGetPrivateMessages(client, message.RequestID, message.Data)
	
	// 世界聊天相关
	case valueobject.ActionSendWorldMessage:
		return h.HandleSendWorldMessage(client, message.RequestID, message.Data)
	case valueobject.ActionJoinWorldChannel:
		return h.HandleJoinWorldChannel(client, message.RequestID, message.Data)
	case valueobject.ActionLeaveWorldChannel:
		return h.HandleLeaveWorldChannel(client, message.RequestID, message.Data)
	case valueobject.ActionGetWorldChannels:
		return h.HandleGetWorldChannels(client, message.RequestID, message.Data)
	
	// 工会聊天相关
	case valueobject.ActionSendUnionMessage:
		return h.HandleSendUnionMessage(client, message.RequestID, message.Data)
	case valueobject.ActionGetUnionMessages:
		return h.HandleGetUnionMessages(client, message.RequestID, message.Data)
	case valueobject.ActionGetRecentUnionMessages:
		return h.HandleGetRecentUnionMessages(client, message.RequestID, message.Data)
	
	default:
		return valueobject.NewErrorResponse(message.RequestID, valueobject.CodeInvalidRequest, "未知的聊天操作")
	}
}

// HandleSendPrivateMessage 处理发送私聊消息
func (h *ChatHandler) HandleSendPrivateMessage(client *Client, requestID string, data json.RawMessage) *valueobject.Response {
	// 1. 检查用户登录状态
	if client.UserID <= 0 {
		return valueobject.NewErrorResponse(requestID, valueobject.CodeUnauthorized, "用户未登录")
	}

	// 2. 解析请求数据
	var request dto.SendPrivateMessageRequest
	if err := json.Unmarshal(data, &request); err != nil {
		return valueobject.NewErrorResponse(requestID, valueobject.CodeInvalidRequest, "请求数据格式错误")
	}

	// 3. 发送消息
	response, err := h.privateChatService.SendMessage(client.UserID, &request)
	if err != nil {
		// 根据错误类型返回不同的错误代码
		if isRateLimitError(err) {
			return valueobject.NewErrorResponse(requestID, valueobject.CodeValidationError, err.Error())
		}
		return valueobject.NewErrorResponse(requestID, valueobject.CodeInternalError, err.Error())
	}

	// 4. 推送消息给接收者（如果在线）
	if targetClient, exists := h.clientManager.GetClient(request.ToUserID); exists {
		notification := &dto.ChatNotificationResponse{
			Type:      "private",
			Action:    "new_message",
			Data:      response,
			Timestamp: response.Timestamp,
		}

		// 构建通知响应
		notifyResponse := valueobject.NewSuccessResponseWithUniqueID(
			valueobject.MessageTypeChat,
			valueobject.ActionSendPrivateMessage,
			notification,
		)

		// 发送通知给接收者
		targetClient.SendResponse(notifyResponse)
	}

	return valueobject.NewSuccessResponse(requestID, response)
}

// HandleGetPrivateMessages 处理获取私聊消息
func (h *ChatHandler) HandleGetPrivateMessages(client *Client, requestID string, data json.RawMessage) *valueobject.Response {
	// 1. 检查用户登录状态
	if client.UserID <= 0 {
		return valueobject.NewErrorResponse(requestID, valueobject.CodeUnauthorized, "用户未登录")
	}

	// 2. 解析查询参数
	var queryParams map[string]interface{}
	if err := json.Unmarshal(data, &queryParams); err != nil {
		return valueobject.NewErrorResponse(requestID, valueobject.CodeInvalidRequest, "请求数据格式错误")
	}

	// 3. 提取参数
	otherUserID, ok := queryParams["other_user_id"]
	if !ok {
		return valueobject.NewErrorResponse(requestID, valueobject.CodeInvalidRequest, "缺少other_user_id参数")
	}

	// 转换other_user_id
	var otherUserIDInt int
	switch v := otherUserID.(type) {
	case float64:
		otherUserIDInt = int(v)
	case int:
		otherUserIDInt = v
	case string:
		var err error
		otherUserIDInt, err = strconv.Atoi(v)
		if err != nil {
			return valueobject.NewErrorResponse(requestID, valueobject.CodeInvalidRequest, "other_user_id格式错误")
		}
	default:
		return valueobject.NewErrorResponse(requestID, valueobject.CodeInvalidRequest, "other_user_id格式错误")
	}

	// 提取分页参数
	page := 1
	limit := 20
	if pageVal, ok := queryParams["page"]; ok {
		if pageFloat, ok := pageVal.(float64); ok {
			page = int(pageFloat)
		}
	}
	if limitVal, ok := queryParams["limit"]; ok {
		if limitFloat, ok := limitVal.(float64); ok {
			limit = int(limitFloat)
		}
	}

	// 4. 获取消息列表
	response, err := h.privateChatService.GetMessages(client.UserID, otherUserIDInt, page, limit)
	if err != nil {
		return valueobject.NewErrorResponse(requestID, valueobject.CodeInternalError, err.Error())
	}

	return valueobject.NewSuccessResponse(requestID, response)
}

// HandleGetUnreadPrivateMessages 处理获取未读私聊消息
func (h *ChatHandler) HandleGetUnreadPrivateMessages(client *Client, requestID string, data json.RawMessage) *valueobject.Response {
	// 1. 检查用户登录状态
	if client.UserID <= 0 {
		return valueobject.NewErrorResponse(requestID, valueobject.CodeUnauthorized, "用户未登录")
	}

	// 2. 获取未读消息
	response, err := h.privateChatService.GetUnreadMessages(client.UserID)
	if err != nil {
		return valueobject.NewErrorResponse(requestID, valueobject.CodeInternalError, err.Error())
	}

	return valueobject.NewSuccessResponse(requestID, response)
}

// HandleMarkMessageAsRead 处理标记消息已读
func (h *ChatHandler) HandleMarkMessageAsRead(client *Client, requestID string, data json.RawMessage) *valueobject.Response {
	// 1. 检查用户登录状态
	if client.UserID <= 0 {
		return valueobject.NewErrorResponse(requestID, valueobject.CodeUnauthorized, "用户未登录")
	}

	// 2. 解析请求数据
	var request map[string]interface{}
	if err := json.Unmarshal(data, &request); err != nil {
		return valueobject.NewErrorResponse(requestID, valueobject.CodeInvalidRequest, "请求数据格式错误")
	}

	// 3. 提取消息ID
	messageIDVal, ok := request["message_id"]
	if !ok {
		return valueobject.NewErrorResponse(requestID, valueobject.CodeInvalidRequest, "缺少message_id参数")
	}

	var messageID int64
	switch v := messageIDVal.(type) {
	case float64:
		messageID = int64(v)
	case int64:
		messageID = v
	case string:
		var err error
		messageID, err = strconv.ParseInt(v, 10, 64)
		if err != nil {
			return valueobject.NewErrorResponse(requestID, valueobject.CodeInvalidRequest, "message_id格式错误")
		}
	default:
		return valueobject.NewErrorResponse(requestID, valueobject.CodeInvalidRequest, "message_id格式错误")
	}

	// 4. 标记消息已读
	if err := h.privateChatService.MarkMessageAsRead(client.UserID, messageID); err != nil {
		return valueobject.NewErrorResponse(requestID, valueobject.CodeInternalError, err.Error())
	}

	return valueobject.NewSuccessResponse(requestID, map[string]interface{}{
		"message": "消息已标记为已读",
	})
}

// HandleMarkAllAsRead 处理标记所有未读消息为已读
func (h *ChatHandler) HandleMarkAllAsRead(client *Client, requestID string, data json.RawMessage) *valueobject.Response {
	// 1. 检查用户登录状态
	if client.UserID <= 0 {
		return valueobject.NewErrorResponse(requestID, valueobject.CodeUnauthorized, "用户未登录")
	}

	// 2. 标记所有未读消息为已读
	if err := h.privateChatService.MarkAllAsRead(client.UserID); err != nil {
		return valueobject.NewErrorResponse(requestID, valueobject.CodeInternalError, err.Error())
	}

	return valueobject.NewSuccessResponse(requestID, map[string]interface{}{
		"message": "所有消息已标记为已读",
	})
}

// HandleGetUnreadCount 处理获取未读消息数量
func (h *ChatHandler) HandleGetUnreadCount(client *Client, requestID string, data json.RawMessage) *valueobject.Response {
	// 1. 检查用户登录状态
	if client.UserID <= 0 {
		return valueobject.NewErrorResponse(requestID, valueobject.CodeUnauthorized, "用户未登录")
	}

	// 2. 获取未读消息数量
	count, err := h.privateChatService.GetUnreadCount(client.UserID)
	if err != nil {
		return valueobject.NewErrorResponse(requestID, valueobject.CodeInternalError, err.Error())
	}

	return valueobject.NewSuccessResponse(requestID, map[string]interface{}{
		"unread_count": count,
	})
}

// HandleDeleteMessage 处理删除消息
func (h *ChatHandler) HandleDeleteMessage(client *Client, requestID string, data json.RawMessage) *valueobject.Response {
	// 1. 检查用户登录状态
	if client.UserID <= 0 {
		return valueobject.NewErrorResponse(requestID, valueobject.CodeUnauthorized, "用户未登录")
	}

	// 2. 解析请求数据
	var request map[string]interface{}
	if err := json.Unmarshal(data, &request); err != nil {
		return valueobject.NewErrorResponse(requestID, valueobject.CodeInvalidRequest, "请求数据格式错误")
	}

	// 3. 提取消息ID
	messageIDVal, ok := request["message_id"]
	if !ok {
		return valueobject.NewErrorResponse(requestID, valueobject.CodeInvalidRequest, "缺少message_id参数")
	}

	var messageID int64
	switch v := messageIDVal.(type) {
	case float64:
		messageID = int64(v)
	case int64:
		messageID = v
	case string:
		var err error
		messageID, err = strconv.ParseInt(v, 10, 64)
		if err != nil {
			return valueobject.NewErrorResponse(requestID, valueobject.CodeInvalidRequest, "message_id格式错误")
		}
	default:
		return valueobject.NewErrorResponse(requestID, valueobject.CodeInvalidRequest, "message_id格式错误")
	}

	// 4. 删除消息
	if err := h.privateChatService.DeleteMessage(client.UserID, messageID); err != nil {
		return valueobject.NewErrorResponse(requestID, valueobject.CodeInternalError, err.Error())
	}

	return valueobject.NewSuccessResponse(requestID, map[string]interface{}{
		"message": "消息删除成功",
	})
}

// HandleGetConversationPreview 处理获取对话预览
func (h *ChatHandler) HandleGetConversationPreview(client *Client, requestID string, data json.RawMessage) *valueobject.Response {
	// 1. 检查用户登录状态
	if client.UserID <= 0 {
		return valueobject.NewErrorResponse(requestID, valueobject.CodeUnauthorized, "用户未登录")
	}

	// 2. 解析请求数据
	var request map[string]interface{}
	if err := json.Unmarshal(data, &request); err != nil {
		return valueobject.NewErrorResponse(requestID, valueobject.CodeInvalidRequest, "请求数据格式错误")
	}

	// 3. 提取other_user_id
	otherUserIDVal, ok := request["other_user_id"]
	if !ok {
		return valueobject.NewErrorResponse(requestID, valueobject.CodeInvalidRequest, "缺少other_user_id参数")
	}

	var otherUserID int
	switch v := otherUserIDVal.(type) {
	case float64:
		otherUserID = int(v)
	case int:
		otherUserID = v
	case string:
		var err error
		otherUserID, err = strconv.Atoi(v)
		if err != nil {
			return valueobject.NewErrorResponse(requestID, valueobject.CodeInvalidRequest, "other_user_id格式错误")
		}
	default:
		return valueobject.NewErrorResponse(requestID, valueobject.CodeInvalidRequest, "other_user_id格式错误")
	}

	// 4. 获取对话预览
	response, err := h.privateChatService.GetConversationPreview(client.UserID, otherUserID)
	if err != nil {
		if err.Error() == "无对话记录" {
			return valueobject.NewErrorResponse(requestID, valueobject.CodeNotFound, err.Error())
		}
		return valueobject.NewErrorResponse(requestID, valueobject.CodeInternalError, err.Error())
	}

	return valueobject.NewSuccessResponse(requestID, response)
}

// 辅助方法

// getOnlineClient 获取在线客户端
func (h *ChatHandler) getOnlineClient(userID int) *Client {
	if h.clientManager != nil {
		client, _ := h.clientManager.GetClient(userID)
		return client
	}
	return nil
}

// isRateLimitError 判断是否是频率限制错误
func isRateLimitError(err error) bool {
	errMsg := err.Error()
	return fmt.Sprintf("%s", errMsg)[:2] == "私聊" && fmt.Sprintf("%s", errMsg)[2:] == "发送过于频繁" ||
		fmt.Sprintf("%s", errMsg)[:2] == "世界" && fmt.Sprintf("%s", errMsg)[2:] == "聊天发送过于频繁" ||
		fmt.Sprintf("%s", errMsg)[:2] == "工会" && fmt.Sprintf("%s", errMsg)[2:] == "聊天发送过于频繁"
}

// ValidateMessageContent 验证消息内容（通用验证逻辑）
func (h *ChatHandler) ValidateMessageContent(content string) error {
	if content == "" {
		return fmt.Errorf("消息内容不能为空")
	}

	// 检查长度限制（30个字符）
	if len([]rune(content)) > 30 {
		return fmt.Errorf("消息内容不能超过30个字符")
	}

	// 检查是否包含敏感词（这里可以扩展）
	// TODO: 实现敏感词过滤

	return nil
}

// FormatTimestamp 格式化时间戳
func (h *ChatHandler) FormatTimestamp(timestamp int64) string {
	return fmt.Sprintf("%d", timestamp)
}

// ========== 世界聊天处理方法 ==========

// HandleSendWorldMessage 处理发送世界聊天消息
func (h *ChatHandler) HandleSendWorldMessage(client *Client, requestID string, data json.RawMessage) *valueobject.Response {
	// 1. 检查用户登录状态
	if client.UserID <= 0 {
		return valueobject.NewErrorResponse(requestID, valueobject.CodeUnauthorized, "用户未登录")
	}

	// 2. 解析请求数据
	var request dto.SendWorldMessageRequest
	if err := json.Unmarshal(data, &request); err != nil {
		return valueobject.NewErrorResponse(requestID, valueobject.CodeInvalidRequest, "请求数据格式错误")
	}

	// 3. 发送世界聊天消息
	response, err := h.worldChatService.SendMessage(client.UserID, &request)
	if err != nil {
		// 根据错误类型返回不同的错误代码
		if isRateLimitError(err) {
			return valueobject.NewErrorResponse(requestID, valueobject.CodeValidationError, err.Error())
		}
		return valueobject.NewErrorResponse(requestID, valueobject.CodeInternalError, err.Error())
	}

	// 4. 广播消息给频道内的所有用户
	if h.clientManager != nil {
		successCount := h.clientManager.BroadcastToWorldChannel(response.ChannelID, response, client.UserID)
		fmt.Printf("世界频道 %d 消息广播成功发送给 %d 个用户\n", response.ChannelID, successCount)
	} else {
		// 如果没有客户端管理器，使用service的广播方法
		if err := h.worldChatService.BroadcastToChannel(response.ChannelID, response); err != nil {
			fmt.Printf("广播世界聊天消息失败: %v\n", err)
		}
	}

	return valueobject.NewSuccessResponse(requestID, response)
}

// HandleJoinWorldChannel 处理加入世界聊天频道
func (h *ChatHandler) HandleJoinWorldChannel(client *Client, requestID string, data json.RawMessage) *valueobject.Response {
	// 1. 检查用户登录状态
	if client.UserID <= 0 {
		return valueobject.NewErrorResponse(requestID, valueobject.CodeUnauthorized, "用户未登录")
	}

	// 2. 解析请求数据
	var request dto.JoinWorldChannelRequest
	if err := json.Unmarshal(data, &request); err != nil {
		return valueobject.NewErrorResponse(requestID, valueobject.CodeInvalidRequest, "请求数据格式错误")
	}

	// 3. 加入频道
	response, err := h.worldChatService.JoinChannel(client.UserID, &request)
	if err != nil {
		return valueobject.NewErrorResponse(requestID, valueobject.CodeInternalError, err.Error())
	}

	// 4. 通知频道内其他用户有新用户加入
	joinNotification := &dto.ChatNotificationResponse{
		Type:      "world",
		Action:    "user_joined",
		Data: map[string]interface{}{
			"user_id":    client.UserID,
			"channel_id": response.ChannelID,
		},
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
	}

	// 构建通知响应
	notifyResponse := valueobject.NewSuccessResponseWithUniqueID(
		valueobject.MessageTypeChat,
		valueobject.ActionJoinWorldChannel,
		joinNotification,
	)

	// TODO: 广播给频道内其他用户
	_ = notifyResponse

	return valueobject.NewSuccessResponse(requestID, response)
}

// HandleLeaveWorldChannel 处理离开世界聊天频道
func (h *ChatHandler) HandleLeaveWorldChannel(client *Client, requestID string, data json.RawMessage) *valueobject.Response {
	// 1. 检查用户登录状态
	if client.UserID <= 0 {
		return valueobject.NewErrorResponse(requestID, valueobject.CodeUnauthorized, "用户未登录")
	}

	// 2. 获取用户当前频道信息（用于通知）
	currentChannelInfo, _ := h.worldChatService.GetUserChannelInfo(client.UserID)

	// 3. 离开频道
	if err := h.worldChatService.LeaveChannel(client.UserID); err != nil {
		return valueobject.NewErrorResponse(requestID, valueobject.CodeInternalError, err.Error())
	}

	// 4. 通知频道内其他用户有用户离开
	if currentChannelInfo != nil {
		leaveNotification := &dto.ChatNotificationResponse{
			Type:   "world",
			Action: "user_left",
			Data: map[string]interface{}{
				"user_id":    client.UserID,
				"channel_id": currentChannelInfo.ChannelID,
			},
			Timestamp: currentChannelInfo.JoinTime,
		}

		// TODO: 广播给频道内其他用户
		_ = leaveNotification
	}

	return valueobject.NewSuccessResponse(requestID, map[string]interface{}{
		"message": "已离开世界聊天频道",
	})
}

// HandleGetWorldChannels 处理获取世界聊天频道列表
func (h *ChatHandler) HandleGetWorldChannels(client *Client, requestID string, data json.RawMessage) *valueobject.Response {
	// 1. 检查用户登录状态
	if client.UserID <= 0 {
		return valueobject.NewErrorResponse(requestID, valueobject.CodeUnauthorized, "用户未登录")
	}

	// 2. 获取频道列表
	response, err := h.worldChatService.GetChannels(client.UserID)
	if err != nil {
		return valueobject.NewErrorResponse(requestID, valueobject.CodeInternalError, err.Error())
	}

	return valueobject.NewSuccessResponse(requestID, response)
}

// HandleGetUserChannelInfo 处理获取用户频道信息
func (h *ChatHandler) HandleGetUserChannelInfo(client *Client, requestID string, data json.RawMessage) *valueobject.Response {
	// 1. 检查用户登录状态
	if client.UserID <= 0 {
		return valueobject.NewErrorResponse(requestID, valueobject.CodeUnauthorized, "用户未登录")
	}

	// 2. 获取用户频道信息
	response, err := h.worldChatService.GetUserChannelInfo(client.UserID)
	if err != nil {
		return valueobject.NewErrorResponse(requestID, valueobject.CodeNotFound, err.Error())
	}

	return valueobject.NewSuccessResponse(requestID, response)
}

// ========== 工会聊天处理方法 ==========

// HandleSendUnionMessage 处理发送工会聊天消息
func (h *ChatHandler) HandleSendUnionMessage(client *Client, requestID string, data json.RawMessage) *valueobject.Response {
	// 1. 检查用户登录状态
	if client.UserID <= 0 {
		return valueobject.NewErrorResponse(requestID, valueobject.CodeUnauthorized, "用户未登录")
	}

	// 2. 解析请求数据
	var request dto.SendUnionMessageRequest
	if err := json.Unmarshal(data, &request); err != nil {
		return valueobject.NewErrorResponse(requestID, valueobject.CodeInvalidRequest, "请求数据格式错误")
	}

	// 3. 发送工会聊天消息
	response, err := h.unionChatService.SendMessage(client.UserID, &request)
	if err != nil {
		// 根据错误类型返回不同的错误代码
		if isRateLimitError(err) {
			return valueobject.NewErrorResponse(requestID, valueobject.CodeValidationError, err.Error())
		}
		return valueobject.NewErrorResponse(requestID, valueobject.CodeInternalError, err.Error())
	}

	// 4. 广播消息给工会内的所有用户
	if h.clientManager != nil {
		successCount := h.clientManager.BroadcastToUnion(response.UnionID, response, client.UserID)
		fmt.Printf("工会 %d 消息广播成功发送给 %d 个用户\n", response.UnionID, successCount)
	} else {
		// 如果没有客户端管理器，使用service的广播方法
		if err := h.unionChatService.BroadcastToUnion(response.UnionID, response); err != nil {
			fmt.Printf("广播工会聊天消息失败: %v\n", err)
		}
	}

	return valueobject.NewSuccessResponse(requestID, response)
}

// HandleGetUnionMessages 处理获取工会聊天消息
func (h *ChatHandler) HandleGetUnionMessages(client *Client, requestID string, data json.RawMessage) *valueobject.Response {
	// 1. 检查用户登录状态
	if client.UserID <= 0 {
		return valueobject.NewErrorResponse(requestID, valueobject.CodeUnauthorized, "用户未登录")
	}

	// 2. 解析请求数据
	var request dto.GetUnionMessagesRequest
	if err := json.Unmarshal(data, &request); err != nil {
		return valueobject.NewErrorResponse(requestID, valueobject.CodeInvalidRequest, "请求数据格式错误")
	}

	// 3. 获取工会聊天消息
	response, err := h.unionChatService.GetMessages(client.UserID, &request)
	if err != nil {
		return valueobject.NewErrorResponse(requestID, valueobject.CodeInternalError, err.Error())
	}

	return valueobject.NewSuccessResponse(requestID, response)
}

// HandleGetRecentUnionMessages 处理获取最近工会聊天消息
func (h *ChatHandler) HandleGetRecentUnionMessages(client *Client, requestID string, data json.RawMessage) *valueobject.Response {
	// 1. 检查用户登录状态
	if client.UserID <= 0 {
		return valueobject.NewErrorResponse(requestID, valueobject.CodeUnauthorized, "用户未登录")
	}

	// 2. 解析请求参数
	var queryParams map[string]interface{}
	if err := json.Unmarshal(data, &queryParams); err != nil {
		return valueobject.NewErrorResponse(requestID, valueobject.CodeInvalidRequest, "请求数据格式错误")
	}

	// 3. 提取limit参数
	limit := 50 // 默认50条
	if limitVal, ok := queryParams["limit"]; ok {
		if limitFloat, ok := limitVal.(float64); ok {
			limit = int(limitFloat)
		}
	}

	// 4. 获取最近工会消息
	response, err := h.unionChatService.GetRecentMessages(client.UserID, limit)
	if err != nil {
		return valueobject.NewErrorResponse(requestID, valueobject.CodeInternalError, err.Error())
	}

	return valueobject.NewSuccessResponse(requestID, response)
}