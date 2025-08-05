package cache

import (
	"GameServer/internal/domain/entity"
	"encoding/json"
	"sync"
	"time"
)

// CacheService defines cache operations
type CacheService interface {
	// User cache operations
	GetUser(key string) (*entity.User, error)
	SetUser(key string, user *entity.User) error
	
	// Player cache operations  
	GetPlayerInfo(key string) (*entity.PlayerInfo, error)
	SetPlayerInfo(key string, playerInfo *entity.PlayerInfo) error
	
	// Equipment cache operations
	GetEquipment(key string) ([]*entity.Equipment, error)
	SetEquipment(key string, equipment []*entity.Equipment) error
	
	// Generic operations
	Get(key string) ([]byte, error)
	Set(key string, value []byte, ttl time.Duration) error
	Delete(key string) error
	Clear() error
}

// memoryCache implements CacheService using in-memory storage
type memoryCache struct {
	data map[string]cacheItem
	mu   sync.RWMutex
}

type cacheItem struct {
	data      []byte
	expiresAt time.Time
}

// NewMemoryCache creates a new memory cache
func NewMemoryCache() CacheService {
	cache := &memoryCache{
		data: make(map[string]cacheItem),
	}
	
	// Start cleanup goroutine
	go cache.cleanup()
	
	return cache
}

// Get retrieves data from cache
func (c *memoryCache) Get(key string) ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	item, exists := c.data[key]
	if !exists {
		return nil, ErrCacheKeyNotFound
	}
	
	if time.Now().After(item.expiresAt) {
		delete(c.data, key)
		return nil, ErrCacheKeyExpired
	}
	
	return item.data, nil
}

// Set stores data in cache
func (c *memoryCache) Set(key string, value []byte, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.data[key] = cacheItem{
		data:      value,
		expiresAt: time.Now().Add(ttl),
	}
	
	return nil
}

// Delete removes data from cache
func (c *memoryCache) Delete(key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	delete(c.data, key)
	return nil
}

// Clear removes all data from cache
func (c *memoryCache) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.data = make(map[string]cacheItem)
	return nil
}

// GetUser retrieves user from cache
func (c *memoryCache) GetUser(key string) (*entity.User, error) {
	data, err := c.Get(key)
	if err != nil {
		return nil, err
	}
	
	var user entity.User
	if err := json.Unmarshal(data, &user); err != nil {
		return nil, err
	}
	
	return &user, nil
}

// SetUser stores user in cache
func (c *memoryCache) SetUser(key string, user *entity.User) error {
	data, err := json.Marshal(user)
	if err != nil {
		return err
	}
	
	return c.Set(key, data, 15*time.Minute) // 15 minutes TTL
}

// GetPlayerInfo retrieves player info from cache
func (c *memoryCache) GetPlayerInfo(key string) (*entity.PlayerInfo, error) {
	data, err := c.Get(key)
	if err != nil {
		return nil, err
	}
	
	var playerInfo entity.PlayerInfo
	if err := json.Unmarshal(data, &playerInfo); err != nil {
		return nil, err
	}
	
	return &playerInfo, nil
}

// SetPlayerInfo stores player info in cache
func (c *memoryCache) SetPlayerInfo(key string, playerInfo *entity.PlayerInfo) error {
	data, err := json.Marshal(playerInfo)
	if err != nil {
		return err
	}
	
	return c.Set(key, data, 10*time.Minute) // 10 minutes TTL
}

// GetEquipment retrieves equipment from cache
func (c *memoryCache) GetEquipment(key string) ([]*entity.Equipment, error) {
	data, err := c.Get(key)
	if err != nil {
		return nil, err
	}
	
	var equipment []*entity.Equipment
	if err := json.Unmarshal(data, &equipment); err != nil {
		return nil, err
	}
	
	return equipment, nil
}

// SetEquipment stores equipment in cache
func (c *memoryCache) SetEquipment(key string, equipment []*entity.Equipment) error {
	data, err := json.Marshal(equipment)
	if err != nil {
		return err
	}
	
	return c.Set(key, data, 5*time.Minute) // 5 minutes TTL
}

// cleanup removes expired items
func (c *memoryCache) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, item := range c.data {
			if now.After(item.expiresAt) {
				delete(c.data, key)
			}
		}
		c.mu.Unlock()
	}
}

// Cache errors
var (
	ErrCacheKeyNotFound = &CacheError{"cache key not found"}
	ErrCacheKeyExpired  = &CacheError{"cache key expired"}
)

type CacheError struct {
	Message string
}

func (e *CacheError) Error() string {
	return e.Message
}