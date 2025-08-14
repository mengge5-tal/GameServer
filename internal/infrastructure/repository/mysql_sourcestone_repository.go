package repository

import (
	"database/sql"
	"GameServer/internal/domain/entity"
	"GameServer/internal/domain/repository"
)

// mysqlSourceStoneRepository implements SourceStoneRepository
type mysqlSourceStoneRepository struct {
	db *sql.DB
}

// NewMySQLSourceStoneRepository creates a new MySQL source stone repository
func NewMySQLSourceStoneRepository(db *sql.DB) repository.SourceStoneRepository {
	return &mysqlSourceStoneRepository{db: db}
}

// GetByID retrieves source stone by ID
func (r *mysqlSourceStoneRepository) GetByID(sourcestoneID int) (*entity.SourceStone, error) {
	stone := &entity.SourceStone{}
	query := `SELECT sourcestoneid, sourcestonename, sourcestonequality, sourcestoneeffect 
			  FROM sourcestone WHERE sourcestoneid = ?`
	
	err := r.db.QueryRow(query, sourcestoneID).Scan(
		&stone.SourceStoneID, &stone.SourceStoneName, 
		&stone.SourceStoneQuality, &stone.SourceStoneEffect,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return stone, nil
}

// GetAll retrieves all source stones
func (r *mysqlSourceStoneRepository) GetAll() ([]*entity.SourceStone, error) {
	query := `SELECT sourcestoneid, sourcestonename, sourcestonequality, sourcestoneeffect 
			  FROM sourcestone ORDER BY sourcestoneid ASC`
	
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sourceStones []*entity.SourceStone
	for rows.Next() {
		stone := &entity.SourceStone{}
		err := rows.Scan(
			&stone.SourceStoneID, &stone.SourceStoneName,
			&stone.SourceStoneQuality, &stone.SourceStoneEffect,
		)
		if err != nil {
			return nil, err
		}
		sourceStones = append(sourceStones, stone)
	}

	return sourceStones, rows.Err()
}