// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package db

import (
	"database/sql"
	"errors"
	"time"
)

const (
	EventActorTypeUser      = "user"
	EventActorTypeAccessKey = "access_key"
	EventActorTypeSystem    = "system"

	EventResourceTypeAgentInstance = "agent_instance"
	EventResourceTypeHealthCheck   = "health_check"
)

// Event is a single row in the events table.
type Event struct {
	Id           Id
	WorkspaceId  Id
	Type         string
	ActorId      *Id
	ActorName    *string
	ActorType    *string
	ResourceType *string
	ResourceId   *Id
	IPAddress    *string
	UserAgent    *string
	Data         *string
	CreatedAt    time.Time
}

// EventRepository is the interface for event persistence.
type EventRepository interface {
	Create(event *Event) error
	GetById(id Id) (*Event, error)
	ListByWorkspaceId(workspaceId Id, limit, offset int) ([]*Event, error)
	CountByWorkspaceId(workspaceId Id) (int64, error)
}

type EventRepositoryPostgres struct {
	db *sql.DB
}

// NewEventRepository returns an event repository.
func NewEventRepository(db *sql.DB) EventRepository {
	return &EventRepositoryPostgres{db: db}
}

// Create inserts an event row.
func (r *EventRepositoryPostgres) Create(event *Event) error {
	id, err := NewId()
	if err != nil {
		return err
	}

	event.Id = id

	err = r.db.QueryRow(
		`INSERT INTO events (
			id, workspace_id, type, actor_id, actor_name, actor_type,
			resource_type, resource_id, ip_address, user_agent, data
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING created_at`,
		event.Id.String(),
		event.WorkspaceId.String(),
		event.Type,
		event.ActorId,
		event.ActorName,
		event.ActorType,
		event.ResourceType,
		event.ResourceId,
		event.IPAddress,
		event.UserAgent,
		event.Data,
	).Scan(&event.CreatedAt)

	return err
}

// GetById returns an event by id.
func (r *EventRepositoryPostgres) GetById(id Id) (*Event, error) {
	event := &Event{}
	err := r.db.QueryRow(
		`SELECT
			id, workspace_id, type, actor_id, actor_name, actor_type,
			resource_type, resource_id, ip_address, user_agent, data, created_at
		FROM events
		WHERE id = $1`,
		id.String(),
	).Scan(
		&event.Id,
		&event.WorkspaceId,
		&event.Type,
		&event.ActorId,
		&event.ActorName,
		&event.ActorType,
		&event.ResourceType,
		&event.ResourceId,
		&event.IPAddress,
		&event.UserAgent,
		&event.Data,
		&event.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	return event, err
}

// ListByWorkspaceId returns the newest events of a workspace.
func (r *EventRepositoryPostgres) ListByWorkspaceId(workspaceId Id, limit, offset int) ([]*Event, error) {
	rows, err := r.db.Query(
		`SELECT
			id, workspace_id, type, actor_id, actor_name, actor_type,
			resource_type, resource_id, ip_address, user_agent, data, created_at
		FROM events
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

	var list []*Event
	for rows.Next() {
		event := &Event{}
		err := rows.Scan(
			&event.Id,
			&event.WorkspaceId,
			&event.Type,
			&event.ActorId,
			&event.ActorName,
			&event.ActorType,
			&event.ResourceType,
			&event.ResourceId,
			&event.IPAddress,
			&event.UserAgent,
			&event.Data,
			&event.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		list = append(list, event)
	}

	return list, rows.Err()
}

// CountByWorkspaceId returns the total number of events in a workspace.
func (r *EventRepositoryPostgres) CountByWorkspaceId(workspaceId Id) (int64, error) {
	var count int64
	err := r.db.QueryRow(
		`SELECT COUNT(*)
		FROM events
		WHERE workspace_id = $1`,
		workspaceId.String(),
	).Scan(&count)

	return count, err
}
