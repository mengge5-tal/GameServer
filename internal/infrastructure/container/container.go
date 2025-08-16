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
	UserSourceStoneRepo repository.UserSourceStoneRepository
	KillCountRepo       repository.KillCountRepository

	// Domain Services
	AuthDomainService domainService.AuthDomainService
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

	return nil
}

// initializeServices initializes all application services
func (c *Container) initializeServices() error {
	// Initialize cache service
	c.CacheService = cache.NewMemoryCache()

	// Initialize domain services
	c.AuthDomainService = domainService.NewAuthDomainService(c.Config.Security.BcryptCost)

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
}

// Close cleans up resources
func (c *Container) Close() error {
	if c.Database != nil {
		return c.Database.Close()
	}
	return nil
}
