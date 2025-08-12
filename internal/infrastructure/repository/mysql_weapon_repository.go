package repository

import (
	"GameServer/internal/domain/entity"
	"GameServer/internal/domain/repository"
	"database/sql"
)

// mysqlWeaponRepository implements WeaponRepository
type mysqlWeaponRepository struct {
	db *sql.DB
}

// NewMySQLWeaponRepository creates a new MySQL weapon repository
func NewMySQLWeaponRepository(db *sql.DB) repository.WeaponRepository {
	return &mysqlWeaponRepository{db: db}
}

// GetByID retrieves weapon by ID
func (r *mysqlWeaponRepository) GetByID(weaponID int) (*entity.Weapon, error) {
	weapon := &entity.Weapon{}
	query := `SELECT weapon_ID, weapon_Name, attack_power, attack_speed, 
			  critical_strike_rate, critical_strike_damage, lucky_value, 
			  enhancement_level, growth_value, quality 
			  FROM weapon WHERE weapon_ID = ?`

	err := r.db.QueryRow(query, weaponID).Scan(
		&weapon.WeaponID, &weapon.WeaponName, &weapon.AttackPower, &weapon.AttackSpeed,
		&weapon.CriticalStrikeRate, &weapon.CriticalStrikeDamage, &weapon.LuckyValue,
		&weapon.EnhancementLevel, &weapon.GrowthValue, &weapon.Quality,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return weapon, nil
}

// GetAll retrieves all weapons
func (r *mysqlWeaponRepository) GetAll() ([]*entity.Weapon, error) {
	query := `SELECT weapon_ID, weapon_Name, attack_power, attack_speed, 
			  critical_strike_rate, critical_strike_damage, lucky_value, 
			  enhancement_level, growth_value, quality 
			  FROM weapon ORDER BY weapon_ID ASC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var weapons []*entity.Weapon
	for rows.Next() {
		weapon := &entity.Weapon{}
		err := rows.Scan(
			&weapon.WeaponID, &weapon.WeaponName, &weapon.AttackPower, &weapon.AttackSpeed,
			&weapon.CriticalStrikeRate, &weapon.CriticalStrikeDamage, &weapon.LuckyValue,
			&weapon.EnhancementLevel, &weapon.GrowthValue, &weapon.Quality,
		)
		if err != nil {
			return nil, err
		}
		weapons = append(weapons, weapon)
	}

	return weapons, rows.Err()
}

// Create creates new weapon
func (r *mysqlWeaponRepository) Create(weapon *entity.Weapon) error {
	query := `INSERT INTO weapon (weapon_Name, attack_power, attack_speed, 
			  critical_strike_rate, critical_strike_damage, lucky_value, 
			  enhancement_level, growth_value, quality) 
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := r.db.Exec(query,
		weapon.WeaponName, weapon.AttackPower, weapon.AttackSpeed,
		weapon.CriticalStrikeRate, weapon.CriticalStrikeDamage, weapon.LuckyValue,
		weapon.EnhancementLevel, weapon.GrowthValue, weapon.Quality,
	)
	if err != nil {
		return err
	}

	// Get the auto-generated weapon ID
	weaponID, err := result.LastInsertId()
	if err != nil {
		return err
	}
	weapon.WeaponID = int(weaponID)

	return nil
}

// Update updates existing weapon
func (r *mysqlWeaponRepository) Update(weapon *entity.Weapon) error {
	query := `UPDATE weapon SET weapon_Name = ?, attack_power = ?, attack_speed = ?, 
			  critical_strike_rate = ?, critical_strike_damage = ?, lucky_value = ?, 
			  enhancement_level = ?, growth_value = ?, quality = ? 
			  WHERE weapon_ID = ?`

	_, err := r.db.Exec(query,
		weapon.WeaponName, weapon.AttackPower, weapon.AttackSpeed,
		weapon.CriticalStrikeRate, weapon.CriticalStrikeDamage, weapon.LuckyValue,
		weapon.EnhancementLevel, weapon.GrowthValue, weapon.Quality, weapon.WeaponID,
	)
	return err
}

// Delete deletes weapon by ID
func (r *mysqlWeaponRepository) Delete(weaponID int) error {
	query := "DELETE FROM weapon WHERE weapon_ID = ?"
	_, err := r.db.Exec(query, weaponID)
	return err
}