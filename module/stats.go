// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"github.com/clivern/sloth/db"

	"github.com/rs/zerolog/log"
)

// Stats is the module for workspace dashboard metrics.
type Stats struct {
	WorkspaceRepository db.WorkspaceRepository
	StatsRepository     db.WorkspaceStatsRepository
}

// NewStats creates a stats module with the given repositories.
func NewStats(workspaces db.WorkspaceRepository, stats db.WorkspaceStatsRepository) *Stats {
	return &Stats{
		WorkspaceRepository: workspaces,
		StatsRepository:     stats,
	}
}

// WorkspaceStatsResponse is workspace metrics shaped for API responses.
type WorkspaceStatsResponse struct{}

// GetWorkspaceStats returns dashboard metrics for a workspace. TODO: remove if not used in future.
func (s *Stats) GetWorkspaceStats(workspaceId db.Id) (*WorkspaceStatsResponse, error) {
	_, err := s.StatsRepository.GetByWorkspaceId(workspaceId)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Msg("Failed to get workspace stats")
		return nil, err
	}

	return &WorkspaceStatsResponse{}, nil
}
