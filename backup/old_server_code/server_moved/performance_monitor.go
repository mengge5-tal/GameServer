package server

import (
	"GameServer/internal/database"
	"GameServer/pkg/logger"
	"runtime"
	"sync"
	"time"
)

// PerformanceMetrics 性能指标结构
type PerformanceMetrics struct {
	mu sync.RWMutex
	
	// 连接相关指标
	ActiveConnections    int64 `json:"active_connections"`
	TotalConnections     int64 `json:"total_connections"`
	DisconnectedClients  int64 `json:"disconnected_clients"`
	
	// 消息处理指标
	MessagesProcessed    int64 `json:"messages_processed"`
	MessagesPerSecond    int64 `json:"messages_per_second"`
	AverageResponseTime  int64 `json:"average_response_time_ms"`
	
	// 错误统计
	AuthFailures         int64 `json:"auth_failures"`
	RateLimitExceeded    int64 `json:"rate_limit_exceeded"`
	DatabaseErrors       int64 `json:"database_errors"`
	
	// 系统资源指标
	MemoryUsageMB        int64 `json:"memory_usage_mb"`
	GoroutineCount       int64 `json:"goroutine_count"`
	GCPauseDuration      int64 `json:"gc_pause_duration_ns"`
	
	// 时间戳
	LastUpdated          int64 `json:"last_updated"`
	
	// 内部计算用
	lastMessageCount     int64
	lastUpdateTime       time.Time
	responseTimes        []int64
	maxResponseSamples   int
}

// NewPerformanceMetrics 创建新的性能指标实例
func NewPerformanceMetrics() *PerformanceMetrics {
	return &PerformanceMetrics{
		maxResponseSamples: 1000, // 保留最近1000个响应时间样本
		responseTimes:      make([]int64, 0, 1000),
		lastUpdateTime:     time.Now(),
	}
}

// RecordConnection 记录新连接
func (pm *PerformanceMetrics) RecordConnection() {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.ActiveConnections++
	pm.TotalConnections++
}

// RecordDisconnection 记录断开连接
func (pm *PerformanceMetrics) RecordDisconnection() {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.ActiveConnections--
	pm.DisconnectedClients++
}

// RecordMessage 记录消息处理
func (pm *PerformanceMetrics) RecordMessage(responseTimeMs int64) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	pm.MessagesProcessed++
	
	// 添加响应时间样本
	if len(pm.responseTimes) >= pm.maxResponseSamples {
		// 移除最老的样本
		pm.responseTimes = pm.responseTimes[1:]
	}
	pm.responseTimes = append(pm.responseTimes, responseTimeMs)
	
	// 计算平均响应时间
	if len(pm.responseTimes) > 0 {
		var total int64
		for _, rt := range pm.responseTimes {
			total += rt
		}
		pm.AverageResponseTime = total / int64(len(pm.responseTimes))
	}
}

// RecordAuthFailure 记录认证失败
func (pm *PerformanceMetrics) RecordAuthFailure() {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.AuthFailures++
}

// RecordRateLimitExceeded 记录限流超限
func (pm *PerformanceMetrics) RecordRateLimitExceeded() {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.RateLimitExceeded++
}

// RecordDatabaseError 记录数据库错误
func (pm *PerformanceMetrics) RecordDatabaseError() {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.DatabaseErrors++
}

// UpdateSystemMetrics 更新系统指标
func (pm *PerformanceMetrics) UpdateSystemMetrics() {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	// 获取内存统计
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	pm.MemoryUsageMB = int64(memStats.Alloc / 1024 / 1024)
	pm.GoroutineCount = int64(runtime.NumGoroutine())
	pm.GCPauseDuration = int64(memStats.PauseTotalNs)
	
	// 计算每秒消息数
	now := time.Now()
	if !pm.lastUpdateTime.IsZero() {
		duration := now.Sub(pm.lastUpdateTime).Seconds()
		if duration > 0 {
			messagesDiff := pm.MessagesProcessed - pm.lastMessageCount
			pm.MessagesPerSecond = int64(float64(messagesDiff) / duration)
		}
	}
	
	pm.lastMessageCount = pm.MessagesProcessed
	pm.lastUpdateTime = now
	pm.LastUpdated = now.Unix()
}

// GetMetrics 获取当前指标快照
func (pm *PerformanceMetrics) GetMetrics() map[string]interface{} {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	
	return map[string]interface{}{
		"connections": map[string]interface{}{
			"active":       pm.ActiveConnections,
			"total":        pm.TotalConnections,
			"disconnected": pm.DisconnectedClients,
		},
		"messages": map[string]interface{}{
			"processed":              pm.MessagesProcessed,
			"per_second":             pm.MessagesPerSecond,
			"average_response_time":  pm.AverageResponseTime,
		},
		"errors": map[string]interface{}{
			"auth_failures":        pm.AuthFailures,
			"rate_limit_exceeded":  pm.RateLimitExceeded,
			"database_errors":      pm.DatabaseErrors,
		},
		"system": map[string]interface{}{
			"memory_usage_mb":      pm.MemoryUsageMB,
			"goroutine_count":      pm.GoroutineCount,
			"gc_pause_duration_ns": pm.GCPauseDuration,
		},
		"timestamp": pm.LastUpdated,
	}
}

// Reset 重置指标（用于测试或定期清理）
func (pm *PerformanceMetrics) Reset() {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	// 保留连接数，重置其他计数器
	pm.MessagesProcessed = 0
	pm.MessagesPerSecond = 0
	pm.AuthFailures = 0
	pm.RateLimitExceeded = 0
	pm.DatabaseErrors = 0
	pm.responseTimes = pm.responseTimes[:0]
	pm.lastMessageCount = 0
	pm.LastUpdated = time.Now().Unix()
}

// PerformanceMonitor 性能监控器
type PerformanceMonitor struct {
	metrics     *PerformanceMetrics
	hub         *Hub
	ticker      *time.Ticker
	done        chan bool
	updateInterval time.Duration
}

// NewPerformanceMonitor 创建新的性能监控器
func NewPerformanceMonitor(hub *Hub, updateInterval time.Duration) *PerformanceMonitor {
	return &PerformanceMonitor{
		metrics:        NewPerformanceMetrics(),
		hub:           hub,
		ticker:        time.NewTicker(updateInterval),
		done:          make(chan bool),
		updateInterval: updateInterval,
	}
}

// Start 启动性能监控
func (pm *PerformanceMonitor) Start() {
	logger.Info("Performance monitor started", map[string]interface{}{
		"update_interval": pm.updateInterval.String(),
	})
	
	go pm.monitor()
}

// Stop 停止性能监控
func (pm *PerformanceMonitor) Stop() {
	pm.ticker.Stop()
	close(pm.done)
	logger.Info("Performance monitor stopped", nil)
}

// monitor 监控主循环
func (pm *PerformanceMonitor) monitor() {
	for {
		select {
		case <-pm.ticker.C:
			pm.updateMetrics()
		case <-pm.done:
			return
		}
	}
}

// updateMetrics 更新所有指标
func (pm *PerformanceMonitor) updateMetrics() {
	// 更新系统指标
	pm.metrics.UpdateSystemMetrics()
	
	// 更新连接数（从Hub获取）
	if pm.hub != nil {
		pm.hub.mutex.RLock()
		activeConnections := int64(len(pm.hub.Clients))
		pm.hub.mutex.RUnlock()
		
		pm.metrics.mu.Lock()
		pm.metrics.ActiveConnections = activeConnections
		pm.metrics.mu.Unlock()
	}
	
	// 定期记录性能日志
	metrics := pm.metrics.GetMetrics()
	logger.Info("Performance metrics updated", metrics)
}

// GetMetrics 获取性能指标
func (pm *PerformanceMonitor) GetMetrics() map[string]interface{} {
	return pm.metrics.GetMetrics()
}

// GetMetricsWithDatabaseStats 获取包含数据库统计的完整指标
func (pm *PerformanceMonitor) GetMetricsWithDatabaseStats() map[string]interface{} {
	metrics := pm.metrics.GetMetrics()
	
	// 添加数据库连接池统计
	if pm.hub != nil && pm.hub.DB != nil {
		dbStats := database.GetConnectionPoolStats(pm.hub.DB)
		metrics["database"] = dbStats
	}
	
	// 添加缓存统计
	if userCache := GetUserCache(); userCache != nil {
		metrics["user_cache"] = userCache.GetStats()
	}
	
	if playerCache := GetPlayerInfoCache(); playerCache != nil {
		metrics["player_cache"] = playerCache.GetStats()
	}
	
	// 添加限流器统计
	if rateLimiter := GetRateLimiter(); rateLimiter != nil {
		metrics["rate_limiter"] = rateLimiter.GetStats()
	}
	
	return metrics
}

// RecordConnection 外部接口：记录连接
func (pm *PerformanceMonitor) RecordConnection() {
	pm.metrics.RecordConnection()
}

// RecordDisconnection 外部接口：记录断开连接
func (pm *PerformanceMonitor) RecordDisconnection() {
	pm.metrics.RecordDisconnection()
}

// RecordMessage 外部接口：记录消息
func (pm *PerformanceMonitor) RecordMessage(responseTimeMs int64) {
	pm.metrics.RecordMessage(responseTimeMs)
}

// RecordAuthFailure 外部接口：记录认证失败
func (pm *PerformanceMonitor) RecordAuthFailure() {
	pm.metrics.RecordAuthFailure()
}

// RecordRateLimitExceeded 外部接口：记录限流超限
func (pm *PerformanceMonitor) RecordRateLimitExceeded() {
	pm.metrics.RecordRateLimitExceeded()
}

// RecordDatabaseError 外部接口：记录数据库错误
func (pm *PerformanceMonitor) RecordDatabaseError() {
	pm.metrics.RecordDatabaseError()
}

// 全局性能监控器实例
var globalPerformanceMonitor *PerformanceMonitor

// InitPerformanceMonitor 初始化全局性能监控器
func InitPerformanceMonitor(hub *Hub, updateInterval time.Duration) {
	globalPerformanceMonitor = NewPerformanceMonitor(hub, updateInterval)
	globalPerformanceMonitor.Start()
}

// GetPerformanceMonitor 获取全局性能监控器
func GetPerformanceMonitor() *PerformanceMonitor {
	return globalPerformanceMonitor
}

// StopPerformanceMonitor 停止全局性能监控器
func StopPerformanceMonitor() {
	if globalPerformanceMonitor != nil {
		globalPerformanceMonitor.Stop()
	}
}