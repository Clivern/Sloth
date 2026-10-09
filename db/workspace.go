// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package db

import (
	"database/sql"
	"time"
)

// Workspace is a single row in the workspaces table.
type Workspace struct {
	Id        Id
	Name      string
	Handle    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// WorkspaceRepository is the interface for workspace CRUD.
type WorkspaceRepository interface {
	Create(workspace *Workspace) error
	GetById(id Id) (*Workspace, error)
	GetByHandle(handle string) (*Workspace, error)
	Update(workspace *Workspace) error
	Delete(id Id) error
	List(limit, offset int, userId Id) ([]*Workspace, error)
	Count(userId Id) (int64, error)
	CountAll() (int64, error)
	GetWorkspaceMembership(workspaceId, userId Id) (*WorkspaceUser, error)
}

type WorkspaceRepositoryPostgres struct {
	db *sql.DB
}

// WorkspaceMeta is a single row in the workspaces_meta table.
type WorkspaceMeta struct {
	Id          Id
	WorkspaceId Id
	Key         string
	Value       string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// WorkspaceMetaRepository is the interface for workspace metadata CRUD.
type WorkspaceMetaRepository interface {
	Create(id Id, key, value string) error
	Get(id Id, key string) (*WorkspaceMeta, error)
	Update(id Id, key, value string) error
	Delete(id Id, key string) error
	ListByWorkspaceId(id Id) ([]*WorkspaceMeta, error)
	Upsert(id Id, key, value string) error
}

type WorkspaceMetaRepositoryPostgres struct {
	db *sql.DB
}

// NewWorkspaceRepository returns the repository for the current driver
func NewWorkspaceRepository(db *sql.DB) WorkspaceRepository {
	return &WorkspaceRepositoryPostgres{db: db}
}

// Create inserts a row and fills Id and timestamps via RETURNING.
func (r *WorkspaceRepositoryPostgres) Create(workspace *Workspace) error {
	id, err := NewId()
	if err != nil {
		return err
	}

	workspace.Id = id

	err = r.db.QueryRow(
		`INSERT INTO workspaces (id, name, handle)
		VALUES ($1, $2, $3)
		RETURNING created_at, updated_at`,
		workspace.Id.String(),
		workspace.Name,
		workspace.Handle,
	).Scan(
		&workspace.CreatedAt,
		&workspace.UpdatedAt,
	)

	return err
}

// GetById returns nil, nil if not found.
func (r *WorkspaceRepositoryPostgres) GetById(id Id) (*Workspace, error) {
	w := &Workspace{}
	err := r.db.QueryRow(
		`SELECT
			id,
			name,
			handle,
			created_at,
			updated_at
		FROM workspaces
		WHERE id = $1`,
		id.String(),
	).Scan(
		&w.Id,
		&w.Name,
		&w.Handle,
		&w.CreatedAt,
		&w.UpdatedAt,
	)
	if isNotFound(err) {
		return nil, nil
	}

	return w, err
}

// GetByHandle returns nil, nil if not found.
func (r *WorkspaceRepositoryPostgres) GetByHandle(handle string) (*Workspace, error) {
	w := &Workspace{}
	err := r.db.QueryRow(
		`SELECT
			id,
			name,
			handle,
			created_at,
			updated_at
		FROM workspaces
		WHERE handle = $1`,
		handle,
	).Scan(
		&w.Id,
		&w.Name,
		&w.Handle,
		&w.CreatedAt,
		&w.UpdatedAt,
	)
	if isNotFound(err) {
		return nil, nil
	}

	return w, err
}

// Update updates a workspace.
func (r *WorkspaceRepositoryPostgres) Update(workspace *Workspace) error {
	_, err := r.db.Exec(
		`UPDATE workspaces
		SET
			name = $1,
			updated_at = $2
		WHERE id = $3`,
		workspace.Name,
		time.Now().UTC(),
		workspace.Id.String(),
	)

	return err
}

// Delete removes a workspace.
func (r *WorkspaceRepositoryPostgres) Delete(id Id) error {
	_, err := r.db.Exec(
		`DELETE FROM workspaces WHERE id = $1`,
		id.String(),
	)

	return err
}

// List returns workspaces the user is a member of, ordered by created_at DESC.
func (r *WorkspaceRepositoryPostgres) List(limit, offset int, userId Id) ([]*Workspace, error) {
	rows, err := r.db.Query(
		`SELECT
			w.id,
			w.name,
			w.handle,
			w.created_at,
			w.updated_at
		FROM workspaces w
		INNER JOIN workspace_users wu ON w.id = wu.workspace_id
		WHERE wu.user_id = $1
		ORDER BY w.created_at DESC
		LIMIT $2 OFFSET $3`,
		userId.String(),
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var list []*Workspace
	for rows.Next() {
		w := &Workspace{}
		err := rows.Scan(
			&w.Id,
			&w.Name,
			&w.Handle,
			&w.CreatedAt,
			&w.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		list = append(list, w)
	}

	return list, rows.Err()
}

// Count returns the number of workspaces the user is a member of.
func (r *WorkspaceRepositoryPostgres) Count(userId Id) (int64, error) {
	var count int64
	err := r.db.QueryRow(
		`SELECT COUNT(*)
		FROM workspaces w
		INNER JOIN workspace_users wu ON w.id = wu.workspace_id
		WHERE wu.user_id = $1`,
		userId.String(),
	).Scan(&count)

	return count, err
}

// CountAll returns the total number of workspaces in the database.
func (r *WorkspaceRepositoryPostgres) CountAll() (int64, error) {
	var count int64
	err := r.db.QueryRow(
		`SELECT COUNT(*)
		FROM workspaces`,
	).Scan(&count)

	return count, err
}

// GetWorkspaceMembership returns the user's membership (role, etc.) for the workspace, or nil if not a member.
func (r *WorkspaceRepositoryPostgres) GetWorkspaceMembership(workspaceId, userId Id) (*WorkspaceUser, error) {
	m := &WorkspaceUser{}
	err := r.db.QueryRow(
		`SELECT
			id,
			workspace_id,
			user_id,
			role,
			created_at,
			updated_at
		FROM workspace_users
		WHERE workspace_id = $1
			AND user_id = $2`,
		workspaceId.String(),
		userId.String(),
	).Scan(
		&m.Id,
		&m.WorkspaceId,
		&m.UserId,
		&m.Role,
		&m.CreatedAt,
		&m.UpdatedAt,
	)
	if isNotFound(err) {
		return nil, nil
	}

	return m, err
}

// NewWorkspaceMetaRepository returns the repository for workspace metadata.
func NewWorkspaceMetaRepository(db *sql.DB) WorkspaceMetaRepository {
	return &WorkspaceMetaRepositoryPostgres{db: db}
}

// Create inserts a workspace metadata row.
func (r *WorkspaceMetaRepositoryPostgres) Create(id Id, key, value string) error {
	metaId, err := NewId()
	if err != nil {
		return err
	}

	_, err = r.db.Exec(
		`INSERT INTO workspaces_meta (id, workspace_id, key, value)
		VALUES ($1, $2, $3, $4)`,
		metaId.String(),
		id.String(),
		key,
		value,
	)

	return err
}

// Get returns workspace metadata by key.
func (r *WorkspaceMetaRepositoryPostgres) Get(id Id, key string) (*WorkspaceMeta, error) {
	meta := &WorkspaceMeta{}
	err := r.db.QueryRow(
		`SELECT id, workspace_id, key, value, created_at, updated_at
		FROM workspaces_meta
		WHERE workspace_id = $1 AND key = $2`,
		id.String(),
		key,
	).Scan(
		&meta.Id,
		&meta.WorkspaceId,
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

// Update updates an existing workspace metadata row.
func (r *WorkspaceMetaRepositoryPostgres) Update(id Id, key, value string) error {
	_, err := r.db.Exec(
		`UPDATE workspaces_meta
		SET value = $1, updated_at = $2
		WHERE workspace_id = $3 AND key = $4`,
		value,
		time.Now().UTC(),
		id.String(),
		key,
	)

	return err
}

// Delete deletes a workspace metadata row.
func (r *WorkspaceMetaRepositoryPostgres) Delete(id Id, key string) error {
	_, err := r.db.Exec(
		`DELETE FROM workspaces_meta
		WHERE workspace_id = $1 AND key = $2`,
		id.String(),
		key,
	)

	return err
}

// ListByWorkspaceId lists workspace metadata rows by workspace id.
func (r *WorkspaceMetaRepositoryPostgres) ListByWorkspaceId(id Id) ([]*WorkspaceMeta, error) {
	rows, err := r.db.Query(
		`SELECT id, workspace_id, key, value, created_at, updated_at
		FROM workspaces_meta
		WHERE workspace_id = $1
		ORDER BY key`,
		id.String(),
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var list []*WorkspaceMeta
	for rows.Next() {
		meta := &WorkspaceMeta{}
		err := rows.Scan(
			&meta.Id,
			&meta.WorkspaceId,
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

// Upsert creates or updates workspace metadata.
func (r *WorkspaceMetaRepositoryPostgres) Upsert(id Id, key, value string) error {
	existing, err := r.Get(id, key)
	if err != nil {
		return err
	}
	if existing == nil {
		return r.Create(id, key, value)
	}

	return r.Update(id, key, value)
}
