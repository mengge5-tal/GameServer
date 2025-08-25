package repository

import (
	"GameServer/internal/domain/entity"
	"GameServer/internal/domain/repository"
	"database/sql"
	"fmt"
)

// mysqlUnionExperienceRepository implements UnionExperienceRepository
type mysqlUnionExperienceRepository struct {
	db *sql.DB
}

// NewMySQLUnionExperienceRepository creates a new MySQL union experience repository
func NewMySQLUnionExperienceRepository(db *sql.DB) repository.UnionExperienceRepository {
	return &mysqlUnionExperienceRepository{db: db}
}

// GetByLevel retrieves union experience requirement by level
func (r *mysqlUnionExperienceRepository) GetByLevel(level int) (*entity.UnionExperience, error) {
	ue := &entity.UnionExperience{}
	query := "SELECT unionlevel, experience FROM unionexperience WHERE unionlevel = ?"
	
	err := r.db.QueryRow(query, level).Scan(&ue.UnionLevel, &ue.Experience)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("获取工会等级经验信息失败: %w", err)
	}
	return ue, nil
}

// GetAllLevels retrieves all union experience levels
func (r *mysqlUnionExperienceRepository) GetAllLevels() ([]*entity.UnionExperience, error) {
	query := "SELECT unionlevel, experience FROM unionexperience ORDER BY unionlevel ASC"
	
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("获取所有工会等级经验信息失败: %w", err)
	}
	defer rows.Close()
	
	var experiences []*entity.UnionExperience
	for rows.Next() {
		ue := &entity.UnionExperience{}
		err := rows.Scan(&ue.UnionLevel, &ue.Experience)
		if err != nil {
			return nil, fmt.Errorf("扫描工会经验数据失败: %w", err)
		}
		experiences = append(experiences, ue)
	}
	
	return experiences, nil
}

// GetNextLevel retrieves the next level experience requirement
func (r *mysqlUnionExperienceRepository) GetNextLevel(currentLevel int) (*entity.UnionExperience, error) {
	ue := &entity.UnionExperience{}
	query := "SELECT unionlevel, experience FROM unionexperience WHERE unionlevel > ? ORDER BY unionlevel ASC LIMIT 1"
	
	err := r.db.QueryRow(query, currentLevel).Scan(&ue.UnionLevel, &ue.Experience)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No next level (max level reached)
		}
		return nil, fmt.Errorf("获取下一等级经验信息失败: %w", err)
	}
	return ue, nil
}