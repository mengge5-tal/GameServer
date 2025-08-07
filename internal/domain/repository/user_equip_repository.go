package repository

import "GameServer/internal/domain/entity"

// UserEquipRepository defines the interface for user equipment data access
type UserEquipRepository interface {
	// GetUserEquippedItems retrieves all equipped items for a user
	GetUserEquippedItems(userID int) ([]*entity.UserEquip, error)
	
	// GetUserEquipBySlot retrieves equipment for a specific slot
	GetUserEquipBySlot(userID int, slot string) (*entity.UserEquip, error)
	
	// UpdateUserEquip updates equipment for a specific slot
	UpdateUserEquip(userEquip *entity.UserEquip) error
	
	// UnequipItem removes equipment from a slot (sets equipid to NULL)
	UnequipItem(userID int, slot string) error
	
	// InitializeUserEquipSlots creates initial empty slots for a new user
	InitializeUserEquipSlots(userID int) error
	
	// GetEquippedItemDetails retrieves full equipment details for equipped items
	GetEquippedItemDetails(userID int) ([]*entity.Equipment, error)
}