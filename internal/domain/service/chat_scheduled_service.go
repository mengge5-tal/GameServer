package service

import (
	"GameServer/internal/infrastructure/cache"
	"fmt"
	"time"
)

// ChatScheduledService 聊天定时任务服务
type ChatScheduledService struct {
	unionChatCache *cache.UnionChatCache
	isRunning      bool
}

// NewChatScheduledService 创建聊天定时任务服务
func NewChatScheduledService(unionChatCache *cache.UnionChatCache) *ChatScheduledService {
	return &ChatScheduledService{
		unionChatCache: unionChatCache,
		isRunning:      false,
	}
}

// StartCleanupTasks 启动清理任务
func (s *ChatScheduledService) StartCleanupTasks() {
	if s.isRunning {
		return
	}
	
	s.isRunning = true
	
	// 启动Redis缓存清理任务（每5分钟清理一次过期数据）
	go s.startRedisCacheCleanup()
	
	fmt.Println("聊天系统定时清理任务已启动")
}

// StopCleanupTasks 停止清理任务
func (s *ChatScheduledService) StopCleanupTasks() {
	s.isRunning = false
	fmt.Println("聊天系统定时清理任务已停止")
}

// startRedisCacheCleanup 启动Redis缓存清理任务
func (s *ChatScheduledService) startRedisCacheCleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for s.isRunning {
		select {
		case <-ticker.C:
			if s.unionChatCache != nil {
				if err := s.unionChatCache.CleanupExpiredData(); err != nil {
					fmt.Printf("Redis缓存清理失败: %v\n", err)
				}
			}
		}
	}
}

// GetCleanupStats 获取清理统计信息
func (s *ChatScheduledService) GetCleanupStats() map[string]interface{} {
	stats := make(map[string]interface{})
	
	stats["is_running"] = s.isRunning
	stats["cleanup_interval"] = "5 minutes"
	
	// 获取Redis缓存统计
	if s.unionChatCache != nil {
		if cacheStats, err := s.unionChatCache.GetCacheStats(); err == nil {
			stats["redis_cache"] = cacheStats
		}
	}
	
	return stats
}