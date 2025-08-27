package entity

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

// MessageStatus 消息状态常量
const (
	MessageStatusUnread = 0 // 未读
	MessageStatusRead   = 1 // 已读
)

// PrivateMessage 私聊消息实体
type PrivateMessage struct {
	ID         int64     `json:"id" db:"id"`
	FromUserID int       `json:"from_user_id" db:"from_user_id"`
	ToUserID   int       `json:"to_user_id" db:"to_user_id"`
	Content    string    `json:"content" db:"content"`
	Status     int       `json:"status" db:"status"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

// Validate 验证私聊消息数据
func (pm *PrivateMessage) Validate() error {
	// 验证发送者ID
	if pm.FromUserID <= 0 {
		return fmt.Errorf("发送者ID无效")
	}

	// 验证接收者ID
	if pm.ToUserID <= 0 {
		return fmt.Errorf("接收者ID无效")
	}

	// 不能给自己发消息
	if pm.FromUserID == pm.ToUserID {
		return fmt.Errorf("不能给自己发送消息")
	}

	// 验证消息内容
	if err := pm.ValidateContent(); err != nil {
		return err
	}

	// 验证消息状态
	if pm.Status < MessageStatusUnread || pm.Status > MessageStatusRead {
		return fmt.Errorf("消息状态无效")
	}

	return nil
}

// ValidateContent 验证消息内容
func (pm *PrivateMessage) ValidateContent() error {
	// 去除前后空白字符
	pm.Content = strings.TrimSpace(pm.Content)

	// 检查内容是否为空
	if pm.Content == "" {
		return fmt.Errorf("消息内容不能为空")
	}

	// 检查字符长度（中文字符按1个字符计算）
	runeCount := utf8.RuneCountInString(pm.Content)
	if runeCount > 30 {
		return fmt.Errorf("消息内容不能超过30个字符")
	}

	// 检查字节长度（防止超出数据库字段限制）
	if len(pm.Content) > 90 {
		return fmt.Errorf("消息内容过长")
	}

	return nil
}

// IsUnread 判断消息是否未读
func (pm *PrivateMessage) IsUnread() bool {
	return pm.Status == MessageStatusUnread
}

// MarkAsRead 标记消息为已读
func (pm *PrivateMessage) MarkAsRead() {
	pm.Status = MessageStatusRead
	pm.UpdatedAt = time.Now()
}

// GetConversationKey 获取对话唯一键（用于缓存或分组）
func (pm *PrivateMessage) GetConversationKey() string {
	// 确保对话键的一致性，使用较小的用户ID在前
	if pm.FromUserID < pm.ToUserID {
		return fmt.Sprintf("conversation:%d_%d", pm.FromUserID, pm.ToUserID)
	}
	return fmt.Sprintf("conversation:%d_%d", pm.ToUserID, pm.FromUserID)
}