// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package db

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegrationWorkspaceKeyValueRepository(t *testing.T) {
	database := openTestDB(t)
	workspace := createTestWorkspace(t, database)
	other := createTestWorkspace(t, database)
	repo := NewWorkspaceKeyValueRepository(database)
	key := "agents/invoice-extractor/model"

	t.Run("upsert and get", func(t *testing.T) {
		item := &WorkspaceKeyValue{WorkspaceId: workspace.Id, Key: key, Value: "gpt-5"}
		require.NoError(t, repo.Upsert(item))
		assert.NotEmpty(t, item.Id.String())

		got, err := repo.Get(workspace.Id, key)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "gpt-5", got.Value)
		assert.Equal(t, workspace.Id, got.WorkspaceId)
		assert.Nil(t, got.ExpiresAt)
	})

	t.Run("replace value", func(t *testing.T) {
		require.NoError(t, repo.Upsert(&WorkspaceKeyValue{
			WorkspaceId: workspace.Id,
			Key:         key,
			Value:       "gpt-4",
		}))
		got, err := repo.Get(workspace.Id, key)
		require.NoError(t, err)
		assert.Equal(t, "gpt-4", got.Value)
	})

	t.Run("keys are isolated per workspace", func(t *testing.T) {
		require.NoError(t, repo.Upsert(&WorkspaceKeyValue{
			WorkspaceId: other.Id,
			Key:         key,
			Value:       "other",
		}))
		got, err := repo.Get(workspace.Id, key)
		require.NoError(t, err)
		assert.Equal(t, "gpt-4", got.Value)

		got, err = repo.Get(other.Id, key)
		require.NoError(t, err)
		assert.Equal(t, "other", got.Value)
	})

	t.Run("list by prefix", func(t *testing.T) {
		require.NoError(t, repo.Upsert(&WorkspaceKeyValue{
			WorkspaceId: workspace.Id,
			Key:         "agents/invoice-extractor/region",
			Value:       "eu",
		}))
		require.NoError(t, repo.Upsert(&WorkspaceKeyValue{
			WorkspaceId: workspace.Id,
			Key:         "flags/beta",
			Value:       "on",
		}))

		list, err := repo.ListByPrefix(workspace.Id, "agents/invoice-extractor/")
		require.NoError(t, err)
		require.Len(t, list, 2)
		assert.Equal(t, "agents/invoice-extractor/model", list[0].Key)
		assert.Equal(t, "agents/invoice-extractor/region", list[1].Key)
	})

	t.Run("expired key is hidden", func(t *testing.T) {
		expiredKey := "agents/invoice-extractor/stale"
		require.NoError(t, repo.Upsert(&WorkspaceKeyValue{
			WorkspaceId: workspace.Id,
			Key:         expiredKey,
			Value:       "gone",
			ExpiresAt:   timePtr(time.Now().UTC().Add(-time.Minute)),
		}))
		got, err := repo.Get(workspace.Id, expiredKey)
		require.NoError(t, err)
		assert.Nil(t, got)

		list, err := repo.ListByPrefix(workspace.Id, "agents/invoice-extractor/")
		require.NoError(t, err)
		assert.Len(t, list, 2)

		deleted, err := repo.DeleteExpired()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, deleted, int64(1))
	})

	t.Run("delete", func(t *testing.T) {
		require.NoError(t, repo.Delete(workspace.Id, key))
		got, err := repo.Get(workspace.Id, key)
		require.NoError(t, err)
		assert.Nil(t, got)

		got, err = repo.Get(other.Id, key)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "other", got.Value)
	})
}
