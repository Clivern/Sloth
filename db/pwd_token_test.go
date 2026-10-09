// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package db

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegrationPasswordResetTokenRepository(t *testing.T) {
	database := openTestDB(t)
	user := createTestUser(t, database)
	repo := NewPasswordResetTokenRepository(database)
	token := "rst-" + uuid.NewString()

	t.Run("create and get", func(t *testing.T) {
		item := &PasswordResetToken{
			UserId:    user.Id,
			Token:     token,
			ExpiresAt: time.Now().UTC().Add(time.Hour),
		}
		require.NoError(t, repo.Create(item))

		got, err := repo.GetByToken(token)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, user.Id, got.UserId)
	})

	t.Run("delete by token", func(t *testing.T) {
		require.NoError(t, repo.DeleteByToken(token))
		got, err := repo.GetByToken(token)
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("delete expired", func(t *testing.T) {
		expired := &PasswordResetToken{
			UserId:    user.Id,
			Token:     "rst-" + uuid.NewString(),
			ExpiresAt: time.Now().UTC().Add(-time.Minute),
		}
		require.NoError(t, repo.Create(expired))
		deleted, err := repo.DeleteExpired()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, deleted, int64(1))
	})

	t.Run("delete by id", func(t *testing.T) {
		live := &PasswordResetToken{
			UserId:    user.Id,
			Token:     "rst-" + uuid.NewString(),
			ExpiresAt: time.Now().UTC().Add(time.Hour),
		}
		require.NoError(t, repo.Create(live))
		require.NoError(t, repo.Delete(live.Id))
		got, err := repo.GetByToken(live.Token)
		require.NoError(t, err)
		assert.Nil(t, got)
	})
}
