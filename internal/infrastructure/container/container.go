package container

import (
	"GameServer/internal/application/service"
	"GameServer/internal/domain/repository"
	domainService "GameServer/internal/domain/service"
	"GameServer/internal/infrastructure/cache"
	"GameServer/internal/infrastructure/config"
	infraRepo "GameServer/internal/infrastructure/repository"
	"GameServer/internal/interfaces/websocket"
	"database/sql"
	"fmt"
	"time"
)

// Container holds all application dependencies
type Container struct {
	Config   *config.Config
	Database *sql.DB

	// Services
	CacheService           cache.CacheService
	AuthService            *service.AuthService
	PlayerService          *service.PlayerService
	FriendService          *service.FriendService
	NotificationService    service.NotificationService
	UserEquipService       *service.UserEquipService
	ExperienceService      *service.ExperienceService
	WeaponService          *service.WeaponService
	UserWeaponService      *service.UserWeaponService
	SourceStoneService     *service.SourceStoneService
	UserSourceStoneService *service.UserSourceStoneService
	KillCountService       *service.KillCountService
	RankingService         *service.RankingService
	UnionService           *service.UnionService
	
	// Chat Services
	PrivateChatService   *service.PrivateChatService
	WorldChatService     *service.WorldChatService
	UnionChatService     *service.UnionChatService
	ChatScheduledService *domainService.ChatScheduledService
	
	// Cache Services
	RedisClient    *cache.RedisClient
	UnionChatCache *cache.UnionChatCache

	// Repositories
	UserRepo            repository.UserRepository
	PlayerRepo          repository.PlayerRepository
	FriendRepo          repository.FriendRepository
	EquipmentRepo       repository.EquipmentRepository
	SourceStoneRepo     repository.SourceStoneRepository
	ExperienceRepo      repository.ExperienceRepository
	UserEquipRepo       repository.UserEquipRepository
	WeaponRepo          repository.WeaponRepository
	UserWeaponRepo      repository.UserWeaponRepository
	UserSourceStoneRepo     repository.UserSourceStoneRepository
	KillCountRepo           repository.KillCountRepository
	UnionRepo               repository.UnionRepository
	UnionMemberRepo         repository.UnionMemberRepository
	UnionRequestRepo        repository.UnionRequestRepository
	UnionExperienceRepo     repository.UnionExperienceRepository
	UnionInviteRepo         repository.UnionInviteRepository
	
	// Chat Repositories
	PrivateChatRepo repository.PrivateMessageRepository
	WorldChatRepo   repository.WorldChatRepository
	UnionChatRepo   repository.UnionChatRepository

	// Domain Services
	AuthDomainService domainService.AuthDomainService
	ChatRateLimiter   *domainService.ChatRateLimiter
}

// NewContainer creates and initializes a new dependency injection container
func NewContainer(cfg *config.Config, db *sql.DB) (*Container, error) {
	container := &Container{
		Config:   cfg,
		Database: db,
	}

	// Initialize dependencies
	if err := container.initializeRepositories(); err != nil {
		return nil, err
	}

	if err := container.initializeServices(); err != nil {
		return nil, err
	}

	return container, nil
}

// initializeRepositories initializes all repository implementations
func (c *Container) initializeRepositories() error {
	c.UserRepo = infraRepo.NewMySQLUserRepository(c.Database)
	c.PlayerRepo = infraRepo.NewMySQLPlayerRepository(c.Database)
	c.FriendRepo = infraRepo.NewMySQLFriendRepository(c.Database)
	c.EquipmentRepo = infraRepo.NewMySQLEquipmentRepository(c.Database)
	c.SourceStoneRepo = infraRepo.NewMySQLSourceStoneRepository(c.Database)
	c.ExperienceRepo = infraRepo.NewMySQLExperienceRepository(c.Database)
	c.UserEquipRepo = infraRepo.NewMySQLUserEquipRepository(c.Database)
	c.WeaponRepo = infraRepo.NewMySQLWeaponRepository(c.Database)
	c.UserWeaponRepo = infraRepo.NewMySQLUserWeaponRepository(c.Database)
	c.UserSourceStoneRepo = infraRepo.NewMySQLUserSourceStoneRepository(c.Database)
	c.KillCountRepo = infraRepo.NewMySQLKillCountRepository(c.Database)
	
	// Union repositories
	c.UnionRepo = infraRepo.NewMySQLUnionRepository(c.Database)
	c.UnionMemberRepo = infraRepo.NewMySQLUnionMemberRepository(c.Database)
	c.UnionRequestRepo = infraRepo.NewMySQLUnionRequestRepository(c.Database)
	c.UnionExperienceRepo = infraRepo.NewMySQLUnionExperienceRepository(c.Database)
	c.UnionInviteRepo = infraRepo.NewMySQLUnionInviteRepository(c.Database)
	
	// Chat repositories
	c.PrivateChatRepo = infraRepo.NewMySQLChatRepository(c.Database)
	c.WorldChatRepo = infraRepo.NewMySQLChatRepository(c.Database)
	c.UnionChatRepo = infraRepo.NewMySQLChatRepository(c.Database)

	return nil
}

// initializeServices initializes all application services
func (c *Container) initializeServices() error {
	// Initialize cache service
	c.CacheService = cache.NewMemoryCache()

	// Initialize domain services
	c.AuthDomainService = domainService.NewAuthDomainService(c.Config.Security.BcryptCost)
	c.ChatRateLimiter = domainService.NewChatRateLimiter()

	// Initialize application services
	c.AuthService = service.NewAuthService(
		c.UserRepo,
		c.PlayerRepo,
		c.AuthDomainService,
		c.CacheService,
	)

	c.PlayerService = service.NewPlayerService(
		c.PlayerRepo,
		c.EquipmentRepo,
		c.CacheService,
	)

	// Note: NotificationService will be initialized later with the WebSocket hub
	c.FriendService = service.NewFriendService(
		c.FriendRepo,
		c.UserRepo,
		c.PlayerRepo,
		nil, // Will be set later when we have the WebSocket hub
	)

	c.UserEquipService = service.NewUserEquipService(
		c.UserEquipRepo,
		c.EquipmentRepo,
		c.UserRepo,
	)

	c.ExperienceService = service.NewExperienceService(
		c.ExperienceRepo,
	)

	c.WeaponService = service.NewWeaponService(
		c.WeaponRepo,
	)

	c.UserWeaponService = service.NewUserWeaponService(
		c.UserWeaponRepo,
		c.WeaponRepo,
		c.UserRepo,
	)

	c.SourceStoneService = service.NewSourceStoneService(
		c.SourceStoneRepo,
	)

	c.UserSourceStoneService = service.NewUserSourceStoneService(
		c.UserSourceStoneRepo,
		c.SourceStoneRepo,
		c.UserRepo,
	)

	c.KillCountService = service.NewKillCountService(
		c.KillCountRepo,
	)

	c.RankingService = service.NewRankingService(
		c.PlayerRepo,
	)

	// UnionService will be initialized later with NotificationService
	c.UnionService = service.NewUnionService(
		c.UnionRepo,
		c.UnionMemberRepo,
		c.UnionRequestRepo,
		c.UnionExperienceRepo,
		c.PlayerRepo,
		c.UserRepo,
		c.CacheService,
		nil, // Will be set later when we have the WebSocket hub
		c.UnionInviteRepo,
	)
	
	// Initialize cache services
	c.initializeCacheServices()
	
	// Initialize chat services (after cache services)
	c.initializeChatServices()

	return nil
}

// GetWebSocketServices returns services formatted for WebSocket handlers
func (c *Container) GetWebSocketServices() *websocket.ServiceContainer {
	return &websocket.ServiceContainer{
		AuthService:            c.AuthService,
		PlayerService:          c.PlayerService,
		FriendService:          c.FriendService,
		UserEquipService:       c.UserEquipService,
		ExperienceService:      c.ExperienceService,
		WeaponService:          c.WeaponService,
		UserWeaponService:      c.UserWeaponService,
		SourceStoneService:     c.SourceStoneService,
		UserSourceStoneService: c.UserSourceStoneService,
		KillCountService:       c.KillCountService,
		RankingService:         c.RankingService,
		UnionService:           c.UnionService,
		
		// Chat Services
		PrivateChatService: c.PrivateChatService,
		WorldChatService:   c.WorldChatService,
		UnionChatService:   c.UnionChatService,
	}
}

// SetNotificationService sets the notification service after WebSocket hub is available
func (c *Container) SetNotificationService(userNotifier service.UserNotifier) {
	c.NotificationService = service.NewNotificationService(userNotifier)

	// Recreate FriendService with notification support
	c.FriendService = service.NewFriendService(
		c.FriendRepo,
		c.UserRepo,
		c.PlayerRepo,
		c.NotificationService,
	)

	// Recreate UnionService with notification support
	c.UnionService = service.NewUnionService(
		c.UnionRepo,
		c.UnionMemberRepo,
		c.UnionRequestRepo,
		c.UnionExperienceRepo,
		c.PlayerRepo,
		c.UserRepo,
		c.CacheService,
		c.NotificationService,
		c.UnionInviteRepo,
	)
}

// initializeCacheServices 初始化缓存服务
func (c *Container) initializeCacheServices() {
	// 初始化Redis客户端（如果配置了Redis）
	if c.Config.Redis != nil && c.Config.Redis.Enabled {
		c.RedisClient = cache.NewRedisClient(&cache.RedisConfig{
			Addr:     fmt.Sprintf("%s:%d", c.Config.Redis.Host, c.Config.Redis.Port),
			Password: c.Config.Redis.Password,
			DB:       c.Config.Redis.DB,
		})
		
		// 初始化工会聊天Redis缓存
		c.UnionChatCache = cache.NewUnionChatCache(c.RedisClient)
	}
}

// initializeChatServices 初始化聊天服务
func (c *Container) initializeChatServices() {
	// 私聊服务
	c.PrivateChatService = service.NewPrivateChatService(
		c.PrivateChatRepo,
		c.UserRepo,
		c.ChatRateLimiter,
	)
	
	// 世界聊天服务（暂时不传ClientManager，稍后设置）
	c.WorldChatService = service.NewWorldChatService(
		c.WorldChatRepo,
		c.UserRepo,
		c.ChatRateLimiter,
		nil, // ClientManager will be set later
	)
	
	// 工会聊天服务配置
	unionChatConfig := &service.UnionChatConfig{
		MaxCacheSize: 100,
		CacheExpiry:  30 * time.Minute,
		UseRedis:     c.RedisClient != nil,
	}
	
	// 工会聊天服务
	c.UnionChatService = service.NewUnionChatService(
		c.UnionChatRepo,
		c.UserRepo,
		c.UnionService,
		c.ChatRateLimiter,
		nil, // ClientManager will be set later
		unionChatConfig,
	)
	
	// 如果启用了Redis，设置Redis缓存
	if c.UnionChatCache != nil {
		c.UnionChatService.SetRedisCache(c.UnionChatCache)
	}
	
	// 聊天定时任务服务
	c.ChatScheduledService = domainService.NewChatScheduledService(c.UnionChatCache)
}

// SetChatClientManager 设置聊天服务的ClientManager
func (c *Container) SetChatClientManager(clientManager websocket.ClientManagerInterface) {
	if c.WorldChatService != nil {
		c.WorldChatService.SetClientManager(clientManager)
	}
	
	if c.UnionChatService != nil {
		c.UnionChatService.SetClientManager(clientManager)
	}
}

// StartChatCleanupTasks 启动聊天清理任务
func (c *Container) StartChatCleanupTasks() {
	if c.ChatScheduledService != nil {
		c.ChatScheduledService.StartCleanupTasks()
	}
}

// Close cleans up resources
func (c *Container) Close() error {
	// 停止聊天清理任务
	if c.ChatScheduledService != nil {
		c.ChatScheduledService.StopCleanupTasks()
	}
	
	// 关闭Redis连接
	if c.RedisClient != nil {
		c.RedisClient.Close()
	}
	
	if c.Database != nil {
		return c.Database.Close()
	}
	return nil
}
