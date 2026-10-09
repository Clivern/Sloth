// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/rs/zerolog/log"
)

// Connection is the database connection.
type Connection struct {
	DB     *sql.DB
	Driver string
}

// DatabaseConfig holds database connection configuration.
type DatabaseConfig struct {
	Driver          string
	Host            string
	Port            int
	Username        string
	Password        string
	Database        string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime int
}

// NewConnection creates a new database connection based on the driver
func NewConnection(config DatabaseConfig) (*Connection, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		config.Host,
		config.Port,
		config.Username,
		config.Password,
		config.Database,
	)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	if config.MaxOpenConns > 0 {
		db.SetMaxOpenConns(config.MaxOpenConns)
	}
	if config.MaxIdleConns > 0 {
		db.SetMaxIdleConns(config.MaxIdleConns)
	}
	if config.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(time.Duration(config.ConnMaxLifetime) * time.Second)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Info().
		Str("driver", config.Driver).
		Str("host", config.Host).
		Int("port", config.Port).
		Str("database", config.Database).
		Msg("Database connection established")

	return &Connection{
		DB:     db,
		Driver: config.Driver,
	}, nil
}

// Close closes the database connection
func (c *Connection) Close() error {
	log.Info().
		Msg("Closing database connection")

	return c.DB.Close()
}

// Ping checks if the database connection is alive
func (c *Connection) Ping() error {
	return c.DB.Ping()
}
