package repository

import (
	"database/sql"
	"fmt"
	"time"
	"GameServer/internal/domain/entity"
	"GameServer/internal/domain/repository"
)

// mysqlKillCountRepository implements KillCountRepository
type mysqlKillCountRepository struct {
	db *sql.DB
}

// NewMySQLKillCountRepository creates a new MySQL kill count repository
func NewMySQLKillCountRepository(db *sql.DB) repository.KillCountRepository {
	return &mysqlKillCountRepository{db: db}
}

// GetByUserIDAndDate retrieves kill count by user ID and date
func (r *mysqlKillCountRepository) GetByUserIDAndDate(userID int, date string) (*entity.KillCount, error) {
	killCount := &entity.KillCount{}
	query := "SELECT id, userid, today, normal, elite, boss FROM killcount WHERE userid = ? AND today = ?"
	err := r.db.QueryRow(query, userID, date).Scan(
		&killCount.ID, &killCount.UserID, &killCount.Today,
		&killCount.Normal, &killCount.Elite, &killCount.Boss,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return killCount, nil
}

// Create creates new kill count record
func (r *mysqlKillCountRepository) Create(killCount *entity.KillCount) error {
	query := "INSERT INTO killcount (userid, today, normal, elite, boss) VALUES (?, ?, ?, ?, ?)"
	result, err := r.db.Exec(query, 
		killCount.UserID, killCount.Today, killCount.Normal, 
		killCount.Elite, killCount.Boss,
	)
	if err != nil {
		return err
	}
	
	// Get the auto-generated ID
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	killCount.ID = int(id)
	
	return nil
}

// Update updates existing kill count record
func (r *mysqlKillCountRepository) Update(killCount *entity.KillCount) error {
	query := "UPDATE killcount SET normal = ?, elite = ?, boss = ? WHERE id = ?"
	_, err := r.db.Exec(query, 
		killCount.Normal, killCount.Elite, killCount.Boss, killCount.ID,
	)
	return err
}

// Delete deletes kill count record by ID
func (r *mysqlKillCountRepository) Delete(id int) error {
	query := "DELETE FROM killcount WHERE id = ?"
	_, err := r.db.Exec(query, id)
	return err
}

// IncrementKill increments kill count for specific monster type
func (r *mysqlKillCountRepository) IncrementKill(userID int, date string, monsterType string, count int) error {
	// First, try to get existing record
	existing, err := r.GetByUserIDAndDate(userID, date)
	if err != nil {
		return err
	}
	
	// If no record exists, create a new one
	if existing == nil {
		newRecord := &entity.KillCount{
			UserID: userID,
			Today:  date,
			Normal: 0,
			Elite:  0,
			Boss:   0,
		}
		
		// Set the appropriate field based on monster type
		switch monsterType {
		case "normal":
			newRecord.Normal = count
		case "elite":
			newRecord.Elite = count
		case "boss":
			newRecord.Boss = count
		default:
			return fmt.Errorf("invalid monster type: %s", monsterType)
		}
		
		return r.Create(newRecord)
	}
	
	// Update existing record
	switch monsterType {
	case "normal":
		existing.Normal += count
	case "elite":
		existing.Elite += count
	case "boss":
		existing.Boss += count
	default:
		return fmt.Errorf("invalid monster type: %s", monsterType)
	}
	
	return r.Update(existing)
}

// ResetAllToday resets all users' kill counts for today
func (r *mysqlKillCountRepository) ResetAllToday() error {
	today := time.Now().Format("2006-01-02")
	query := "DELETE FROM killcount WHERE today = ?"
	_, err := r.db.Exec(query, today)
	return err
}

// GetTodayKillCount retrieves today's kill count for a user
func (r *mysqlKillCountRepository) GetTodayKillCount(userID int) (*entity.KillCount, error) {
	today := time.Now().Format("2006-01-02")
	return r.GetByUserIDAndDate(userID, today)
}