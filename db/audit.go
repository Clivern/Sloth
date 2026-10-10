// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package db

import (
	"database/sql"
	"errors"
	"time"

	"github.com/samber/lo"
)

// AuditEvent is a single row in the audit table.
type AuditEvent struct {
	Id           Id
	WorkspaceId  Id
	UserId       *Id
	Action       string
	ResourceType *string
	ResourceId   *Id
	IPAddress    *string
	UserAgent    *string
	Meta         *string
	CreatedAt    time.Time
}

// AuditEventRepository is the interface for audit event persistence.
type AuditEventRepository interface {
	Create(event *AuditEvent) error
	GetById(id Id) (*AuditEvent, error)
	ListByWorkspaceId(workspaceId Id, limit, offset int) ([]*AuditEvent, error)
	CountByWorkspaceId(workspaceId Id) (int64, error)
}

type AuditEventRepositoryPostgres struct {
	db *sql.DB
}

// NewAuditEventRepository returns an audit event repository.
func NewAuditEventRepository(db *sql.DB) AuditEventRepository {
	return &AuditEventRepositoryPostgres{db: db}
}

// Create inserts an audit event row. TODO: remove if not used in future.
func (r *AuditEventRepositoryPostgres) Create(event *AuditEvent) error {
	id, err := NewId()
	if err != nil {
		return err
	}

	event.Id = id

	err = r.db.QueryRow(
		`INSERT INTO audit (
			id, workspace_id, actor_id, action, resource_type, resource_id,
			ip_address, user_agent, meta
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING created_at`,
		event.Id.String(),
		event.WorkspaceId.String(),
		lo.FromPtr(event.UserId),
		event.Action,
		event.ResourceType,
		lo.FromPtr(event.ResourceId),
		event.IPAddress,
		event.UserAgent,
		event.Meta,
	).Scan(&event.CreatedAt)

	return err
}

// GetById returns an audit event by id.
func (r *AuditEventRepositoryPostgres) GetById(id Id) (*AuditEvent, error) {
	event := &AuditEvent{}
	err := r.db.QueryRow(
		`SELECT
			id, workspace_id, actor_id, action, resource_type, resource_id,
			ip_address, user_agent, meta, created_at
		FROM audit
		WHERE id = $1`,
		id.String(),
	).Scan(
		&event.Id,
		&event.WorkspaceId,
		&event.UserId,
		&event.Action,
		&event.ResourceType,
		&event.ResourceId,
		&event.IPAddress,
		&event.UserAgent,
		&event.Meta,
		&event.CreatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	return event, err
}

// ListByWorkspaceId lists audit event rows by workspace id.
func (r *AuditEventRepositoryPostgres) ListByWorkspaceId(workspaceId Id, limit, offset int) ([]*AuditEvent, error) {
	rows, err := r.db.Query(
		`SELECT
			id, workspace_id, actor_id, action, resource_type, resource_id,
			ip_address, user_agent, meta, created_at
		FROM audit
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

	var events []*AuditEvent
	for rows.Next() {
		event := &AuditEvent{}
		if err := rows.Scan(
			&event.Id,
			&event.WorkspaceId,
			&event.UserId,
			&event.Action,
			&event.ResourceType,
			&event.ResourceId,
			&event.IPAddress,
			&event.UserAgent,
			&event.Meta,
			&event.CreatedAt,
		); err != nil {
			return nil, err
		}

		events = append(events, event)
	}

	return events, rows.Err()
}

// CountByWorkspaceId counts audit event rows by workspace id.
func (r *AuditEventRepositoryPostgres) CountByWorkspaceId(workspaceId Id) (int64, error) {
	var count int64
	err := r.db.QueryRow(
		`SELECT COUNT(*)
		FROM audit
		WHERE workspace_id = $1`,
		workspaceId.String(),
	).Scan(&count)

	return count, err
}
