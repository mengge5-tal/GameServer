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

// GetByUserID retrieves all source stones for a user
func (r *mysqlSourceStoneRepository) GetByUserID(userID int) ([]*entity.SourceStone, error) {
	query := `SELECT equipid, sourcetype, count, quality, userid 
			  FROM sourcestone WHERE userid = ?`
	
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sourceStones []*entity.SourceStone
	for rows.Next() {
		stone := &entity.SourceStone{}
		err := rows.Scan(
			&stone.EquipID, &stone.SourceType, &stone.Count,
			&stone.Quality, &stone.UserID,
		)
		if err != nil {
			return nil, err
		}
		sourceStones = append(sourceStones, stone)
	}

	return sourceStones, rows.Err()
}

// GetByEquipID retrieves source stone by equipment ID
func (r *mysqlSourceStoneRepository) GetByEquipID(equipID int) (*entity.SourceStone, error) {
	stone := &entity.SourceStone{}
	query := `SELECT equipid, sourcetype, count, quality, userid 
			  FROM sourcestone WHERE equipid = ?`
	
	err := r.db.QueryRow(query, equipID).Scan(
		&stone.EquipID, &stone.SourceType, &stone.Count,
		&stone.Quality, &stone.UserID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return stone, nil
}

// Create creates a new source stone
func (r *mysqlSourceStoneRepository) Create(sourceStone *entity.SourceStone) error {
	query := `INSERT INTO sourcestone (equipid, sourcetype, count, quality, userid) 
			  VALUES (?, ?, ?, ?, ?)`
	
	_, err := r.db.Exec(query,
		sourceStone.EquipID, sourceStone.SourceType, sourceStone.Count,
		sourceStone.Quality, sourceStone.UserID,
	)
	return err
}

// Update updates an existing source stone
func (r *mysqlSourceStoneRepository) Update(sourceStone *entity.SourceStone) error {
	query := `UPDATE sourcestone SET sourcetype = ?, count = ?, quality = ?, userid = ? 
			  WHERE equipid = ?`
	
	_, err := r.db.Exec(query,
		sourceStone.SourceType, sourceStone.Count, sourceStone.Quality,
		sourceStone.UserID, sourceStone.EquipID,
	)
	return err
}

// Delete deletes a source stone by equipment ID
func (r *mysqlSourceStoneRepository) Delete(equipID int) error {
	query := "DELETE FROM sourcestone WHERE equipid = ?"
	_, err := r.db.Exec(query, equipID)
	return err
}

// UpdateCount updates the count of a source stone
func (r *mysqlSourceStoneRepository) UpdateCount(equipID, count int) error {
	query := "UPDATE sourcestone SET count = ? WHERE equipid = ?"
	_, err := r.db.Exec(query, count, equipID)
	return err
}