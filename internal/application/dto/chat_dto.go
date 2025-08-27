package dto

import "time"

// ========== 私聊相关DTO ==========

// SendPrivateMessageRequest 发送私聊消息请求
type SendPrivateMessageRequest struct {
	ToUserID int    `json:"to_user_id"`
	Content  string `json:"content"`
}

// PrivateMessageResponse 私聊消息响应
type PrivateMessageResponse struct {
	ID           int64  `json:"id"`
	FromUserID   int    `json:"from_user_id"`
	FromUsername string `json:"from_username"`
	ToUserID     int    `json:"to_user_id"`
	ToUsername   string `json:"to_username"`
	Content      string `json:"content"`
	Status       int    `json:"status"`       // 0:未读 1:已读
	Timestamp    string `json:"timestamp"`    // 格式化时间字符串
	IsFromSelf   bool   `json:"is_from_self"` // 是否是自己发送的消息
}

// GetPrivateMessagesResponse 获取私聊消息列表响应
type GetPrivateMessagesResponse struct {
	Messages []PrivateMessageResponse `json:"messages"`
	Total    int                      `json:"total"`
	HasMore  bool                     `json:"has_more"`
}

// ========== 世界聊天相关DTO ==========

// SendWorldMessageRequest 发送世界消息请求
type SendWorldMessageRequest struct {
	Content string `json:"content"`
}

// WorldMessageResponse 世界聊天消息响应
type WorldMessageResponse struct {
	ChannelID int    `json:"channel_id"`
	UserID    int    `json:"user_id"`
	Username  string `json:"username"`
	Content   string `json:"content"`
	Timestamp string `json:"timestamp"`
	MessageID string `json:"message_id"`
}

// JoinWorldChannelRequest 加入世界频道请求
type JoinWorldChannelRequest struct {
	ChannelID int `json:"channel_id,omitempty"` // 可选，不指定则自动分配
}

// WorldChannelResponse 世界频道响应
type WorldChannelResponse struct {
	ChannelID    int    `json:"channel_id"`
	ChannelName  string `json:"channel_name"`
	MaxUsers     int    `json:"max_users"`
	CurrentUsers int    `json:"current_users"`
	IsActive     bool   `json:"is_active"`
}

// GetWorldChannelsResponse 获取世界频道列表响应
type GetWorldChannelsResponse struct {
	Channels      []WorldChannelResponse `json:"channels"`
	CurrentUser   *UserChannelInfo       `json:"current_user,omitempty"`
	TotalChannels int                    `json:"total_channels"`
}

// UserChannelInfo 用户频道信息
type UserChannelInfo struct {
	UserID    int    `json:"user_id"`
	ChannelID int    `json:"channel_id"`
	JoinTime  string `json:"join_time"`
}

// ========== 工会聊天相关DTO ==========

// SendUnionMessageRequest 发送工会消息请求
type SendUnionMessageRequest struct {
	Content string `json:"content"`
}

// UnionMessageResponse 工会聊天消息响应
type UnionMessageResponse struct {
	ID        int64  `json:"id"`
	UnionID   int    `json:"union_id"`
	UserID    int    `json:"user_id"`
	Username  string `json:"username"`
	Content   string `json:"content"`
	Timestamp string `json:"timestamp"`
	MessageID string `json:"message_id"`
}

// GetUnionMessagesRequest 获取工会消息请求
type GetUnionMessagesRequest struct {
	Page      int    `json:"page,omitempty"`       // 页码，默认1
	Limit     int    `json:"limit,omitempty"`      // 每页数量，默认20
	YearMonth string `json:"year_month,omitempty"` // 指定年月，格式：2006-01
}

// GetUnionMessagesResponse 获取工会消息响应
type GetUnionMessagesResponse struct {
	Messages   []UnionMessageResponse `json:"messages"`
	Total      int                    `json:"total"`
	Page       int                    `json:"page"`
	Limit      int                    `json:"limit"`
	TotalPages int                    `json:"total_pages"`
	HasMore    bool                   `json:"has_more"`
	UnionID    int                    `json:"union_id"`
	UnionName  string                 `json:"union_name"`
}

// ========== 通用聊天响应DTO ==========

// ChatNotificationResponse 聊天通知响应（用于实时推送）
type ChatNotificationResponse struct {
	Type      string      `json:"type"`       // "private", "world", "union"
	Action    string      `json:"action"`     // "new_message", "user_join", "user_leave"
	Data      interface{} `json:"data"`       // 具体消息内容
	Timestamp string      `json:"timestamp"`  // 通知时间
	UnreadCount int       `json:"unread_count,omitempty"` // 未读消息数（仅私聊）
}

// ========== 聊天统计DTO ==========

// ChatStatsResponse 聊天统计响应
type ChatStatsResponse struct {
	PrivateUnreadCount int                      `json:"private_unread_count"` // 私聊未读数
	WorldChannelInfo   *UserChannelInfo         `json:"world_channel_info"`   // 当前世界频道信息
	UnionChatAvailable bool                     `json:"union_chat_available"` // 是否可使用工会聊天
	LastActivity       map[string]time.Time     `json:"last_activity"`        // 各类聊天最后活跃时间
}

// ========== 错误响应DTO ==========

// ChatErrorResponse 聊天错误响应
type ChatErrorResponse struct {
	ErrorCode    string `json:"error_code"`    // 错误代码
	ErrorMessage string `json:"error_message"` // 错误消息
	RetryAfter   int    `json:"retry_after,omitempty"` // 重试等待时间（秒）
}

// 常用错误代码常量
const (
	ChatErrorRateLimit        = "RATE_LIMIT_EXCEEDED"      // 发送频率超限
	ChatErrorContentTooLong   = "CONTENT_TOO_LONG"         // 内容过长
	ChatErrorUserNotFound     = "USER_NOT_FOUND"           // 用户不存在
	ChatErrorChannelFull      = "CHANNEL_FULL"             // 频道已满
	ChatErrorNotInUnion       = "NOT_IN_UNION"             // 不在工会中
	ChatErrorUserOffline      = "USER_OFFLINE"             // 用户离线
	ChatErrorInvalidContent   = "INVALID_CONTENT"          // 内容无效
	ChatErrorPermissionDenied = "PERMISSION_DENIED"        // 权限不足
)