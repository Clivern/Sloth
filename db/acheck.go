// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package db

import (
	"database/sql"
	"time"
)

// AgentCheck is a single row in the agent_checks table.
type AgentCheck struct {
	Id         Id
	AgentId    Id
	CheckId    string
	Name       string
	Type       string
	Definition string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// AgentCheckRepository is the interface for agent check template CRUD.
type AgentCheckRepository interface {
	Create(check *AgentCheck) error
	GetByAgentAndCheckId(agentId Id, checkId string) (*AgentCheck, error)
	Update(check *AgentCheck) error
	Delete(id Id) error
	ListByAgentId(agentId Id) ([]*AgentCheck, error)
}

type AgentCheckRepositoryPostgres struct {
	db *sql.DB
}

const agentCheckColumns = `id, agent_id, check_id, name, type, definition, created_at, updated_at`

// NewAgentCheckRepository returns the repository for agent check templates.
func NewAgentCheckRepository(db *sql.DB) AgentCheckRepository {
	return &AgentCheckRepositoryPostgres{db: db}
}

// Create inserts an agent check template.
func (r *AgentCheckRepositoryPostgres) Create(check *AgentCheck) error {
	id, err := NewId()
	if err != nil {
		return err
	}

	check.Id = id

	return r.db.QueryRow(
		`INSERT INTO agent_checks (id, agent_id, check_id, name, type, definition)
		VALUES ($1, $2, $3, $4, $5, $6::jsonb)
		RETURNING created_at, updated_at`,
		check.Id.String(),
		check.AgentId.String(),
		check.CheckId,
		check.Name,
		check.Type,
		check.Definition,
	).Scan(&check.CreatedAt, &check.UpdatedAt)
}

// GetByAgentAndCheckId returns an agent check template by its check id.
func (r *AgentCheckRepositoryPostgres) GetByAgentAndCheckId(agentId Id, checkId string) (*AgentCheck, error) {
	check := &AgentCheck{}
	err := r.db.QueryRow(
		`SELECT `+agentCheckColumns+`
		FROM agent_checks
		WHERE agent_id = $1 AND check_id = $2`,
		agentId.String(),
		checkId,
	).Scan(
		&check.Id,
		&check.AgentId,
		&check.CheckId,
		&check.Name,
		&check.Type,
		&check.Definition,
		&check.CreatedAt,
		&check.UpdatedAt,
	)
	if isNotFound(err) {
		return nil, nil
	}

	return check, err
}

// Update updates an agent check template.
func (r *AgentCheckRepositoryPostgres) Update(check *AgentCheck) error {
	check.UpdatedAt = time.Now().UTC()
	_, err := r.db.Exec(
		`UPDATE agent_checks
		SET name = $1, type = $2, definition = $3::jsonb, updated_at = $4
		WHERE id = $5`,
		check.Name,
		check.Type,
		check.Definition,
		check.UpdatedAt,
		check.Id.String(),
	)

	return err
}

// Delete removes an agent check template.
func (r *AgentCheckRepositoryPostgres) Delete(id Id) error {
	_, err := r.db.Exec(`DELETE FROM agent_checks WHERE id = $1`, id.String())

	return err
}

// ListByAgentId lists the check templates of an agent.
func (r *AgentCheckRepositoryPostgres) ListByAgentId(agentId Id) ([]*AgentCheck, error) {
	rows, err := r.db.Query(
		`SELECT `+agentCheckColumns+`
		FROM agent_checks
		WHERE agent_id = $1
		ORDER BY check_id`,
		agentId.String(),
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var list []*AgentCheck
	for rows.Next() {
		check := &AgentCheck{}
		err := rows.Scan(
			&check.Id,
			&check.AgentId,
			&check.CheckId,
			&check.Name,
			&check.Type,
			&check.Definition,
			&check.CreatedAt,
			&check.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		list = append(list, check)
	}

	return list, rows.Err()
}
