// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegrationAsyncTaskRepository(t *testing.T) {
	database := openTestDB(t)
	workspace := createTestWorkspace(t, database)
	repo := NewAsyncTaskRepository(database)

	t.Run("create pending", func(t *testing.T) {
		task := &AsyncTask{
			WorkspaceId: workspace.Id,
			Type:        AsyncTaskTypeNoop,
			Payload:     stringPtr(`{"n":1}`),
			Priority:    10,
		}
		require.NoError(t, repo.Create(task))
		assert.Equal(t, AsyncTaskStatusPending, task.Status)

		pending, err := repo.ListByStatus(AsyncTaskStatusPending)
		require.NoError(t, err)
		found := false
		for _, item := range pending {
			if item.Id == task.Id {
				found = true
				break
			}
		}

		assert.True(t, found)

		require.NoError(t, repo.MarkRunning(task.Id))
		running, err := repo.CountByStatus(AsyncTaskStatusRunning)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, running, int64(1))

		require.NoError(t, repo.Complete(task.Id, `{"ok":true}`))
		completed, err := repo.CountByStatus(AsyncTaskStatusCompleted)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, completed, int64(1))
	})

	t.Run("fail", func(t *testing.T) {
		failing := &AsyncTask{
			WorkspaceId: workspace.Id,
			Type:        AsyncTaskTypeNoop,
			Payload:     stringPtr(`{}`),
		}
		require.NoError(t, repo.Create(failing))
		require.NoError(t, repo.MarkRunning(failing.Id))
		require.NoError(t, repo.Fail(failing.Id, "boom"))
		failed, err := repo.CountByStatus(AsyncTaskStatusFailed)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, failed, int64(1))
	})
}

func TestIntegrationAsyncTaskMetaRepository(t *testing.T) {
	database := openTestDB(t)
	workspace := createTestWorkspace(t, database)
	task := &AsyncTask{WorkspaceId: workspace.Id, Type: AsyncTaskTypeNoop}
	require.NoError(t, NewAsyncTaskRepository(database).Create(task))
	repo := NewAsyncTaskMetaRepository(database)

	t.Run("create and get", func(t *testing.T) {
		require.NoError(t, repo.Create(task.Id, "worker", "w1"))
		got, err := repo.Get(task.Id, "worker")
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "w1", got.Value)
	})

	t.Run("upsert", func(t *testing.T) {
		require.NoError(t, repo.Upsert(task.Id, "worker", "w2"))
		got, err := repo.Get(task.Id, "worker")
		require.NoError(t, err)
		assert.Equal(t, "w2", got.Value)
	})

	t.Run("list and delete", func(t *testing.T) {
		list, err := repo.ListByAsyncTaskId(task.Id)
		require.NoError(t, err)
		assert.Len(t, list, 1)

		require.NoError(t, repo.Delete(task.Id, "worker"))
		got, err := repo.Get(task.Id, "worker")
		require.NoError(t, err)
		assert.Nil(t, got)
	})
}
