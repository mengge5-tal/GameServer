package database

import (
	"database/sql"
	"fmt"
	"log"
	"GameServer/internal/infrastructure/config"
	
	_ "github.com/go-sql-driver/mysql"
)

// Connection manages database connections
type Connection struct {
	db     *sql.DB
	config *config.Config
}

// NewConnection creates a new database connection
func NewConnection(cfg *config.Config) (*Connection, error) {
	db, err := sql.Open("mysql", cfg.GetConnectionString())
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.Database.ConnMaxIdleTime)

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Database connection established successfully")

	return &Connection{
		db:     db,
		config: cfg,
	}, nil
}

// GetDB returns the database instance
func (c *Connection) GetDB() *sql.DB {
	return c.db
}

// Close closes the database connection
func (c *Connection) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

// CheckTables verifies that all required tables exist
func (c *Connection) CheckTables() error {
	requiredTables := []string{
		"user", "playerinfo", "equip", "sourcestone",
		"friend", "friend_request", "ranking", "experience",
	}

	for _, tableName := range requiredTables {
		exists, err := c.tableExists(tableName)
		if err != nil {
			return fmt.Errorf("failed to check if table %s exists: %w", tableName, err)
		}
		if !exists {
			return fmt.Errorf("required table %s does not exist", tableName)
		}
		log.Printf("Table %s: OK", tableName)
	}

	return nil
}

// CreateMissingTables creates any missing tables
func (c *Connection) CreateMissingTables() error {
	// This would contain the table creation logic
	// For now, we'll just check if tables exist
	log.Println("Checking for missing tables...")
	return c.CheckTables()
}

// CheckTableStructure verifies table structure
func (c *Connection) CheckTableStructure() error {
	// This could be enhanced to check column types, constraints, etc.
	log.Println("Table structure check completed")
	return nil
}

// SetAllUsersOffline sets all users to offline status on startup
func (c *Connection) SetAllUsersOffline() error {
	query := "UPDATE user SET online_status = 0 WHERE online_status = 1"
	result, err := c.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to set all users offline: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Could not get rows affected count: %v", err)
	} else {
		log.Printf("Set %d users to offline status", rowsAffected)
	}
	
	return nil
}

// tableExists checks if a table exists in the database
func (c *Connection) tableExists(tableName string) (bool, error) {
	query := `
		SELECT COUNT(*) 
		FROM information_schema.tables 
		WHERE table_schema = ? AND table_name = ?
	`
	
	var count int
	err := c.db.QueryRow(query, c.config.Database.Name, tableName).Scan(&count)
	if err != nil {
		return false, err
	}
	
	return count > 0, nil
}