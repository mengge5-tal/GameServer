package repository

import (
	"GameServer/internal/domain/entity"
	"GameServer/internal/domain/repository"
	"database/sql"
)

// mysqlEquipmentRepository implements EquipmentRepository
type mysqlEquipmentRepository struct {
	db *sql.DB
}

// NewMySQLEquipmentRepository creates a new MySQL equipment repository
func NewMySQLEquipmentRepository(db *sql.DB) repository.EquipmentRepository {
	return &mysqlEquipmentRepository{db: db}
}

// GetByUserID retrieves all equipment for a user
func (r *mysqlEquipmentRepository) GetByUserID(userID int) ([]*entity.Equipment, error) {
	query := `SELECT equipid, quality, damage, crit, critdamage, damagespeed, 
			  bloodsuck, hp, movespeed, equipname, suitid, suitname, equip_type_id, equip_type_name,
			  userid, defense, goodfortune, type 
			  FROM equip WHERE userid = ?`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var equipment []*entity.Equipment
	for rows.Next() {
		equip := &entity.Equipment{}
		err := rows.Scan(
			&equip.EquipID, &equip.Quality, &equip.Damage, &equip.Crit,
			&equip.CritDamage, &equip.DamageSpeed, &equip.BloodSuck, &equip.HP,
			&equip.MoveSpeed, &equip.EquipName, &equip.SuitID, &equip.SuitName, &equip.EquipTypeID,
			&equip.EquipTypeName, &equip.UserID, &equip.Defense, &equip.GoodFortune, &equip.Type,
		)
		if err != nil {
			return nil, err
		}
		equipment = append(equipment, equip)
	}

	return equipment, rows.Err()
}

// GetByEquipID retrieves equipment by ID
func (r *mysqlEquipmentRepository) GetByEquipID(equipID int) (*entity.Equipment, error) {
	equip := &entity.Equipment{}
	query := `SELECT equipid, quality, damage, crit, critdamage, damagespeed, 
			  bloodsuck, hp, movespeed, equipname, suitid, suitname, equip_type_id, equip_type_name,
			  userid, defense, goodfortune, type 
			  FROM equip WHERE equipid = ?`

	err := r.db.QueryRow(query, equipID).Scan(
		&equip.EquipID, &equip.Quality, &equip.Damage, &equip.Crit,
		&equip.CritDamage, &equip.DamageSpeed, &equip.BloodSuck, &equip.HP,
		&equip.MoveSpeed, &equip.EquipName, &equip.SuitID, &equip.SuitName, &equip.EquipTypeID,
		&equip.EquipTypeName, &equip.UserID, &equip.Defense, &equip.GoodFortune, &equip.Type,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return equip, nil
}

// Create creates new equipment
func (r *mysqlEquipmentRepository) Create(equipment *entity.Equipment) error {
	query := `INSERT INTO equip (quality, damage, crit, critdamage, damagespeed, 
			  bloodsuck, hp, movespeed, equipname, suitid, suitname, equip_type_id, equip_type_name,
			  userid, defense, goodfortune, type) 
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := r.db.Exec(query,
		equipment.Quality, equipment.Damage, equipment.Crit,
		equipment.CritDamage, equipment.DamageSpeed, equipment.BloodSuck, equipment.HP,
		equipment.MoveSpeed, equipment.EquipName, equipment.SuitID, equipment.SuitName, equipment.EquipTypeID,
		equipment.EquipTypeName, equipment.UserID, equipment.Defense, equipment.GoodFortune, equipment.Type,
	)
	if err != nil {
		return err
	}

	// Get the auto-generated equipment ID
	equipID, err := result.LastInsertId()
	if err != nil {
		return err
	}
	equipment.EquipID = int(equipID)

	return nil
}

// Update updates existing equipment
func (r *mysqlEquipmentRepository) Update(equipment *entity.Equipment) error {
	query := `UPDATE equip SET quality = ?, damage = ?, crit = ?, critdamage = ?, 
			  damagespeed = ?, bloodsuck = ?, hp = ?, movespeed = ?, equipname = ?, suitid = ?, suitname = ?,
			  equip_type_id = ?, equip_type_name = ?, userid = ?, defense = ?, goodfortune = ?, type = ? 
			  WHERE equipid = ?`

	_, err := r.db.Exec(query,
		equipment.Quality, equipment.Damage, equipment.Crit, equipment.CritDamage,
		equipment.DamageSpeed, equipment.BloodSuck, equipment.HP, equipment.MoveSpeed,
		equipment.EquipName, equipment.SuitID, equipment.SuitName, equipment.EquipTypeID, equipment.EquipTypeName,
		equipment.UserID, equipment.Defense, equipment.GoodFortune, equipment.Type, equipment.EquipID,
	)
	return err
}

// Delete deletes equipment by ID
func (r *mysqlEquipmentRepository) Delete(equipID int) error {
	query := "DELETE FROM equip WHERE equipid = ?"
	_, err := r.db.Exec(query, equipID)
	return err
}

// BatchDelete deletes multiple equipment by IDs
func (r *mysqlEquipmentRepository) BatchDelete(equipIDs []int) (int, []int, error) {
	if len(equipIDs) == 0 {
		return 0, nil, nil
	}

	var deletedCount int
	var failedIDs []int

	// Begin transaction
	tx, err := r.db.Begin()
	if err != nil {
		return 0, nil, err
	}
	defer tx.Rollback()

	for _, equipID := range equipIDs {
		query := "DELETE FROM equip WHERE equipid = ?"
		result, err := tx.Exec(query, equipID)
		if err != nil {
			failedIDs = append(failedIDs, equipID)
			continue
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			failedIDs = append(failedIDs, equipID)
			continue
		}

		if rowsAffected > 0 {
			deletedCount++
		} else {
			// Equipment not found
			failedIDs = append(failedIDs, equipID)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return 0, nil, err
	}

	return deletedCount, failedIDs, nil
}

// GetUserEquipmentCount returns the count of equipment for a user
func (r *mysqlEquipmentRepository) GetUserEquipmentCount(userID int) (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM equip WHERE userid = ?"
	err := r.db.QueryRow(query, userID).Scan(&count)
	return count, err
}

