package repository

import (
	"database/sql"
	"fmt"

	"GameServer/internal/domain/entity"
	"GameServer/internal/domain/repository"
)

// MySQLUserEquipRepository implements UserEquipRepository using MySQL
type MySQLUserEquipRepository struct {
	db *sql.DB
}

// NewMySQLUserEquipRepository creates a new MySQL user equipment repository
func NewMySQLUserEquipRepository(db *sql.DB) repository.UserEquipRepository {
	return &MySQLUserEquipRepository{db: db}
}

// GetUserEquippedItems retrieves all equipped items for a user
func (r *MySQLUserEquipRepository) GetUserEquippedItems(userID int) ([]*entity.UserEquip, error) {
	query := `SELECT id, userid, equip_slot, equipid FROM user_equip WHERE userid = ?`
	
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user equipped items: %w", err)
	}
	defer rows.Close()
	
	var userEquips []*entity.UserEquip
	for rows.Next() {
		var ue entity.UserEquip
		var equipID sql.NullInt32
		
		err := rows.Scan(&ue.ID, &ue.UserID, &ue.EquipSlot, &equipID)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user equip: %w", err)
		}
		
		if equipID.Valid {
			equipIDInt := int(equipID.Int32)
			ue.EquipID = &equipIDInt
		}
		
		userEquips = append(userEquips, &ue)
	}
	
	return userEquips, nil
}

// GetUserEquipBySlot retrieves equipment for a specific slot
func (r *MySQLUserEquipRepository) GetUserEquipBySlot(userID int, slot string) (*entity.UserEquip, error) {
	query := `SELECT id, userid, equip_slot, equipid FROM user_equip WHERE userid = ? AND equip_slot = ?`
	
	var ue entity.UserEquip
	var equipID sql.NullInt32
	
	err := r.db.QueryRow(query, userID, slot).Scan(&ue.ID, &ue.UserID, &ue.EquipSlot, &equipID)
	if err == sql.ErrNoRows {
		return nil, nil // No equipment in this slot
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user equip by slot: %w", err)
	}
	
	if equipID.Valid {
		equipIDInt := int(equipID.Int32)
		ue.EquipID = &equipIDInt
	}
	
	return &ue, nil
}

// UpdateUserEquip updates equipment for a specific slot
func (r *MySQLUserEquipRepository) UpdateUserEquip(userEquip *entity.UserEquip) error {
	// First check if a record exists for this user and slot
	existing, err := r.GetUserEquipBySlot(userEquip.UserID, userEquip.EquipSlot)
	if err != nil {
		return fmt.Errorf("failed to check existing user equip: %w", err)
	}
	
	if existing == nil {
		// Insert new record
		query := `INSERT INTO user_equip (userid, equip_slot, equipid) VALUES (?, ?, ?)`
		_, err = r.db.Exec(query, userEquip.UserID, userEquip.EquipSlot, userEquip.EquipID)
		if err != nil {
			return fmt.Errorf("failed to insert user equip: %w", err)
		}
	} else {
		// Update existing record
		query := `UPDATE user_equip SET equipid = ? WHERE userid = ? AND equip_slot = ?`
		_, err = r.db.Exec(query, userEquip.EquipID, userEquip.UserID, userEquip.EquipSlot)
		if err != nil {
			return fmt.Errorf("failed to update user equip: %w", err)
		}
	}
	
	return nil
}

// UnequipItem removes equipment from a slot (sets equipid to NULL)
func (r *MySQLUserEquipRepository) UnequipItem(userID int, slot string) error {
	query := `UPDATE user_equip SET equipid = NULL WHERE userid = ? AND equip_slot = ?`
	
	result, err := r.db.Exec(query, userID, slot)
	if err != nil {
		return fmt.Errorf("failed to unequip item: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	
	if rowsAffected == 0 {
		// No existing record, create one with NULL equipid
		insertQuery := `INSERT INTO user_equip (userid, equip_slot, equipid) VALUES (?, ?, NULL)`
		_, err = r.db.Exec(insertQuery, userID, slot)
		if err != nil {
			return fmt.Errorf("failed to create empty equip slot: %w", err)
		}
	}
	
	return nil
}

// InitializeUserEquipSlots creates initial empty slots for a new user
func (r *MySQLUserEquipRepository) InitializeUserEquipSlots(userID int) error {
	// Begin transaction
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()
	
	// Insert records for all equipment slots with NULL equipid
	query := `INSERT INTO user_equip (userid, equip_slot, equipid) VALUES (?, ?, NULL)`
	
	for _, slot := range entity.ValidEquipSlots {
		_, err = tx.Exec(query, userID, slot)
		if err != nil {
			return fmt.Errorf("failed to initialize equip slot %s: %w", slot, err)
		}
	}
	
	// Commit transaction
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	return nil
}

// GetEquippedItemDetails retrieves full equipment details for equipped items
func (r *MySQLUserEquipRepository) GetEquippedItemDetails(userID int) ([]*entity.Equipment, error) {
	query := `
		SELECT 
			e.equipid, e.quality, e.damage, e.crit, e.critdamage, e.damagespeed,
			e.bloodsuck, e.hp, e.movespeed, e.suitid, e.suitname, 
			e.equip_type_id, e.equip_type_name, e.userid, e.defense, e.goodfortune, e.type
		FROM user_equip ue
		INNER JOIN equip e ON ue.equipid = e.equipid
		WHERE ue.userid = ? AND ue.equipid IS NOT NULL
	`
	
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get equipped item details: %w", err)
	}
	defer rows.Close()
	
	var equipments []*entity.Equipment
	for rows.Next() {
		var e entity.Equipment
		
		err := rows.Scan(
			&e.EquipID, &e.Quality, &e.Damage, &e.Crit, &e.CritDamage, &e.DamageSpeed,
			&e.BloodSuck, &e.HP, &e.MoveSpeed, &e.SuitID, &e.SuitName,
			&e.EquipTypeID, &e.EquipTypeName, &e.UserID, &e.Defense, &e.GoodFortune, &e.Type,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan equipment: %w", err)
		}
		
		equipments = append(equipments, &e)
	}
	
	return equipments, nil
}