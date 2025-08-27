package service

import (
	"fmt"
	"sync"
	"time"
)

// RateLimiter 频率限制器接口
type RateLimiter interface {
	// Allow 检查是否允许操作，返回是否允许和剩余等待时间
	Allow(key string) (bool, time.Duration)
	// Reset 重置指定key的限制
	Reset(key string)
	// GetRemaining 获取剩余操作次数
	GetRemaining(key string) int
}

// TokenBucketLimiter 令牌桶限制器
type TokenBucketLimiter struct {
	mutex    sync.RWMutex
	buckets  map[string]*tokenBucket
	rate     int           // 每秒生成的令牌数
	capacity int           // 桶容量
	interval time.Duration // 令牌生成间隔
}

// tokenBucket 令牌桶
type tokenBucket struct {
	tokens     int       // 当前令牌数
	lastUpdate time.Time // 最后更新时间
	capacity   int       // 桶容量
	rate       int       // 令牌生成速率
}

// NewTokenBucketLimiter 创建令牌桶限制器
func NewTokenBucketLimiter(rate int, capacity int) *TokenBucketLimiter {
	return &TokenBucketLimiter{
		buckets:  make(map[string]*tokenBucket),
		rate:     rate,
		capacity: capacity,
		interval: time.Second / time.Duration(rate),
	}
}

// Allow 检查是否允许操作
func (tbl *TokenBucketLimiter) Allow(key string) (bool, time.Duration) {
	tbl.mutex.Lock()
	defer tbl.mutex.Unlock()

	bucket := tbl.getBucket(key)
	now := time.Now()

	// 计算从上次更新到现在应该添加的令牌数
	elapsed := now.Sub(bucket.lastUpdate)
	tokensToAdd := int(elapsed.Seconds()) * bucket.rate
	
	// 更新令牌数，不超过容量
	bucket.tokens = min(bucket.capacity, bucket.tokens+tokensToAdd)
	bucket.lastUpdate = now

	// 检查是否有令牌可用
	if bucket.tokens > 0 {
		bucket.tokens--
		return true, 0
	}

	// 计算需要等待的时间（生成下一个令牌的时间）
	waitTime := tbl.interval
	return false, waitTime
}

// Reset 重置指定key的限制
func (tbl *TokenBucketLimiter) Reset(key string) {
	tbl.mutex.Lock()
	defer tbl.mutex.Unlock()
	
	delete(tbl.buckets, key)
}

// GetRemaining 获取剩余令牌数
func (tbl *TokenBucketLimiter) GetRemaining(key string) int {
	tbl.mutex.RLock()
	defer tbl.mutex.RUnlock()

	bucket := tbl.getBucket(key)
	now := time.Now()

	// 计算当前应该有的令牌数
	elapsed := now.Sub(bucket.lastUpdate)
	tokensToAdd := int(elapsed.Seconds()) * bucket.rate
	currentTokens := min(bucket.capacity, bucket.tokens+tokensToAdd)

	return currentTokens
}

// getBucket 获取或创建令牌桶（需要在锁保护下调用）
func (tbl *TokenBucketLimiter) getBucket(key string) *tokenBucket {
	bucket, exists := tbl.buckets[key]
	if !exists {
		bucket = &tokenBucket{
			tokens:     tbl.capacity,
			lastUpdate: time.Now(),
			capacity:   tbl.capacity,
			rate:       tbl.rate,
		}
		tbl.buckets[key] = bucket
	}
	return bucket
}

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// FixedWindowLimiter 固定窗口限制器
type FixedWindowLimiter struct {
	mutex     sync.RWMutex
	windows   map[string]*fixedWindow
	limit     int           // 窗口内允许的最大请求数
	window    time.Duration // 窗口大小
}

// fixedWindow 固定窗口
type fixedWindow struct {
	count      int       // 当前窗口内的请求数
	windowStart time.Time // 窗口开始时间
	limit      int       // 窗口限制
	window     time.Duration // 窗口大小
}

// NewFixedWindowLimiter 创建固定窗口限制器
func NewFixedWindowLimiter(limit int, window time.Duration) *FixedWindowLimiter {
	return &FixedWindowLimiter{
		windows: make(map[string]*fixedWindow),
		limit:   limit,
		window:  window,
	}
}

// Allow 检查是否允许操作
func (fwl *FixedWindowLimiter) Allow(key string) (bool, time.Duration) {
	fwl.mutex.Lock()
	defer fwl.mutex.Unlock()

	window := fwl.getWindow(key)
	now := time.Now()

	// 检查是否需要重置窗口
	if now.Sub(window.windowStart) >= window.window {
		window.count = 0
		window.windowStart = now
	}

	// 检查是否超过限制
	if window.count >= window.limit {
		// 计算到下个窗口的等待时间
		nextWindow := window.windowStart.Add(window.window)
		waitTime := nextWindow.Sub(now)
		return false, waitTime
	}

	// 增加计数
	window.count++
	return true, 0
}

// Reset 重置指定key的限制
func (fwl *FixedWindowLimiter) Reset(key string) {
	fwl.mutex.Lock()
	defer fwl.mutex.Unlock()
	
	delete(fwl.windows, key)
}

// GetRemaining 获取剩余次数
func (fwl *FixedWindowLimiter) GetRemaining(key string) int {
	fwl.mutex.RLock()
	defer fwl.mutex.RUnlock()

	window := fwl.getWindow(key)
	now := time.Now()

	// 检查是否需要重置窗口
	if now.Sub(window.windowStart) >= window.window {
		return window.limit
	}

	remaining := window.limit - window.count
	if remaining < 0 {
		return 0
	}
	return remaining
}

// getWindow 获取或创建固定窗口（需要在锁保护下调用）
func (fwl *FixedWindowLimiter) getWindow(key string) *fixedWindow {
	window, exists := fwl.windows[key]
	if !exists {
		window = &fixedWindow{
			count:      0,
			windowStart: time.Now(),
			limit:      fwl.limit,
			window:     fwl.window,
		}
		fwl.windows[key] = window
	}
	return window
}

// ChatRateLimiter 聊天专用频率限制器
type ChatRateLimiter struct {
	privateChat RateLimiter // 私聊限制器
	worldChat   RateLimiter // 世界聊天限制器
	unionChat   RateLimiter // 工会聊天限制器
}

// NewChatRateLimiter 创建聊天频率限制器
func NewChatRateLimiter() *ChatRateLimiter {
	return &ChatRateLimiter{
		// 私聊：每秒最多1条消息，桶容量2（允许短时burst）
		privateChat: NewTokenBucketLimiter(1, 2),
		// 世界聊天：每2秒最多1条消息，桶容量1
		worldChat: NewFixedWindowLimiter(1, 2*time.Second),
		// 工会聊天：每2秒最多1条消息，桶容量1  
		unionChat: NewFixedWindowLimiter(1, 2*time.Second),
	}
}

// CheckPrivateChat 检查私聊频率限制
func (crl *ChatRateLimiter) CheckPrivateChat(userID int) (bool, time.Duration, error) {
	key := fmt.Sprintf("private:%d", userID)
	allowed, waitTime := crl.privateChat.Allow(key)
	if !allowed {
		return false, waitTime, fmt.Errorf("私聊发送过于频繁，请等待%v后重试", waitTime)
	}
	return true, 0, nil
}

// CheckWorldChat 检查世界聊天频率限制
func (crl *ChatRateLimiter) CheckWorldChat(userID int) (bool, time.Duration, error) {
	key := fmt.Sprintf("world:%d", userID)
	allowed, waitTime := crl.worldChat.Allow(key)
	if !allowed {
		return false, waitTime, fmt.Errorf("世界聊天发送过于频繁，请等待%v后重试", waitTime)
	}
	return true, 0, nil
}

// CheckUnionChat 检查工会聊天频率限制
func (crl *ChatRateLimiter) CheckUnionChat(userID int) (bool, time.Duration, error) {
	key := fmt.Sprintf("union:%d", userID)
	allowed, waitTime := crl.unionChat.Allow(key)
	if !allowed {
		return false, waitTime, fmt.Errorf("工会聊天发送过于频繁，请等待%v后重试", waitTime)
	}
	return true, 0, nil
}

// ResetUserLimits 重置用户的所有频率限制（用于管理员操作）
func (crl *ChatRateLimiter) ResetUserLimits(userID int) {
	privateKey := fmt.Sprintf("private:%d", userID)
	worldKey := fmt.Sprintf("world:%d", userID)
	unionKey := fmt.Sprintf("union:%d", userID)

	crl.privateChat.Reset(privateKey)
	crl.worldChat.Reset(worldKey)
	crl.unionChat.Reset(unionKey)
}

// GetRemainingLimits 获取用户各类聊天的剩余次数
func (crl *ChatRateLimiter) GetRemainingLimits(userID int) (private, world, union int) {
	privateKey := fmt.Sprintf("private:%d", userID)
	worldKey := fmt.Sprintf("world:%d", userID)
	unionKey := fmt.Sprintf("union:%d", userID)

	private = crl.privateChat.GetRemaining(privateKey)
	world = crl.worldChat.GetRemaining(worldKey)
	union = crl.unionChat.GetRemaining(unionKey)

	return
}