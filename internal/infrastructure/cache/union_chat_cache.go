package cache

import (
	"GameServer/internal/domain/entity"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// UnionChatCache 工会聊天缓存服务
type UnionChatCache struct {
	redis *RedisClient
}

// NewUnionChatCache 创建工会聊天缓存服务
func NewUnionChatCache(redisClient *RedisClient) *UnionChatCache {
	return &UnionChatCache{
		redis: redisClient,
	}
}

// ========== 消息缓存操作 ==========

// AddMessage 添加工会消息到缓存
func (ucc *UnionChatCache) AddMessage(unionID int, message *entity.UnionChatMessage) error {
	// 1. 序列化消息
	messageJSON, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("消息序列化失败: %v", err)
	}

	// 2. 添加到消息列表（左侧插入，保持最新消息在前）
	messagesKey := ucc.getMessagesKey(unionID)
	if err := ucc.redis.LPush(messagesKey, messageJSON); err != nil {
		return fmt.Errorf("添加消息到缓存失败: %v", err)
	}

	// 3. 维护列表大小（最多100条）
	if err := ucc.redis.LTrim(messagesKey, 0, 99); err != nil {
		return fmt.Errorf("裁剪消息列表失败: %v", err)
	}

	// 4. 设置过期时间（30分钟）
	if err := ucc.redis.Expire(messagesKey, 30*time.Minute); err != nil {
		return fmt.Errorf("设置消息列表过期时间失败: %v", err)
	}

	// 5. 更新消息计数
	if err := ucc.incrementMessageCount(unionID); err != nil {
		// 计数更新失败不影响主要功能
		fmt.Printf("更新工会 %d 消息计数失败: %v\n", unionID, err)
	}

	// 6. 设置批量写入标记（用于定时任务检查）
	flushKey := ucc.getFlushKey(unionID)
	if err := ucc.redis.Set(flushKey, "pending", 35*time.Minute); err != nil {
		// 设置批量写入标记失败不影响主要功能
		fmt.Printf("设置工会 %d 批量写入标记失败: %v\n", unionID, err)
	}

	return nil
}

// GetMessages 获取工会消息列表
func (ucc *UnionChatCache) GetMessages(unionID int, limit int) ([]*entity.UnionChatMessage, error) {
	if limit <= 0 || limit > 100 {
		limit = 100
	}

	// 获取消息列表
	messagesKey := ucc.getMessagesKey(unionID)
	messageStrings, err := ucc.redis.LRange(messagesKey, 0, int64(limit-1))
	if err != nil {
		if err == redis.Nil {
			return []*entity.UnionChatMessage{}, nil
		}
		return nil, fmt.Errorf("获取消息列表失败: %v", err)
	}

	// 反序列化消息
	messages := make([]*entity.UnionChatMessage, 0, len(messageStrings))
	for _, messageStr := range messageStrings {
		var message entity.UnionChatMessage
		if err := json.Unmarshal([]byte(messageStr), &message); err != nil {
			fmt.Printf("反序列化消息失败: %v, 消息内容: %s\n", err, messageStr)
			continue
		}
		messages = append(messages, &message)
	}

	return messages, nil
}

// GetMessageCount 获取工会消息总数
func (ucc *UnionChatCache) GetMessageCount(unionID int) (int64, error) {
	countKey := ucc.getCountKey(unionID)
	countStr, err := ucc.redis.Get(countKey)
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		return 0, fmt.Errorf("获取消息计数失败: %v", err)
	}

	count, err := strconv.ParseInt(countStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("解析消息计数失败: %v", err)
	}

	return count, nil
}

// ClearMessages 清空工会消息缓存
func (ucc *UnionChatCache) ClearMessages(unionID int) error {
	messagesKey := ucc.getMessagesKey(unionID)
	countKey := ucc.getCountKey(unionID)
	
	return ucc.redis.Del(messagesKey, countKey)
}

// ========== 用户在线状态缓存 ==========

// SetUserOnline 设置用户在线状态
func (ucc *UnionChatCache) SetUserOnline(userID int, unionID int) error {
	// 添加到工会在线用户集合
	onlineKey := ucc.getOnlineUsersKey(unionID)
	if err := ucc.redis.SAdd(onlineKey, userID); err != nil {
		return fmt.Errorf("添加在线用户失败: %v", err)
	}

	// 设置过期时间（5分钟，需要定期刷新）
	if err := ucc.redis.Expire(onlineKey, 5*time.Minute); err != nil {
		return fmt.Errorf("设置在线用户过期时间失败: %v", err)
	}

	// 设置用户到工会的映射
	userUnionKey := ucc.getUserUnionKey(userID)
	if err := ucc.redis.Set(userUnionKey, unionID, 5*time.Minute); err != nil {
		return fmt.Errorf("设置用户工会映射失败: %v", err)
	}

	return nil
}

// SetUserOffline 设置用户离线状态
func (ucc *UnionChatCache) SetUserOffline(userID int) error {
	// 获取用户所在工会
	unionID, err := ucc.getUserUnion(userID)
	if err != nil {
		return err
	}
	if unionID == 0 {
		return nil // 用户不在任何工会
	}

	// 从工会在线用户集合中移除
	onlineKey := ucc.getOnlineUsersKey(unionID)
	if err := ucc.redis.SRem(onlineKey, userID); err != nil {
		return fmt.Errorf("移除在线用户失败: %v", err)
	}

	// 删除用户工会映射
	userUnionKey := ucc.getUserUnionKey(userID)
	return ucc.redis.Del(userUnionKey)
}

// GetOnlineUsers 获取工会在线用户列表
func (ucc *UnionChatCache) GetOnlineUsers(unionID int) ([]int, error) {
	onlineKey := ucc.getOnlineUsersKey(unionID)
	userStrings, err := ucc.redis.SMembers(onlineKey)
	if err != nil {
		if err == redis.Nil {
			return []int{}, nil
		}
		return nil, fmt.Errorf("获取在线用户失败: %v", err)
	}

	userIDs := make([]int, 0, len(userStrings))
	for _, userStr := range userStrings {
		userID, err := strconv.Atoi(userStr)
		if err != nil {
			fmt.Printf("解析用户ID失败: %v, 用户ID: %s\n", err, userStr)
			continue
		}
		userIDs = append(userIDs, userID)
	}

	return userIDs, nil
}

// IsUserOnline 检查用户是否在线
func (ucc *UnionChatCache) IsUserOnline(userID int, unionID int) (bool, error) {
	onlineKey := ucc.getOnlineUsersKey(unionID)
	return ucc.redis.SIsMember(onlineKey, userID)
}

// RefreshUserOnlineStatus 刷新用户在线状态（心跳）
func (ucc *UnionChatCache) RefreshUserOnlineStatus(userID int) error {
	// 获取用户所在工会
	unionID, err := ucc.getUserUnion(userID)
	if err != nil {
		return err
	}
	if unionID == 0 {
		return nil // 用户不在任何工会
	}

	// 刷新在线状态
	return ucc.SetUserOnline(userID, unionID)
}

// ========== 批量操作 ==========

// BatchAddMessages 批量添加消息（用于数据库同步）
func (ucc *UnionChatCache) BatchAddMessages(unionID int, messages []*entity.UnionChatMessage) error {
	if len(messages) == 0 {
		return nil
	}

	// 使用管道批量操作
	pipe := ucc.redis.Pipeline()
	messagesKey := ucc.getMessagesKey(unionID)

	for _, message := range messages {
		messageJSON, err := json.Marshal(message)
		if err != nil {
			return fmt.Errorf("序列化消息失败: %v", err)
		}
		pipe.RPush(ucc.redis.GetContext(), messagesKey, messageJSON)
	}

	// 维护列表大小
	pipe.LTrim(ucc.redis.GetContext(), messagesKey, 0, 99)
	pipe.Expire(ucc.redis.GetContext(), messagesKey, 30*time.Minute)

	// 执行管道
	_, err := pipe.Exec(ucc.redis.GetContext())
	if err != nil {
		return fmt.Errorf("批量添加消息失败: %v", err)
	}

	return nil
}

// ========== 统计功能 ==========

// GetCacheStats 获取缓存统计信息
func (ucc *UnionChatCache) GetCacheStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 获取所有工会消息键
	pattern := "union:messages:*"
	keys, err := ucc.redis.Keys(pattern)
	if err != nil {
		return nil, fmt.Errorf("获取工会消息键失败: %v", err)
	}

	totalMessages := int64(0)
	totalUnions := len(keys)

	for _, key := range keys {
		count, err := ucc.redis.LLen(key)
		if err != nil {
			continue
		}
		totalMessages += count
	}

	// 获取在线用户统计
	onlinePattern := "union:online:*"
	onlineKeys, err := ucc.redis.Keys(onlinePattern)
	if err == nil {
		totalOnlineUsers := int64(0)
		for _, key := range onlineKeys {
			count, err := ucc.redis.SCard(key)
			if err != nil {
				continue
			}
			totalOnlineUsers += count
		}
		stats["online_users"] = totalOnlineUsers
	}

	stats["total_unions"] = totalUnions
	stats["total_messages"] = totalMessages
	stats["avg_messages_per_union"] = float64(totalMessages) / float64(totalUnions)

	return stats, nil
}

// ========== 内部辅助方法 ==========

// getMessagesKey 获取工会消息列表键
func (ucc *UnionChatCache) getMessagesKey(unionID int) string {
	return fmt.Sprintf("union:messages:%d", unionID)
}

// getCountKey 获取工会消息计数键
func (ucc *UnionChatCache) getCountKey(unionID int) string {
	return fmt.Sprintf("union:count:%d", unionID)
}

// getOnlineUsersKey 获取工会在线用户键
func (ucc *UnionChatCache) getOnlineUsersKey(unionID int) string {
	return fmt.Sprintf("union:online:%d", unionID)
}

// getUserUnionKey 获取用户工会映射键
func (ucc *UnionChatCache) getUserUnionKey(userID int) string {
	return fmt.Sprintf("user:union:%d", userID)
}

// getFlushKey 获取工会批量写入标记键
func (ucc *UnionChatCache) getFlushKey(unionID int) string {
	return fmt.Sprintf("union:flush:%d", unionID)
}

// incrementMessageCount 增加消息计数
func (ucc *UnionChatCache) incrementMessageCount(unionID int) error {
	countKey := ucc.getCountKey(unionID)
	
	// 使用事务确保原子性
	pipe := ucc.redis.TxPipeline()
	pipe.Incr(ucc.redis.GetContext(), countKey)
	pipe.Expire(ucc.redis.GetContext(), countKey, 24*time.Hour) // 计数器24小时过期
	
	_, err := pipe.Exec(ucc.redis.GetContext())
	return err
}

// getUserUnion 获取用户所在工会
func (ucc *UnionChatCache) getUserUnion(userID int) (int, error) {
	userUnionKey := ucc.getUserUnionKey(userID)
	unionStr, err := ucc.redis.Get(userUnionKey)
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		return 0, fmt.Errorf("获取用户工会映射失败: %v", err)
	}

	unionID, err := strconv.Atoi(unionStr)
	if err != nil {
		return 0, fmt.Errorf("解析工会ID失败: %v", err)
	}

	return unionID, nil
}

// ========== 清理功能 ==========

// CleanupExpiredData 清理过期数据（定时任务）
func (ucc *UnionChatCache) CleanupExpiredData() error {
	// 清理过期的在线用户数据
	pattern := "union:online:*"
	keys, err := ucc.redis.Keys(pattern)
	if err != nil {
		return fmt.Errorf("获取在线用户键失败: %v", err)
	}

	expiredKeys := make([]string, 0)
	for _, key := range keys {
		ttl, err := ucc.redis.TTL(key)
		if err != nil {
			continue
		}
		
		// 如果TTL为-1（永不过期）或小于1分钟，标记为需要清理
		if ttl == -1 || (ttl > 0 && ttl < time.Minute) {
			expiredKeys = append(expiredKeys, key)
		}
	}

	if len(expiredKeys) > 0 {
		if err := ucc.redis.Del(expiredKeys...); err != nil {
			return fmt.Errorf("清理过期在线用户数据失败: %v", err)
		}
		fmt.Printf("清理了 %d 个过期的在线用户键\n", len(expiredKeys))
	}

	return nil
}

// GetPendingFlushUnions 获取需要批量写入的工会列表
func (ucc *UnionChatCache) GetPendingFlushUnions() ([]int, error) {
	pattern := "union:flush:*"
	keys, err := ucc.redis.Keys(pattern)
	if err != nil {
		return nil, fmt.Errorf("获取批量写入标记键失败: %v", err)
	}

	unionIDs := make([]int, 0, len(keys))
	for _, key := range keys {
		// 从 "union:flush:123" 提取工会ID
		if len(key) > 12 {
			unionIDStr := key[12:] // 去掉 "union:flush:" 前缀
			if unionID, err := strconv.Atoi(unionIDStr); err == nil {
				unionIDs = append(unionIDs, unionID)
			}
		}
	}

	return unionIDs, nil
}

// ClearFlushMark 清除批量写入标记
func (ucc *UnionChatCache) ClearFlushMark(unionID int) error {
	flushKey := ucc.getFlushKey(unionID)
	return ucc.redis.Del(flushKey)
}