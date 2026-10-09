// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package db

import (
	"database/sql"
	"errors"
	"time"
)

const (
	AgentInstanceStatusPassing  = "passing"
	AgentInstanceStatusWarning  = "warning"
	AgentInstanceStatusCritical = "critical"
)

// AgentInstance is a single row in the agent_instances table.
type AgentInstance struct {
	Id         Id
	AgentId    Id
	InstanceId string
	Address    string
	Port       int
	Datacenter *string
	Meta       *string
	Status     string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// AgentInstanceRepository is the interface for agent instance persistence.
type AgentInstanceRepository interface {
	Create(instance *AgentInstance) error
	GetById(id Id) (*AgentInstance, error)
	GetByAgentAndInstanceId(agentId Id, instanceId string) (*AgentInstance, error)
	Update(instance *AgentInstance) error
	Delete(id Id) error
	ListByAgentId(agentId Id) ([]*AgentInstance, error)
	ListLiveByAgentId(agentId Id, now time.Time) ([]*AgentInstance, error)
}

type AgentInstanceRepositoryPostgres struct {
	db *sql.DB
}

// AgentInstanceMeta is a single row in the agent_instances_meta table.
type AgentInstanceMeta struct {
	Id              Id
	AgentInstanceId Id
	Key             string
	Value           string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// AgentInstanceMetaRepository is the interface for agent instance metadata CRUD.
type AgentInstanceMetaRepository interface {
	Create(id Id, key, value string) error
	Get(id Id, key string) (*AgentInstanceMeta, error)
	Update(id Id, key, value string) error
	Delete(id Id, key string) error
	ListByAgentInstanceId(id Id) ([]*AgentInstanceMeta, error)
	Upsert(id Id, key, value string) error
}

type AgentInstanceMetaRepositoryPostgres struct {
	db *sql.DB
}

// NewAgentInstanceRepository returns the repository for agent instances.
func NewAgentInstanceRepository(db *sql.DB) AgentInstanceRepository {
	return &AgentInstanceRepositoryPostgres{db: db}
}

// Create inserts an agent instance row.
func (r *AgentInstanceRepositoryPostgres) Create(instance *AgentInstance) error {
	id, err := NewId()
	if err != nil {
		return err
	}

	instance.Id = id
	if instance.Status == "" {
		instance.Status = AgentInstanceStatusPassing
	}

	return r.db.QueryRow(
		`INSERT INTO agent_instances
		(id, agent_id, instance_id, address, port, datacenter, meta, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb, $8)
		RETURNING created_at, updated_at`,
		instance.Id.String(),
		instance.AgentId.String(),
		instance.InstanceId,
		instance.Address,
		instance.Port,
		instance.Datacenter,
		instance.Meta,
		instance.Status,
	).Scan(&instance.CreatedAt, &instance.UpdatedAt)
}

// GetById returns an agent instance by id. TODO: remove if not used in future.
func (r *AgentInstanceRepositoryPostgres) GetById(id Id) (*AgentInstance, error) {
	instance := &AgentInstance{}
	err := r.db.QueryRow(
		`SELECT id, agent_id, instance_id, address, port, datacenter, meta, status, created_at, updated_at
		FROM agent_instances
		WHERE id = $1`,
		id.String(),
	).Scan(
		&instance.Id,
		&instance.AgentId,
		&instance.InstanceId,
		&instance.Address,
		&instance.Port,
		&instance.Datacenter,
		&instance.Meta,
		&instance.Status,
		&instance.CreatedAt,
		&instance.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	return instance, err
}

// GetByAgentAndInstanceId returns an agent instance by agent and instance id.
func (r *AgentInstanceRepositoryPostgres) GetByAgentAndInstanceId(agentId Id, instanceId string) (*AgentInstance, error) {
	instance := &AgentInstance{}
	err := r.db.QueryRow(
		`SELECT id, agent_id, instance_id, address, port, datacenter, meta, status, created_at, updated_at
		FROM agent_instances
		WHERE agent_id = $1 AND instance_id = $2`,
		agentId.String(),
		instanceId,
	).Scan(
		&instance.Id,
		&instance.AgentId,
		&instance.InstanceId,
		&instance.Address,
		&instance.Port,
		&instance.Datacenter,
		&instance.Meta,
		&instance.Status,
		&instance.CreatedAt,
		&instance.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	return instance, err
}

// Update updates an agent instance address and status.
func (r *AgentInstanceRepositoryPostgres) Update(instance *AgentInstance) error {
	_, err := r.db.Exec(
		`UPDATE agent_instances
		SET address = $1, port = $2, datacenter = $3, meta = $4::jsonb, status = $5, updated_at = $6
		WHERE id = $7`,
		instance.Address,
		instance.Port,
		instance.Datacenter,
		instance.Meta,
		instance.Status,
		time.Now().UTC(),
		instance.Id.String(),
	)

	return err
}

// Delete removes an agent instance row.
func (r *AgentInstanceRepositoryPostgres) Delete(id Id) error {
	_, err := r.db.Exec(`DELETE FROM agent_instances WHERE id = $1`, id.String())

	return err
}

// ListByAgentId lists instances for an agent.
func (r *AgentInstanceRepositoryPostgres) ListByAgentId(agentId Id) ([]*AgentInstance, error) {
	rows, err := r.db.Query(
		`SELECT id, agent_id, instance_id, address, port, datacenter, meta, status, created_at, updated_at
		FROM agent_instances
		WHERE agent_id = $1
		ORDER BY instance_id`,
		agentId.String(),
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var list []*AgentInstance
	for rows.Next() {
		instance := &AgentInstance{}
		err := rows.Scan(
			&instance.Id,
			&instance.AgentId,
			&instance.InstanceId,
			&instance.Address,
			&instance.Port,
			&instance.Datacenter,
			&instance.Meta,
			&instance.Status,
			&instance.CreatedAt,
			&instance.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		list = append(list, instance)
	}

	return list, rows.Err()
}

// ListLiveByAgentId lists instances with a live lease and no critical or expired checks.
func (r *AgentInstanceRepositoryPostgres) ListLiveByAgentId(agentId Id, now time.Time) ([]*AgentInstance, error) {
	rows, err := r.db.Query(
		`SELECT id, agent_id, instance_id, address, port, datacenter, meta, status, created_at, updated_at
		FROM agent_instances
		WHERE agent_id = $1
			AND EXISTS (
				SELECT 1 FROM health_checks lease
				WHERE lease.agent_instance_id = agent_instances.id
					AND lease.source = $5
					AND lease.ttl_expires_at > $2
			)
			AND NOT EXISTS (
				SELECT 1 FROM health_checks hc
				WHERE hc.agent_instance_id = agent_instances.id
					AND (
						hc.status = $4
						OR (hc.type = $3 AND (hc.ttl_expires_at IS NULL OR hc.ttl_expires_at <= $2))
					)
			)
		ORDER BY instance_id`,
		agentId.String(),
		now,
		HealthCheckTypeTTL,
		HealthCheckStatusCritical,
		HealthCheckSourceLease,
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var list []*AgentInstance
	for rows.Next() {
		instance := &AgentInstance{}
		err := rows.Scan(
			&instance.Id,
			&instance.AgentId,
			&instance.InstanceId,
			&instance.Address,
			&instance.Port,
			&instance.Datacenter,
			&instance.Meta,
			&instance.Status,
			&instance.CreatedAt,
			&instance.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		list = append(list, instance)
	}

	return list, rows.Err()
}

// NewAgentInstanceMetaRepository returns the repository for agent instance metadata.
func NewAgentInstanceMetaRepository(db *sql.DB) AgentInstanceMetaRepository {
	return &AgentInstanceMetaRepositoryPostgres{db: db}
}

// Create inserts an agent instance metadata row. TODO: remove if not used in future.
func (r *AgentInstanceMetaRepositoryPostgres) Create(id Id, key, value string) error {
	metaId, err := NewId()
	if err != nil {
		return err
	}

	_, err = r.db.Exec(
		`INSERT INTO agent_instances_meta (id, agent_instance_id, key, value)
		VALUES ($1, $2, $3, to_jsonb($4::text))`,
		metaId.String(),
		id.String(),
		key,
		value,
	)

	return err
}

// Get returns agent instance metadata by key. TODO: remove if not used in future.
func (r *AgentInstanceMetaRepositoryPostgres) Get(id Id, key string) (*AgentInstanceMeta, error) {
	meta := &AgentInstanceMeta{}
	err := r.db.QueryRow(
		`SELECT id, agent_instance_id, key, value #>> '{}', created_at, updated_at
		FROM agent_instances_meta
		WHERE agent_instance_id = $1 AND key = $2`,
		id.String(),
		key,
	).Scan(
		&meta.Id,
		&meta.AgentInstanceId,
		&meta.Key,
		&meta.Value,
		&meta.CreatedAt,
		&meta.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	return meta, err
}

// Update updates an existing agent instance metadata row. TODO: remove if not used in future.
func (r *AgentInstanceMetaRepositoryPostgres) Update(id Id, key, value string) error {
	_, err := r.db.Exec(
		`UPDATE agent_instances_meta
		SET value = to_jsonb($1::text), updated_at = $2
		WHERE agent_instance_id = $3 AND key = $4`,
		value,
		time.Now().UTC(),
		id.String(),
		key,
	)

	return err
}

// Delete deletes an agent instance metadata row. TODO: remove if not used in future.
func (r *AgentInstanceMetaRepositoryPostgres) Delete(id Id, key string) error {
	_, err := r.db.Exec(
		`DELETE FROM agent_instances_meta WHERE agent_instance_id = $1 AND key = $2`,
		id.String(),
		key,
	)

	return err
}

// ListByAgentInstanceId lists agent instance metadata rows. TODO: remove if not used in future.
func (r *AgentInstanceMetaRepositoryPostgres) ListByAgentInstanceId(id Id) ([]*AgentInstanceMeta, error) {
	rows, err := r.db.Query(
		`SELECT id, agent_instance_id, key, value #>> '{}', created_at, updated_at
		FROM agent_instances_meta
		WHERE agent_instance_id = $1
		ORDER BY key`,
		id.String(),
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var list []*AgentInstanceMeta
	for rows.Next() {
		meta := &AgentInstanceMeta{}
		err := rows.Scan(
			&meta.Id,
			&meta.AgentInstanceId,
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

// Upsert creates or updates agent instance metadata. TODO: remove if not used in future.
func (r *AgentInstanceMetaRepositoryPostgres) Upsert(id Id, key, value string) error {
	existing, err := r.Get(id, key)
	if err != nil {
		return err
	}
	if existing == nil {
		return r.Create(id, key, value)
	}

	return r.Update(id, key, value)
}
