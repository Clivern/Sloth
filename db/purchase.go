// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package db

import (
	"database/sql"
	"time"
)

// TokenPurchase is a Stripe payment that credited AI tokens to a workspace.
type TokenPurchase struct {
	Id              Id
	WorkspaceId     Id
	StripeSessionId string
	AmountCents     int64
	Tokens          int64
	CreatedAt       time.Time
}

// TokenPurchaseRepository is the interface for token purchase persistence.
type TokenPurchaseRepository interface {
	Create(purchase *TokenPurchase) (bool, error)
}

type TokenPurchaseRepositoryPostgres struct {
	db *sql.DB
}

// NewTokenPurchaseRepository returns the repository for token purchases.
func NewTokenPurchaseRepository(db *sql.DB) TokenPurchaseRepository {
	return &TokenPurchaseRepositoryPostgres{db: db}
}

// Create inserts a token purchase. Returns false when the Stripe session was already credited.
func (r *TokenPurchaseRepositoryPostgres) Create(purchase *TokenPurchase) (bool, error) {
	id, err := NewId()
	if err != nil {
		return false, err
	}

	err = r.db.QueryRow(
		`INSERT INTO workspace_token_purchases
		(id, workspace_id, stripe_session_id, amount_cents, tokens)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (stripe_session_id) DO NOTHING
		RETURNING id, created_at`,
		id.String(),
		purchase.WorkspaceId.String(),
		purchase.StripeSessionId,
		purchase.AmountCents,
		purchase.Tokens,
	).Scan(&purchase.Id, &purchase.CreatedAt)
	if isNotFound(err) {
		return false, nil
	}

	return true, err
}
