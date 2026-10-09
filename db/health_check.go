// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package db

import (
	"database/sql"
	"time"
)

const (
	HealthCheckStatusPassing  = "passing"
	HealthCheckStatusWarning  = "warning"
	HealthCheckStatusCritical = "critical"

	HealthCheckTypeTTL   = "ttl"
	HealthCheckTypeHTTP  = "http"
	HealthCheckTypeTCP   = "tcp"
	HealthCheckTypeA2A   = "a2a"
	HealthCheckTypeCard  = "card"
	HealthCheckTypeSkill = "skill"

	// HealthCheckIDTTL is the check_id for the instance lease.
	HealthCheckIDTTL = "ttl"
)

// HealthCheck is a single row in the health_checks table.
type HealthCheck struct {
	Id              Id
	AgentInstanceId Id
	CheckId         string
	Name            string
	Type            string
	Status          string
	Definition      *string
	Output          *string
	TTLExpiresAt    *time.Time
	LastRunAt       *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// HealthCheckRepository is the interface for health check persistence.
type HealthCheckRepository interface {
	Create(check *HealthCheck) error
	GetById(id Id) (*HealthCheck, error)
	GetByInstanceAndCheckId(agentInstanceId Id, checkId string) (*HealthCheck, error)
	Update(check *HealthCheck) error
	Delete(id Id) error
	ListByAgentInstanceId(agentInstanceId Id) ([]*HealthCheck, error)
	Pass(id Id, output string, ttlExpiresAt *time.Time) error
	Warn(id Id, output string) error
	Fail(id Id, output string) error
}

type HealthCheckRepositoryPostgres struct {
	db *sql.DB
}

// HealthCheckMeta is a single row in the health_checks_meta table.
type HealthCheckMeta struct {
	Id            Id
	HealthCheckId Id
	Key           string
	Value         string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// HealthCheckMetaRepository is the interface for health check metadata CRUD.
type HealthCheckMetaRepository interface {
	Create(id Id, key, value string) error
	Get(id Id, key string) (*HealthCheckMeta, error)
	Update(id Id, key, value string) error
	Delete(id Id, key string) error
	ListByHealthCheckId(id Id) ([]*HealthCheckMeta, error)
	Upsert(id Id, key, value string) error
}

type HealthCheckMetaRepositoryPostgres struct {
	db *sql.DB
}

const healthCheckColumns = `id, agent_instance_id, check_id, name, type, status, definition, output, ttl_expires_at, last_run_at, created_at, updated_at`

// NewHealthCheckRepository returns the repository for health checks.
func NewHealthCheckRepository(db *sql.DB) HealthCheckRepository {
	return &HealthCheckRepositoryPostgres{db: db}
}

// Create inserts a health check row. A new check starts critical until it passes.
func (r *HealthCheckRepositoryPostgres) Create(check *HealthCheck) error {
	id, err := NewId()
	if err != nil {
		return err
	}
	check.Id = id
	if check.Status == "" {
		check.Status = HealthCheckStatusCritical
	}

	return r.db.QueryRow(
		`INSERT INTO health_checks
		(id, agent_instance_id, check_id, name, type, status, definition, output, ttl_expires_at, last_run_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb, $8, $9, $10)
		RETURNING created_at, updated_at`,
		check.Id.String(),
		check.AgentInstanceId.String(),
		check.CheckId,
		check.Name,
		check.Type,
		check.Status,
		check.Definition,
		check.Output,
		check.TTLExpiresAt,
		check.LastRunAt,
	).Scan(&check.CreatedAt, &check.UpdatedAt)
}

// GetById returns a health check by id.
func (r *HealthCheckRepositoryPostgres) GetById(id Id) (*HealthCheck, error) {
	check := &HealthCheck{}
	err := r.db.QueryRow(
		`SELECT `+healthCheckColumns+`
		FROM health_checks
		WHERE id = $1`,
		id.String(),
	).Scan(scanHealthCheck(check)...)
	if isNotFound(err) {
		return nil, nil
	}
	return check, err
}

// GetByInstanceAndCheckId returns a health check by instance and check id.
func (r *HealthCheckRepositoryPostgres) GetByInstanceAndCheckId(agentInstanceId Id, checkId string) (*HealthCheck, error) {
	check := &HealthCheck{}
	err := r.db.QueryRow(
		`SELECT `+healthCheckColumns+`
		FROM health_checks
		WHERE agent_instance_id = $1 AND check_id = $2`,
		agentInstanceId.String(),
		checkId,
	).Scan(scanHealthCheck(check)...)
	if isNotFound(err) {
		return nil, nil
	}
	return check, err
}

// Update updates a health check definition.
func (r *HealthCheckRepositoryPostgres) Update(check *HealthCheck) error {
	_, err := r.db.Exec(
		`UPDATE health_checks
		SET name = $1, type = $2, definition = $3::jsonb, updated_at = $4
		WHERE id = $5`,
		check.Name,
		check.Type,
		check.Definition,
		time.Now().UTC(),
		check.Id.String(),
	)
	return err
}

// Delete removes a health check row.
func (r *HealthCheckRepositoryPostgres) Delete(id Id) error {
	_, err := r.db.Exec(`DELETE FROM health_checks WHERE id = $1`, id.String())
	return err
}

// ListByAgentInstanceId lists health checks for an instance.
func (r *HealthCheckRepositoryPostgres) ListByAgentInstanceId(agentInstanceId Id) ([]*HealthCheck, error) {
	rows, err := r.db.Query(
		`SELECT `+healthCheckColumns+`
		FROM health_checks
		WHERE agent_instance_id = $1
		ORDER BY check_id`,
		agentInstanceId.String(),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*HealthCheck
	for rows.Next() {
		check := &HealthCheck{}
		err := rows.Scan(scanHealthCheck(check)...)
		if err != nil {
			return nil, err
		}
		list = append(list, check)
	}
	return list, rows.Err()
}

// Pass marks a check passing and optionally extends a TTL.
func (r *HealthCheckRepositoryPostgres) Pass(id Id, output string, ttlExpiresAt *time.Time) error {
	return r.setStatus(id, HealthCheckStatusPassing, output, ttlExpiresAt)
}

// Warn marks a check warning. The instance stays in discovery.
func (r *HealthCheckRepositoryPostgres) Warn(id Id, output string) error {
	return r.setStatus(id, HealthCheckStatusWarning, output, nil)
}

// Fail marks a check critical and drops the instance from discovery.
func (r *HealthCheckRepositoryPostgres) Fail(id Id, output string) error {
	return r.setStatus(id, HealthCheckStatusCritical, output, nil)
}

func (r *HealthCheckRepositoryPostgres) setStatus(id Id, status, output string, ttlExpiresAt *time.Time) error {
	now := time.Now().UTC()
	var raw *string
	if output != "" {
		raw = &output
	}

	_, err := r.db.Exec(
		`UPDATE health_checks
		SET status = $1, output = $2, ttl_expires_at = COALESCE($3, ttl_expires_at),
			last_run_at = $4, updated_at = $4
		WHERE id = $5`,
		status,
		raw,
		ttlExpiresAt,
		now,
		id.String(),
	)
	return err
}

func scanHealthCheck(check *HealthCheck) []any {
	return []any{
		&check.Id,
		&check.AgentInstanceId,
		&check.CheckId,
		&check.Name,
		&check.Type,
		&check.Status,
		&check.Definition,
		&check.Output,
		&check.TTLExpiresAt,
		&check.LastRunAt,
		&check.CreatedAt,
		&check.UpdatedAt,
	}
}

// NewHealthCheckMetaRepository returns the repository for health check metadata.
func NewHealthCheckMetaRepository(db *sql.DB) HealthCheckMetaRepository {
	return &HealthCheckMetaRepositoryPostgres{db: db}
}

// Create inserts a health check metadata row.
func (r *HealthCheckMetaRepositoryPostgres) Create(id Id, key, value string) error {
	metaId, err := NewId()
	if err != nil {
		return err
	}
	_, err = r.db.Exec(
		`INSERT INTO health_checks_meta (id, health_check_id, key, value)
		VALUES ($1, $2, $3, to_jsonb($4::text))`,
		metaId.String(),
		id.String(),
		key,
		value,
	)
	return err
}

// Get returns health check metadata by key.
func (r *HealthCheckMetaRepositoryPostgres) Get(id Id, key string) (*HealthCheckMeta, error) {
	meta := &HealthCheckMeta{}
	err := r.db.QueryRow(
		`SELECT id, health_check_id, key, value #>> '{}', created_at, updated_at
		FROM health_checks_meta
		WHERE health_check_id = $1 AND key = $2`,
		id.String(),
		key,
	).Scan(
		&meta.Id,
		&meta.HealthCheckId,
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

// Update updates an existing health check metadata row.
func (r *HealthCheckMetaRepositoryPostgres) Update(id Id, key, value string) error {
	_, err := r.db.Exec(
		`UPDATE health_checks_meta
		SET value = to_jsonb($1::text), updated_at = $2
		WHERE health_check_id = $3 AND key = $4`,
		value,
		time.Now().UTC(),
		id.String(),
		key,
	)
	return err
}

// Delete deletes a health check metadata row.
func (r *HealthCheckMetaRepositoryPostgres) Delete(id Id, key string) error {
	_, err := r.db.Exec(
		`DELETE FROM health_checks_meta WHERE health_check_id = $1 AND key = $2`,
		id.String(),
		key,
	)
	return err
}

// ListByHealthCheckId lists health check metadata rows.
func (r *HealthCheckMetaRepositoryPostgres) ListByHealthCheckId(id Id) ([]*HealthCheckMeta, error) {
	rows, err := r.db.Query(
		`SELECT id, health_check_id, key, value #>> '{}', created_at, updated_at
		FROM health_checks_meta
		WHERE health_check_id = $1
		ORDER BY key`,
		id.String(),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*HealthCheckMeta
	for rows.Next() {
		meta := &HealthCheckMeta{}
		err := rows.Scan(
			&meta.Id,
			&meta.HealthCheckId,
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

// Upsert creates or updates health check metadata.
func (r *HealthCheckMetaRepositoryPostgres) Upsert(id Id, key, value string) error {
	existing, err := r.Get(id, key)
	if err != nil {
		return err
	}
	if existing == nil {
		return r.Create(id, key, value)
	}
	return r.Update(id, key, value)
}
