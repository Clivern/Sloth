// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package db

import (
	"database/sql"
)

const UsageTypeAITokens = "ai_tokens"
const UsageTypeAICost = "ai_cost"
const UsageUnitTokens = "tokens"
const UsageUnitNanoUSD = "nano_usd"

// WorkspaceStats holds aggregate metrics for a workspace dashboard.
type WorkspaceStats struct{}

// WorkspaceStatsRepository loads workspace dashboard metrics.
type WorkspaceStatsRepository interface {
	GetByWorkspaceId(workspaceId Id) (*WorkspaceStats, error)
}

type WorkspaceStatsRepositoryPostgres struct {
	db *sql.DB
}

// NewWorkspaceStatsRepository returns a workspace stats repository.
func NewWorkspaceStatsRepository(db *sql.DB) WorkspaceStatsRepository {
	return &WorkspaceStatsRepositoryPostgres{db: db}
}

// GetByWorkspaceId returns a workspace stats by workspace id.
func (r *WorkspaceStatsRepositoryPostgres) GetByWorkspaceId(_ Id) (*WorkspaceStats, error) {
	return &WorkspaceStats{}, nil
}
