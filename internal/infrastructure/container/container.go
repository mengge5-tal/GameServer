package container

import (
	"database/sql"
	"GameServer/internal/application/service"
	"GameServer/internal/domain/repository"
	domainService "GameServer/internal/domain/service"
	"GameServer/internal/infrastructure/cache"
	"GameServer/internal/infrastructure/config"
	infraRepo "GameServer/internal/infrastructure/repository"
	"GameServer/internal/interfaces/websocket"
)

// Container holds all application dependencies
type Container struct {
	Config   *config.Config
	Database *sql.DB
	
	// Services
	CacheService     cache.CacheService
	AuthService      *service.AuthService
	PlayerService    *service.PlayerService
	FriendService    *service.FriendService
	RankingService   *service.RankingService
	UserEquipService *service.UserEquipService
	
	// Repositories
	UserRepo        repository.UserRepository
	PlayerRepo      repository.PlayerRepository
	FriendRepo      repository.FriendRepository
	RankingRepo     repository.RankingRepository
	EquipmentRepo   repository.EquipmentRepository
	SourceStoneRepo repository.SourceStoneRepository
	ExperienceRepo  repository.ExperienceRepository
	UserEquipRepo   repository.UserEquipRepository
	
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
	c.RankingRepo = infraRepo.NewMySQLRankingRepository(c.Database)
	c.EquipmentRepo = infraRepo.NewMySQLEquipmentRepository(c.Database)
	c.SourceStoneRepo = infraRepo.NewMySQLSourceStoneRepository(c.Database)
	c.ExperienceRepo = infraRepo.NewMySQLExperienceRepository(c.Database)
	c.UserEquipRepo = infraRepo.NewMySQLUserEquipRepository(c.Database)
	
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
		c.SourceStoneRepo,
		c.CacheService,
	)
	
	c.FriendService = service.NewFriendService(
		c.FriendRepo,
		c.UserRepo,
		c.PlayerRepo,
	)
	
	c.RankingService = service.NewRankingService(
		c.RankingRepo,
		c.UserRepo,
		c.PlayerRepo,
	)
	
	c.UserEquipService = service.NewUserEquipService(
		c.UserEquipRepo,
		c.EquipmentRepo,
		c.UserRepo,
	)
	
	return nil
}

// GetWebSocketServices returns services formatted for WebSocket handlers
func (c *Container) GetWebSocketServices() *websocket.ServiceContainer {
	return &websocket.ServiceContainer{
		AuthService:      c.AuthService,
		PlayerService:    c.PlayerService,
		FriendService:    c.FriendService,
		RankingService:   c.RankingService,
		UserEquipService: c.UserEquipService,
	}
}

// Close cleans up resources
func (c *Container) Close() error {
	if c.Database != nil {
		return c.Database.Close()
	}
	return nil
}