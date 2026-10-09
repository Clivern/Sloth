// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegrationWorkspaceStatsRepository(t *testing.T) {
	database := openTestDB(t)
	workspace := createTestWorkspace(t, database)

	t.Run("returns empty stats", func(t *testing.T) {
		got, err := NewWorkspaceStatsRepository(database).GetByWorkspaceId(workspace.Id)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, &WorkspaceStats{}, got)
	})
}
