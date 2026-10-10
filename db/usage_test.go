// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package db

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegrationUsageRepository(t *testing.T) {
	database := openTestDB(t)
	workspace := createTestWorkspace(t, database)
	repo := NewUsageRepository(database)
	start := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	usage := &Usage{
		WorkspaceId: workspace.Id,
		Type:        UsageTypeAITokens,
		Quantity:    10,
		Unit:        stringPtr(UsageUnitTokens),
		PeriodStart: start,
		PeriodEnd:   end,
	}

	t.Run("create and get", func(t *testing.T) {
		require.NoError(t, repo.Create(usage))
		got, err := repo.GetById(usage.Id)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, int64(10), got.Quantity)
	})

	t.Run("increment period", func(t *testing.T) {
		require.NoError(
			t,
			repo.IncrementByPeriod(workspace.Id, UsageTypeAITokens, start, end, 5, UsageUnitTokens),
		)
		qty, err := repo.GetQuantityByPeriod(workspace.Id, UsageTypeAITokens, start)
		require.NoError(t, err)
		assert.Equal(t, int64(15), qty)
	})

	t.Run("list and count", func(t *testing.T) {
		list, err := repo.ListByWorkspaceId(workspace.Id, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(list), 1)

		count, err := repo.CountByWorkspaceId(workspace.Id)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, int64(1))
	})

	t.Run("missing period", func(t *testing.T) {
		qty, err := repo.GetQuantityByPeriod(workspace.Id, "missing", start)
		require.NoError(t, err)
		assert.Equal(t, int64(0), qty)
	})
}
