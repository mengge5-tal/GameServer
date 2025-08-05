package server

import (
	"GameServer/pkg/logger"
	"sync"
	"time"
)

// ClientRateInfo 客户端限流信息
type ClientRateInfo struct {
	requests   []time.Time
	lastAccess time.Time
}

// RateLimiter 线程安全的限流器
type RateLimiter struct {
	mu                   sync.RWMutex
	clients              map[string]*ClientRateInfo
	maxRequestsPerMinute int
	cleanupInterval      time.Duration
	ticker               *time.Ticker
	done                 chan bool
}

// NewRateLimiter 创建新的限流器
func NewRateLimiter(maxRequestsPerMinute int, cleanupInterval time.Duration) *RateLimiter {
	rl := &RateLimiter{
		clients:              make(map[string]*ClientRateInfo),
		maxRequestsPerMinute: maxRequestsPerMinute,
		cleanupInterval:      cleanupInterval,
		ticker:               time.NewTicker(cleanupInterval),
		done:                 make(chan bool),
	}
	
	// 启动清理 goroutine
	go rl.cleanup()
	return rl
}

// IsAllowed 检查是否允许请求
func (rl *RateLimiter) IsAllowed(clientID string) bool {
	now := time.Now()
	
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	// 获取或创建客户端信息
	clientInfo, exists := rl.clients[clientID]
	if !exists {
		clientInfo = &ClientRateInfo{
			requests:   make([]time.Time, 0),
			lastAccess: now,
		}
		rl.clients[clientID] = clientInfo
	}
	
	clientInfo.lastAccess = now
	
	// 清理过期的请求记录（1分钟前的）
	validRequests := make([]time.Time, 0, len(clientInfo.requests))
	for _, requestTime := range clientInfo.requests {
		if now.Sub(requestTime) < time.Minute {
			validRequests = append(validRequests, requestTime)
		}
	}
	clientInfo.requests = validRequests
	
	// 检查是否超过限制
	if len(clientInfo.requests) >= rl.maxRequestsPerMinute {
		return false
	}
	
	// 记录当前请求
	clientInfo.requests = append(clientInfo.requests, now)
	return true
}

// RemoveClient 移除客户端（当客户端断开连接时调用）
func (rl *RateLimiter) RemoveClient(clientID string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	delete(rl.clients, clientID)
}

// cleanup 定期清理不活跃的客户端
func (rl *RateLimiter) cleanup() {
	for {
		select {
		case <-rl.ticker.C:
			rl.performCleanup()
		case <-rl.done:
			return
		}
	}
}

// performCleanup 执行清理操作
func (rl *RateLimiter) performCleanup() {
	now := time.Now()
	inactiveThreshold := 5 * time.Minute // 5分钟无活动则清理
	
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	var removedCount int
	for clientID, clientInfo := range rl.clients {
		if now.Sub(clientInfo.lastAccess) > inactiveThreshold {
			delete(rl.clients, clientID)
			removedCount++
		}
	}
	
	if removedCount > 0 {
		logger.Info("Rate limiter cleanup completed", map[string]interface{}{
			"removed_clients": removedCount,
			"active_clients":  len(rl.clients),
		})
	}
}

// GetStats 获取限流器统计信息
func (rl *RateLimiter) GetStats() map[string]interface{} {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	
	var totalRequests int
	for _, clientInfo := range rl.clients {
		totalRequests += len(clientInfo.requests)
	}
	
	return map[string]interface{}{
		"active_clients":  len(rl.clients),
		"total_requests":  totalRequests,
		"max_per_minute":  rl.maxRequestsPerMinute,
	}
}

// Stop 停止限流器
func (rl *RateLimiter) Stop() {
	rl.ticker.Stop()
	close(rl.done)
}

// 全局限流器实例
var globalRateLimiter *RateLimiter

// InitRateLimiter 初始化全局限流器
func InitRateLimiter(maxRequestsPerMinute int, cleanupInterval time.Duration) {
	globalRateLimiter = NewRateLimiter(maxRequestsPerMinute, cleanupInterval)
}

// GetRateLimiter 获取全局限流器
func GetRateLimiter() *RateLimiter {
	return globalRateLimiter
}