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
	query := "SELECT id, userid, today, normal, elite, boss, count FROM killcount WHERE userid = ? AND today = ?"
	err := r.db.QueryRow(query, userID, date).Scan(
		&killCount.ID, &killCount.UserID, &killCount.Today,
		&killCount.Normal, &killCount.Elite, &killCount.Boss, &killCount.Count,
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
	// Calculate total count before inserting
	killCount.CalculateTotalKillCount()
	
	query := "INSERT INTO killcount (userid, today, normal, elite, boss, count) VALUES (?, ?, ?, ?, ?, ?)"
	result, err := r.db.Exec(query, 
		killCount.UserID, killCount.Today, killCount.Normal, 
		killCount.Elite, killCount.Boss, killCount.Count,
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
	// Calculate total count before updating
	killCount.CalculateTotalKillCount()
	
	query := "UPDATE killcount SET normal = ?, elite = ?, boss = ?, count = ? WHERE id = ?"
	_, err := r.db.Exec(query, 
		killCount.Normal, killCount.Elite, killCount.Boss, killCount.Count, killCount.ID,
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

// GetKillRanking retrieves kill count ranking top N players
func (r *mysqlKillCountRepository) GetKillRanking(limit int) ([]*entity.KillRankingEntry, error) {
	if limit <= 0 {
		limit = 100
	}
	
	query := `
		SELECT 
			k.userid, 
			u.username, 
			p.level, 
			k.count,
			RANK() OVER (ORDER BY k.count DESC) as rank_position
		FROM killcount k
		JOIN user u ON k.userid = u.userid
		JOIN playerinfo p ON k.userid = p.userid
		WHERE k.today = ?
		ORDER BY k.count DESC
		LIMIT ?
	`
	
	today := time.Now().Format("2006-01-02")
	rows, err := r.db.Query(query, today, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var rankings []*entity.KillRankingEntry
	for rows.Next() {
		entry := &entity.KillRankingEntry{}
		err := rows.Scan(&entry.UserID, &entry.Username, &entry.Level, &entry.Count, &entry.Rank)
		if err != nil {
			return nil, err
		}
		rankings = append(rankings, entry)
	}
	
	return rankings, nil
}

// GetUserKillRank retrieves specific user's kill count ranking
func (r *mysqlKillCountRepository) GetUserKillRank(userID int) (*entity.KillRankingEntry, error) {
	query := `
		SELECT 
			k.userid, 
			u.username, 
			p.level, 
			k.count,
			(
				SELECT COUNT(*) + 1 
				FROM killcount k2 
				WHERE k2.today = ? AND k2.count > k.count
			) as rank_position
		FROM killcount k
		JOIN user u ON k.userid = u.userid
		JOIN playerinfo p ON k.userid = p.userid
		WHERE k.userid = ? AND k.today = ?
	`
	
	today := time.Now().Format("2006-01-02")
	entry := &entity.KillRankingEntry{}
	err := r.db.QueryRow(query, today, userID, today).Scan(
		&entry.UserID, &entry.Username, &entry.Level, &entry.Count, &entry.Rank,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			// User has no kill count record, return with rank 0
			return &entity.KillRankingEntry{
				UserID:   userID,
				Username: "",
				Level:    0,
				Count:    0,
				Rank:     0,
			}, nil
		}
		return nil, err
	}
	
	return entry, nil
}