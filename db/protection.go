// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package db

import (
	"database/sql"
	"time"
)

const (
	AgentProtectionTypeAPIKey            = "api_key"
	AgentProtectionTypeBasicAuth         = "basic_auth"
	AgentProtectionTypeClientCredentials = "client_credentials"
)

// AgentProtection is a single row in the agent_protection table.
type AgentProtection struct {
	Id        Id
	AgentId   Id
	Name      string
	Type      string
	Config    string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// AgentProtectionRepository is the interface for agent protection CRUD.
type AgentProtectionRepository interface {
	Create(protection *AgentProtection) error
	GetById(id Id) (*AgentProtection, error)
	GetByAgentAndName(agentId Id, name string) (*AgentProtection, error)
	Update(protection *AgentProtection) error
	Delete(id Id) error
	ListByAgentId(agentId Id) ([]*AgentProtection, error)
	ListActiveByAgentId(agentId Id) ([]*AgentProtection, error)
}

type AgentProtectionRepositoryPostgres struct {
	db *sql.DB
}

// AgentProtectionToken is a single row in the agent_protection_tokens table.
type AgentProtectionToken struct {
	Id           Id
	ProtectionId Id
	TokenHash    string
	ExpiresAt    time.Time
	RevokedAt    *time.Time
	Meta         *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// AgentProtectionTokenRepository is the interface for issued protection token CRUD.
type AgentProtectionTokenRepository interface {
	Create(token *AgentProtectionToken) error
	GetById(id Id) (*AgentProtectionToken, error)
	GetByTokenHash(tokenHash string) (*AgentProtectionToken, error)
	ListByProtectionId(protectionId Id) ([]*AgentProtectionToken, error)
	Revoke(id Id) error
	DeleteExpired() (int64, error)
}

type AgentProtectionTokenRepositoryPostgres struct {
	db *sql.DB
}

// NewAgentProtectionRepository returns the repository for the agent_protection table.
func NewAgentProtectionRepository(db *sql.DB) AgentProtectionRepository {
	return &AgentProtectionRepositoryPostgres{db: db}
}

// Create inserts an agent protection row.
func (r *AgentProtectionRepositoryPostgres) Create(protection *AgentProtection) error {
	id, err := NewId()
	if err != nil {
		return err
	}

	protection.Id = id

	return r.db.QueryRow(
		`INSERT INTO agent_protection (id, agent_id, name, type, config, is_active)
		VALUES ($1, $2, $3, $4, $5::jsonb, $6)
		RETURNING created_at, updated_at`,
		protection.Id.String(),
		protection.AgentId.String(),
		protection.Name,
		protection.Type,
		protection.Config,
		protection.IsActive,
	).Scan(&protection.CreatedAt, &protection.UpdatedAt)
}

// GetById returns an agent protection by id.
func (r *AgentProtectionRepositoryPostgres) GetById(id Id) (*AgentProtection, error) {
	protection := &AgentProtection{}
	err := r.db.QueryRow(
		`SELECT id, agent_id, name, type, config, is_active, created_at, updated_at
		FROM agent_protection
		WHERE id = $1`,
		id.String(),
	).Scan(
		&protection.Id,
		&protection.AgentId,
		&protection.Name,
		&protection.Type,
		&protection.Config,
		&protection.IsActive,
		&protection.CreatedAt,
		&protection.UpdatedAt,
	)
	if isNotFound(err) {
		return nil, nil
	}

	return protection, err
}

// GetByAgentAndName returns an agent protection by agent and name.
func (r *AgentProtectionRepositoryPostgres) GetByAgentAndName(agentId Id, name string) (*AgentProtection, error) {
	protection := &AgentProtection{}
	err := r.db.QueryRow(
		`SELECT id, agent_id, name, type, config, is_active, created_at, updated_at
		FROM agent_protection
		WHERE agent_id = $1 AND name = $2`,
		agentId.String(),
		name,
	).Scan(
		&protection.Id,
		&protection.AgentId,
		&protection.Name,
		&protection.Type,
		&protection.Config,
		&protection.IsActive,
		&protection.CreatedAt,
		&protection.UpdatedAt,
	)
	if isNotFound(err) {
		return nil, nil
	}

	return protection, err
}

// Update updates an agent protection row.
func (r *AgentProtectionRepositoryPostgres) Update(protection *AgentProtection) error {
	_, err := r.db.Exec(
		`UPDATE agent_protection
		SET name = $1, type = $2, config = $3::jsonb, is_active = $4, updated_at = $5
		WHERE id = $6`,
		protection.Name,
		protection.Type,
		protection.Config,
		protection.IsActive,
		time.Now().UTC(),
		protection.Id.String(),
	)

	return err
}

// Delete removes an agent protection row.
func (r *AgentProtectionRepositoryPostgres) Delete(id Id) error {
	_, err := r.db.Exec(`DELETE FROM agent_protection WHERE id = $1`, id.String())

	return err
}

// ListByAgentId lists the protection rows of an agent.
func (r *AgentProtectionRepositoryPostgres) ListByAgentId(agentId Id) ([]*AgentProtection, error) {
	rows, err := r.db.Query(
		`SELECT id, agent_id, name, type, config, is_active, created_at, updated_at
		FROM agent_protection
		WHERE agent_id = $1
		ORDER BY name`,
		agentId.String(),
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var list []*AgentProtection
	for rows.Next() {
		protection := &AgentProtection{}
		err := rows.Scan(
			&protection.Id,
			&protection.AgentId,
			&protection.Name,
			&protection.Type,
			&protection.Config,
			&protection.IsActive,
			&protection.CreatedAt,
			&protection.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		list = append(list, protection)
	}

	return list, rows.Err()
}

// ListActiveByAgentId lists the active protection rows of an agent.
func (r *AgentProtectionRepositoryPostgres) ListActiveByAgentId(agentId Id) ([]*AgentProtection, error) {
	rows, err := r.db.Query(
		`SELECT id, agent_id, name, type, config, is_active, created_at, updated_at
		FROM agent_protection
		WHERE agent_id = $1 AND is_active = TRUE
		ORDER BY name`,
		agentId.String(),
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var list []*AgentProtection
	for rows.Next() {
		protection := &AgentProtection{}
		err := rows.Scan(
			&protection.Id,
			&protection.AgentId,
			&protection.Name,
			&protection.Type,
			&protection.Config,
			&protection.IsActive,
			&protection.CreatedAt,
			&protection.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		list = append(list, protection)
	}

	return list, rows.Err()
}

// NewAgentProtectionTokenRepository returns the repository for agent_protection_tokens.
func NewAgentProtectionTokenRepository(db *sql.DB) AgentProtectionTokenRepository {
	return &AgentProtectionTokenRepositoryPostgres{db: db}
}

// Create inserts an agent protection token row.
func (r *AgentProtectionTokenRepositoryPostgres) Create(token *AgentProtectionToken) error {
	id, err := NewId()
	if err != nil {
		return err
	}

	token.Id = id

	return r.db.QueryRow(
		`INSERT INTO agent_protection_tokens
		(id, protection_id, token_hash, expires_at, revoked_at, meta)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at, updated_at`,
		token.Id.String(),
		token.ProtectionId.String(),
		token.TokenHash,
		token.ExpiresAt,
		token.RevokedAt,
		token.Meta,
	).Scan(&token.CreatedAt, &token.UpdatedAt)
}

// GetById returns an agent protection token by id.
func (r *AgentProtectionTokenRepositoryPostgres) GetById(id Id) (*AgentProtectionToken, error) {
	token := &AgentProtectionToken{}
	err := r.db.QueryRow(
		`SELECT id, protection_id, token_hash, expires_at, revoked_at, meta, created_at, updated_at
		FROM agent_protection_tokens
		WHERE id = $1`,
		id.String(),
	).Scan(
		&token.Id,
		&token.ProtectionId,
		&token.TokenHash,
		&token.ExpiresAt,
		&token.RevokedAt,
		&token.Meta,
		&token.CreatedAt,
		&token.UpdatedAt,
	)
	if isNotFound(err) {
		return nil, nil
	}

	return token, err
}

// GetByTokenHash returns a non-revoked, non-expired token by hash.
func (r *AgentProtectionTokenRepositoryPostgres) GetByTokenHash(tokenHash string) (*AgentProtectionToken, error) {
	token := &AgentProtectionToken{}
	err := r.db.QueryRow(
		`SELECT id, protection_id, token_hash, expires_at, revoked_at, meta, created_at, updated_at
		FROM agent_protection_tokens
		WHERE token_hash = $1
			AND revoked_at IS NULL
			AND expires_at > $2`,
		tokenHash,
		time.Now().UTC(),
	).Scan(
		&token.Id,
		&token.ProtectionId,
		&token.TokenHash,
		&token.ExpiresAt,
		&token.RevokedAt,
		&token.Meta,
		&token.CreatedAt,
		&token.UpdatedAt,
	)
	if isNotFound(err) {
		return nil, nil
	}

	return token, err
}

// ListByProtectionId lists tokens for a protection row.
func (r *AgentProtectionTokenRepositoryPostgres) ListByProtectionId(protectionId Id) ([]*AgentProtectionToken, error) {
	rows, err := r.db.Query(
		`SELECT id, protection_id, token_hash, expires_at, revoked_at, meta, created_at, updated_at
		FROM agent_protection_tokens
		WHERE protection_id = $1
		ORDER BY created_at DESC`,
		protectionId.String(),
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var list []*AgentProtectionToken
	for rows.Next() {
		token := &AgentProtectionToken{}
		err := rows.Scan(
			&token.Id,
			&token.ProtectionId,
			&token.TokenHash,
			&token.ExpiresAt,
			&token.RevokedAt,
			&token.Meta,
			&token.CreatedAt,
			&token.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		list = append(list, token)
	}

	return list, rows.Err()
}

// Revoke marks a token as revoked.
func (r *AgentProtectionTokenRepositoryPostgres) Revoke(id Id) error {
	now := time.Now().UTC()
	_, err := r.db.Exec(
		`UPDATE agent_protection_tokens
		SET revoked_at = $1, updated_at = $1
		WHERE id = $2 AND revoked_at IS NULL`,
		now,
		id.String(),
	)

	return err
}

// DeleteExpired removes expired token rows.
func (r *AgentProtectionTokenRepositoryPostgres) DeleteExpired() (int64, error) {
	result, err := r.db.Exec(
		`DELETE FROM agent_protection_tokens WHERE expires_at <= $1`,
		time.Now().UTC(),
	)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}
