package service

import (
	"GameServer/internal/application/dto"
	"GameServer/internal/domain/entity"
	"GameServer/internal/domain/repository"
)

// WeaponService handles weapon-related business logic
type WeaponService struct {
	weaponRepo repository.WeaponRepository
}

// NewWeaponService creates a new weapon service
func NewWeaponService(weaponRepo repository.WeaponRepository) *WeaponService {
	return &WeaponService{
		weaponRepo: weaponRepo,
	}
}

// GetWeaponByID retrieves weapon by ID
func (s *WeaponService) GetWeaponByID(weaponID int) (*dto.WeaponResponse, error) {
	if weaponID <= 0 {
		return nil, entity.NewDomainError("weapon ID must be positive")
	}

	weapon, err := s.weaponRepo.GetByID(weaponID)
	if err != nil {
		return nil, err
	}
	if weapon == nil {
		return nil, entity.NewDomainError("weapon not found")
	}

	return &dto.WeaponResponse{
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
	}, nil
}

// GetAllWeapons retrieves all weapons
func (s *WeaponService) GetAllWeapons() ([]*dto.WeaponResponse, error) {
	weapons, err := s.weaponRepo.GetAll()
	if err != nil {
		return nil, err
	}

	var response []*dto.WeaponResponse
	for _, weapon := range weapons {
		response = append(response, &dto.WeaponResponse{
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
		})
	}

	return response, nil
}

// CreateWeapon creates a new weapon
func (s *WeaponService) CreateWeapon(req *dto.CreateWeaponRequest) (*dto.WeaponResponse, error) {
	// Validate input
	if err := s.validateWeaponData(req.WeaponName, req.AttackPower, req.Quality); err != nil {
		return nil, err
	}

	weapon := &entity.Weapon{
		WeaponName:          req.WeaponName,
		AttackPower:         req.AttackPower,
		AttackSpeed:         req.AttackSpeed,
		CriticalStrikeRate:  req.CriticalStrikeRate,
		CriticalStrikeDamage: req.CriticalStrikeDamage,
		LuckyValue:          req.LuckyValue,
		EnhancementLevel:    req.EnhancementLevel,
		GrowthValue:         req.GrowthValue,
		Quality:             req.Quality,
	}

	if err := s.weaponRepo.Create(weapon); err != nil {
		return nil, err
	}

	return &dto.WeaponResponse{
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
	}, nil
}

// UpdateWeapon updates an existing weapon
func (s *WeaponService) UpdateWeapon(req *dto.UpdateWeaponRequest) (*dto.WeaponResponse, error) {
	// Validate input
	if req.WeaponID <= 0 {
		return nil, entity.NewDomainError("weapon ID must be positive")
	}
	if err := s.validateWeaponData(req.WeaponName, req.AttackPower, req.Quality); err != nil {
		return nil, err
	}

	// Check if weapon exists
	existingWeapon, err := s.weaponRepo.GetByID(req.WeaponID)
	if err != nil {
		return nil, err
	}
	if existingWeapon == nil {
		return nil, entity.NewDomainError("weapon not found")
	}

	weapon := &entity.Weapon{
		WeaponID:            req.WeaponID,
		WeaponName:          req.WeaponName,
		AttackPower:         req.AttackPower,
		AttackSpeed:         req.AttackSpeed,
		CriticalStrikeRate:  req.CriticalStrikeRate,
		CriticalStrikeDamage: req.CriticalStrikeDamage,
		LuckyValue:          req.LuckyValue,
		EnhancementLevel:    req.EnhancementLevel,
		GrowthValue:         req.GrowthValue,
		Quality:             req.Quality,
	}

	if err := s.weaponRepo.Update(weapon); err != nil {
		return nil, err
	}

	return &dto.WeaponResponse{
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
	}, nil
}

// DeleteWeapon deletes a weapon by ID
func (s *WeaponService) DeleteWeapon(weaponID int) error {
	if weaponID <= 0 {
		return entity.NewDomainError("weapon ID must be positive")
	}

	// Check if weapon exists
	existingWeapon, err := s.weaponRepo.GetByID(weaponID)
	if err != nil {
		return err
	}
	if existingWeapon == nil {
		return entity.NewDomainError("weapon not found")
	}

	return s.weaponRepo.Delete(weaponID)
}

// validateWeaponData validates weapon data
func (s *WeaponService) validateWeaponData(weaponName string, attackPower, quality int) error {
	if weaponName == "" {
		return entity.NewDomainError("weapon name is required")
	}
	if len(weaponName) > 45 {
		return entity.NewDomainError("weapon name must not exceed 45 characters")
	}
	if attackPower < 0 {
		return entity.NewDomainError("attack power must be non-negative")
	}
	if quality < 0 {
		return entity.NewDomainError("quality must be non-negative")
	}
	return nil
}