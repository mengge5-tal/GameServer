package service

import (
	"GameServer/internal/application/dto"
	"GameServer/internal/domain/entity"
	"GameServer/internal/domain/repository"
	"GameServer/internal/domain/service"
	"GameServer/internal/infrastructure/cache"
	"GameServer/internal/interfaces/websocket"
	"fmt"
	"sync"
	"time"
)

// UnionChatService 工会聊天服务
type UnionChatService struct {
	unionRepo     repository.UnionChatRepository
	userRepo      repository.UserRepository
	unionSvc      websocket.UnionServiceInterface // 用于验证用户工会关系
	rateLimiter   *service.ChatRateLimiter
	clientManager websocket.ClientManagerInterface // 用于消息广播
	
	// Redis缓存（优先使用）
	redisCache *cache.UnionChatCache
	
	// 内存缓存（Redis不可用时的备选方案）
	messageCache map[string][]*entity.UnionChatMessage
	cacheMutex   sync.RWMutex
	
	// 缓存配置
	maxCacheSize int
	cacheExpiry  time.Duration
	useRedis     bool // 是否使用Redis
}

// UnionChatConfig 工会聊天配置
type UnionChatConfig struct {
	MaxCacheSize int           // 每个工会最大缓存消息数
	CacheExpiry  time.Duration // 缓存过期时间
	UseRedis     bool          // 是否使用Redis缓存
	RedisConfig  *cache.RedisConfig // Redis配置
}

// DefaultUnionChatConfig 默认配置
func DefaultUnionChatConfig() *UnionChatConfig {
	return &UnionChatConfig{
		MaxCacheSize: 100,              // 每个工会缓存100条消息
		CacheExpiry:  30 * time.Minute, // 30分钟过期
		UseRedis:     false,            // 默认不使用Redis
		RedisConfig:  nil,              // 使用Redis默认配置
	}
}

// NewUnionChatService 创建工会聊天服务
func NewUnionChatService(
	unionRepo repository.UnionChatRepository,
	userRepo repository.UserRepository,
	unionSvc websocket.UnionServiceInterface,
	rateLimiter *service.ChatRateLimiter,
	clientManager websocket.ClientManagerInterface,
	config *UnionChatConfig,
) *UnionChatService {
	if config == nil {
		config = DefaultUnionChatConfig()
	}

	service := &UnionChatService{
		unionRepo:     unionRepo,
		userRepo:      userRepo,
		unionSvc:      unionSvc,
		rateLimiter:   rateLimiter,
		clientManager: clientManager,
		messageCache:  make(map[string][]*entity.UnionChatMessage),
		maxCacheSize:  config.MaxCacheSize,
		cacheExpiry:   config.CacheExpiry,
		useRedis:      config.UseRedis,
	}

	// 初始化Redis缓存
	if config.UseRedis {
		redisClient := cache.NewRedisClient(config.RedisConfig)
		
		// 测试Redis连接
		if err := redisClient.Ping(); err != nil {
			fmt.Printf("Redis连接失败，将使用内存缓存: %v\n", err)
			service.useRedis = false
		} else {
			service.redisCache = cache.NewUnionChatCache(redisClient)
			fmt.Println("Redis缓存已启用")
		}
	}

	// 启动缓存清理任务
	go service.startCacheCleanup()

	return service
}

// SendMessage 发送工会聊天消息
func (s *UnionChatService) SendMessage(userID int, request *dto.SendUnionMessageRequest) (*dto.UnionMessageResponse, error) {
	// 1. 验证请求参数
	if request.Content == "" {
		return nil, fmt.Errorf("消息内容不能为空")
	}

	// 2. 检查发送频率限制
	allowed, waitTime, err := s.rateLimiter.CheckUnionChat(userID)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, fmt.Errorf("工会聊天发送过于频繁，请等待%v后重试", waitTime)
	}

	// 3. 验证用户是否存在
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("用户不存在: %v", err)
	}

	// 4. 获取用户的工会信息
	unionInfo, err := s.unionSvc.GetMyUnionInfo(userID)
	if err != nil {
		return nil, fmt.Errorf("获取用户工会信息失败: %v", err)
	}
	if unionInfo == nil {
		return nil, fmt.Errorf("用户未加入工会")
	}

	// 5. 验证消息内容长度
	if len([]rune(request.Content)) > 30 {
		return nil, fmt.Errorf("消息内容不能超过30个字符")
	}

	// 6. 创建工会聊天消息
	message := &entity.UnionChatMessage{
		UnionID:   unionInfo.UnionID,
		UserID:    userID,
		Username:  user.Username,
		Content:   request.Content,
		CreatedAt: time.Now(),
	}

	// 7. 验证消息
	if err := message.Validate(); err != nil {
		return nil, err
	}

	// 8. 获取当前活跃的分表
	activeTable, err := s.unionRepo.GetActiveTable()
	if err != nil {
		return nil, fmt.Errorf("获取活跃分表失败: %v", err)
	}

	// 9. 保存消息到缓存（不立即写数据库）
	var needFlush bool
	if s.useRedis && s.redisCache != nil {
		// 使用Redis缓存
		if err := s.redisCache.AddMessage(message.UnionID, message); err != nil {
			fmt.Printf("Redis缓存添加消息失败: %v\n", err)
			// 降级到内存缓存
			cacheKey := message.GetCacheKey()
			needFlush = s.addToCache(cacheKey, message)
		} else {
			// 检查Redis缓存是否需要刷新到数据库
			messageCount, _ := s.redisCache.GetMessageCount(message.UnionID)
			if messageCount >= 100 {
				needFlush = true
			}
		}
	} else {
		// 使用内存缓存
		cacheKey := message.GetCacheKey()
		needFlush = s.addToCache(cacheKey, message)
	}

	// 10. 如果缓存已满，批量写入数据库
	if needFlush {
		go s.flushCacheToDatabase(message.UnionID, activeTable.TableName)
	}

	// 11. 构建响应
	response := &dto.UnionMessageResponse{
		UnionID:   message.UnionID,
		UserID:    message.UserID,
		Username:  message.Username,
		Content:   message.Content,
		Timestamp: message.CreatedAt.Format("2006-01-02 15:04:05"),
		MessageID: s.generateMessageID(userID, unionInfo.UnionID),
	}

	// 12. 更新用户在线状态到Redis缓存
	if s.useRedis && s.redisCache != nil {
		if err := s.redisCache.SetUserOnline(userID, unionInfo.UnionID); err != nil {
			fmt.Printf("更新用户在线状态失败: %v\n", err)
		}
	}

	return response, nil
}

// GetMessages 获取工会聊天消息
func (s *UnionChatService) GetMessages(userID int, request *dto.GetUnionMessagesRequest) (*dto.GetUnionMessagesResponse, error) {
	// 1. 验证用户是否存在
	_, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("用户不存在: %v", err)
	}

	// 2. 获取用户的工会信息
	unionInfo, err := s.unionSvc.GetMyUnionInfo(userID)
	if err != nil {
		return nil, fmt.Errorf("获取用户工会信息失败: %v", err)
	}
	if unionInfo == nil {
		return nil, fmt.Errorf("用户未加入工会")
	}

	// 3. 设置默认分页参数
	page := request.Page
	limit := request.Limit
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 50 {
		limit = 20 // 默认每页20条，最大50条
	}

	var messages []*entity.UnionChatMessage
	var total int
	var useCache bool

	// 4. 首先尝试从缓存获取

	if s.useRedis && s.redisCache != nil {
		// 使用Redis缓存
		if cachedMessages, err := s.redisCache.GetMessages(unionInfo.UnionID, limit); err == nil && len(cachedMessages) > 0 {
			if page == 1 {
				// 第一页优先使用缓存数据
				messages = cachedMessages
				if len(messages) > limit {
					messages = messages[:limit]
				}
				total = len(cachedMessages)
				useCache = true
			}
		}
	} else {
		// 使用内存缓存
		tempMessage := &entity.UnionChatMessage{UnionID: unionInfo.UnionID}
		cacheKey := tempMessage.GetCacheKey()
		cachedMessages := s.getFromCache(cacheKey)

		if len(cachedMessages) > 0 && page == 1 {
			// 第一页且缓存中有消息，直接使用缓存
			messages = cachedMessages
			if len(messages) > limit {
				messages = messages[:limit]
			}
			total = len(cachedMessages)
			useCache = true
		}
	}

	if !useCache {
		// 从数据库获取
		var tableName string
		if request.YearMonth != "" {
			// 指定月份的消息
			table, err := s.unionRepo.GetTableByYearMonth(request.YearMonth)
			if err != nil {
				return nil, fmt.Errorf("获取指定月份分表失败: %v", err)
			}
			tableName = table.TableName
		} else {
			// 当前月份的消息
			activeTable, err := s.unionRepo.GetActiveTable()
			if err != nil {
				return nil, fmt.Errorf("获取活跃分表失败: %v", err)
			}
			tableName = activeTable.TableName
		}

		messages, total, err = s.unionRepo.GetMessagesByUnionID(tableName, unionInfo.UnionID, page, limit)
		if err != nil {
			return nil, fmt.Errorf("获取工会消息失败: %v", err)
		}
	}

	// 5. 构建响应消息列表
	var messageResponses []dto.UnionMessageResponse
	for _, message := range messages {
		messageResponse := dto.UnionMessageResponse{
			UnionID:   message.UnionID,
			UserID:    message.UserID,
			Username:  message.Username,
			Content:   message.Content,
			Timestamp: message.CreatedAt.Format("2006-01-02 15:04:05"),
			MessageID: s.generateMessageID(message.UserID, message.UnionID),
		}
		messageResponses = append(messageResponses, messageResponse)
	}

	// 计算是否还有更多消息
	totalPages := (total + limit - 1) / limit
	hasMore := page < totalPages

	response := &dto.GetUnionMessagesResponse{
		Messages:  messageResponses,
		Total:     total,
		HasMore:   hasMore,
		UnionID:   unionInfo.UnionID,
		UnionName: unionInfo.UnionName,
	}

	return response, nil
}

// GetRecentMessages 获取最近的工会聊天消息
func (s *UnionChatService) GetRecentMessages(userID int, limit int) (*dto.GetUnionMessagesResponse, error) {
	// 1. 验证参数
	if limit <= 0 || limit > 100 {
		limit = 50 // 默认50条，最大100条
	}

	// 2. 验证用户是否存在
	_, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("用户不存在: %v", err)
	}

	// 3. 获取用户的工会信息
	unionInfo, err := s.unionSvc.GetMyUnionInfo(userID)
	if err != nil {
		return nil, fmt.Errorf("获取用户工会信息失败: %v", err)
	}
	if unionInfo == nil {
		return nil, fmt.Errorf("用户未加入工会")
	}

	var messages []*entity.UnionChatMessage

	// 4. 首先尝试从Redis缓存获取
	if s.useRedis && s.redisCache != nil {
		if cachedMessages, err := s.redisCache.GetMessages(unionInfo.UnionID, limit); err == nil && len(cachedMessages) > 0 {
			// 如果缓存中有消息，优先使用缓存
			messages = cachedMessages
			if len(messages) > limit {
				messages = messages[:limit]
			}
		} else {
			// Redis获取失败，降级到内存缓存
			tempMessage := &entity.UnionChatMessage{UnionID: unionInfo.UnionID}
			cacheKey := tempMessage.GetCacheKey()
			cachedMessages := s.getFromCache(cacheKey)

			if len(cachedMessages) > 0 {
				// 如果内存缓存中有消息，优先使用缓存
				messages = cachedMessages
				if len(messages) > limit {
					messages = messages[:limit]
				}
			} else {
				// 从数据库获取
				activeTable, err := s.unionRepo.GetActiveTable()
				if err != nil {
					return nil, fmt.Errorf("获取活跃分表失败: %v", err)
				}

				messages, err = s.unionRepo.GetRecentMessages(activeTable.TableName, unionInfo.UnionID, limit)
				if err != nil {
					return nil, fmt.Errorf("获取最近工会消息失败: %v", err)
				}
			}
		}
	} else {
		// 使用内存缓存
		tempMessage := &entity.UnionChatMessage{UnionID: unionInfo.UnionID}
		cacheKey := tempMessage.GetCacheKey()
		cachedMessages := s.getFromCache(cacheKey)

		if len(cachedMessages) > 0 {
			// 如果缓存中有消息，优先使用缓存
			messages = cachedMessages
			if len(messages) > limit {
				messages = messages[:limit]
			}
		} else {
			// 从数据库获取
			activeTable, err := s.unionRepo.GetActiveTable()
			if err != nil {
				return nil, fmt.Errorf("获取活跃分表失败: %v", err)
			}

			messages, err = s.unionRepo.GetRecentMessages(activeTable.TableName, unionInfo.UnionID, limit)
			if err != nil {
				return nil, fmt.Errorf("获取最近工会消息失败: %v", err)
			}
		}
	}

	// 5. 构建响应消息列表
	var messageResponses []dto.UnionMessageResponse
	for _, message := range messages {
		messageResponse := dto.UnionMessageResponse{
			UnionID:   message.UnionID,
			UserID:    message.UserID,
			Username:  message.Username,
			Content:   message.Content,
			Timestamp: message.CreatedAt.Format("2006-01-02 15:04:05"),
			MessageID: s.generateMessageID(message.UserID, message.UnionID),
		}
		messageResponses = append(messageResponses, messageResponse)
	}

	response := &dto.GetUnionMessagesResponse{
		Messages:  messageResponses,
		Total:     len(messageResponses),
		HasMore:   false, // 最近消息一次性返回
		UnionID:   unionInfo.UnionID,
		UnionName: unionInfo.UnionName,
	}

	return response, nil
}

// BroadcastToUnion 向工会广播消息
func (s *UnionChatService) BroadcastToUnion(unionID int, message *dto.UnionMessageResponse) error {
	// 使用ClientManager进行真实的消息广播
	if s.clientManager != nil {
		successCount := s.clientManager.BroadcastToUnion(unionID, message, message.UserID)
		fmt.Printf("工会 %d 消息广播成功发送给 %d 个用户\n", unionID, successCount)
	} else {
		fmt.Printf("向工会 %d 广播消息: %s (来自用户 %s)\n", unionID, message.Content, message.Username)
	}
	
	// 更新用户在线状态到Redis
	if s.useRedis && s.redisCache != nil {
		if err := s.redisCache.SetUserOnline(message.UserID, unionID); err != nil {
			fmt.Printf("更新用户在线状态失败: %v\n", err)
		}
	}
	
	return nil
}

// 内部方法

// addToCache 添加消息到缓存，返回是否需要刷新到数据库
func (s *UnionChatService) addToCache(cacheKey string, message *entity.UnionChatMessage) bool {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()

	// 获取现有缓存
	messages, exists := s.messageCache[cacheKey]
	if !exists {
		messages = make([]*entity.UnionChatMessage, 0, s.maxCacheSize)
	}

	// 添加新消息到开头
	messages = append([]*entity.UnionChatMessage{message}, messages...)

	// 检查是否达到缓存上限
	needFlush := len(messages) >= s.maxCacheSize

	// 限制缓存大小
	if len(messages) > s.maxCacheSize {
		messages = messages[:s.maxCacheSize]
	}

	// 更新缓存
	s.messageCache[cacheKey] = messages
	
	return needFlush
}

// getFromCache 从缓存获取消息
func (s *UnionChatService) getFromCache(cacheKey string) []*entity.UnionChatMessage {
	s.cacheMutex.RLock()
	defer s.cacheMutex.RUnlock()

	messages, exists := s.messageCache[cacheKey]
	if !exists {
		return nil
	}

	// 复制切片避免并发修改问题
	result := make([]*entity.UnionChatMessage, len(messages))
	copy(result, messages)
	
	return result
}

// flushCacheToDatabase 批量刷新缓存到数据库
func (s *UnionChatService) flushCacheToDatabase(unionID int, tableName string) {
	var messages []*entity.UnionChatMessage
	
	// 从缓存获取消息
	if s.useRedis && s.redisCache != nil {
		// Redis缓存
		cachedMessages, err := s.redisCache.GetMessages(unionID, s.maxCacheSize)
		if err != nil {
			fmt.Printf("获取Redis缓存失败: %v\n", err)
			return
		}
		messages = cachedMessages
		
		// 清空Redis缓存
		if err := s.redisCache.ClearMessages(unionID); err != nil {
			fmt.Printf("清空Redis缓存失败: %v\n", err)
		}
	} else {
		// 内存缓存
		tempMessage := &entity.UnionChatMessage{UnionID: unionID}
		cacheKey := tempMessage.GetCacheKey()
		
		s.cacheMutex.Lock()
		cachedMessages, exists := s.messageCache[cacheKey]
		if exists {
			// 复制消息列表
			messages = make([]*entity.UnionChatMessage, len(cachedMessages))
			copy(messages, cachedMessages)
			// 清空缓存
			delete(s.messageCache, cacheKey)
		}
		s.cacheMutex.Unlock()
	}
	
	if len(messages) == 0 {
		return
	}
	
	// 批量写入数据库
	if err := s.unionRepo.BatchCreateMessages(tableName, messages); err != nil {
		fmt.Printf("批量写入工会消息失败: %v，消息数量: %d\n", err, len(messages))
	} else {
		fmt.Printf("成功批量写入 %d 条工会消息到数据库\n", len(messages))
	}
}

// generateMessageID 生成消息ID
func (s *UnionChatService) generateMessageID(userID, unionID int) string {
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("union_%d_%d_%d", unionID, userID, timestamp)
}

// startCacheCleanup 启动缓存清理和定时写入任务
func (s *UnionChatService) startCacheCleanup() {
	ticker := time.NewTicker(s.cacheExpiry)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.cleanupExpiredCache()
		}
	}
}

// cleanupExpiredCache 清理过期缓存并批量写入数据库
func (s *UnionChatService) cleanupExpiredCache() {
	// 获取活跃分表信息
	activeTable, err := s.unionRepo.GetActiveTable()
	if err != nil {
		fmt.Printf("获取活跃分表失败，跳过缓存清理: %v\n", err)
		return
	}

	if s.useRedis && s.redisCache != nil {
		// Redis缓存：检查需要批量写入的工会
		s.handleRedisCacheFlush(activeTable.TableName)
		return
	}

	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()

	now := time.Now()
	for cacheKey, messages := range s.messageCache {
		// 分离过期和有效的消息
		var expiredMessages []*entity.UnionChatMessage
		var validMessages []*entity.UnionChatMessage
		
		for _, message := range messages {
			if now.Sub(message.CreatedAt) >= s.cacheExpiry {
				expiredMessages = append(expiredMessages, message)
			} else {
				validMessages = append(validMessages, message)
			}
		}

		// 如果有过期消息，批量写入数据库
		if len(expiredMessages) > 0 {
			go func(tableName string, msgs []*entity.UnionChatMessage) {
				if err := s.unionRepo.BatchCreateMessages(tableName, msgs); err != nil {
					fmt.Printf("批量写入过期消息失败: %v，消息数量: %d\n", err, len(msgs))
				} else {
					fmt.Printf("成功批量写入 %d 条过期工会消息到数据库\n", len(msgs))
				}
			}(activeTable.TableName, expiredMessages)
		}

		if len(validMessages) == 0 {
			// 删除空缓存
			delete(s.messageCache, cacheKey)
		} else {
			// 更新缓存（只保留未过期的消息）
			s.messageCache[cacheKey] = validMessages
		}
	}
}

// GetCacheStats 获取缓存统计信息
func (s *UnionChatService) GetCacheStats() map[string]interface{} {
	s.cacheMutex.RLock()
	defer s.cacheMutex.RUnlock()

	totalMessages := 0
	for _, messages := range s.messageCache {
		totalMessages += len(messages)
	}

	return map[string]interface{}{
		"total_unions":    len(s.messageCache),
		"total_messages":  totalMessages,
		"max_cache_size":  s.maxCacheSize,
		"cache_expiry":    s.cacheExpiry.String(),
		"memory_usage":    fmt.Sprintf("%.2f KB", float64(totalMessages*200)/1024), // 估算内存使用
	}
}

// FlushCache 清空指定工会的缓存（管理员操作）
func (s *UnionChatService) FlushCache(unionID int) error {
	// 清理Redis缓存
	if s.useRedis && s.redisCache != nil {
		if err := s.redisCache.ClearMessages(unionID); err != nil {
			fmt.Printf("清理Redis缓存失败: %v\n", err)
		}
	}

	// 清理内存缓存
	tempMessage := &entity.UnionChatMessage{UnionID: unionID}
	cacheKey := tempMessage.GetCacheKey()
	
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()
	
	delete(s.messageCache, cacheKey)
	return nil
}

// FlushAllCache 清空所有缓存（管理员操作）
func (s *UnionChatService) FlushAllCache() error {
	// 清理Redis缓存（这里只清理内存，Redis清理需要额外实现）
	if s.useRedis && s.redisCache != nil {
		fmt.Println("注意：Redis缓存清理需要在Redis层面实现")
	}

	// 清理内存缓存
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()
	
	s.messageCache = make(map[string][]*entity.UnionChatMessage)
	return nil
}

// PreloadCacheFromDB 从数据库预加载缓存（系统启动时调用）
func (s *UnionChatService) PreloadCacheFromDB(unionID int, limit int) error {
	if limit <= 0 || limit > s.maxCacheSize {
		limit = s.maxCacheSize
	}

	// 获取活跃分表
	activeTable, err := s.unionRepo.GetActiveTable()
	if err != nil {
		return fmt.Errorf("获取活跃分表失败: %v", err)
	}

	// 从数据库获取最近的消息
	messages, err := s.unionRepo.GetRecentMessages(activeTable.TableName, unionID, limit)
	if err != nil {
		return fmt.Errorf("从数据库预加载消息失败: %v", err)
	}

	// 优先添加到Redis缓存
	if s.useRedis && s.redisCache != nil {
		if err := s.redisCache.BatchAddMessages(unionID, messages); err != nil {
			fmt.Printf("Redis缓存预加载失败: %v\n", err)
			// 降级到内存缓存
			tempMessage := &entity.UnionChatMessage{UnionID: unionID}
			cacheKey := tempMessage.GetCacheKey()
			s.cacheMutex.Lock()
			s.messageCache[cacheKey] = messages
			s.cacheMutex.Unlock()
		}
	} else {
		// 使用内存缓存
		tempMessage := &entity.UnionChatMessage{UnionID: unionID}
		cacheKey := tempMessage.GetCacheKey()
		s.cacheMutex.Lock()
		s.messageCache[cacheKey] = messages
		s.cacheMutex.Unlock()
	}

	return nil
}

// SetRedisCache 设置Redis缓存
func (s *UnionChatService) SetRedisCache(redisCache *cache.UnionChatCache) {
	s.redisCache = redisCache
	s.useRedis = redisCache != nil
}

// handleRedisCacheFlush 处理Redis缓存的批量写入
func (s *UnionChatService) handleRedisCacheFlush(tableName string) {
	// 获取需要批量写入的工会列表
	unionIDs, err := s.redisCache.GetPendingFlushUnions()
	if err != nil {
		fmt.Printf("获取待刷新工会列表失败: %v\n", err)
		return
	}

	for _, unionID := range unionIDs {
		// 检查缓存是否达到写入条件
		messageCount, err := s.redisCache.GetMessageCount(unionID)
		if err != nil {
			continue
		}

		// 获取缓存中的消息
		messages, err := s.redisCache.GetMessages(unionID, int(messageCount))
		if err != nil {
			continue
		}

		// 检查是否需要批量写入（100条消息或有消息且30分钟过期）
		shouldFlush := false
		if len(messages) >= 100 {
			shouldFlush = true
		} else if len(messages) > 0 {
			// 检查最老的消息是否超过30分钟
			oldestMessage := messages[len(messages)-1]
			if time.Since(oldestMessage.CreatedAt) >= 30*time.Minute {
				shouldFlush = true
			}
		}

		if shouldFlush {
			go func(uid int, msgs []*entity.UnionChatMessage) {
				// 批量写入数据库
				if err := s.unionRepo.BatchCreateMessages(tableName, msgs); err != nil {
					fmt.Printf("Redis缓存批量写入失败: %v，工会ID: %d，消息数量: %d\n", err, uid, len(msgs))
				} else {
					fmt.Printf("成功从Redis批量写入 %d 条工会消息到数据库，工会ID: %d\n", len(msgs), uid)
					// 清除缓存和批量写入标记
					if err := s.redisCache.ClearMessages(uid); err != nil {
						fmt.Printf("清除Redis缓存失败: %v\n", err)
					}
					if err := s.redisCache.ClearFlushMark(uid); err != nil {
						fmt.Printf("清除批量写入标记失败: %v\n", err)
					}
				}
			}(unionID, messages)
		}
	}
}

// SetClientManager 设置客户端管理器
func (s *UnionChatService) SetClientManager(clientManager websocket.ClientManagerInterface) {
	s.clientManager = clientManager
}