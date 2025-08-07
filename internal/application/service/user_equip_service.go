package service

import (
	"fmt"

	"GameServer/internal/domain/entity"
	"GameServer/internal/domain/repository"
)

// UserEquipService handles user equipment business logic
type UserEquipService struct {
	userEquipRepo repository.UserEquipRepository
	equipmentRepo repository.EquipmentRepository
	userRepo      repository.UserRepository
}

// NewUserEquipService creates a new user equipment service
func NewUserEquipService(userEquipRepo repository.UserEquipRepository, equipmentRepo repository.EquipmentRepository, userRepo repository.UserRepository) *UserEquipService {
	return &UserEquipService{
		userEquipRepo: userEquipRepo,
		equipmentRepo: equipmentRepo,
		userRepo:      userRepo,
	}
}

// GetUserEquippedItems retrieves all equipped items for a user with detailed information
func (s *UserEquipService) GetUserEquippedItems(userID int) (map[string]interface{}, error) {
	// Verify user exists
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Get user equipped items
	userEquips, err := s.userEquipRepo.GetUserEquippedItems(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user equipped items: %w", err)
	}

	// If no equipment slots exist, initialize them
	if len(userEquips) == 0 {
		err = s.userEquipRepo.InitializeUserEquipSlots(userID)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize equipment slots: %w", err)
		}
		// Get the newly initialized slots
		userEquips, err = s.userEquipRepo.GetUserEquippedItems(userID)
		if err != nil {
			return nil, fmt.Errorf("failed to get user equipped items after initialization: %w", err)
		}
	}

	// Create result map with all slots
	result := make(map[string]interface{})
	for _, slot := range entity.ValidEquipSlots {
		result[slot] = nil // Initialize all slots as empty
	}

	// Fill in equipped items
	for _, userEquip := range userEquips {
		if userEquip.EquipID != nil {
			// Get equipment details
			equipment, err := s.equipmentRepo.GetByEquipID(*userEquip.EquipID)
			if err != nil {
				return nil, fmt.Errorf("failed to get equipment details for ID %d: %w", *userEquip.EquipID, err)
			}
			if equipment != nil {
				result[userEquip.EquipSlot] = equipment
			}
		}
	}

	return result, nil
}

// EquipItem equips an item to a specific slot
func (s *UserEquipService) EquipItem(userID int, slot string, equipID int) error {
	// Validate slot type
	isValidSlot := false
	for _, validSlot := range entity.ValidEquipSlots {
		if slot == validSlot {
			isValidSlot = true
			break
		}
	}
	if !isValidSlot {
		return fmt.Errorf("invalid equipment slot: %s", slot)
	}

	// Verify user exists
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("failed to verify user: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	// Verify equipment exists and belongs to user
	equipment, err := s.equipmentRepo.GetByEquipID(equipID)
	if err != nil {
		return fmt.Errorf("failed to get equipment: %w", err)
	}
	if equipment == nil {
		return fmt.Errorf("equipment not found")
	}
	if equipment.UserID != userID {
		return fmt.Errorf("equipment does not belong to user")
	}

	// Check if equipment is already equipped in another slot
	userEquips, err := s.userEquipRepo.GetUserEquippedItems(userID)
	if err != nil {
		return fmt.Errorf("failed to get user equipped items: %w", err)
	}

	for _, ue := range userEquips {
		if ue.EquipID != nil && *ue.EquipID == equipID && ue.EquipSlot != slot {
			return fmt.Errorf("equipment is already equipped in slot: %s", ue.EquipSlot)
		}
	}

	// Update user equipment
	userEquip := &entity.UserEquip{
		UserID:    userID,
		EquipSlot: slot,
		EquipID:   &equipID,
	}

	err = userEquip.Validate()
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	err = s.userEquipRepo.UpdateUserEquip(userEquip)
	if err != nil {
		return fmt.Errorf("failed to equip item: %w", err)
	}

	return nil
}

// UnequipItem removes equipment from a specific slot
func (s *UserEquipService) UnequipItem(userID int, slot string) error {
	// Validate slot type
	isValidSlot := false
	for _, validSlot := range entity.ValidEquipSlots {
		if slot == validSlot {
			isValidSlot = true
			break
		}
	}
	if !isValidSlot {
		return fmt.Errorf("invalid equipment slot: %s", slot)
	}

	// Verify user exists
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("failed to verify user: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	// Unequip item
	err = s.userEquipRepo.UnequipItem(userID, slot)
	if err != nil {
		return fmt.Errorf("failed to unequip item: %w", err)
	}

	return nil
}

// GetEquippedItemsBySlot retrieves equipment for a specific slot
func (s *UserEquipService) GetEquippedItemsBySlot(userID int, slot string) (interface{}, error) {
	// Validate slot type
	isValidSlot := false
	for _, validSlot := range entity.ValidEquipSlots {
		if slot == validSlot {
			isValidSlot = true
			break
		}
	}
	if !isValidSlot {
		return nil, fmt.Errorf("invalid equipment slot: %s", slot)
	}

	// Get user equipment for the slot
	userEquip, err := s.userEquipRepo.GetUserEquipBySlot(userID, slot)
	if err != nil {
		return nil, fmt.Errorf("failed to get user equipment: %w", err)
	}

	if userEquip == nil || userEquip.EquipID == nil {
		return nil, nil // No equipment in this slot
	}

	// Get equipment details
	equipment, err := s.equipmentRepo.GetByEquipID(*userEquip.EquipID)
	if err != nil {
		return nil, fmt.Errorf("failed to get equipment details: %w", err)
	}

	return equipment, nil
}

// InitializeUserEquipSlots initializes equipment slots for a new user
func (s *UserEquipService) InitializeUserEquipSlots(userID int) error {
	return s.userEquipRepo.InitializeUserEquipSlots(userID)
}

// GetEquipmentStats calculates total stats from all equipped items
func (s *UserEquipService) GetEquipmentStats(userID int) (map[string]int, error) {
	equippedItems, err := s.GetUserEquippedItems(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get equipped items: %w", err)
	}

	stats := map[string]int{
		"damage":      0,
		"crit":        0,
		"critdamage":  0,
		"damagespeed": 0,
		"bloodsuck":   0,
		"hp":          0,
		"movespeed":   0,
		"defense":     0,
		"goodfortune": 0,
	}

	for _, equipmentInterface := range equippedItems {
		if equipmentInterface != nil {
			equipment, ok := equipmentInterface.(*entity.Equipment)
			if ok {
				stats["damage"] += equipment.Damage
				stats["crit"] += equipment.Crit
				stats["critdamage"] += equipment.CritDamage
				stats["damagespeed"] += equipment.DamageSpeed
				stats["bloodsuck"] += equipment.BloodSuck
				stats["hp"] += equipment.HP
				stats["movespeed"] += equipment.MoveSpeed
				stats["defense"] += equipment.Defense
				stats["goodfortune"] += equipment.GoodFortune
			}
		}
	}

	return stats, nil
}