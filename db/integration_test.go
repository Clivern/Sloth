// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegrationIntegrationsRepository(t *testing.T) {
	database := openTestDB(t)
	workspace := createTestWorkspace(t, database)
	repo := NewIntegrationRepository(database)
	item := &Integration{
		WorkspaceId: workspace.Id,
		Type:        "webhook",
		Name:        "alerts",
		Config:      stringPtr(`{"url":"https://example.test"}`),
		IsActive:    true,
	}

	t.Run("create and get", func(t *testing.T) {
		require.NoError(t, repo.Create(item))
		got, err := repo.GetById(item.Id)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "webhook", got.Type)
		assert.True(t, got.IsActive)
	})

	t.Run("update", func(t *testing.T) {
		item.Name = "alerts-prod"
		item.IsActive = false
		require.NoError(t, repo.Update(item))
		got, err := repo.GetById(item.Id)
		require.NoError(t, err)
		assert.Equal(t, "alerts-prod", got.Name)
		assert.False(t, got.IsActive)
	})

	t.Run("list and count", func(t *testing.T) {
		list, err := repo.ListByWorkspaceId(workspace.Id, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(list), 1)

		count, err := repo.CountByWorkspaceId(workspace.Id)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})

	t.Run("list active by event", func(t *testing.T) {
		subscribed := &Integration{
			WorkspaceId: workspace.Id,
			Type:        "slack",
			Name:        "ops",
			Config:      stringPtr(`{"url":"https://example.test","events":["catalog.changed","audit.log"]}`),
			IsActive:    true,
		}
		require.NoError(t, repo.Create(subscribed))

		list, err := repo.ListActiveByEvent(workspace.Id, "catalog.changed")
		require.NoError(t, err)
		require.Len(t, list, 1)
		assert.Equal(t, subscribed.Id, list[0].Id)

		list, err = repo.ListActiveByEvent(workspace.Id, "gateway.call.denied")
		require.NoError(t, err)
		assert.Empty(t, list)

		require.NoError(t, repo.Delete(subscribed.Id))
	})

	t.Run("delete", func(t *testing.T) {
		require.NoError(t, repo.Delete(item.Id))
		got, err := repo.GetById(item.Id)
		require.NoError(t, err)
		assert.Nil(t, got)
	})
}

func TestIntegrationIntegrationMetaRepository(t *testing.T) {
	database := openTestDB(t)
	workspace := createTestWorkspace(t, database)
	item := &Integration{
		WorkspaceId: workspace.Id,
		Type:        "slack",
		Name:        "ops",
		Config:      stringPtr(`{}`),
	}
	require.NoError(t, NewIntegrationRepository(database).Create(item))
	repo := NewIntegrationMetaRepository(database)

	t.Run("create and get", func(t *testing.T) {
		require.NoError(t, repo.Create(item.Id, "channel", `{"v":"alerts"}`))
		got, err := repo.Get(item.Id, "channel")
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Contains(t, got.Value, "alerts")
	})

	t.Run("upsert", func(t *testing.T) {
		require.NoError(t, repo.Upsert(item.Id, "channel", `{"v":"incidents"}`))
		got, err := repo.Get(item.Id, "channel")
		require.NoError(t, err)
		assert.Contains(t, got.Value, "incidents")
	})

	t.Run("list and delete", func(t *testing.T) {
		list, err := repo.ListByIntegrationId(item.Id)
		require.NoError(t, err)
		assert.Len(t, list, 1)

		require.NoError(t, repo.Delete(item.Id, "channel"))
		got, err := repo.Get(item.Id, "channel")
		require.NoError(t, err)
		assert.Nil(t, got)
	})
}
