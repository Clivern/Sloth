// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package db

import (
	"database/sql"
	"time"
)

const (
	TrafficDecisionAllow = "allow"
	TrafficDecisionDeny  = "deny"

	TrafficResultCompleted  = "completed"
	TrafficResultFailed     = "failed"
	TrafficResultDenied     = "denied"
	TrafficResultTimeout    = "timeout"
	TrafficResultNoInstance = "no_instance"
)

// TrafficCall is a single row in the traffic table.
type TrafficCall struct {
	Id                  Id
	WorkspaceId         Id
	SourceInstance      string
	DestinationInstance string
	Skill               string
	Verb                string
	Decision            string
	Result              string
	LatencyMs           int
	TaskId              *string
	Meta                *string
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// TrafficRepository is the interface for gateway traffic persistence.
type TrafficRepository interface {
	Create(call *TrafficCall) error
	GetById(id Id) (*TrafficCall, error)
	ListByWorkspaceId(workspaceId Id, limit, offset int) ([]*TrafficCall, error)
	CountByWorkspaceId(workspaceId Id) (int64, error)
}

type TrafficRepositoryPostgres struct {
	db *sql.DB
}

// NewTrafficRepository returns the repository for traffic.
func NewTrafficRepository(db *sql.DB) TrafficRepository {
	return &TrafficRepositoryPostgres{db: db}
}

// Create inserts a traffic call row.
func (r *TrafficRepositoryPostgres) Create(call *TrafficCall) error {
	id, err := NewId()
	if err != nil {
		return err
	}

	call.Id = id

	return r.db.QueryRow(
		`INSERT INTO traffic (
			id, workspace_id, source_instance, destination_instance, skill, verb,
			decision, result, latency_ms, task_id, meta
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11::jsonb)
		RETURNING created_at, updated_at`,
		call.Id.String(),
		call.WorkspaceId.String(),
		call.SourceInstance,
		call.DestinationInstance,
		call.Skill,
		call.Verb,
		call.Decision,
		call.Result,
		call.LatencyMs,
		call.TaskId,
		call.Meta,
	).Scan(&call.CreatedAt, &call.UpdatedAt)
}

// GetById returns a traffic call by id.
func (r *TrafficRepositoryPostgres) GetById(id Id) (*TrafficCall, error) {
	call := &TrafficCall{}
	err := r.db.QueryRow(
		`SELECT
			id, workspace_id, source_instance, destination_instance, skill, verb,
			decision, result, latency_ms, task_id, meta,
			created_at, updated_at
		FROM traffic
		WHERE id = $1`,
		id.String(),
	).Scan(
		&call.Id,
		&call.WorkspaceId,
		&call.SourceInstance,
		&call.DestinationInstance,
		&call.Skill,
		&call.Verb,
		&call.Decision,
		&call.Result,
		&call.LatencyMs,
		&call.TaskId,
		&call.Meta,
		&call.CreatedAt,
		&call.UpdatedAt,
	)
	if isNotFound(err) {
		return nil, nil
	}

	return call, err
}

// ListByWorkspaceId lists traffic calls for a workspace, newest first.
func (r *TrafficRepositoryPostgres) ListByWorkspaceId(workspaceId Id, limit, offset int) ([]*TrafficCall, error) {
	rows, err := r.db.Query(
		`SELECT
			id, workspace_id, source_instance, destination_instance, skill, verb,
			decision, result, latency_ms, task_id, meta,
			created_at, updated_at
		FROM traffic
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

	var list []*TrafficCall
	for rows.Next() {
		call := &TrafficCall{}
		err := rows.Scan(
			&call.Id,
			&call.WorkspaceId,
			&call.SourceInstance,
			&call.DestinationInstance,
			&call.Skill,
			&call.Verb,
			&call.Decision,
			&call.Result,
			&call.LatencyMs,
			&call.TaskId,
			&call.Meta,
			&call.CreatedAt,
			&call.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		list = append(list, call)
	}

	return list, rows.Err()
}

// CountByWorkspaceId returns the number of traffic calls in a workspace.
func (r *TrafficRepositoryPostgres) CountByWorkspaceId(workspaceId Id) (int64, error) {
	var count int64
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM traffic WHERE workspace_id = $1`,
		workspaceId.String(),
	).Scan(&count)

	return count, err
}
