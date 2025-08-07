package service

import (
	"GameServer/internal/application/dto"
	"GameServer/internal/domain/entity"
	"GameServer/internal/domain/repository"
	"GameServer/internal/domain/service"
	"GameServer/internal/infrastructure/cache"
	"golang.org/x/crypto/bcrypt"
)

// AuthService handles authentication business logic
type AuthService struct {
	userRepo       repository.UserRepository
	playerRepo     repository.PlayerRepository
	authDomain     service.AuthDomainService
	cacheService   cache.CacheService
}

// NewAuthService creates a new auth service
func NewAuthService(
	userRepo repository.UserRepository,
	playerRepo repository.PlayerRepository,
	authDomain service.AuthDomainService,
	cacheService cache.CacheService,
) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		playerRepo:   playerRepo,
		authDomain:   authDomain,
		cacheService: cacheService,
	}
}

// Login handles user login
func (s *AuthService) Login(req *dto.LoginRequest) (*dto.LoginResponse, error) {
	// Validate input
	if err := s.authDomain.ValidateUsername(req.Username); err != nil {
		return nil, err
	}

	// Always check database for login to get latest online status
	// Cache is only used for subsequent operations, not for login verification
	cacheKey := "user:" + req.Username

	// Get user from database
	user, err := s.userRepo.VerifyCredentials(req.Username, req.Password)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, entity.NewDomainError("invalid username or password")
	}

	// Verify password
	if !s.verifyPassword(req.Password, user.Password) {
		return nil, entity.NewDomainError("invalid username or password")
	}

	// Check if user is already online (prevent duplicate login)
	if user.OnlineStatus == 1 {
		return nil, entity.NewDomainError("user is already logged in")
	}

	// Update online status to 1 (online)
	if err := s.userRepo.UpdateOnlineStatus(user.ID, 1); err != nil {
		// Log error but don't fail login
		// In production, you might want to handle this differently
	}

	// Cache user
	s.cacheService.SetUser(cacheKey, user)

	return &dto.LoginResponse{
		UserID:   user.ID,
		Username: user.Username,
	}, nil
}

// Register handles user registration
func (s *AuthService) Register(req *dto.RegisterRequest) (*dto.RegisterResponse, error) {
	// Validate input
	if err := s.authDomain.ValidateUsername(req.Username); err != nil {
		return nil, err
	}
	if err := s.authDomain.ValidatePassword(req.Password); err != nil {
		return nil, err
	}

	// Check if user already exists
	exists, err := s.userRepo.Exists(req.Username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, entity.NewDomainError("username already exists")
	}

	// Hash password
	hashedPassword, err := s.hashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// Create user entity
	user := &entity.User{
		Username: req.Username,
		Password: hashedPassword,
	}

	// Save user
	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	// Create default player info
	playerInfo := &entity.PlayerInfo{
		UserID:      user.ID,
		Level:       1,
		Experience:  0,
		GameLevel:   1,
		BloodEnergy: 100,
	}
	if err := s.playerRepo.Create(playerInfo); err != nil {
		// Log error but don't fail registration
		// In production, you might want to handle this with compensation
	}

	return &dto.RegisterResponse{
		UserID:   user.ID,
		Username: user.Username,
		Message:  "Registration successful",
	}, nil
}

// GetUserProfile gets user profile with player info
func (s *AuthService) GetUserProfile(userID int) (*dto.UserProfile, error) {
	// Get user info
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, entity.NewDomainError("user not found")
	}

	// Get player info
	playerInfo, err := s.playerRepo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}
	if playerInfo == nil {
		// Create default player info if not exists
		playerInfo = &entity.PlayerInfo{
			UserID:      userID,
			Level:       1,
			Experience:  0,
			GameLevel:   1,
			BloodEnergy: 100,
		}
		s.playerRepo.Create(playerInfo)
	}

	return &dto.UserProfile{
		UserID:      user.ID,
		Username:    user.Username,
		Level:       playerInfo.Level,
		Experience:  playerInfo.Experience,
		GameLevel:   playerInfo.GameLevel,
		BloodEnergy: playerInfo.BloodEnergy,
	}, nil
}

// Logout handles user logout
func (s *AuthService) Logout(userID int) error {
	// Update online status to 0 (offline)
	if err := s.userRepo.UpdateOnlineStatus(userID, 0); err != nil {
		// Log error but continue with logout process
	}

	// Clear cache
	user, err := s.userRepo.GetByID(userID)
	if err == nil && user != nil {
		cacheKey := "user:" + user.Username
		s.cacheService.Delete(cacheKey)
	}
	
	return nil
}

// hashPassword hashes password using bcrypt
func (s *AuthService) hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// verifyPassword verifies password against hash
func (s *AuthService) verifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}