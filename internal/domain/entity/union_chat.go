package entity

import (
	"fmt"
	"time"
)

// UnionChatMessage 工会聊天消息实体
type UnionChatMessage struct {
	ID        int64     `json:"id" db:"id"`
	UnionID   int       `json:"union_id" db:"union_id"`
	UserID    int       `json:"user_id" db:"user_id"`
	Username  string    `json:"username" db:"username"`
	Content   string    `json:"content" db:"content"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// UnionChatTable 工会聊天分表管理实体
type UnionChatTable struct {
	ID        int       `json:"id" db:"id"`
	TableName string    `json:"table_name" db:"table_name"`
	YearMonth string    `json:"year_month" db:"year_month"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	IsActive  bool      `json:"is_active" db:"is_active"`
}

// Validate 验证工会聊天消息
func (ucm *UnionChatMessage) Validate() error {
	if ucm.UnionID <= 0 {
		return fmt.Errorf("工会ID无效")
	}

	if ucm.UserID <= 0 {
		return fmt.Errorf("用户ID无效")
	}

	if ucm.Username == "" {
		return fmt.Errorf("用户名不能为空")
	}

	// 复用私聊消息的内容验证逻辑
	pm := &PrivateMessage{Content: ucm.Content}
	if err := pm.ValidateContent(); err != nil {
		return err
	}

	return nil
}

// GetCacheKey 获取Redis缓存键
func (ucm *UnionChatMessage) GetCacheKey() string {
	return fmt.Sprintf("union:messages:%d", ucm.UnionID)
}

// GetCountCacheKey 获取消息计数缓存键
func (ucm *UnionChatMessage) GetCountCacheKey() string {
	return fmt.Sprintf("union:count:%d", ucm.UnionID)
}

// Validate 验证工会聊天分表管理数据
func (uct *UnionChatTable) Validate() error {
	if uct.TableName == "" {
		return fmt.Errorf("表名不能为空")
	}

	if uct.YearMonth == "" {
		return fmt.Errorf("年月不能为空")
	}

	// 验证年月格式（YYYY-MM）
	if len(uct.YearMonth) != 7 || uct.YearMonth[4] != '-' {
		return fmt.Errorf("年月格式错误，应为YYYY-MM")
	}

	return nil
}

// GetTableName 根据年月获取表名
func (uct *UnionChatTable) GetTableName(yearMonth string) string {
	// 将 "2025-08" 转换为 "union_messages_202508"
	tableMonth := yearMonth[:4] + yearMonth[5:]
	return fmt.Sprintf("union_messages_%s", tableMonth)
}

// IsCurrentMonth 判断是否为当前月份的表
func (uct *UnionChatTable) IsCurrentMonth() bool {
	currentYearMonth := time.Now().Format("2006-01")
	return uct.YearMonth == currentYearMonth && uct.IsActive
}