package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisClient Redis客户端包装器
type RedisClient struct {
	client *redis.Client
	ctx    context.Context
}

// RedisConfig Redis配置
type RedisConfig struct {
	Addr     string `json:"addr"`     // Redis地址，默认 localhost:6379
	Password string `json:"password"` // Redis密码，默认为空
	DB       int    `json:"db"`       // Redis数据库，默认为0
}

// NewRedisClient 创建Redis客户端
func NewRedisClient(config *RedisConfig) *RedisClient {
	if config == nil {
		config = &RedisConfig{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		}
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       config.DB,
	})

	return &RedisClient{
		client: rdb,
		ctx:    context.Background(),
	}
}

// Ping 测试Redis连接
func (r *RedisClient) Ping() error {
	return r.client.Ping(r.ctx).Err()
}

// Close 关闭Redis连接
func (r *RedisClient) Close() error {
	return r.client.Close()
}

// ========== 基础操作 ==========

// Set 设置键值对
func (r *RedisClient) Set(key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(r.ctx, key, value, expiration).Err()
}

// Get 获取键对应的值
func (r *RedisClient) Get(key string) (string, error) {
	return r.client.Get(r.ctx, key).Result()
}

// Del 删除键
func (r *RedisClient) Del(keys ...string) error {
	return r.client.Del(r.ctx, keys...).Err()
}

// Exists 检查键是否存在
func (r *RedisClient) Exists(keys ...string) (int64, error) {
	return r.client.Exists(r.ctx, keys...).Result()
}

// Expire 设置键过期时间
func (r *RedisClient) Expire(key string, expiration time.Duration) error {
	return r.client.Expire(r.ctx, key, expiration).Err()
}

// TTL 获取键的剩余过期时间
func (r *RedisClient) TTL(key string) (time.Duration, error) {
	return r.client.TTL(r.ctx, key).Result()
}

// ========== JSON操作 ==========

// SetJSON 设置JSON数据
func (r *RedisClient) SetJSON(key string, value interface{}, expiration time.Duration) error {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("JSON序列化失败: %v", err)
	}
	return r.client.Set(r.ctx, key, jsonData, expiration).Err()
}

// GetJSON 获取JSON数据
func (r *RedisClient) GetJSON(key string, dest interface{}) error {
	jsonData, err := r.client.Get(r.ctx, key).Result()
	if err != nil {
		return err
	}
	
	return json.Unmarshal([]byte(jsonData), dest)
}

// ========== 列表操作 ==========

// LPush 从左侧插入列表
func (r *RedisClient) LPush(key string, values ...interface{}) error {
	return r.client.LPush(r.ctx, key, values...).Err()
}

// RPush 从右侧插入列表
func (r *RedisClient) RPush(key string, values ...interface{}) error {
	return r.client.RPush(r.ctx, key, values...).Err()
}

// LPop 从左侧弹出元素
func (r *RedisClient) LPop(key string) (string, error) {
	return r.client.LPop(r.ctx, key).Result()
}

// RPop 从右侧弹出元素
func (r *RedisClient) RPop(key string) (string, error) {
	return r.client.RPop(r.ctx, key).Result()
}

// LRange 获取列表范围内的元素
func (r *RedisClient) LRange(key string, start, stop int64) ([]string, error) {
	return r.client.LRange(r.ctx, key, start, stop).Result()
}

// LLen 获取列表长度
func (r *RedisClient) LLen(key string) (int64, error) {
	return r.client.LLen(r.ctx, key).Result()
}

// LTrim 修剪列表
func (r *RedisClient) LTrim(key string, start, stop int64) error {
	return r.client.LTrim(r.ctx, key, start, stop).Err()
}

// ========== 哈希操作 ==========

// HSet 设置哈希字段
func (r *RedisClient) HSet(key string, values ...interface{}) error {
	return r.client.HSet(r.ctx, key, values...).Err()
}

// HGet 获取哈希字段
func (r *RedisClient) HGet(key, field string) (string, error) {
	return r.client.HGet(r.ctx, key, field).Result()
}

// HGetAll 获取哈希所有字段
func (r *RedisClient) HGetAll(key string) (map[string]string, error) {
	return r.client.HGetAll(r.ctx, key).Result()
}

// HDel 删除哈希字段
func (r *RedisClient) HDel(key string, fields ...string) error {
	return r.client.HDel(r.ctx, key, fields...).Err()
}

// HExists 检查哈希字段是否存在
func (r *RedisClient) HExists(key, field string) (bool, error) {
	return r.client.HExists(r.ctx, key, field).Result()
}

// ========== 集合操作 ==========

// SAdd 添加集合成员
func (r *RedisClient) SAdd(key string, members ...interface{}) error {
	return r.client.SAdd(r.ctx, key, members...).Err()
}

// SMembers 获取集合所有成员
func (r *RedisClient) SMembers(key string) ([]string, error) {
	return r.client.SMembers(r.ctx, key).Result()
}

// SIsMember 检查是否是集合成员
func (r *RedisClient) SIsMember(key string, member interface{}) (bool, error) {
	return r.client.SIsMember(r.ctx, key, member).Result()
}

// SRem 移除集合成员
func (r *RedisClient) SRem(key string, members ...interface{}) error {
	return r.client.SRem(r.ctx, key, members...).Err()
}

// SCard 获取集合成员数量
func (r *RedisClient) SCard(key string) (int64, error) {
	return r.client.SCard(r.ctx, key).Result()
}

// ========== 有序集合操作 ==========

// ZAdd 添加有序集合成员
func (r *RedisClient) ZAdd(key string, members ...redis.Z) error {
	return r.client.ZAdd(r.ctx, key, members...).Err()
}

// ZRange 按索引范围获取有序集合成员
func (r *RedisClient) ZRange(key string, start, stop int64) ([]string, error) {
	return r.client.ZRange(r.ctx, key, start, stop).Result()
}

// ZRangeWithScores 按索引范围获取有序集合成员及分数
func (r *RedisClient) ZRangeWithScores(key string, start, stop int64) ([]redis.Z, error) {
	return r.client.ZRangeWithScores(r.ctx, key, start, stop).Result()
}

// ZRem 移除有序集合成员
func (r *RedisClient) ZRem(key string, members ...interface{}) error {
	return r.client.ZRem(r.ctx, key, members...).Err()
}

// ZCard 获取有序集合成员数量
func (r *RedisClient) ZCard(key string) (int64, error) {
	return r.client.ZCard(r.ctx, key).Result()
}

// ========== 批量操作 ==========

// Pipeline 创建管道
func (r *RedisClient) Pipeline() redis.Pipeliner {
	return r.client.Pipeline()
}

// TxPipeline 创建事务管道
func (r *RedisClient) TxPipeline() redis.Pipeliner {
	return r.client.TxPipeline()
}

// ========== 工具方法 ==========

// Keys 获取匹配模式的所有键
func (r *RedisClient) Keys(pattern string) ([]string, error) {
	return r.client.Keys(r.ctx, pattern).Result()
}

// FlushDB 清空当前数据库
func (r *RedisClient) FlushDB() error {
	return r.client.FlushDB(r.ctx).Err()
}

// Info 获取Redis服务器信息
func (r *RedisClient) Info(sections ...string) (string, error) {
	if len(sections) == 0 {
		return r.client.Info(r.ctx).Result()
	}
	return r.client.Info(r.ctx, sections...).Result()
}

// GetClient 获取原始Redis客户端（用于高级操作）
func (r *RedisClient) GetClient() *redis.Client {
	return r.client
}

// GetContext 获取上下文
func (r *RedisClient) GetContext() context.Context {
	return r.ctx
}