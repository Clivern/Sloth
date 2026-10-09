// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package db

import (
	"database/sql"
	"time"
)

// PasswordResetToken is the DB row for a password reset token.
type PasswordResetToken struct {
	Id        Id
	UserId    Id
	Token     string
	ExpiresAt time.Time
	CreatedAt time.Time
}

// PasswordResetTokenRepository is the interface for password reset token CRUD.
type PasswordResetTokenRepository interface {
	Create(t *PasswordResetToken) error
	GetByToken(token string) (*PasswordResetToken, error)
	Delete(id Id) error
	DeleteByToken(token string) error
	DeleteExpired() (int64, error)
}

type PasswordResetTokenRepositoryPostgres struct {
	db *sql.DB
}

// NewPasswordResetTokenRepository returns the repository for the current driver
func NewPasswordResetTokenRepository(db *sql.DB) PasswordResetTokenRepository {
	return &PasswordResetTokenRepositoryPostgres{db: db}
}

// --- Postgres ---

// Create inserts a password reset token row.
func (r *PasswordResetTokenRepositoryPostgres) Create(t *PasswordResetToken) error {
	id, err := NewId()
	if err != nil {
		return err
	}
	t.Id = id

	_, err = r.db.Exec(
		`INSERT INTO password_reset_tokens (id, user_id, token, expires_at)
		VALUES ($1, $2, $3, $4)`,
		t.Id.String(),
		t.UserId.String(),
		t.Token,
		t.ExpiresAt,
	)
	return err
}

// GetByToken returns a token by its string value.
func (r *PasswordResetTokenRepositoryPostgres) GetByToken(token string) (*PasswordResetToken, error) {
	tok := &PasswordResetToken{}
	err := r.db.QueryRow(
		`SELECT
			id,
			user_id,
			token,
			expires_at,
			created_at
		FROM password_reset_tokens
		WHERE token = $1`,
		token,
	).Scan(
		&tok.Id,
		&tok.UserId,
		&tok.Token,
		&tok.ExpiresAt,
		&tok.CreatedAt,
	)
	if isNotFound(err) {
		return nil, nil
	}
	return tok, err
}

// Delete removes a password reset token by Id.
func (r *PasswordResetTokenRepositoryPostgres) Delete(id Id) error {
	_, err := r.db.Exec(
		`DELETE FROM password_reset_tokens WHERE id = $1`,
		id.String(),
	)
	return err
}

// DeleteByToken removes a password reset token by token string.
func (r *PasswordResetTokenRepositoryPostgres) DeleteByToken(token string) error {
	_, err := r.db.Exec(
		`DELETE FROM password_reset_tokens WHERE token = $1`,
		token,
	)
	return err
}

// DeleteExpired removes expired tokens and returns the count deleted.
func (r *PasswordResetTokenRepositoryPostgres) DeleteExpired() (int64, error) {
	result, err := r.db.Exec(
		`DELETE FROM password_reset_tokens
		WHERE expires_at < $1`,
		time.Now().UTC(),
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
