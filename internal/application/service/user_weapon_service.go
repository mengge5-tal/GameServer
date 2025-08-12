package service

import (
	"GameServer/internal/application/dto"
	"GameServer/internal/domain/entity"
	"GameServer/internal/domain/repository"
	"database/sql"
)

// UserWeaponService handles user weapon ownership business logic
type UserWeaponService struct {
	userWeaponRepo repository.UserWeaponRepository
	weaponRepo     repository.WeaponRepository
	userRepo       repository.UserRepository
}

// NewUserWeaponService creates a new user weapon service
func NewUserWeaponService(
	userWeaponRepo repository.UserWeaponRepository,
	weaponRepo repository.WeaponRepository,
	userRepo repository.UserRepository,
) *UserWeaponService {
	return &UserWeaponService{
		userWeaponRepo: userWeaponRepo,
		weaponRepo:     weaponRepo,
		userRepo:       userRepo,
	}
}

// GetUserWeapons retrieves all weapons owned by a user
func (s *UserWeaponService) GetUserWeapons(userID int, withDetails bool) (interface{}, error) {
	if userID <= 0 {
		return nil, entity.NewDomainError("user ID must be positive")
	}

	// Check if user exists
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, entity.NewDomainError("user not found")
	}

	userWeapons, err := s.userWeaponRepo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}

	if !withDetails {
		// Return simple response without weapon details
		var response []*dto.UserWeaponResponse
		for _, uw := range userWeapons {
			response = append(response, &dto.UserWeaponResponse{
				ID:       uw.ID,
				UserID:   uw.UserID,
				WeaponID: uw.WeaponID,
			})
		}
		return response, nil
	}

	// Return detailed response with weapon information
	var response []*dto.UserWeaponDetailResponse
	for _, uw := range userWeapons {
		detailResp := &dto.UserWeaponDetailResponse{
			ID:       uw.ID,
			UserID:   uw.UserID,
			WeaponID: uw.WeaponID,
		}

		// Get weapon details
		weapon, err := s.weaponRepo.GetByID(uw.WeaponID)
		if err != nil && err != sql.ErrNoRows {
			return nil, err
		}
		if weapon != nil {
			detailResp.Weapon = &dto.WeaponResponse{
				WeaponID:            weapon.WeaponID,
				WeaponName:          weapon.WeaponName,
				AttackPower:         weapon.AttackPower,
				AttackSpeed:         weapon.AttackSpeed,
				CriticalStrikeRate:  weapon.CriticalStrikeRate,
				CriticalStrikeDamage: weapon.CriticalStrikeDamage,
				LuckyValue:          weapon.LuckyValue,
				EnhancementLevel:    weapon.EnhancementLevel,
				GrowthValue:         weapon.GrowthValue,
				Quality:             weapon.Quality,
			}
		}

		response = append(response, detailResp)
	}

	return response, nil
}

// AddUserWeapon adds a weapon to user's inventory
func (s *UserWeaponService) AddUserWeapon(req *dto.AddUserWeaponRequest) (*dto.UserWeaponResponse, error) {
	if req.UserID <= 0 {
		return nil, entity.NewDomainError("user ID must be positive")
	}
	if req.WeaponID <= 0 {
		return nil, entity.NewDomainError("weapon ID must be positive")
	}

	// Check if user exists
	user, err := s.userRepo.GetByID(req.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, entity.NewDomainError("user not found")
	}

	// Check if weapon exists
	weapon, err := s.weaponRepo.GetByID(req.WeaponID)
	if err != nil {
		return nil, err
	}
	if weapon == nil {
		return nil, entity.NewDomainError("weapon not found")
	}

	// Check if user already owns this weapon
	owns, err := s.userWeaponRepo.UserOwnsWeapon(req.UserID, req.WeaponID)
	if err != nil {
		return nil, err
	}
	if owns {
		return nil, entity.NewDomainError("user already owns this weapon")
	}

	userWeapon := &entity.UserWeapon{
		UserID:   req.UserID,
		WeaponID: req.WeaponID,
	}

	if err := s.userWeaponRepo.Create(userWeapon); err != nil {
		return nil, err
	}

	return &dto.UserWeaponResponse{
		ID:       userWeapon.ID,
		UserID:   userWeapon.UserID,
		WeaponID: userWeapon.WeaponID,
	}, nil
}

// RemoveUserWeapon removes a weapon from user's inventory
func (s *UserWeaponService) RemoveUserWeapon(req *dto.RemoveUserWeaponRequest) error {
	if req.UserID <= 0 {
		return entity.NewDomainError("user ID must be positive")
	}
	if req.WeaponID <= 0 {
		return entity.NewDomainError("weapon ID must be positive")
	}

	// Check if user owns this weapon
	userWeapon, err := s.userWeaponRepo.GetByUserAndWeapon(req.UserID, req.WeaponID)
	if err != nil {
		return err
	}
	if userWeapon == nil {
		return entity.NewDomainError("user does not own this weapon")
	}

	return s.userWeaponRepo.DeleteByUserAndWeapon(req.UserID, req.WeaponID)
}

// RemoveUserWeaponByID removes a user weapon ownership by ID
func (s *UserWeaponService) RemoveUserWeaponByID(id int) error {
	if id <= 0 {
		return entity.NewDomainError("ID must be positive")
	}

	// Check if user weapon exists
	userWeapon, err := s.userWeaponRepo.GetByID(id)
	if err != nil {
		return err
	}
	if userWeapon == nil {
		return entity.NewDomainError("user weapon ownership not found")
	}

	return s.userWeaponRepo.Delete(id)
}

// CheckUserWeapon checks if a user owns a specific weapon
func (s *UserWeaponService) CheckUserWeapon(req *dto.CheckUserWeaponRequest) (*dto.CheckUserWeaponResponse, error) {
	if req.UserID <= 0 {
		return nil, entity.NewDomainError("user ID must be positive")
	}
	if req.WeaponID <= 0 {
		return nil, entity.NewDomainError("weapon ID must be positive")
	}

	owns, err := s.userWeaponRepo.UserOwnsWeapon(req.UserID, req.WeaponID)
	if err != nil {
		return nil, err
	}

	return &dto.CheckUserWeaponResponse{
		UserID:     req.UserID,
		WeaponID:   req.WeaponID,
		OwnsWeapon: owns,
	}, nil
}