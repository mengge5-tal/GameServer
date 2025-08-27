package repository

import (
	"GameServer/internal/domain/entity"
	"GameServer/internal/domain/repository"
	"database/sql"
	"fmt"
	"strings"
)

// MySQLChatRepository MySQL聊天数据访问实现
type MySQLChatRepository struct {
	db *sql.DB
}

// NewMySQLChatRepository 创建MySQL聊天Repository
func NewMySQLChatRepository(db *sql.DB) *MySQLChatRepository {
	return &MySQLChatRepository{
		db: db,
	}
}

// ========== PrivateMessageRepository 实现 ==========

// Create 创建私聊消息
func (r *MySQLChatRepository) Create(message *entity.PrivateMessage) error {
	if err := message.Validate(); err != nil {
		return err
	}

	query := `
		INSERT INTO private_messages (from_user_id, to_user_id, content, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, NOW(), NOW())
	`
	result, err := r.db.Exec(query, message.FromUserID, message.ToUserID, message.Content, message.Status)
	if err != nil {
		return fmt.Errorf("创建私聊消息失败: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("获取插入ID失败: %v", err)
	}
	message.ID = id

	return nil
}

// GetByID 根据ID获取私聊消息
func (r *MySQLChatRepository) GetByID(id int64) (*entity.PrivateMessage, error) {
	message := &entity.PrivateMessage{}
	query := `
		SELECT id, from_user_id, to_user_id, content, status, created_at, updated_at
		FROM private_messages WHERE id = ?
	`
	err := r.db.QueryRow(query, id).Scan(
		&message.ID, &message.FromUserID, &message.ToUserID,
		&message.Content, &message.Status, &message.CreatedAt, &message.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("消息不存在")
		}
		return nil, fmt.Errorf("查询消息失败: %v", err)
	}

	return message, nil
}

// Update 更新私聊消息
func (r *MySQLChatRepository) Update(message *entity.PrivateMessage) error {
	if err := message.Validate(); err != nil {
		return err
	}

	query := `
		UPDATE private_messages 
		SET content = ?, status = ?, updated_at = NOW()
		WHERE id = ?
	`
	result, err := r.db.Exec(query, message.Content, message.Status, message.ID)
	if err != nil {
		return fmt.Errorf("更新消息失败: %v", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取影响行数失败: %v", err)
	}
	if affected == 0 {
		return fmt.Errorf("消息不存在或无变更")
	}

	return nil
}

// Delete 删除私聊消息
func (r *MySQLChatRepository) Delete(id int64) error {
	query := `DELETE FROM private_messages WHERE id = ?`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("删除消息失败: %v", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取影响行数失败: %v", err)
	}
	if affected == 0 {
		return fmt.Errorf("消息不存在")
	}

	return nil
}

// GetUnreadMessages 获取用户的未读私聊消息
func (r *MySQLChatRepository) GetUnreadMessages(userID int) ([]*entity.PrivateMessage, error) {
	query := `
		SELECT id, from_user_id, to_user_id, content, status, created_at, updated_at
		FROM private_messages 
		WHERE to_user_id = ? AND status = 0
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("查询未读消息失败: %v", err)
	}
	defer rows.Close()

	var messages []*entity.PrivateMessage
	for rows.Next() {
		message := &entity.PrivateMessage{}
		err := rows.Scan(
			&message.ID, &message.FromUserID, &message.ToUserID,
			&message.Content, &message.Status, &message.CreatedAt, &message.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描未读消息失败: %v", err)
		}
		messages = append(messages, message)
	}

	return messages, nil
}

// GetConversationHistory 获取两个用户之间的对话历史
func (r *MySQLChatRepository) GetConversationHistory(user1ID, user2ID int, page, limit int) ([]*entity.PrivateMessage, int, error) {
	// 计算offset
	offset := (page - 1) * limit

	// 查询总数
	countQuery := `
		SELECT COUNT(*) 
		FROM private_messages 
		WHERE (from_user_id = ? AND to_user_id = ?) 
		   OR (from_user_id = ? AND to_user_id = ?)
	`
	var total int
	err := r.db.QueryRow(countQuery, user1ID, user2ID, user2ID, user1ID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("查询对话总数失败: %v", err)
	}

	// 查询消息列表
	query := `
		SELECT id, from_user_id, to_user_id, content, status, created_at, updated_at
		FROM private_messages 
		WHERE (from_user_id = ? AND to_user_id = ?) 
		   OR (from_user_id = ? AND to_user_id = ?)
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	rows, err := r.db.Query(query, user1ID, user2ID, user2ID, user1ID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("查询对话历史失败: %v", err)
	}
	defer rows.Close()

	var messages []*entity.PrivateMessage
	for rows.Next() {
		message := &entity.PrivateMessage{}
		err := rows.Scan(
			&message.ID, &message.FromUserID, &message.ToUserID,
			&message.Content, &message.Status, &message.CreatedAt, &message.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("扫描对话历史失败: %v", err)
		}
		messages = append(messages, message)
	}

	return messages, total, nil
}

// MarkAsRead 标记消息为已读
func (r *MySQLChatRepository) MarkAsRead(messageID int64) error {
	query := `UPDATE private_messages SET status = 1, updated_at = NOW() WHERE id = ?`
	result, err := r.db.Exec(query, messageID)
	if err != nil {
		return fmt.Errorf("标记消息已读失败: %v", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取影响行数失败: %v", err)
	}
	if affected == 0 {
		return fmt.Errorf("消息不存在或已是已读状态")
	}

	return nil
}

// MarkAllAsReadForUser 标记用户的所有未读消息为已读
func (r *MySQLChatRepository) MarkAllAsReadForUser(userID int) error {
	query := `
		UPDATE private_messages 
		SET status = 1, updated_at = NOW() 
		WHERE to_user_id = ? AND status = 0
	`
	_, err := r.db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("标记用户所有消息已读失败: %v", err)
	}

	return nil
}

// MarkConversationAsRead 标记特定对话的所有未读消息为已读
func (r *MySQLChatRepository) MarkConversationAsRead(userID, otherUserID int) error {
	query := `
		UPDATE private_messages 
		SET status = 1, updated_at = NOW() 
		WHERE to_user_id = ? AND from_user_id = ? AND status = 0
	`
	_, err := r.db.Exec(query, userID, otherUserID)
	if err != nil {
		return fmt.Errorf("标记对话消息已读失败: %v", err)
	}

	return nil
}

// GetUnreadCount 获取用户未读消息数量
func (r *MySQLChatRepository) GetUnreadCount(userID int) (int, error) {
	query := `
		SELECT COUNT(*) 
		FROM private_messages 
		WHERE to_user_id = ? AND status = 0
	`
	var count int
	err := r.db.QueryRow(query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("查询未读消息数量失败: %v", err)
	}

	return count, nil
}

// DeleteReadMessages 删除已读消息（定时清理用）
func (r *MySQLChatRepository) DeleteReadMessages(beforeTime string) (int64, error) {
	query := `DELETE FROM private_messages WHERE status = 1 AND created_at < ?`
	result, err := r.db.Exec(query, beforeTime)
	if err != nil {
		return 0, fmt.Errorf("删除已读消息失败: %v", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("获取影响行数失败: %v", err)
	}

	return affected, nil
}

// DeleteOldUnreadMessages 删除超过指定时间的未读消息
func (r *MySQLChatRepository) DeleteOldUnreadMessages(beforeTime string) (int64, error) {
	query := `DELETE FROM private_messages WHERE status = 0 AND created_at < ?`
	result, err := r.db.Exec(query, beforeTime)
	if err != nil {
		return 0, fmt.Errorf("删除旧未读消息失败: %v", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("获取影响行数失败: %v", err)
	}

	return affected, nil
}

// ========== WorldChatRepository 实现 ==========

// GetChannelByID 根据ID获取频道
func (r *MySQLChatRepository) GetChannelByID(channelID int) (*entity.WorldChatChannel, error) {
	channel := &entity.WorldChatChannel{}
	query := `
		SELECT id, channel_id, channel_name, max_users, current_users, is_active, created_at, updated_at
		FROM world_chat_channels WHERE channel_id = ?
	`
	err := r.db.QueryRow(query, channelID).Scan(
		&channel.ID, &channel.ChannelID, &channel.ChannelName,
		&channel.MaxUsers, &channel.CurrentUsers, &channel.IsActive,
		&channel.CreatedAt, &channel.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("频道不存在")
		}
		return nil, fmt.Errorf("查询频道失败: %v", err)
	}

	return channel, nil
}

// GetAllChannels 获取所有频道
func (r *MySQLChatRepository) GetAllChannels() ([]*entity.WorldChatChannel, error) {
	query := `
		SELECT id, channel_id, channel_name, max_users, current_users, is_active, created_at, updated_at
		FROM world_chat_channels ORDER BY channel_id
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("查询所有频道失败: %v", err)
	}
	defer rows.Close()

	var channels []*entity.WorldChatChannel
	for rows.Next() {
		channel := &entity.WorldChatChannel{}
		err := rows.Scan(
			&channel.ID, &channel.ChannelID, &channel.ChannelName,
			&channel.MaxUsers, &channel.CurrentUsers, &channel.IsActive,
			&channel.CreatedAt, &channel.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描频道数据失败: %v", err)
		}
		channels = append(channels, channel)
	}

	return channels, nil
}

// GetActiveChannels 获取活跃频道
func (r *MySQLChatRepository) GetActiveChannels() ([]*entity.WorldChatChannel, error) {
	query := `
		SELECT id, channel_id, channel_name, max_users, current_users, is_active, created_at, updated_at
		FROM world_chat_channels WHERE is_active = true ORDER BY channel_id
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("查询活跃频道失败: %v", err)
	}
	defer rows.Close()

	var channels []*entity.WorldChatChannel
	for rows.Next() {
		channel := &entity.WorldChatChannel{}
		err := rows.Scan(
			&channel.ID, &channel.ChannelID, &channel.ChannelName,
			&channel.MaxUsers, &channel.CurrentUsers, &channel.IsActive,
			&channel.CreatedAt, &channel.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描活跃频道数据失败: %v", err)
		}
		channels = append(channels, channel)
	}

	return channels, nil
}

// UpdateChannel 更新频道
func (r *MySQLChatRepository) UpdateChannel(channel *entity.WorldChatChannel) error {
	if err := channel.Validate(); err != nil {
		return err
	}

	query := `
		UPDATE world_chat_channels 
		SET channel_name = ?, max_users = ?, current_users = ?, is_active = ?, updated_at = NOW()
		WHERE channel_id = ?
	`
	result, err := r.db.Exec(query, channel.ChannelName, channel.MaxUsers, 
		channel.CurrentUsers, channel.IsActive, channel.ChannelID)
	if err != nil {
		return fmt.Errorf("更新频道失败: %v", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取影响行数失败: %v", err)
	}
	if affected == 0 {
		return fmt.Errorf("频道不存在或无变更")
	}

	return nil
}

// CreateChannel 创建频道
func (r *MySQLChatRepository) CreateChannel(channel *entity.WorldChatChannel) error {
	if err := channel.Validate(); err != nil {
		return err
	}

	query := `
		INSERT INTO world_chat_channels (channel_id, channel_name, max_users, current_users, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, NOW(), NOW())
	`
	result, err := r.db.Exec(query, channel.ChannelID, channel.ChannelName, 
		channel.MaxUsers, channel.CurrentUsers, channel.IsActive)
	if err != nil {
		return fmt.Errorf("创建频道失败: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("获取插入ID失败: %v", err)
	}
	channel.ID = int(id)

	return nil
}

// GetLeastUsersChannel 获取用户数最少的活跃频道
func (r *MySQLChatRepository) GetLeastUsersChannel() (*entity.WorldChatChannel, error) {
	channel := &entity.WorldChatChannel{}
	query := `
		SELECT id, channel_id, channel_name, max_users, current_users, is_active, created_at, updated_at
		FROM world_chat_channels 
		WHERE is_active = true AND current_users < max_users
		ORDER BY current_users ASC, channel_id ASC
		LIMIT 1
	`
	err := r.db.QueryRow(query).Scan(
		&channel.ID, &channel.ChannelID, &channel.ChannelName,
		&channel.MaxUsers, &channel.CurrentUsers, &channel.IsActive,
		&channel.CreatedAt, &channel.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("没有可用的频道")
		}
		return nil, fmt.Errorf("查询最少用户频道失败: %v", err)
	}

	return channel, nil
}

// IncrementChannelUsers 增加频道用户数
func (r *MySQLChatRepository) IncrementChannelUsers(channelID int) error {
	query := `
		UPDATE world_chat_channels 
		SET current_users = current_users + 1, updated_at = NOW()
		WHERE channel_id = ? AND current_users < max_users
	`
	result, err := r.db.Exec(query, channelID)
	if err != nil {
		return fmt.Errorf("增加频道用户数失败: %v", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取影响行数失败: %v", err)
	}
	if affected == 0 {
		return fmt.Errorf("频道不存在或已满")
	}

	return nil
}

// DecrementChannelUsers 减少频道用户数
func (r *MySQLChatRepository) DecrementChannelUsers(channelID int) error {
	query := `
		UPDATE world_chat_channels 
		SET current_users = GREATEST(current_users - 1, 0), updated_at = NOW()
		WHERE channel_id = ?
	`
	result, err := r.db.Exec(query, channelID)
	if err != nil {
		return fmt.Errorf("减少频道用户数失败: %v", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取影响行数失败: %v", err)
	}
	if affected == 0 {
		return fmt.Errorf("频道不存在")
	}

	return nil
}

// UpdateChannelUserCount 更新频道用户数
func (r *MySQLChatRepository) UpdateChannelUserCount(channelID int, count int) error {
	query := `
		UPDATE world_chat_channels 
		SET current_users = ?, updated_at = NOW()
		WHERE channel_id = ?
	`
	result, err := r.db.Exec(query, count, channelID)
	if err != nil {
		return fmt.Errorf("更新频道用户数失败: %v", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取影响行数失败: %v", err)
	}
	if affected == 0 {
		return fmt.Errorf("频道不存在")
	}

	return nil
}

// ========== UnionChatRepository 实现 ==========

// CreateMessage 创建工会聊天消息
func (r *MySQLChatRepository) CreateMessage(tableName string, message *entity.UnionChatMessage) error {
	if err := message.Validate(); err != nil {
		return err
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (union_id, user_id, username, content, created_at)
		VALUES (?, ?, ?, ?, NOW())
	`, tableName)
	
	result, err := r.db.Exec(query, message.UnionID, message.UserID, message.Username, message.Content)
	if err != nil {
		return fmt.Errorf("创建工会消息失败: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("获取插入ID失败: %v", err)
	}
	message.ID = id

	return nil
}

// GetMessagesByUnionID 根据工会ID获取消息（分页）
func (r *MySQLChatRepository) GetMessagesByUnionID(tableName string, unionID int, page, limit int) ([]*entity.UnionChatMessage, int, error) {
	// 计算offset
	offset := (page - 1) * limit

	// 查询总数
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE union_id = ?`, tableName)
	var total int
	err := r.db.QueryRow(countQuery, unionID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("查询工会消息总数失败: %v", err)
	}

	// 查询消息列表
	query := fmt.Sprintf(`
		SELECT id, union_id, user_id, username, content, created_at
		FROM %s WHERE union_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, tableName)
	
	rows, err := r.db.Query(query, unionID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("查询工会消息失败: %v", err)
	}
	defer rows.Close()

	var messages []*entity.UnionChatMessage
	for rows.Next() {
		message := &entity.UnionChatMessage{}
		err := rows.Scan(
			&message.ID, &message.UnionID, &message.UserID,
			&message.Username, &message.Content, &message.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("扫描工会消息失败: %v", err)
		}
		messages = append(messages, message)
	}

	return messages, total, nil
}

// GetRecentMessages 获取最近的工会消息
func (r *MySQLChatRepository) GetRecentMessages(tableName string, unionID int, limit int) ([]*entity.UnionChatMessage, error) {
	query := fmt.Sprintf(`
		SELECT id, union_id, user_id, username, content, created_at
		FROM %s WHERE union_id = ?
		ORDER BY created_at DESC
		LIMIT ?
	`, tableName)
	
	rows, err := r.db.Query(query, unionID, limit)
	if err != nil {
		return nil, fmt.Errorf("查询最近工会消息失败: %v", err)
	}
	defer rows.Close()

	var messages []*entity.UnionChatMessage
	for rows.Next() {
		message := &entity.UnionChatMessage{}
		err := rows.Scan(
			&message.ID, &message.UnionID, &message.UserID,
			&message.Username, &message.Content, &message.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描最近工会消息失败: %v", err)
		}
		messages = append(messages, message)
	}

	return messages, nil
}

// BatchCreateMessages 批量创建工会消息
func (r *MySQLChatRepository) BatchCreateMessages(tableName string, messages []*entity.UnionChatMessage) error {
	if len(messages) == 0 {
		return nil
	}

	// 构建批量插入SQL
	placeholders := make([]string, len(messages))
	values := make([]interface{}, 0, len(messages)*4)
	
	for i, message := range messages {
		if err := message.Validate(); err != nil {
			return fmt.Errorf("消息验证失败[%d]: %v", i, err)
		}
		placeholders[i] = "(?, ?, ?, ?, ?)"
		values = append(values, message.UnionID, message.UserID, message.Username, message.Content, message.CreatedAt)
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (union_id, user_id, username, content, created_at)
		VALUES %s
	`, tableName, strings.Join(placeholders, ","))

	_, err := r.db.Exec(query, values...)
	if err != nil {
		return fmt.Errorf("批量创建工会消息失败: %v", err)
	}

	return nil
}

// GetActiveTable 获取当前活跃的分表
func (r *MySQLChatRepository) GetActiveTable() (*entity.UnionChatTable, error) {
	table := &entity.UnionChatTable{}
	query := `SELECT id, table_name, ` + "`year_month`" + ` FROM union_chat_tables WHERE is_active = 1 LIMIT 1`
	err := r.db.QueryRow(query).Scan(
		&table.ID, &table.TableName, &table.YearMonth,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("没有活跃的分表")
		}
		return nil, fmt.Errorf("查询活跃分表失败: %v", err)
	}

	table.IsActive = true
	return table, nil
}

// GetTableByYearMonth 根据年月获取分表
func (r *MySQLChatRepository) GetTableByYearMonth(yearMonth string) (*entity.UnionChatTable, error) {
	table := &entity.UnionChatTable{}
	query := "SELECT id, table_name, year_month, created_at, is_active FROM union_chat_tables WHERE year_month = ?"
	err := r.db.QueryRow(query, yearMonth).Scan(
		&table.ID, &table.TableName, &table.YearMonth,
		&table.CreatedAt, &table.IsActive,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("分表不存在")
		}
		return nil, fmt.Errorf("查询分表失败: %v", err)
	}

	return table, nil
}

// CreateTable 创建分表记录
func (r *MySQLChatRepository) CreateTable(table *entity.UnionChatTable) error {
	if err := table.Validate(); err != nil {
		return err
	}

	query := `
		INSERT INTO union_chat_tables (table_name, year_month, created_at, is_active)
		VALUES (?, ?, NOW(), ?)
	`
	result, err := r.db.Exec(query, table.TableName, table.YearMonth, table.IsActive)
	if err != nil {
		return fmt.Errorf("创建分表记录失败: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("获取插入ID失败: %v", err)
	}
	table.ID = int(id)

	return nil
}

// UpdateTable 更新分表记录
func (r *MySQLChatRepository) UpdateTable(table *entity.UnionChatTable) error {
	if err := table.Validate(); err != nil {
		return err
	}

	query := `
		UPDATE union_chat_tables 
		SET table_name = ?, is_active = ?
		WHERE year_month = ?
	`
	result, err := r.db.Exec(query, table.TableName, table.IsActive, table.YearMonth)
	if err != nil {
		return fmt.Errorf("更新分表记录失败: %v", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取影响行数失败: %v", err)
	}
	if affected == 0 {
		return fmt.Errorf("分表记录不存在或无变更")
	}

	return nil
}

// TableExists 检查表是否存在
func (r *MySQLChatRepository) TableExists(tableName string) (bool, error) {
	query := `
		SELECT COUNT(*) 
		FROM information_schema.tables 
		WHERE table_schema = DATABASE() AND table_name = ?
	`
	var count int
	err := r.db.QueryRow(query, tableName).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("检查表存在性失败: %v", err)
	}

	return count > 0, nil
}

// CreateMonthlyTable 创建新的月度聊天表
func (r *MySQLChatRepository) CreateMonthlyTable(tableName string) error {
	query := fmt.Sprintf(`
		CREATE TABLE %s (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			union_id INT NOT NULL,
			user_id INT NOT NULL,
			username VARCHAR(50) NOT NULL,
			content VARCHAR(30) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_union_created (union_id, created_at),
			INDEX idx_created_at (created_at)
		)
	`, tableName)

	_, err := r.db.Exec(query)
	if err != nil {
		return fmt.Errorf("创建月度聊天表失败: %v", err)
	}

	return nil
}

// GetAllTables 获取所有分表记录
func (r *MySQLChatRepository) GetAllTables() ([]*entity.UnionChatTable, error) {
	query := `
		SELECT id, table_name, year_month, created_at, is_active
		FROM union_chat_tables ORDER BY year_month DESC
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("查询所有分表记录失败: %v", err)
	}
	defer rows.Close()

	var tables []*entity.UnionChatTable
	for rows.Next() {
		table := &entity.UnionChatTable{}
		err := rows.Scan(
			&table.ID, &table.TableName, &table.YearMonth,
			&table.CreatedAt, &table.IsActive,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描分表记录失败: %v", err)
		}
		tables = append(tables, table)
	}

	return tables, nil
}

// DeactivateTable 设置表为非活跃状态
func (r *MySQLChatRepository) DeactivateTable(yearMonth string) error {
	query := `UPDATE union_chat_tables SET is_active = false WHERE year_month = ?`
	result, err := r.db.Exec(query, yearMonth)
	if err != nil {
		return fmt.Errorf("停用分表失败: %v", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取影响行数失败: %v", err)
	}
	if affected == 0 {
		return fmt.Errorf("分表不存在")
	}

	return nil
}

// 确保实现了所有接口
var _ repository.PrivateMessageRepository = (*MySQLChatRepository)(nil)
var _ repository.WorldChatRepository = (*MySQLChatRepository)(nil)
var _ repository.UnionChatRepository = (*MySQLChatRepository)(nil)