package auth

import (
	"fmt"
	"os"
	"path/filepath"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DatabaseType represents the type of database
type DatabaseType string

const (
	DatabaseTypeSQLite    DatabaseType = "sqlite"
	DatabaseTypePostgres  DatabaseType = "postgres"
	DatabaseTypeFile      DatabaseType = "file"
)

// OpenDatabase opens a database connection based on the database type
func OpenDatabase(dbType DatabaseType, dsn string) (*gorm.DB, error) {
	var dialector gorm.Dialector

	switch dbType {
	case DatabaseTypeSQLite:
		// Ensure directory exists for SQLite
		if dsn != ":memory:" {
			dir := filepath.Dir(dsn)
			if dir != "." && dir != "" {
				if err := os.MkdirAll(dir, 0755); err != nil {
					return nil, fmt.Errorf("failed to create database directory: %w", err)
				}
			}
		}
		dialector = sqlite.Open(dsn)

	case DatabaseTypePostgres:
		dialector = postgres.Open(dsn)

	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}

	// Configure GORM
	config := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Change to logger.Info for SQL logging
	}

	db, err := gorm.Open(dialector, config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying SQL DB for connection pool settings
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	return db, nil
}

// NewUserStore creates a user store based on the database type
func NewUserStore(dbType DatabaseType, dsn string, dataDir string) (UserStore, error) {
	switch dbType {
	case DatabaseTypeSQLite, DatabaseTypePostgres:
		db, err := OpenDatabase(dbType, dsn)
		if err != nil {
			return nil, err
		}
		return NewDBUserStore(db)

	case DatabaseTypeFile:
		return NewFileUserStore(dataDir)

	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}
}

