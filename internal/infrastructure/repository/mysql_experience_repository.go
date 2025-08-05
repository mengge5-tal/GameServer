package repository

import (
	"database/sql"
	"GameServer/internal/domain/entity"
	"GameServer/internal/domain/repository"
)

// mysqlExperienceRepository implements ExperienceRepository
type mysqlExperienceRepository struct {
	db *sql.DB
}

// NewMySQLExperienceRepository creates a new MySQL experience repository
func NewMySQLExperienceRepository(db *sql.DB) repository.ExperienceRepository {
	return &mysqlExperienceRepository{db: db}
}

// GetByLevel retrieves experience info by level
func (r *mysqlExperienceRepository) GetByLevel(level int) (*entity.Experience, error) {
	exp := &entity.Experience{}
	query := "SELECT level, value FROM experience WHERE level = ?"
	
	err := r.db.QueryRow(query, level).Scan(&exp.Level, &exp.Value)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return exp, nil
}

// GetAllLevels retrieves all experience levels
func (r *mysqlExperienceRepository) GetAllLevels() ([]*entity.Experience, error) {
	query := "SELECT level, value FROM experience ORDER BY level ASC"
	
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var experiences []*entity.Experience
	for rows.Next() {
		exp := &entity.Experience{}
		err := rows.Scan(&exp.Level, &exp.Value)
		if err != nil {
			return nil, err
		}
		experiences = append(experiences, exp)
	}

	return experiences, rows.Err()
}