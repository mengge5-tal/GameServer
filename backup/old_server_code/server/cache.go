package server

import (
	"GameServer/internal/models"
	"GameServer/pkg/logger"
	"sync"
	"time"
)

// CachedUser 缓存的用户信息
type CachedUser struct {
	User      *models.User
	LastAccess time.Time
	CreatedAt  time.Time
}

// UserCache 用户信息缓存
type UserCache struct {
	mu       sync.RWMutex
	users    map[int]*CachedUser
	ttl      time.Duration
	maxSize  int
	ticker   *time.Ticker
	done     chan bool
}

// NewUserCache 创建新的用户缓存
func NewUserCache(ttl time.Duration, maxSize int, cleanupInterval time.Duration) *UserCache {
	uc := &UserCache{
		users:   make(map[int]*CachedUser),
		ttl:     ttl,
		maxSize: maxSize,
		ticker:  time.NewTicker(cleanupInterval),
		done:    make(chan bool),
	}
	
	// 启动清理 goroutine
	go uc.cleanup()
	return uc
}

// Get 获取用户信息
func (uc *UserCache) Get(userID int) (*models.User, bool) {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	
	cachedUser, exists := uc.users[userID]
	if !exists {
		return nil, false
	}
	
	// 检查是否过期
	if time.Since(cachedUser.CreatedAt) > uc.ttl {
		delete(uc.users, userID)
		return nil, false
	}
	
	// 更新访问时间
	cachedUser.LastAccess = time.Now()
	return cachedUser.User, true
}

// Set 设置用户信息
func (uc *UserCache) Set(userID int, user *models.User) {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	
	// 如果缓存已满，移除最老的条目
	if len(uc.users) >= uc.maxSize {
		uc.evictOldest()
	}
	
	now := time.Now()
	uc.users[userID] = &CachedUser{
		User:       user,
		LastAccess: now,
		CreatedAt:  now,
	}
}

// Delete 删除用户信息
func (uc *UserCache) Delete(userID int) {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	delete(uc.users, userID)
}

// evictOldest 移除最老的条目（需要在锁保护下调用）
func (uc *UserCache) evictOldest() {
	var oldestUserID int
	var oldestTime time.Time
	first := true
	
	for userID, cachedUser := range uc.users {
		if first || cachedUser.LastAccess.Before(oldestTime) {
			oldestUserID = userID
			oldestTime = cachedUser.LastAccess
			first = false
		}
	}
	
	if !first {
		delete(uc.users, oldestUserID)
	}
}

// cleanup 定期清理过期条目
func (uc *UserCache) cleanup() {
	for {
		select {
		case <-uc.ticker.C:
			uc.performCleanup()
		case <-uc.done:
			return
		}
	}
}

// performCleanup 执行清理操作
func (uc *UserCache) performCleanup() {
	now := time.Now()
	
	uc.mu.Lock()
	defer uc.mu.Unlock()
	
	var expiredCount int
	for userID, cachedUser := range uc.users {
		if now.Sub(cachedUser.CreatedAt) > uc.ttl {
			delete(uc.users, userID)
			expiredCount++
		}
	}
	
	if expiredCount > 0 {
		logger.Info("User cache cleanup completed", map[string]interface{}{
			"expired_entries": expiredCount,
			"active_entries":  len(uc.users),
		})
	}
}

// GetStats 获取缓存统计信息
func (uc *UserCache) GetStats() map[string]interface{} {
	uc.mu.RLock()
	defer uc.mu.RUnlock()
	
	var hitCount, missCount int64
	totalEntries := len(uc.users)
	
	return map[string]interface{}{
		"total_entries": totalEntries,
		"max_size":      uc.maxSize,
		"ttl_seconds":   int(uc.ttl.Seconds()),
		"hit_count":     hitCount,
		"miss_count":    missCount,
	}
}

// Stop 停止缓存
func (uc *UserCache) Stop() {
	uc.ticker.Stop()
	close(uc.done)
}

// PlayerInfoCache 玩家信息缓存
type PlayerInfoCache struct {
	mu       sync.RWMutex
	players  map[int]*CachedPlayerInfo
	ttl      time.Duration
	maxSize  int
	ticker   *time.Ticker
	done     chan bool
}

// CachedPlayerInfo 缓存的玩家信息
type CachedPlayerInfo struct {
	Level       int
	Experience  int
	GameLevel   int
	BloodEnergy int
	LastAccess  time.Time
	CreatedAt   time.Time
}

// NewPlayerInfoCache 创建新的玩家信息缓存
func NewPlayerInfoCache(ttl time.Duration, maxSize int, cleanupInterval time.Duration) *PlayerInfoCache {
	pic := &PlayerInfoCache{
		players: make(map[int]*CachedPlayerInfo),
		ttl:     ttl,
		maxSize: maxSize,
		ticker:  time.NewTicker(cleanupInterval),
		done:    make(chan bool),
	}
	
	go pic.cleanup()
	return pic
}

// Get 获取玩家信息
func (pic *PlayerInfoCache) Get(userID int) (*CachedPlayerInfo, bool) {
	pic.mu.Lock()
	defer pic.mu.Unlock()
	
	cachedPlayer, exists := pic.players[userID]
	if !exists {
		return nil, false
	}
	
	if time.Since(cachedPlayer.CreatedAt) > pic.ttl {
		delete(pic.players, userID)
		return nil, false
	}
	
	cachedPlayer.LastAccess = time.Now()
	return cachedPlayer, true
}

// Set 设置玩家信息
func (pic *PlayerInfoCache) Set(userID, level, experience, gameLevel, bloodEnergy int) {
	pic.mu.Lock()
	defer pic.mu.Unlock()
	
	if len(pic.players) >= pic.maxSize {
		pic.evictOldest()
	}
	
	now := time.Now()
	pic.players[userID] = &CachedPlayerInfo{
		Level:       level,
		Experience:  experience,
		GameLevel:   gameLevel,
		BloodEnergy: bloodEnergy,
		LastAccess:  now,
		CreatedAt:   now,
	}
}

// Delete 删除玩家信息
func (pic *PlayerInfoCache) Delete(userID int) {
	pic.mu.Lock()
	defer pic.mu.Unlock()
	delete(pic.players, userID)
}

// evictOldest 移除最老的条目
func (pic *PlayerInfoCache) evictOldest() {
	var oldestUserID int
	var oldestTime time.Time
	first := true
	
	for userID, cachedPlayer := range pic.players {
		if first || cachedPlayer.LastAccess.Before(oldestTime) {
			oldestUserID = userID
			oldestTime = cachedPlayer.LastAccess
			first = false
		}
	}
	
	if !first {
		delete(pic.players, oldestUserID)
	}
}

// cleanup 定期清理过期条目
func (pic *PlayerInfoCache) cleanup() {
	for {
		select {
		case <-pic.ticker.C:
			pic.performCleanup()
		case <-pic.done:
			return
		}
	}
}

// performCleanup 执行清理操作
func (pic *PlayerInfoCache) performCleanup() {
	now := time.Now()
	
	pic.mu.Lock()
	defer pic.mu.Unlock()
	
	var expiredCount int
	for userID, cachedPlayer := range pic.players {
		if now.Sub(cachedPlayer.CreatedAt) > pic.ttl {
			delete(pic.players, userID)
			expiredCount++
		}
	}
	
	if expiredCount > 0 {
		logger.Info("Player info cache cleanup completed", map[string]interface{}{
			"expired_entries": expiredCount,
			"active_entries":  len(pic.players),
		})
	}
}

// GetStats 获取缓存统计信息
func (pic *PlayerInfoCache) GetStats() map[string]interface{} {
	pic.mu.RLock()
	defer pic.mu.RUnlock()
	
	return map[string]interface{}{
		"total_entries": len(pic.players),
		"max_size":      pic.maxSize,
		"ttl_seconds":   int(pic.ttl.Seconds()),
	}
}

// Stop 停止缓存
func (pic *PlayerInfoCache) Stop() {
	pic.ticker.Stop()
	close(pic.done)
}

// 全局缓存实例
var (
	globalUserCache       *UserCache
	globalPlayerInfoCache *PlayerInfoCache
)

// InitCaches 初始化全局缓存
func InitCaches() {
	// 用户缓存：5分钟TTL，最大1000条目，每分钟清理一次
	globalUserCache = NewUserCache(5*time.Minute, 1000, time.Minute)
	
	// 玩家信息缓存：3分钟TTL，最大2000条目，每30秒清理一次
	globalPlayerInfoCache = NewPlayerInfoCache(3*time.Minute, 2000, 30*time.Second)
	
	logger.Info("Caches initialized successfully", map[string]interface{}{
		"user_cache_ttl":         "5m",
		"user_cache_max_size":    1000,
		"player_cache_ttl":       "3m",
		"player_cache_max_size":  2000,
	})
}

// GetUserCache 获取用户缓存
func GetUserCache() *UserCache {
	return globalUserCache
}

// GetPlayerInfoCache 获取玩家信息缓存
func GetPlayerInfoCache() *PlayerInfoCache {
	return globalPlayerInfoCache
}

// StopCaches 停止所有缓存
func StopCaches() {
	if globalUserCache != nil {
		globalUserCache.Stop()
	}
	if globalPlayerInfoCache != nil {
		globalPlayerInfoCache.Stop()
	}
}