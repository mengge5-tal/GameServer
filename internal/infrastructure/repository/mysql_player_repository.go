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

// GetPlayerRanking retrieves player ranking by type with limit
func (r *mysqlPlayerRepository) GetPlayerRanking(rankType string, limit int) ([]*entity.PlayerRankingEntry, error) {
	var query string
	
	// Build query based on rank type
	switch rankType {
	case "level":
		query = `SELECT p.userid, u.username, p.level as value
				 FROM playerinfo p 
				 INNER JOIN user u ON p.userid = u.userid
				 ORDER BY p.level DESC 
				 LIMIT ?`
	case "experience":
		query = `SELECT p.userid, u.username, p.experience as value
				 FROM playerinfo p 
				 INNER JOIN user u ON p.userid = u.userid
				 ORDER BY p.experience DESC 
				 LIMIT ?`
	case "gamelevel":
		query = `SELECT p.userid, u.username, p.gamelevel as value
				 FROM playerinfo p 
				 INNER JOIN user u ON p.userid = u.userid
				 ORDER BY p.gamelevel DESC 
				 LIMIT ?`
	case "bloodenergy":
		query = `SELECT p.userid, u.username, p.bloodenergy as value
				 FROM playerinfo p 
				 INNER JOIN user u ON p.userid = u.userid
				 ORDER BY p.bloodenergy DESC 
				 LIMIT ?`
	default:
		return nil, sql.ErrNoRows
	}
	
	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rankings []*entity.PlayerRankingEntry
	position := 1
	for rows.Next() {
		var userID, value int
		var username string
		
		err := rows.Scan(&userID, &username, &value)
		if err != nil {
			return nil, err
		}
		
		ranking := &entity.PlayerRankingEntry{
			UserID:   userID,
			Username: username,
			Value:    value,
			Position: position,
		}
		rankings = append(rankings, ranking)
		position++
	}

	return rankings, rows.Err()
}

// GetUserRank retrieves a specific user's rank for a given rank type
func (r *mysqlPlayerRepository) GetUserRank(userID int, rankType string) (*entity.PlayerRankingEntry, error) {
	var valueQuery string
	var countQuery string
	
	// Build queries based on rank type
	switch rankType {
	case "level":
		valueQuery = `SELECT p.level, u.username FROM playerinfo p 
					  INNER JOIN user u ON p.userid = u.userid 
					  WHERE p.userid = ?`
		countQuery = `SELECT COUNT(*) + 1 FROM playerinfo WHERE level > 
					  (SELECT level FROM playerinfo WHERE userid = ?)`
	case "experience":
		valueQuery = `SELECT p.experience, u.username FROM playerinfo p 
					  INNER JOIN user u ON p.userid = u.userid 
					  WHERE p.userid = ?`
		countQuery = `SELECT COUNT(*) + 1 FROM playerinfo WHERE experience > 
					  (SELECT experience FROM playerinfo WHERE userid = ?)`
	case "gamelevel":
		valueQuery = `SELECT p.gamelevel, u.username FROM playerinfo p 
					  INNER JOIN user u ON p.userid = u.userid 
					  WHERE p.userid = ?`
		countQuery = `SELECT COUNT(*) + 1 FROM playerinfo WHERE gamelevel > 
					  (SELECT gamelevel FROM playerinfo WHERE userid = ?)`
	case "bloodenergy":
		valueQuery = `SELECT p.bloodenergy, u.username FROM playerinfo p 
					  INNER JOIN user u ON p.userid = u.userid 
					  WHERE p.userid = ?`
		countQuery = `SELECT COUNT(*) + 1 FROM playerinfo WHERE bloodenergy > 
					  (SELECT bloodenergy FROM playerinfo WHERE userid = ?)`
	default:
		return nil, sql.ErrNoRows
	}
	
	// Get user's value and username
	var value int
	var username string
	err := r.db.QueryRow(valueQuery, userID).Scan(&value, &username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	// Calculate user's position
	var position int
	err = r.db.QueryRow(countQuery, userID).Scan(&position)
	if err != nil {
		return nil, err
	}
	
	return &entity.PlayerRankingEntry{
		UserID:   userID,
		Username: username,
		Value:    value,
		Position: position,
	}, nil
}