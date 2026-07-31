package database

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	// Register supported database drivers
	_ "github.com/go-sql-driver/mysql" // MySQL and MariaDB
	_ "modernc.org/sqlite"             // Pure Go SQLite
)

// DBWrapper wraps sql.DB to provide simpler access
type DBWrapper struct {
	*sql.DB
}

// Config for database connection
type Config struct {
	Driver          string
	DSN             string // Data Source Name
	MaxConns        int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// Connect initializes the database connection
func Connect(cfg Config) (*DBWrapper, error) {
	dsn := cfg.DSN

	// Set critical PRAGMAs for SQLite to handle concurrency properly
	if cfg.Driver == "sqlite" {
		// Only append if not already configured by user
		if !strings.Contains(dsn, "_pragma") {
			if strings.Contains(dsn, "?") {
				dsn += "&_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=foreign_keys(ON)"
			} else {
				dsn += "?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=foreign_keys(ON)"
			}
		}
	}

	db, err := sql.Open(cfg.Driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	if cfg.MaxConns > 0 {
		db.SetMaxOpenConns(cfg.MaxConns)
	}
	if cfg.MaxIdleConns > 0 {
		db.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	return &DBWrapper{DB: db}, nil
}
