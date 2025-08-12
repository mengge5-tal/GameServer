package repository

import (
	"GameServer/internal/domain/entity"
	"GameServer/internal/domain/repository"
	"database/sql"
)

// mysqlUserWeaponRepository implements UserWeaponRepository
type mysqlUserWeaponRepository struct {
	db *sql.DB
}

// NewMySQLUserWeaponRepository creates a new MySQL user weapon repository
func NewMySQLUserWeaponRepository(db *sql.DB) repository.UserWeaponRepository {
	return &mysqlUserWeaponRepository{db: db}
}

// GetByID retrieves user weapon ownership by ID
func (r *mysqlUserWeaponRepository) GetByID(id int) (*entity.UserWeapon, error) {
	userWeapon := &entity.UserWeapon{}
	query := "SELECT id, user_ID, weapon_ID FROM user_weapon WHERE id = ?"

	err := r.db.QueryRow(query, id).Scan(
		&userWeapon.ID, &userWeapon.UserID, &userWeapon.WeaponID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return userWeapon, nil
}

// GetByUserID retrieves all weapons owned by a user
func (r *mysqlUserWeaponRepository) GetByUserID(userID int) ([]*entity.UserWeapon, error) {
	query := "SELECT id, user_ID, weapon_ID FROM user_weapon WHERE user_ID = ? ORDER BY id ASC"

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userWeapons []*entity.UserWeapon
	for rows.Next() {
		userWeapon := &entity.UserWeapon{}
		err := rows.Scan(
			&userWeapon.ID, &userWeapon.UserID, &userWeapon.WeaponID,
		)
		if err != nil {
			return nil, err
		}
		userWeapons = append(userWeapons, userWeapon)
	}

	return userWeapons, rows.Err()
}

// GetByUserAndWeapon retrieves user weapon ownership by user ID and weapon ID
func (r *mysqlUserWeaponRepository) GetByUserAndWeapon(userID, weaponID int) (*entity.UserWeapon, error) {
	userWeapon := &entity.UserWeapon{}
	query := "SELECT id, user_ID, weapon_ID FROM user_weapon WHERE user_ID = ? AND weapon_ID = ?"

	err := r.db.QueryRow(query, userID, weaponID).Scan(
		&userWeapon.ID, &userWeapon.UserID, &userWeapon.WeaponID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return userWeapon, nil
}

// Create creates new user weapon ownership
func (r *mysqlUserWeaponRepository) Create(userWeapon *entity.UserWeapon) error {
	query := "INSERT INTO user_weapon (user_ID, weapon_ID) VALUES (?, ?)"

	result, err := r.db.Exec(query, userWeapon.UserID, userWeapon.WeaponID)
	if err != nil {
		return err
	}

	// Get the auto-generated ID
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	userWeapon.ID = int(id)

	return nil
}

// Delete deletes user weapon ownership by ID
func (r *mysqlUserWeaponRepository) Delete(id int) error {
	query := "DELETE FROM user_weapon WHERE id = ?"
	_, err := r.db.Exec(query, id)
	return err
}

// DeleteByUserAndWeapon deletes user weapon ownership by user ID and weapon ID
func (r *mysqlUserWeaponRepository) DeleteByUserAndWeapon(userID, weaponID int) error {
	query := "DELETE FROM user_weapon WHERE user_ID = ? AND weapon_ID = ?"
	_, err := r.db.Exec(query, userID, weaponID)
	return err
}

// UserOwnsWeapon checks if a user owns a specific weapon
func (r *mysqlUserWeaponRepository) UserOwnsWeapon(userID, weaponID int) (bool, error) {
	var count int
	query := "SELECT COUNT(*) FROM user_weapon WHERE user_ID = ? AND weapon_ID = ?"
	err := r.db.QueryRow(query, userID, weaponID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}