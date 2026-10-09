// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package db

import (
	"database/sql"
	"errors"
	"time"
)

// KeyValue is a key/value row with an optional expiry.
type KeyValue struct {
	Id        Id
	Key       string
	Value     string
	ExpiresAt *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

// KeyValueRepository is a generic expiry-aware key/value store.
type KeyValueRepository interface {
	Upsert(item *KeyValue) error
	Get(key string) (*KeyValue, error)
	Delete(key string) error
	DeleteExpired() (int64, error)
}

type KeyValueRepositoryPostgres struct {
	db *sql.DB
}

// NewKeyValueRepository returns the repository for the kv table.
func NewKeyValueRepository(db *sql.DB) KeyValueRepository {
	return &KeyValueRepositoryPostgres{db: db}
}

// Upsert inserts or replaces a key/value row. TODO: remove if not used in future.
func (r *KeyValueRepositoryPostgres) Upsert(item *KeyValue) error {
	id, err := NewId()
	if err != nil {
		return err
	}

	return r.db.QueryRow(
		`INSERT INTO kv (id, key, value, expires_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (key) DO UPDATE SET
			value = EXCLUDED.value,
			expires_at = EXCLUDED.expires_at,
			updated_at = CURRENT_TIMESTAMP AT TIME ZONE 'UTC'
		RETURNING id, created_at, updated_at`,
		id.String(),
		item.Key,
		item.Value,
		item.ExpiresAt,
	).Scan(
		&item.Id,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
}

// Get returns a non-expired value for key. A null expires_at never expires. TODO: remove if not used in future.
func (r *KeyValueRepositoryPostgres) Get(key string) (*KeyValue, error) {
	item := &KeyValue{}
	err := r.db.QueryRow(
		`SELECT id, key, value, expires_at, created_at, updated_at
		FROM kv
		WHERE key = $1 AND (expires_at IS NULL OR expires_at > $2)`,
		key,
		time.Now().UTC(),
	).Scan(
		&item.Id,
		&item.Key,
		&item.Value,
		&item.ExpiresAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	return item, err
}

// Delete removes a key/value row. TODO: remove if not used in future.
func (r *KeyValueRepositoryPostgres) Delete(key string) error {
	_, err := r.db.Exec(`DELETE FROM kv WHERE key = $1`, key)

	return err
}

// DeleteExpired removes rows that have passed their expiry. TODO: remove if not used in future.
func (r *KeyValueRepositoryPostgres) DeleteExpired() (int64, error) {
	result, err := r.db.Exec(
		`DELETE FROM kv WHERE expires_at IS NOT NULL AND expires_at <= $1`,
		time.Now().UTC(),
	)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}
