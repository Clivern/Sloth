// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegrationWorkspaceRepository(t *testing.T) {
	database := openTestDB(t)
	repo := NewWorkspaceRepository(database)
	user := createTestUser(t, database)
	workspace := createTestWorkspace(t, database)

	t.Run("get by id and handle", func(t *testing.T) {
		got, err := repo.GetById(workspace.Id)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, workspace.Handle, got.Handle)

		byHandle, err := repo.GetByHandle(workspace.Handle)
		require.NoError(t, err)
		require.NotNil(t, byHandle)
		assert.Equal(t, workspace.Id, byHandle.Id)
	})

	t.Run("update", func(t *testing.T) {
		workspace.Name = "renamed"
		require.NoError(t, repo.Update(workspace))
		got, err := repo.GetById(workspace.Id)
		require.NoError(t, err)
		assert.Equal(t, "renamed", got.Name)
	})

	t.Run("list membership and counts", func(t *testing.T) {
		require.NoError(t, NewWorkspaceUserRepository(database).Create(&WorkspaceUser{
			WorkspaceId: workspace.Id,
			UserId:      user.Id,
			Role:        UserRoleOwner,
		}))

		list, err := repo.List(10, 0, user.Id)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(list), 1)

		count, err := repo.Count(user.Id)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, int64(1))

		all, err := repo.CountAll()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, all, int64(1))

		member, err := repo.GetWorkspaceMembership(workspace.Id, user.Id)
		require.NoError(t, err)
		require.NotNil(t, member)
		assert.Equal(t, UserRoleOwner, member.Role)
	})

	t.Run("missing handle", func(t *testing.T) {
		missing, err := repo.GetByHandle("missing-handle")
		require.NoError(t, err)
		assert.Nil(t, missing)
	})
}

func TestIntegrationWorkspaceMetaRepository(t *testing.T) {
	database := openTestDB(t)
	workspace := createTestWorkspace(t, database)
	repo := NewWorkspaceMetaRepository(database)

	t.Run("create and get", func(t *testing.T) {
		require.NoError(t, repo.Create(workspace.Id, "plan", `{"v":"pro"}`))
		got, err := repo.Get(workspace.Id, "plan")
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Contains(t, got.Value, "pro")
	})

	t.Run("upsert", func(t *testing.T) {
		require.NoError(t, repo.Upsert(workspace.Id, "plan", `{"v":"team"}`))
		got, err := repo.Get(workspace.Id, "plan")
		require.NoError(t, err)
		assert.Contains(t, got.Value, "team")
	})

	t.Run("list and delete", func(t *testing.T) {
		list, err := repo.ListByWorkspaceId(workspace.Id)
		require.NoError(t, err)
		assert.Len(t, list, 1)

		require.NoError(t, repo.Delete(workspace.Id, "plan"))
		got, err := repo.Get(workspace.Id, "plan")
		require.NoError(t, err)
		assert.Nil(t, got)
	})
}
