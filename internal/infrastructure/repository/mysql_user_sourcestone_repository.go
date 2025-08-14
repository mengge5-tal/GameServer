package repository

import (
	"GameServer/internal/domain/entity"
	"GameServer/internal/domain/repository"
	"database/sql"
)

// mysqlUserSourceStoneRepository implements UserSourceStoneRepository
type mysqlUserSourceStoneRepository struct {
	db *sql.DB
}

// NewMySQLUserSourceStoneRepository creates a new MySQL user source stone repository
func NewMySQLUserSourceStoneRepository(db *sql.DB) repository.UserSourceStoneRepository {
	return &mysqlUserSourceStoneRepository{db: db}
}

// GetByID retrieves user source stone ownership by ID
func (r *mysqlUserSourceStoneRepository) GetByID(id int) (*entity.UserSourceStone, error) {
	userSourceStone := &entity.UserSourceStone{}
	query := "SELECT id, userid, sourcestoneid, sourcestonecount FROM user_sourcestone WHERE id = ?"

	err := r.db.QueryRow(query, id).Scan(
		&userSourceStone.ID, &userSourceStone.UserID, &userSourceStone.SourceStoneID, &userSourceStone.SourceStoneCount,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return userSourceStone, nil
}

// GetByUserID retrieves all source stones owned by a user
func (r *mysqlUserSourceStoneRepository) GetByUserID(userID int) ([]*entity.UserSourceStone, error) {
	query := "SELECT id, userid, sourcestoneid, sourcestonecount FROM user_sourcestone WHERE userid = ? ORDER BY id ASC"

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userSourceStones []*entity.UserSourceStone
	for rows.Next() {
		userSourceStone := &entity.UserSourceStone{}
		err := rows.Scan(
			&userSourceStone.ID, &userSourceStone.UserID, &userSourceStone.SourceStoneID, &userSourceStone.SourceStoneCount,
		)
		if err != nil {
			return nil, err
		}
		userSourceStones = append(userSourceStones, userSourceStone)
	}

	return userSourceStones, rows.Err()
}

// GetByUserAndSourceStone retrieves user source stone ownership by user ID and source stone ID
func (r *mysqlUserSourceStoneRepository) GetByUserAndSourceStone(userID, sourcestoneID int) (*entity.UserSourceStone, error) {
	userSourceStone := &entity.UserSourceStone{}
	query := "SELECT id, userid, sourcestoneid, sourcestonecount FROM user_sourcestone WHERE userid = ? AND sourcestoneid = ?"

	err := r.db.QueryRow(query, userID, sourcestoneID).Scan(
		&userSourceStone.ID, &userSourceStone.UserID, &userSourceStone.SourceStoneID, &userSourceStone.SourceStoneCount,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return userSourceStone, nil
}

// Create creates new user source stone ownership
func (r *mysqlUserSourceStoneRepository) Create(userSourceStone *entity.UserSourceStone) error {
	query := "INSERT INTO user_sourcestone (userid, sourcestoneid, sourcestonecount) VALUES (?, ?, ?)"

	result, err := r.db.Exec(query, userSourceStone.UserID, userSourceStone.SourceStoneID, userSourceStone.SourceStoneCount)
	if err != nil {
		return err
	}

	// Get the auto-generated ID
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	userSourceStone.ID = int(id)

	return nil
}

// Update updates user source stone ownership
func (r *mysqlUserSourceStoneRepository) Update(userSourceStone *entity.UserSourceStone) error {
	query := "UPDATE user_sourcestone SET sourcestonecount = ? WHERE id = ?"
	_, err := r.db.Exec(query, userSourceStone.SourceStoneCount, userSourceStone.ID)
	return err
}

// Delete deletes user source stone ownership by ID
func (r *mysqlUserSourceStoneRepository) Delete(id int) error {
	query := "DELETE FROM user_sourcestone WHERE id = ?"
	_, err := r.db.Exec(query, id)
	return err
}

// DeleteByUserAndSourceStone deletes user source stone ownership by user ID and source stone ID
func (r *mysqlUserSourceStoneRepository) DeleteByUserAndSourceStone(userID, sourcestoneID int) error {
	query := "DELETE FROM user_sourcestone WHERE userid = ? AND sourcestoneid = ?"
	_, err := r.db.Exec(query, userID, sourcestoneID)
	return err
}

// UserOwnsSourceStone checks if a user owns a specific source stone
func (r *mysqlUserSourceStoneRepository) UserOwnsSourceStone(userID, sourcestoneID int) (bool, error) {
	var count int
	query := "SELECT COUNT(*) FROM user_sourcestone WHERE userid = ? AND sourcestoneid = ?"
	err := r.db.QueryRow(query, userID, sourcestoneID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}