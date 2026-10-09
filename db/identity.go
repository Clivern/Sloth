// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package db

import (
	"database/sql"
	"time"
)

const (
	AgentIdentityTypeAPIKey    = "api_key"
	AgentIdentityTypeBasicAuth = "basic_auth"
)

// AgentIdentity is a single row in the agent_identity table.
type AgentIdentity struct {
	Id        Id
	AgentId   Id
	Name      string
	Type      string
	Config    string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// AgentIdentityRepository is the interface for agent identity CRUD.
type AgentIdentityRepository interface {
	Create(identity *AgentIdentity) error
	GetById(id Id) (*AgentIdentity, error)
	GetByAgentAndName(agentId Id, name string) (*AgentIdentity, error)
	Update(identity *AgentIdentity) error
	Delete(id Id) error
	ListByAgentId(agentId Id) ([]*AgentIdentity, error)
	ListActiveByAgentId(agentId Id) ([]*AgentIdentity, error)
}

type AgentIdentityRepositoryPostgres struct {
	db *sql.DB
}

// NewAgentIdentityRepository returns the repository for the agent_identity table.
func NewAgentIdentityRepository(db *sql.DB) AgentIdentityRepository {
	return &AgentIdentityRepositoryPostgres{db: db}
}

// Create inserts an agent identity row. TODO: remove if not used in future.
func (r *AgentIdentityRepositoryPostgres) Create(identity *AgentIdentity) error {
	id, err := NewId()
	if err != nil {
		return err
	}

	identity.Id = id

	return r.db.QueryRow(
		`INSERT INTO agent_identity (id, agent_id, name, type, config, is_active)
		VALUES ($1, $2, $3, $4, $5::jsonb, $6)
		RETURNING created_at, updated_at`,
		identity.Id.String(),
		identity.AgentId.String(),
		identity.Name,
		identity.Type,
		identity.Config,
		identity.IsActive,
	).Scan(&identity.CreatedAt, &identity.UpdatedAt)
}

// GetById returns an agent identity by id. TODO: remove if not used in future.
func (r *AgentIdentityRepositoryPostgres) GetById(id Id) (*AgentIdentity, error) {
	identity := &AgentIdentity{}
	err := r.db.QueryRow(
		`SELECT id, agent_id, name, type, config, is_active, created_at, updated_at
		FROM agent_identity
		WHERE id = $1`,
		id.String(),
	).Scan(
		&identity.Id,
		&identity.AgentId,
		&identity.Name,
		&identity.Type,
		&identity.Config,
		&identity.IsActive,
		&identity.CreatedAt,
		&identity.UpdatedAt,
	)
	if isNotFound(err) {
		return nil, nil
	}

	return identity, err
}

// GetByAgentAndName returns an agent identity by agent and name. TODO: remove if not used in future.
func (r *AgentIdentityRepositoryPostgres) GetByAgentAndName(agentId Id, name string) (*AgentIdentity, error) {
	identity := &AgentIdentity{}
	err := r.db.QueryRow(
		`SELECT id, agent_id, name, type, config, is_active, created_at, updated_at
		FROM agent_identity
		WHERE agent_id = $1 AND name = $2`,
		agentId.String(),
		name,
	).Scan(
		&identity.Id,
		&identity.AgentId,
		&identity.Name,
		&identity.Type,
		&identity.Config,
		&identity.IsActive,
		&identity.CreatedAt,
		&identity.UpdatedAt,
	)
	if isNotFound(err) {
		return nil, nil
	}

	return identity, err
}

// Update updates an agent identity row. TODO: remove if not used in future.
func (r *AgentIdentityRepositoryPostgres) Update(identity *AgentIdentity) error {
	_, err := r.db.Exec(
		`UPDATE agent_identity
		SET name = $1, type = $2, config = $3::jsonb, is_active = $4, updated_at = $5
		WHERE id = $6`,
		identity.Name,
		identity.Type,
		identity.Config,
		identity.IsActive,
		time.Now().UTC(),
		identity.Id.String(),
	)

	return err
}

// Delete removes an agent identity row. TODO: remove if not used in future.
func (r *AgentIdentityRepositoryPostgres) Delete(id Id) error {
	_, err := r.db.Exec(`DELETE FROM agent_identity WHERE id = $1`, id.String())

	return err
}

// ListByAgentId lists the identity rows of an agent. TODO: remove if not used in future.
func (r *AgentIdentityRepositoryPostgres) ListByAgentId(agentId Id) ([]*AgentIdentity, error) {
	rows, err := r.db.Query(
		`SELECT id, agent_id, name, type, config, is_active, created_at, updated_at
		FROM agent_identity
		WHERE agent_id = $1
		ORDER BY name`,
		agentId.String(),
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var list []*AgentIdentity
	for rows.Next() {
		identity := &AgentIdentity{}
		err := rows.Scan(
			&identity.Id,
			&identity.AgentId,
			&identity.Name,
			&identity.Type,
			&identity.Config,
			&identity.IsActive,
			&identity.CreatedAt,
			&identity.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		list = append(list, identity)
	}

	return list, rows.Err()
}

// ListActiveByAgentId lists the active identity rows of an agent. TODO: remove if not used in future.
func (r *AgentIdentityRepositoryPostgres) ListActiveByAgentId(agentId Id) ([]*AgentIdentity, error) {
	rows, err := r.db.Query(
		`SELECT id, agent_id, name, type, config, is_active, created_at, updated_at
		FROM agent_identity
		WHERE agent_id = $1 AND is_active = TRUE
		ORDER BY name`,
		agentId.String(),
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var list []*AgentIdentity
	for rows.Next() {
		identity := &AgentIdentity{}
		err := rows.Scan(
			&identity.Id,
			&identity.AgentId,
			&identity.Name,
			&identity.Type,
			&identity.Config,
			&identity.IsActive,
			&identity.CreatedAt,
			&identity.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		list = append(list, identity)
	}

	return list, rows.Err()
}
