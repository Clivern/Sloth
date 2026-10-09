// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package db

import (
	"database/sql"
	"time"
)

// APIKey is the DB row for a user API key.
type APIKey struct {
	Id        Id
	UserId    Id
	Name      string
	Key       string
	ExpiresAt *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

// APIKeyRepository is the interface for API key CRUD.
type APIKeyRepository interface {
	Create(apiKey *APIKey) error
	GetById(id Id) (*APIKey, error)
	GetByKey(key string) (*APIKey, error)
	ListByUserId(userId Id, limit, offset int) ([]*APIKey, error)
	Delete(id Id) error
	DeleteExpired() (int64, error)
	Count() (int64, error)
	CountByUserId(userId Id) (int64, error)
}

// APIKeyRepositoryPostgres implements APIKeyRepository for PostgreSQL
type APIKeyRepositoryPostgres struct {
	db *sql.DB
}

// NewAPIKeyRepository returns the repository for the current driver
func NewAPIKeyRepository(db *sql.DB) APIKeyRepository {
	return &APIKeyRepositoryPostgres{db: db}
}

// Create inserts a row
func (r *APIKeyRepositoryPostgres) Create(apiKey *APIKey) error {
	id, err := NewId()
	if err != nil {
		return err
	}

	apiKey.Id = id

	err = r.db.QueryRow(
		`INSERT INTO user_api_keys
		(id, user_id, name, token, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at, updated_at`,
		apiKey.Id.String(),
		apiKey.UserId.String(),
		apiKey.Name,
		apiKey.Key,
		apiKey.ExpiresAt,
	).Scan(
		&apiKey.CreatedAt,
		&apiKey.UpdatedAt,
	)

	return err
}

// GetById returns a API key by Id
func (r *APIKeyRepositoryPostgres) GetById(id Id) (*APIKey, error) {
	k := &APIKey{}
	err := r.db.QueryRow(
		`SELECT
			id, user_id, name, token, expires_at, created_at, updated_at
		FROM user_api_keys
		WHERE id = $1`,
		id.String(),
	).Scan(
		&k.Id,
		&k.UserId,
		&k.Name,
		&k.Key,
		&k.ExpiresAt,
		&k.CreatedAt,
		&k.UpdatedAt,
	)
	if isNotFound(err) {
		return nil, nil
	}

	return k, err
}

// GetByKey returns a API key by raw key
func (r *APIKeyRepositoryPostgres) GetByKey(key string) (*APIKey, error) {
	k := &APIKey{}
	err := r.db.QueryRow(
		`SELECT
			id, user_id, name, token, expires_at, created_at, updated_at
		FROM user_api_keys
		WHERE token = $1 AND (expires_at IS NULL OR expires_at > $2)`,
		key, time.Now().UTC(),
	).Scan(
		&k.Id,
		&k.UserId,
		&k.Name,
		&k.Key,
		&k.ExpiresAt,
		&k.CreatedAt,
		&k.UpdatedAt,
	)
	if isNotFound(err) {
		return nil, nil
	}

	return k, err
}

// ListByUserId returns a paginated list of API keys by user Id.
func (r *APIKeyRepositoryPostgres) ListByUserId(userId Id, limit, offset int) ([]*APIKey, error) {
	rows, err := r.db.Query(
		`SELECT
			id, user_id, name, token, expires_at, created_at, updated_at
		FROM user_api_keys
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`,
		userId.String(),
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var list []*APIKey
	for rows.Next() {
		k := &APIKey{}
		if err := rows.Scan(
			&k.Id,
			&k.UserId,
			&k.Name,
			&k.Key,
			&k.ExpiresAt,
			&k.CreatedAt,
			&k.UpdatedAt,
		); err != nil {
			return nil, err
		}

		list = append(list, k)
	}

	return list, rows.Err()
}

// Delete removes a API key
func (r *APIKeyRepositoryPostgres) Delete(id Id) error {
	_, err := r.db.Exec(
		`DELETE FROM user_api_keys WHERE id = $1`,
		id.String(),
	)

	return err
}

// DeleteExpired removes expired API keys
func (r *APIKeyRepositoryPostgres) DeleteExpired() (int64, error) {
	result, err := r.db.Exec(
		`DELETE FROM user_api_keys
		WHERE expires_at IS NOT NULL AND expires_at < $1`,
		time.Now().UTC(),
	)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// Count returns the total number of API keys
func (r *APIKeyRepositoryPostgres) Count() (int64, error) {
	var count int64
	err := r.db.QueryRow(
		`SELECT COUNT(*)
		FROM user_api_keys
		WHERE (expires_at IS NULL OR expires_at > $1)`,
		time.Now().UTC(),
	).Scan(&count)

	return count, err
}

// CountByUserId returns the total number of API keys by user Id
func (r *APIKeyRepositoryPostgres) CountByUserId(userId Id) (int64, error) {
	var count int64
	err := r.db.QueryRow(
		`SELECT COUNT(*)
		FROM user_api_keys
		WHERE user_id = $1 AND (expires_at IS NULL OR expires_at > $2)`,
		userId.String(), time.Now().UTC(),
	).Scan(&count)

	return count, err
}
