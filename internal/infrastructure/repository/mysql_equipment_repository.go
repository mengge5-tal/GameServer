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
			  bloodsuck, hp, movespeed, equipname, userid, defense, goodfortune, type 
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
			&equip.MoveSpeed, &equip.EquipName, &equip.UserID, &equip.Defense,
			&equip.GoodFortune, &equip.Type,
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
			  bloodsuck, hp, movespeed, equipname, userid, defense, goodfortune, type 
			  FROM equip WHERE equipid = ?`

	err := r.db.QueryRow(query, equipID).Scan(
		&equip.EquipID, &equip.Quality, &equip.Damage, &equip.Crit,
		&equip.CritDamage, &equip.DamageSpeed, &equip.BloodSuck, &equip.HP,
		&equip.MoveSpeed, &equip.EquipName, &equip.UserID, &equip.Defense,
		&equip.GoodFortune, &equip.Type,
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
	query := `INSERT INTO equip (equipid, quality, damage, crit, critdamage, damagespeed, 
			  bloodsuck, hp, movespeed, equipname, userid, defense, goodfortune, type) 
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.Exec(query,
		equipment.EquipID, equipment.Quality, equipment.Damage, equipment.Crit,
		equipment.CritDamage, equipment.DamageSpeed, equipment.BloodSuck, equipment.HP,
		equipment.MoveSpeed, equipment.EquipName, equipment.UserID, equipment.Defense,
		equipment.GoodFortune, equipment.Type,
	)
	return err
}

// Update updates existing equipment
func (r *mysqlEquipmentRepository) Update(equipment *entity.Equipment) error {
	query := `UPDATE equip SET quality = ?, damage = ?, crit = ?, critdamage = ?, 
			  damagespeed = ?, bloodsuck = ?, hp = ?, movespeed = ?, equipname = ?, 
			  userid = ?, defense = ?, goodfortune = ?, type = ? WHERE equipid = ?`

	_, err := r.db.Exec(query,
		equipment.Quality, equipment.Damage, equipment.Crit, equipment.CritDamage,
		equipment.DamageSpeed, equipment.BloodSuck, equipment.HP, equipment.MoveSpeed,
		equipment.EquipName, equipment.UserID, equipment.Defense, equipment.GoodFortune,
		equipment.Type, equipment.EquipID,
	)
	return err
}

// Delete deletes equipment by ID
func (r *mysqlEquipmentRepository) Delete(equipID int) error {
	query := "DELETE FROM equip WHERE equipid = ?"
	_, err := r.db.Exec(query, equipID)
	return err
}

// GetUserEquipmentCount returns the count of equipment for a user
func (r *mysqlEquipmentRepository) GetUserEquipmentCount(userID int) (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM equip WHERE userid = ?"
	err := r.db.QueryRow(query, userID).Scan(&count)
	return count, err
}

// GetMaxSequenceByTypeAndQuality returns the maximum sequence number for given type and quality
func (r *mysqlEquipmentRepository) GetMaxSequenceByTypeAndQuality(equipType, quality int) (int, error) {
	// Equipment ID format: [type][quality][6-digit sequence]
	// For type=4, quality=1: 41000001, 41000002, ...
	// We need to find the maximum sequence number (last 6 digits)
	
	prefix := equipType*10 + quality // For type=4, quality=1: prefix=41
	minID := prefix * 1000000        // 41000000
	maxID := minID + 999999          // 41999999
	
	var maxEquipID sql.NullInt64
	query := `SELECT MAX(equipid) FROM equip 
			  WHERE equipid >= ? AND equipid <= ? 
			  AND type = ? AND quality = ?`
	
	err := r.db.QueryRow(query, minID, maxID, equipType, quality).Scan(&maxEquipID)
	if err != nil {
		return 0, err
	}
	
	if !maxEquipID.Valid {
		// No equipment found, return 0 so the first sequence will be 1
		return 0, nil
	}
	
	// Extract the sequence number (last 6 digits)
	sequence := int(maxEquipID.Int64) % 1000000
	return sequence, nil
}
