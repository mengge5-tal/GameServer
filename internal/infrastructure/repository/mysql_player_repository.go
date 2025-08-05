package repository

import (
	"database/sql"
	"GameServer/internal/domain/entity"
	"GameServer/internal/domain/repository"
)

// mysqlPlayerRepository implements PlayerRepository
type mysqlPlayerRepository struct {
	db *sql.DB
}

// NewMySQLPlayerRepository creates a new MySQL player repository
func NewMySQLPlayerRepository(db *sql.DB) repository.PlayerRepository {
	return &mysqlPlayerRepository{db: db}
}

// GetByUserID retrieves player info by user ID
func (r *mysqlPlayerRepository) GetByUserID(userID int) (*entity.PlayerInfo, error) {
	player := &entity.PlayerInfo{}
	query := "SELECT userid, level, experience, gamelevel, bloodenergy FROM playerinfo WHERE userid = ?"
	err := r.db.QueryRow(query, userID).Scan(
		&player.UserID, &player.Level, &player.Experience, 
		&player.GameLevel, &player.BloodEnergy,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return player, nil
}

// Create creates new player info
func (r *mysqlPlayerRepository) Create(player *entity.PlayerInfo) error {
	query := "INSERT INTO playerinfo (userid, level, experience, gamelevel, bloodenergy) VALUES (?, ?, ?, ?, ?)"
	_, err := r.db.Exec(query, 
		player.UserID, player.Level, player.Experience, 
		player.GameLevel, player.BloodEnergy,
	)
	return err
}

// Update updates existing player info
func (r *mysqlPlayerRepository) Update(player *entity.PlayerInfo) error {
	query := "UPDATE playerinfo SET level = ?, experience = ?, gamelevel = ?, bloodenergy = ? WHERE userid = ?"
	_, err := r.db.Exec(query, 
		player.Level, player.Experience, player.GameLevel, 
		player.BloodEnergy, player.UserID,
	)
	return err
}

// Delete deletes player info by user ID
func (r *mysqlPlayerRepository) Delete(userID int) error {
	query := "DELETE FROM playerinfo WHERE userid = ?"
	_, err := r.db.Exec(query, userID)
	return err
}

// UpdateExperience updates player experience
func (r *mysqlPlayerRepository) UpdateExperience(userID, experience int) error {
	query := "UPDATE playerinfo SET experience = ? WHERE userid = ?"
	_, err := r.db.Exec(query, experience, userID)
	return err
}

// UpdateLevel updates player level
func (r *mysqlPlayerRepository) UpdateLevel(userID, level int) error {
	query := "UPDATE playerinfo SET level = ? WHERE userid = ?"
	_, err := r.db.Exec(query, level, userID)
	return err
}

// UpdateBloodEnergy updates player blood energy
func (r *mysqlPlayerRepository) UpdateBloodEnergy(userID, bloodEnergy int) error {
	query := "UPDATE playerinfo SET bloodenergy = ? WHERE userid = ?"
	_, err := r.db.Exec(query, bloodEnergy, userID)
	return err
}