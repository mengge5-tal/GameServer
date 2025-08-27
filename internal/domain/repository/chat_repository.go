package repository

import "GameServer/internal/domain/entity"

// PrivateMessageRepository 私聊消息数据访问接口
type PrivateMessageRepository interface {
	// 基础CRUD操作
	Create(message *entity.PrivateMessage) error
	GetByID(id int64) (*entity.PrivateMessage, error)
	Update(message *entity.PrivateMessage) error
	Delete(id int64) error

	// 获取用户的未读私聊消息
	GetUnreadMessages(userID int) ([]*entity.PrivateMessage, error)

	// 获取两个用户之间的对话历史（分页）
	GetConversationHistory(user1ID, user2ID int, page, limit int) ([]*entity.PrivateMessage, int, error)

	// 标记消息为已读
	MarkAsRead(messageID int64) error

	// 标记用户的所有未读消息为已读
	MarkAllAsReadForUser(userID int) error

	// 标记特定对话的所有未读消息为已读
	MarkConversationAsRead(userID, otherUserID int) error

	// 获取用户未读消息数量
	GetUnreadCount(userID int) (int, error)

	// 删除已读消息（定时清理用）
	DeleteReadMessages(beforeTime string) (int64, error)

	// 删除超过指定天数的未读消息（防止垃圾数据）
	DeleteOldUnreadMessages(beforeTime string) (int64, error)
}

// WorldChatRepository 世界聊天数据访问接口
type WorldChatRepository interface {
	// 频道管理
	GetChannelByID(channelID int) (*entity.WorldChatChannel, error)
	GetAllChannels() ([]*entity.WorldChatChannel, error)
	GetActiveChannels() ([]*entity.WorldChatChannel, error)
	UpdateChannel(channel *entity.WorldChatChannel) error
	CreateChannel(channel *entity.WorldChatChannel) error

	// 获取最少用户数的频道（用于自动分配）
	GetLeastUsersChannel() (*entity.WorldChatChannel, error)

	// 更新频道用户数
	IncrementChannelUsers(channelID int) error
	DecrementChannelUsers(channelID int) error

	// 批量更新频道用户数（用于服务重启时重新统计）
	UpdateChannelUserCount(channelID int, count int) error
}

// UnionChatRepository 工会聊天数据访问接口  
type UnionChatRepository interface {
	// 消息操作
	CreateMessage(tableName string, message *entity.UnionChatMessage) error
	GetMessagesByUnionID(tableName string, unionID int, page, limit int) ([]*entity.UnionChatMessage, int, error)
	GetRecentMessages(tableName string, unionID int, limit int) ([]*entity.UnionChatMessage, error)

	// 批量插入消息（从缓存刷新到数据库时使用）
	BatchCreateMessages(tableName string, messages []*entity.UnionChatMessage) error

	// 分表管理
	GetActiveTable() (*entity.UnionChatTable, error)
	GetTableByYearMonth(yearMonth string) (*entity.UnionChatTable, error)
	CreateTable(table *entity.UnionChatTable) error
	UpdateTable(table *entity.UnionChatTable) error

	// 检查表是否存在
	TableExists(tableName string) (bool, error)

	// 创建新的月度聊天表
	CreateMonthlyTable(tableName string) error

	// 获取表列表（用于数据清理和归档）
	GetAllTables() ([]*entity.UnionChatTable, error)

	// 设置表为非活跃状态（用于归档）
	DeactivateTable(yearMonth string) error
}