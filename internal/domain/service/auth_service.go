package service

import (
	"GameServer/internal/domain/entity"
	"regexp"
	"unicode"
)

// AuthDomainService defines authentication business rules
type AuthDomainService interface {
	ValidatePassword(password string) error
	ValidateUsername(username string) error
	HashPassword(password string) (string, error)
	VerifyPassword(password, hash string) bool
}

// authDomainService implements AuthDomainService
type authDomainService struct {
	bcryptCost int
}

// NewAuthDomainService creates a new auth domain service
func NewAuthDomainService(bcryptCost int) AuthDomainService {
	return &authDomainService{
		bcryptCost: bcryptCost,
	}
}

// ValidatePassword validates password strength
func (s *authDomainService) ValidatePassword(password string) error {
	if len(password) < 8 {
		return entity.NewDomainError("password must be at least 8 characters long")
	}
	if len(password) > 128 {
		return entity.NewDomainError("password must be less than 128 characters")
	}

	var hasUpper, hasLower, hasNumber, hasSpecial bool
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper || !hasLower || !hasNumber || !hasSpecial {
		return entity.NewDomainError("password must contain uppercase, lowercase, number and special character")
	}

	return nil
}

// ValidateUsername validates username format
func (s *authDomainService) ValidateUsername(username string) error {
	if len(username) < 3 || len(username) > 20 {
		return entity.NewDomainError("username must be 3-20 characters")
	}

	// Only allow letters, numbers and underscores
	matched, _ := regexp.MatchString("^[a-zA-Z0-9_]+$", username)
	if !matched {
		return entity.NewDomainError("username can only contain letters, numbers and underscores")
	}

	return nil
}

// HashPassword hashes password with bcrypt
func (s *authDomainService) HashPassword(password string) (string, error) {
	// This will be implemented by infrastructure layer
	// Domain service defines the interface, infrastructure provides implementation
	return "", entity.NewDomainError("not implemented - should be provided by infrastructure")
}

// VerifyPassword verifies password against hash
func (s *authDomainService) VerifyPassword(password, hash string) bool {
	// This will be implemented by infrastructure layer
	// Domain service defines the interface, infrastructure provides implementation
	return false
}