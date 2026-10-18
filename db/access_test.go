// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package db

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegrationAccessKeyRepository(t *testing.T) {
	database := openTestDB(t)
	workspace := createTestWorkspace(t, database)
	repo := NewAccessKeyRepository(database)
	token := "wk-" + uuid.NewString()
	key := &AccessKey{
		WorkspaceId: workspace.Id,
		Name:        "gateway",
		Key:         token,
		Meta:        stringPtr(`{"scope":"a2a"}`),
	}

	t.Run("create and get", func(t *testing.T) {
		require.NoError(t, repo.Create(key))

		got, err := repo.GetById(key.Id)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, token, got.Key)

		byKey, err := repo.GetByKey(token)
		require.NoError(t, err)
		require.NotNil(t, byKey)
		assert.Equal(t, key.Id, byKey.Id)
	})

	t.Run("list and count", func(t *testing.T) {
		list, err := repo.ListByWorkspaceId(workspace.Id, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(list), 1)

		count, err := repo.CountByWorkspaceId(workspace.Id)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, int64(1))
	})

	t.Run("delete expired", func(t *testing.T) {
		expired := &AccessKey{
			WorkspaceId: workspace.Id,
			Name:        "old",
			Key:         "wk-" + uuid.NewString(),
			ExpiresAt:   timePtr(time.Now().UTC().Add(-time.Minute)),
		}
		require.NoError(t, repo.Create(expired))
		deleted, err := repo.DeleteExpired()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, deleted, int64(1))

		got, err := repo.GetByKey(expired.Key)
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("delete", func(t *testing.T) {
		require.NoError(t, repo.Delete(key.Id))
		got, err := repo.GetById(key.Id)
		require.NoError(t, err)
		assert.Nil(t, got)
	})
}
