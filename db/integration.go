// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package db

import (
	"database/sql"
	"time"
)

// Integration is the DB row for a workspace integration (webhook, Slack, etc.).
type Integration struct {
	Id          Id
	WorkspaceId Id
	Type        string
	Name        string
	Config      *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// IntegrationRepository is the interface for integration CRUD.
type IntegrationRepository interface {
	Create(integration *Integration) error
	GetById(id Id) (*Integration, error)
	Update(integration *Integration) error
	Delete(id Id) error
	ListByWorkspaceId(workspaceId Id, limit, offset int) ([]*Integration, error)
	CountByWorkspaceId(workspaceId Id) (int64, error)
}

// IntegrationRepositoryPostgres implements IntegrationRepository for PostgreSQL
type IntegrationRepositoryPostgres struct {
	db *sql.DB
}

// IntegrationMeta is a single row in the integrations_meta table.
type IntegrationMeta struct {
	Id            Id
	IntegrationId Id
	Key           string
	Value         string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// IntegrationMetaRepository is the interface for integration metadata CRUD.
type IntegrationMetaRepository interface {
	Create(id Id, key, value string) error
	Get(id Id, key string) (*IntegrationMeta, error)
	Update(id Id, key, value string) error
	Delete(id Id, key string) error
	ListByIntegrationId(id Id) ([]*IntegrationMeta, error)
	Upsert(id Id, key, value string) error
}

type IntegrationMetaRepositoryPostgres struct {
	db *sql.DB
}

// NewIntegrationRepository returns the repository for the current driver
func NewIntegrationRepository(db *sql.DB) IntegrationRepository {
	return &IntegrationRepositoryPostgres{db: db}
}

// --- Postgres ---

// Create inserts a row
func (r *IntegrationRepositoryPostgres) Create(integration *Integration) error {
	id, err := NewId()
	if err != nil {
		return err
	}

	integration.Id = id

	err = r.db.QueryRow(
		`INSERT INTO integrations (id, workspace_id, type, name, config)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at, updated_at`,
		integration.Id.String(),
		integration.WorkspaceId.String(),
		integration.Type,
		integration.Name,
		integration.Config,
	).Scan(&integration.CreatedAt, &integration.UpdatedAt)

	return err
}

// GetById returns a integration by Id
func (r *IntegrationRepositoryPostgres) GetById(id Id) (*Integration, error) {
	inv := &Integration{}
	err := r.db.QueryRow(
		`SELECT
			id, workspace_id, type, name, config, created_at, updated_at
		FROM integrations
		WHERE id = $1`,
		id.String(),
	).Scan(
		&inv.Id,
		&inv.WorkspaceId,
		&inv.Type,
		&inv.Name,
		&inv.Config,
		&inv.CreatedAt,
		&inv.UpdatedAt,
	)
	if isNotFound(err) {
		return nil, nil
	}

	return inv, err
}

// Update updates a integration
func (r *IntegrationRepositoryPostgres) Update(integration *Integration) error {
	_, err := r.db.Exec(
		`UPDATE integrations
		SET
			workspace_id = $1,
			type = $2,
			name = $3,
			config = $4,
			updated_at = $5
		WHERE id = $6`,
		integration.WorkspaceId.String(),
		integration.Type,
		integration.Name,
		integration.Config,
		time.Now().UTC(),
		integration.Id.String(),
	)

	return err
}

// Delete removes a integration
func (r *IntegrationRepositoryPostgres) Delete(id Id) error {
	_, err := r.db.Exec(
		`DELETE FROM integrations WHERE id = $1`,
		id.String(),
	)

	return err
}

// ListByWorkspaceId returns a list of integrations by workspace Id
func (r *IntegrationRepositoryPostgres) ListByWorkspaceId(workspaceId Id, limit, offset int) ([]*Integration, error) {
	rows, err := r.db.Query(
		`SELECT id, workspace_id, type, name, config, created_at, updated_at
		FROM integrations
		WHERE workspace_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`,
		workspaceId.String(),
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var list []*Integration
	for rows.Next() {
		inv := &Integration{}
		if err := rows.Scan(
			&inv.Id,
			&inv.WorkspaceId,
			&inv.Type,
			&inv.Name,
			&inv.Config,
			&inv.CreatedAt,
			&inv.UpdatedAt,
		); err != nil {
			return nil, err
		}

		list = append(list, inv)
	}

	return list, rows.Err()
}

// CountByWorkspaceId returns the total number of integrations by workspace Id
func (r *IntegrationRepositoryPostgres) CountByWorkspaceId(workspaceId Id) (int64, error) {
	var count int64
	err := r.db.QueryRow(
		`SELECT COUNT(*)
		FROM integrations
		WHERE workspace_id = $1`,
		workspaceId.String(),
	).Scan(&count)

	return count, err
}

// NewIntegrationMetaRepository returns the repository for integration metadata.
func NewIntegrationMetaRepository(db *sql.DB) IntegrationMetaRepository {
	return &IntegrationMetaRepositoryPostgres{db: db}
}

// Create inserts an integration metadata row.
func (r *IntegrationMetaRepositoryPostgres) Create(id Id, key, value string) error {
	metaId, err := NewId()
	if err != nil {
		return err
	}

	_, err = r.db.Exec(
		`INSERT INTO integrations_meta (id, integration_id, key, value)
		VALUES ($1, $2, $3, $4)`,
		metaId.String(), id.String(), key, value,
	)

	return err
}

// Get returns integration metadata by key.
func (r *IntegrationMetaRepositoryPostgres) Get(id Id, key string) (*IntegrationMeta, error) {
	meta := &IntegrationMeta{}
	err := r.db.QueryRow(
		`SELECT id, integration_id, key, value, created_at, updated_at
		FROM integrations_meta
		WHERE integration_id = $1 AND key = $2`,
		id.String(), key,
	).Scan(&meta.Id, &meta.IntegrationId, &meta.Key, &meta.Value, &meta.CreatedAt, &meta.UpdatedAt)
	if isNotFound(err) {
		return nil, nil
	}

	return meta, err
}

// Update updates an existing integration metadata row.
func (r *IntegrationMetaRepositoryPostgres) Update(id Id, key, value string) error {
	_, err := r.db.Exec(
		`UPDATE integrations_meta
		SET value = $1, updated_at = $2
		WHERE integration_id = $3 AND key = $4`,
		value, time.Now().UTC(), id.String(), key,
	)

	return err
}

// Delete deletes an integration metadata row.
func (r *IntegrationMetaRepositoryPostgres) Delete(id Id, key string) error {
	_, err := r.db.Exec(
		`DELETE FROM integrations_meta
		WHERE integration_id = $1 AND key = $2`,
		id.String(), key,
	)

	return err
}

// ListByIntegrationId lists integration metadata rows by integration id.
func (r *IntegrationMetaRepositoryPostgres) ListByIntegrationId(id Id) ([]*IntegrationMeta, error) {
	rows, err := r.db.Query(
		`SELECT id, integration_id, key, value, created_at, updated_at
		FROM integrations_meta
		WHERE integration_id = $1
		ORDER BY key`,
		id.String(),
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var list []*IntegrationMeta
	for rows.Next() {
		meta := &IntegrationMeta{}
		err := rows.Scan(&meta.Id, &meta.IntegrationId, &meta.Key, &meta.Value, &meta.CreatedAt, &meta.UpdatedAt)
		if err != nil {
			return nil, err
		}

		list = append(list, meta)
	}

	return list, rows.Err()
}

// Upsert creates or updates integration metadata.
func (r *IntegrationMetaRepositoryPostgres) Upsert(id Id, key, value string) error {
	existing, err := r.Get(id, key)
	if err != nil {
		return err
	}
	if existing == nil {
		return r.Create(id, key, value)
	}

	return r.Update(id, key, value)
}
