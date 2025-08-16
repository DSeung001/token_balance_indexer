package config

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"os"
)

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	DSN string
}

// NewDatabaseConfig creates a new database config
func NewDatabaseConfig() *DatabaseConfig {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		panic("DATABASE_URL is empty")
	}
	return &DatabaseConfig{DSN: dsn}
}

// Connect establishes database connection
func (c *DatabaseConfig) Connect() (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(c.DSN), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}
