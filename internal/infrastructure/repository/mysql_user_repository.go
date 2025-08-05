package repository

import (
	"database/sql"
	"GameServer/internal/domain/entity"
	"GameServer/internal/domain/repository"
)

// mysqlUserRepository implements UserRepository
type mysqlUserRepository struct {
	db *sql.DB
}

// NewMySQLUserRepository creates a new MySQL user repository
func NewMySQLUserRepository(db *sql.DB) repository.UserRepository {
	return &mysqlUserRepository{db: db}
}

// GetByID retrieves a user by ID
func (r *mysqlUserRepository) GetByID(id int) (*entity.User, error) {
	user := &entity.User{}
	query := "SELECT userid, username FROM user WHERE userid = ?"
	err := r.db.QueryRow(query, id).Scan(&user.ID, &user.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

// GetByUsername retrieves a user by username
func (r *mysqlUserRepository) GetByUsername(username string) (*entity.User, error) {
	user := &entity.User{}
	query := "SELECT userid, username FROM user WHERE username = ?"
	err := r.db.QueryRow(query, username).Scan(&user.ID, &user.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

// Create creates a new user
func (r *mysqlUserRepository) Create(user *entity.User) error {
	query := "INSERT INTO user (username, passward) VALUES (?, ?)"
	result, err := r.db.Exec(query, user.Username, user.Password)
	if err != nil {
		return err
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	
	user.ID = int(id)
	return nil
}

// Update updates an existing user
func (r *mysqlUserRepository) Update(user *entity.User) error {
	query := "UPDATE user SET username = ?, passward = ? WHERE userid = ?"
	_, err := r.db.Exec(query, user.Username, user.Password, user.ID)
	return err
}

// Delete deletes a user by ID
func (r *mysqlUserRepository) Delete(id int) error {
	query := "DELETE FROM user WHERE userid = ?"
	_, err := r.db.Exec(query, id)
	return err
}

// Exists checks if a user with the given username exists
func (r *mysqlUserRepository) Exists(username string) (bool, error) {
	var count int
	query := "SELECT COUNT(*) FROM user WHERE username = ?"
	err := r.db.QueryRow(query, username).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// VerifyCredentials verifies user credentials and returns user if valid
func (r *mysqlUserRepository) VerifyCredentials(username, password string) (*entity.User, error) {
	user := &entity.User{}
	var storedPassword string
	query := "SELECT userid, username, passward FROM user WHERE username = ?"
	err := r.db.QueryRow(query, username).Scan(&user.ID, &user.Username, &storedPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	// Store the hashed password for verification
	user.Password = storedPassword
	return user, nil
}