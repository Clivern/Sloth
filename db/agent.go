// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package db

import (
	"database/sql"
	"time"
)

// Agent is a single row in the agents table.
type Agent struct {
	Id           Id
	WorkspaceId  Id
	Name         string
	Card         string
	CardChecksum string
	Version      string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// AgentRepository is the interface for agent catalog CRUD.
type AgentRepository interface {
	Create(agent *Agent) error
	GetById(id Id) (*Agent, error)
	GetByWorkspaceAndName(workspaceId Id, name string) (*Agent, error)
	Update(agent *Agent) error
	Delete(id Id) error
	ListByWorkspaceId(workspaceId Id, limit, offset int) ([]*Agent, error)
	CountByWorkspaceId(workspaceId Id) (int64, error)
}

type AgentRepositoryPostgres struct {
	db *sql.DB
}

// AgentMeta is a single row in the agents_meta table.
type AgentMeta struct {
	Id        Id
	AgentId   Id
	Key       string
	Value     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// AgentMetaRepository is the interface for agent metadata CRUD.
type AgentMetaRepository interface {
	Create(id Id, key, value string) error
	Get(id Id, key string) (*AgentMeta, error)
	Update(id Id, key, value string) error
	Delete(id Id, key string) error
	ListByAgentId(id Id) ([]*AgentMeta, error)
	Upsert(id Id, key, value string) error
}

type AgentMetaRepositoryPostgres struct {
	db *sql.DB
}

// NewAgentRepository returns the repository for the agents table.
func NewAgentRepository(db *sql.DB) AgentRepository {
	return &AgentRepositoryPostgres{db: db}
}

// Create inserts an agent row.
func (r *AgentRepositoryPostgres) Create(agent *Agent) error {
	id, err := NewId()
	if err != nil {
		return err
	}

	agent.Id = id

	return r.db.QueryRow(
		`INSERT INTO agents (id, workspace_id, name, card, card_checksum, version)
		VALUES ($1, $2, $3, $4::jsonb, $5, $6)
		RETURNING created_at, updated_at`,
		agent.Id.String(),
		agent.WorkspaceId.String(),
		agent.Name,
		agent.Card,
		agent.CardChecksum,
		agent.Version,
	).Scan(&agent.CreatedAt, &agent.UpdatedAt)
}

// GetById returns an agent by id.
func (r *AgentRepositoryPostgres) GetById(id Id) (*Agent, error) {
	agent := &Agent{}
	err := r.db.QueryRow(
		`SELECT id, workspace_id, name, card, card_checksum, version, created_at, updated_at
		FROM agents
		WHERE id = $1`,
		id.String(),
	).Scan(
		&agent.Id,
		&agent.WorkspaceId,
		&agent.Name,
		&agent.Card,
		&agent.CardChecksum,
		&agent.Version,
		&agent.CreatedAt,
		&agent.UpdatedAt,
	)
	if isNotFound(err) {
		return nil, nil
	}

	return agent, err
}

// GetByWorkspaceAndName returns an agent by workspace and name.
func (r *AgentRepositoryPostgres) GetByWorkspaceAndName(workspaceId Id, name string) (*Agent, error) {
	agent := &Agent{}
	err := r.db.QueryRow(
		`SELECT id, workspace_id, name, card, card_checksum, version, created_at, updated_at
		FROM agents
		WHERE workspace_id = $1 AND name = $2`,
		workspaceId.String(),
		name,
	).Scan(
		&agent.Id,
		&agent.WorkspaceId,
		&agent.Name,
		&agent.Card,
		&agent.CardChecksum,
		&agent.Version,
		&agent.CreatedAt,
		&agent.UpdatedAt,
	)
	if isNotFound(err) {
		return nil, nil
	}

	return agent, err
}

// Update updates an agent card and version.
func (r *AgentRepositoryPostgres) Update(agent *Agent) error {
	_, err := r.db.Exec(
		`UPDATE agents
		SET name = $1, card = $2::jsonb, card_checksum = $3, version = $4, updated_at = $5
		WHERE id = $6`,
		agent.Name,
		agent.Card,
		agent.CardChecksum,
		agent.Version,
		time.Now().UTC(),
		agent.Id.String(),
	)

	return err
}

// Delete removes an agent row.
func (r *AgentRepositoryPostgres) Delete(id Id) error {
	_, err := r.db.Exec(`DELETE FROM agents WHERE id = $1`, id.String())

	return err
}

// ListByWorkspaceId lists agents in a workspace.
func (r *AgentRepositoryPostgres) ListByWorkspaceId(workspaceId Id, limit, offset int) ([]*Agent, error) {
	rows, err := r.db.Query(
		`SELECT id, workspace_id, name, card, card_checksum, version, created_at, updated_at
		FROM agents
		WHERE workspace_id = $1
		ORDER BY name
		LIMIT $2 OFFSET $3`,
		workspaceId.String(),
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var list []*Agent
	for rows.Next() {
		agent := &Agent{}
		err := rows.Scan(
			&agent.Id,
			&agent.WorkspaceId,
			&agent.Name,
			&agent.Card,
			&agent.CardChecksum,
			&agent.Version,
			&agent.CreatedAt,
			&agent.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		list = append(list, agent)
	}

	return list, rows.Err()
}

// CountByWorkspaceId returns the number of agents in a workspace.
func (r *AgentRepositoryPostgres) CountByWorkspaceId(workspaceId Id) (int64, error) {
	var count int64
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM agents WHERE workspace_id = $1`,
		workspaceId.String(),
	).Scan(&count)

	return count, err
}

// NewAgentMetaRepository returns the repository for agent metadata.
func NewAgentMetaRepository(db *sql.DB) AgentMetaRepository {
	return &AgentMetaRepositoryPostgres{db: db}
}

// Create inserts an agent metadata row. TODO: remove if not used in future.
func (r *AgentMetaRepositoryPostgres) Create(id Id, key, value string) error {
	metaId, err := NewId()
	if err != nil {
		return err
	}

	_, err = r.db.Exec(
		`INSERT INTO agents_meta (id, agent_id, key, value)
		VALUES ($1, $2, $3, to_jsonb($4::text))`,
		metaId.String(),
		id.String(),
		key,
		value,
	)

	return err
}

// Get returns agent metadata by key. TODO: remove if not used in future.
func (r *AgentMetaRepositoryPostgres) Get(id Id, key string) (*AgentMeta, error) {
	meta := &AgentMeta{}
	err := r.db.QueryRow(
		`SELECT id, agent_id, key, value #>> '{}', created_at, updated_at
		FROM agents_meta
		WHERE agent_id = $1 AND key = $2`,
		id.String(),
		key,
	).Scan(
		&meta.Id,
		&meta.AgentId,
		&meta.Key,
		&meta.Value,
		&meta.CreatedAt,
		&meta.UpdatedAt,
	)
	if isNotFound(err) {
		return nil, nil
	}

	return meta, err
}

// Update updates an existing agent metadata row. TODO: remove if not used in future.
func (r *AgentMetaRepositoryPostgres) Update(id Id, key, value string) error {
	_, err := r.db.Exec(
		`UPDATE agents_meta
		SET value = to_jsonb($1::text), updated_at = $2
		WHERE agent_id = $3 AND key = $4`,
		value,
		time.Now().UTC(),
		id.String(),
		key,
	)

	return err
}

// Delete deletes an agent metadata row. TODO: remove if not used in future.
func (r *AgentMetaRepositoryPostgres) Delete(id Id, key string) error {
	_, err := r.db.Exec(
		`DELETE FROM agents_meta WHERE agent_id = $1 AND key = $2`,
		id.String(),
		key,
	)

	return err
}

// ListByAgentId lists agent metadata rows. TODO: remove if not used in future.
func (r *AgentMetaRepositoryPostgres) ListByAgentId(id Id) ([]*AgentMeta, error) {
	rows, err := r.db.Query(
		`SELECT id, agent_id, key, value #>> '{}', created_at, updated_at
		FROM agents_meta
		WHERE agent_id = $1
		ORDER BY key`,
		id.String(),
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var list []*AgentMeta
	for rows.Next() {
		meta := &AgentMeta{}
		err := rows.Scan(
			&meta.Id,
			&meta.AgentId,
			&meta.Key,
			&meta.Value,
			&meta.CreatedAt,
			&meta.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		list = append(list, meta)
	}

	return list, rows.Err()
}

// Upsert creates or updates agent metadata. TODO: remove if not used in future.
func (r *AgentMetaRepositoryPostgres) Upsert(id Id, key, value string) error {
	existing, err := r.Get(id, key)
	if err != nil {
		return err
	}
	if existing == nil {
		return r.Create(id, key, value)
	}

	return r.Update(id, key, value)
}
