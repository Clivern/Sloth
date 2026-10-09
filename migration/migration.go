// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package migration

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// Migration is one DB migration (version + up/down).
type Migration struct {
	Version     string
	Description string
	Up          func(*sql.DB) error
	Down        func(*sql.DB) error
}

// Manager runs and tracks DB migrations.
type Manager struct {
	db         *sql.DB
	migrations []Migration
}

// NewManager creates a new migration manager
func NewManager(db *sql.DB) *Manager {
	return &Manager{
		db:         db,
		migrations: []Migration{},
	}
}

// Register adds a migration to the manager
func (m *Manager) Register(migration Migration) {
	m.migrations = append(m.migrations, migration)
}

// createMigrationsTable creates the migrations tracking table if it doesn't exist
func (m *Manager) createMigrationsTable() error {
	if err := exec(m.db, `
		CREATE TABLE IF NOT EXISTS migrations (
			id UUID PRIMARY KEY,
			version VARCHAR(255) NOT NULL UNIQUE,
			description TEXT,
			applied_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')
		);
		CREATE INDEX IF NOT EXISTS idx_migrations_version ON migrations(version)`); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	return nil
}

// isApplied checks if a migration version has been applied
func (m *Manager) isApplied(version string) (bool, error) {
	var count int
	err := m.db.QueryRow("SELECT COUNT(*) FROM migrations WHERE version = $1", version).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check migration status: %w", err)
	}

	return count > 0, nil
}

// recordMigration records a migration as applied
func (m *Manager) recordMigration(version, description string) error {
	id, err := uuid.NewRandom()
	if err != nil {
		return fmt.Errorf("failed to generate migration id: %w", err)
	}

	if err := exec(
		m.db,
		"INSERT INTO migrations (id, version, description, applied_at) VALUES ($1, $2, $3, $4)",
		id.String(),
		version,
		description,
		time.Now().UTC(),
	); err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	return nil
}

// removeMigration removes a migration record
func (m *Manager) removeMigration(version string) error {
	err := exec(m.db, "DELETE FROM migrations WHERE version = $1", version)
	if err != nil {
		return fmt.Errorf("failed to remove migration record: %w", err)
	}

	return nil
}

// Up runs all pending migrations
func (m *Manager) Up() error {
	err := m.createMigrationsTable()
	if err != nil {
		return err
	}

	sort.Slice(m.migrations, func(i, j int) bool {
		return m.migrations[i].Version < m.migrations[j].Version
	})

	appliedCount := 0

	for _, migration := range m.migrations {
		applied, err := m.isApplied(migration.Version)
		if err != nil {
			return err
		}

		if applied {
			log.Debug().
				Str("version", migration.Version).
				Msg("Migration already applied, skipping")
			continue
		}

		log.Info().
			Str("version", migration.Version).
			Str("description", migration.Description).
			Msg("Running migration")

		tx, err := m.db.Begin()
		if err != nil {
			return fmt.Errorf("failed to start transaction: %w", err)
		}

		err = migration.Up(m.db)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("migration %s failed: %w", migration.Version, err)
		}

		err = m.recordMigration(migration.Version, migration.Description)
		if err != nil {
			tx.Rollback()
			return err
		}

		err = tx.Commit()
		if err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}

		log.Info().
			Str("version", migration.Version).
			Msg("Migration applied successfully")

		appliedCount++
	}

	if appliedCount == 0 {
		log.Info().
			Msg("No pending migrations to apply")
	} else {
		log.Info().
			Int("count", appliedCount).
			Msg("Migrations applied successfully")
	}

	return nil
}

// Down rolls back the last migration
func (m *Manager) Down() error {
	err := m.createMigrationsTable()
	if err != nil {
		return err
	}

	var version, description string
	err = m.db.QueryRow(`
		SELECT version, description
		FROM migrations
		ORDER BY applied_at DESC, id DESC
		LIMIT 1
	`).Scan(&version, &description)

	if err == sql.ErrNoRows {
		log.Info().
			Msg("No migrations to roll back")
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to get last migration: %w", err)
	}

	var migration *Migration
	for i := range m.migrations {
		if m.migrations[i].Version == version {
			migration = &m.migrations[i]
			break
		}
	}

	if migration == nil {
		return fmt.Errorf("migration %s not found in registered migrations", version)
	}
	if migration.Down == nil {
		return fmt.Errorf("migration %s has no down function", version)
	}

	log.Info().
		Str("version", version).
		Str("description", description).
		Msg("Rolling back migration")

	tx, err := m.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	err = migration.Down(m.db)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("rollback of migration %s failed: %w", version, err)
	}

	err = m.removeMigration(version)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Info().
		Str("version", version).
		Msg("Migration rolled back successfully")

	return nil
}

// Status shows the status of all migrations
func (m *Manager) Status() error {
	err := m.createMigrationsTable()
	if err != nil {
		return err
	}

	sort.Slice(m.migrations, func(i, j int) bool {
		return m.migrations[i].Version < m.migrations[j].Version
	})

	log.Info().
		Msg("Migration Status:")
	log.Info().
		Msg("==================")

	for _, migration := range m.migrations {
		applied, err := m.isApplied(migration.Version)
		if err != nil {
			return err
		}

		status := "Pending"
		if applied {
			status = "Applied"
		}

		log.Info().
			Str("version", migration.Version).
			Str("description", migration.Description).
			Str("status", status).
			Msg("")
	}

	return nil
}

// exec runs SQL statements within a migration.
func exec(db *sql.DB, query string, args ...any) error {
	var version string
	err := db.QueryRow("SELECT version()").Scan(&version)
	if err != nil {
		return fmt.Errorf("detect database driver: %w", err)
	}
	if !strings.Contains(strings.ToLower(version), "postgresql") {
		return fmt.Errorf("unsupported database driver")
	}

	_, err = db.Exec(query, args...)

	return err
}
