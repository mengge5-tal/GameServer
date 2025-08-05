package repository

import (
	"database/sql"
	"GameServer/internal/domain/entity"
	"GameServer/internal/domain/repository"
)

// mysqlRankingRepository implements RankingRepository
type mysqlRankingRepository struct {
	db *sql.DB
}

// NewMySQLRankingRepository creates a new MySQL ranking repository
func NewMySQLRankingRepository(db *sql.DB) repository.RankingRepository {
	return &mysqlRankingRepository{db: db}
}

// GetRankingByType retrieves ranking by type with limit
func (r *mysqlRankingRepository) GetRankingByType(rankType string, limit int) ([]*entity.Ranking, error) {
	query := `SELECT id, userid, rank_type, rank_value, rank_position, updated_at 
			  FROM ranking WHERE rank_type = ? ORDER BY rank_position ASC LIMIT ?`
	
	rows, err := r.db.Query(query, rankType, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rankings []*entity.Ranking
	for rows.Next() {
		ranking := &entity.Ranking{}
		err := rows.Scan(
			&ranking.ID, &ranking.UserID, &ranking.RankType,
			&ranking.RankValue, &ranking.RankPosition, &ranking.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		rankings = append(rankings, ranking)
	}

	return rankings, rows.Err()
}

// UpdateUserRanking updates or creates user ranking
func (r *mysqlRankingRepository) UpdateUserRanking(userID int, rankType string, value int) error {
	// Use INSERT ... ON DUPLICATE KEY UPDATE for upsert
	query := `INSERT INTO ranking (userid, rank_type, rank_value, rank_position) 
			  VALUES (?, ?, ?, 0) 
			  ON DUPLICATE KEY UPDATE rank_value = ?, updated_at = CURRENT_TIMESTAMP`
	
	_, err := r.db.Exec(query, userID, rankType, value, value)
	return err
}

// GetUserRanking retrieves user's ranking for a specific type
func (r *mysqlRankingRepository) GetUserRanking(userID int, rankType string) (*entity.Ranking, error) {
	ranking := &entity.Ranking{}
	query := `SELECT id, userid, rank_type, rank_value, rank_position, updated_at 
			  FROM ranking WHERE userid = ? AND rank_type = ?`
	
	err := r.db.QueryRow(query, userID, rankType).Scan(
		&ranking.ID, &ranking.UserID, &ranking.RankType,
		&ranking.RankValue, &ranking.RankPosition, &ranking.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return ranking, nil
}

// RefreshRankings recalculates and updates rank positions for a specific type
func (r *mysqlRankingRepository) RefreshRankings(rankType string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Get all rankings for this type ordered by value descending
	query := `SELECT id FROM ranking WHERE rank_type = ? ORDER BY rank_value DESC`
	rows, err := tx.Query(query, rankType)
	if err != nil {
		return err
	}
	defer rows.Close()

	// Update rank positions
	position := 1
	updateQuery := "UPDATE ranking SET rank_position = ? WHERE id = ?"
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return err
		}
		
		if _, err := tx.Exec(updateQuery, position, id); err != nil {
			return err
		}
		position++
	}

	return tx.Commit()
}