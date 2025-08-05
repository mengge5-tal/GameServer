package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	Database  DatabaseConfig  `json:"database"`
	Server    ServerConfig    `json:"server"`
	WebSocket WebSocketConfig `json:"websocket"`
	Security  SecurityConfig  `json:"security"`
	Logging   LoggingConfig   `json:"logging"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host            string        `json:"host"`
	Port            string        `json:"port"`
	Name            string        `json:"name"`
	User            string        `json:"user"`
	Password        string        `json:"password"`
	MaxOpenConns    int           `json:"max_open_conns"`
	MaxIdleConns    int           `json:"max_idle_conns"`
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `json:"conn_max_idle_time"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

// WebSocketConfig holds WebSocket configuration
type WebSocketConfig struct {
	AllowedOrigins   []string      `json:"allowed_origins"`
	ReadBufferSize   int           `json:"read_buffer_size"`
	WriteBufferSize  int           `json:"write_buffer_size"`
	HandshakeTimeout time.Duration `json:"handshake_timeout"`
}

// SecurityConfig holds security configuration
type SecurityConfig struct {
	BcryptCost int `json:"bcrypt_cost"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string `json:"level"`
	Format string `json:"format"`
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	config := &Config{
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "rm-2zevr95ez9rrid70uho.mysql.rds.aliyuncs.com"),
			Port:            getEnv("DB_PORT", "3306"),
			Name:            getEnv("DB_NAME", "Vampire"),
			User:            getEnv("DB_USER", "wwk18255113901"),
			Password:        getEnv("DB_PASSWORD", "BaiChen123456+"),
			MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvDuration("DB_CONN_MAX_LIFETIME", "300s"),
			ConnMaxIdleTime: getEnvDuration("DB_CONN_MAX_IDLE_TIME", "60s"),
		},
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "localhost"),
			Port: getEnvInt("SERVER_PORT", 8080),
		},
		WebSocket: WebSocketConfig{
			AllowedOrigins:   getEnvStringArray("WS_ALLOWED_ORIGINS", []string{"http://localhost:3000", "http://localhost:8080"}),
			ReadBufferSize:   getEnvInt("WS_READ_BUFFER_SIZE", 1024),
			WriteBufferSize:  getEnvInt("WS_WRITE_BUFFER_SIZE", 1024),
			HandshakeTimeout: getEnvDuration("WS_HANDSHAKE_TIMEOUT", "10s"),
		},
		Security: SecurityConfig{
			BcryptCost: getEnvInt("BCRYPT_COST", 12),
		},
		Logging: LoggingConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Database validation
	if c.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if c.Database.Name == "" {
		return fmt.Errorf("database name is required")
	}
	if c.Database.User == "" {
		return fmt.Errorf("database user is required")
	}

	// Server validation
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("server port must be between 1 and 65535")
	}

	// Security validation
	if c.Security.BcryptCost < 4 || c.Security.BcryptCost > 31 {
		return fmt.Errorf("bcrypt cost must be between 4 and 31")
	}

	// Logging validation
	validLogLevels := []string{"debug", "info", "warn", "error"}
	if !contains(validLogLevels, c.Logging.Level) {
		return fmt.Errorf("log level must be one of: %s", strings.Join(validLogLevels, ", "))
	}

	return nil
}

// GetConnectionString returns the database connection string
func (c *Config) GetConnectionString() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.Database.User, c.Database.Password, c.Database.Host, c.Database.Port, c.Database.Name)
}

// Helper functions

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return fallback
}

func getEnvDuration(key, fallback string) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	if duration, err := time.ParseDuration(fallback); err == nil {
		return duration
	}
	return 5 * time.Minute // safe fallback
}

func getEnvStringArray(key string, fallback []string) []string {
	if value := os.Getenv(key); value != "" {
		parts := strings.Split(value, ",")
		for i, part := range parts {
			parts[i] = strings.TrimSpace(part)
		}
		return parts
	}
	return fallback
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
