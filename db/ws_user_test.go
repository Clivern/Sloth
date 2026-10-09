// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegrationWorkspaceUserRepository(t *testing.T) {
	database := openTestDB(t)
	workspace := createTestWorkspace(t, database)
	user := createTestUser(t, database)
	repo := NewWorkspaceUserRepository(database)
	member := &WorkspaceUser{
		WorkspaceId: workspace.Id,
		UserId:      user.Id,
		Role:        UserRoleRegular,
	}

	t.Run("create and get", func(t *testing.T) {
		require.NoError(t, repo.Create(member))

		got, err := repo.GetById(member.Id)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, UserRoleRegular, got.Role)

		byPair, err := repo.GetByWorkspaceAndUser(workspace.Id, user.Id)
		require.NoError(t, err)
		require.NotNil(t, byPair)
		assert.Equal(t, member.Id, byPair.Id)
	})

	t.Run("update", func(t *testing.T) {
		member.Role = UserRoleOwner
		require.NoError(t, repo.Update(member))
		got, err := repo.GetById(member.Id)
		require.NoError(t, err)
		assert.Equal(t, UserRoleOwner, got.Role)
	})

	t.Run("list and count", func(t *testing.T) {
		list, err := repo.ListByWorkspaceId(workspace.Id, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(list), 1)

		count, err := repo.CountByWorkspaceId(workspace.Id)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})

	t.Run("delete", func(t *testing.T) {
		require.NoError(t, repo.Delete(member.Id))
		got, err := repo.GetById(member.Id)
		require.NoError(t, err)
		assert.Nil(t, got)
	})
}
