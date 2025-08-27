package entity

import (
	"fmt"
	"time"
)

// WorldChatChannel 世界聊天频道实体
type WorldChatChannel struct {
	ID           int       `json:"id" db:"id"`
	ChannelID    int       `json:"channel_id" db:"channel_id"`
	ChannelName  string    `json:"channel_name" db:"channel_name"`
	MaxUsers     int       `json:"max_users" db:"max_users"`
	CurrentUsers int       `json:"current_users" db:"current_users"`
	IsActive     bool      `json:"is_active" db:"is_active"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// WorldChatMessage 世界聊天消息（仅用于传输，不存储）
type WorldChatMessage struct {
	ChannelID  int       `json:"channel_id"`
	UserID     int       `json:"user_id"`
	Username   string    `json:"username"`
	Content    string    `json:"content"`
	Timestamp  time.Time `json:"timestamp"`
	MessageID  string    `json:"message_id"` // 用于去重的临时ID
}

// Validate 验证世界聊天频道数据
func (wcc *WorldChatChannel) Validate() error {
	if wcc.ChannelID <= 0 {
		return fmt.Errorf("频道ID无效")
	}

	if wcc.ChannelName == "" {
		return fmt.Errorf("频道名称不能为空")
	}

	if wcc.MaxUsers <= 0 || wcc.MaxUsers > 500 {
		return fmt.Errorf("频道最大用户数必须在1-500之间")
	}

	if wcc.CurrentUsers < 0 {
		return fmt.Errorf("当前用户数不能为负数")
	}

	if wcc.CurrentUsers > wcc.MaxUsers {
		return fmt.Errorf("当前用户数不能超过最大用户数")
	}

	return nil
}

// IsFull 判断频道是否已满
func (wcc *WorldChatChannel) IsFull() bool {
	return wcc.CurrentUsers >= wcc.MaxUsers
}

// CanJoin 判断是否可以加入频道
func (wcc *WorldChatChannel) CanJoin() bool {
	return wcc.IsActive && !wcc.IsFull()
}

// IncrementUsers 增加用户数量
func (wcc *WorldChatChannel) IncrementUsers() error {
	if wcc.IsFull() {
		return fmt.Errorf("频道已满")
	}
	wcc.CurrentUsers++
	wcc.UpdatedAt = time.Now()
	return nil
}

// DecrementUsers 减少用户数量
func (wcc *WorldChatChannel) DecrementUsers() {
	if wcc.CurrentUsers > 0 {
		wcc.CurrentUsers--
		wcc.UpdatedAt = time.Now()
	}
}

// Validate 验证世界聊天消息
func (wcm *WorldChatMessage) Validate() error {
	if wcm.ChannelID <= 0 {
		return fmt.Errorf("频道ID无效")
	}

	if wcm.UserID <= 0 {
		return fmt.Errorf("用户ID无效")
	}

	if wcm.Username == "" {
		return fmt.Errorf("用户名不能为空")
	}

	// 复用私聊消息的内容验证逻辑
	pm := &PrivateMessage{Content: wcm.Content}
	if err := pm.ValidateContent(); err != nil {
		return err
	}

	return nil
}