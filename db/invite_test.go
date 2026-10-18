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

func TestIntegrationUserInviteRepository(t *testing.T) {
	database := openTestDB(t)
	workspace := createTestWorkspace(t, database)
	inviter := createTestUser(t, database)
	repo := NewUserInviteRepository(database)
	email := "inv-" + uuid.NewString()[:8] + "@t.test"
	token := "inv-" + uuid.NewString()
	invite := &UserInvite{
		Email:         email,
		Role:          UserRoleRegular,
		Token:         token,
		Status:        "pending",
		InviterUserId: inviter.Id,
		WorkspaceId:   workspace.Id,
		ExpiresAt:     time.Now().UTC().Add(time.Hour),
	}

	t.Run("create and get", func(t *testing.T) {
		require.NoError(t, repo.Create(invite))

		got, err := repo.GetById(invite.Id)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, email, got.Email)

		byToken, err := repo.GetByToken(token)
		require.NoError(t, err)
		require.NotNil(t, byToken)
	})

	t.Run("pending by email", func(t *testing.T) {
		pending, err := repo.ListPendingByEmail(email)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(pending), 1)

		count, err := repo.CountPendingByEmailInWorkspace(workspace.Id, email)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})

	t.Run("accept", func(t *testing.T) {
		now := time.Now().UTC()
		require.NoError(t, repo.UpdateStatus(invite.Id, "accepted", &now))
		got, err := repo.GetById(invite.Id)
		require.NoError(t, err)
		assert.Equal(t, "accepted", got.Status)
	})

	t.Run("mark expired", func(t *testing.T) {
		expired := &UserInvite{
			Email:         "exp-" + uuid.NewString()[:8] + "@t.test",
			Role:          UserRoleRegular,
			Token:         "inv-" + uuid.NewString(),
			Status:        "pending",
			InviterUserId: inviter.Id,
			WorkspaceId:   workspace.Id,
			ExpiresAt:     time.Now().UTC().Add(-time.Minute),
		}
		require.NoError(t, repo.Create(expired))
		marked, err := repo.MarkExpiredAsExpired()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, marked, int64(1))
	})

	t.Run("list and delete", func(t *testing.T) {
		list, err := repo.ListByWorkspaceId(workspace.Id, 20, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(list), 1)

		require.NoError(t, repo.Delete(invite.Id))
		got, err := repo.GetById(invite.Id)
		require.NoError(t, err)
		assert.Nil(t, got)
	})
}
