// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegrationEventRepository(t *testing.T) {
	database := openTestDB(t)
	workspace := createTestWorkspace(t, database)
	repo := NewEventRepository(database)

	event := &Event{
		WorkspaceId:  workspace.Id,
		Type:         "catalog.changed",
		ActorName:    stringPtr("Ada"),
		ActorType:    stringPtr(EventActorTypeUser),
		ResourceType: stringPtr("agent"),
		IPAddress:    stringPtr("10.0.0.1"),
		Data:         stringPtr(`{"agent":"weather"}`),
	}

	t.Run("create and get", func(t *testing.T) {
		require.NoError(t, repo.Create(event))
		got, err := repo.GetById(event.Id)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "catalog.changed", got.Type)
		assert.Equal(t, "Ada", *got.ActorName)
		assert.Equal(t, "10.0.0.1", *got.IPAddress)
		assert.Nil(t, got.ActorId)
		assert.JSONEq(t, `{"agent":"weather"}`, *got.Data)
	})

	t.Run("list and count", func(t *testing.T) {
		list, err := repo.ListByWorkspaceId(workspace.Id, 10, 0)
		require.NoError(t, err)
		require.Len(t, list, 1)
		assert.Equal(t, event.Id, list[0].Id)

		count, err := repo.CountByWorkspaceId(workspace.Id)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})

	t.Run("get unknown returns nil", func(t *testing.T) {
		got, err := repo.GetById(Id("00000000-0000-0000-0000-000000000000"))
		require.NoError(t, err)
		assert.Nil(t, got)
	})
}
