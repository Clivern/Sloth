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

func TestIntegrationSessionRepository(t *testing.T) {
	database := openTestDB(t)
	user := createTestUser(t, database)
	repo := NewSessionRepository(database)
	token := "sess-" + uuid.NewString()
	session := &Session{
		Token:     token,
		UserId:    user.Id,
		IPAddress: stringPtr("127.0.0.1"),
		UserAgent: stringPtr("test"),
		ExpiresAt: time.Now().UTC().Add(time.Hour),
	}

	t.Run("create and get", func(t *testing.T) {
		require.NoError(t, repo.Create(session))

		got, err := repo.GetByToken(token)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, user.Id, got.UserId)

		byId, err := repo.GetById(session.Id)
		require.NoError(t, err)
		require.NotNil(t, byId)

		list, err := repo.GetByUserId(user.Id)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(list), 1)
	})

	t.Run("is valid then expire", func(t *testing.T) {
		valid, err := repo.IsValid(token)
		require.NoError(t, err)
		assert.True(t, valid)

		require.NoError(t, repo.UpdateExpiration(session.Id, time.Now().UTC().Add(-time.Minute)))
		valid, err = repo.IsValid(token)
		require.NoError(t, err)
		assert.False(t, valid)

		deleted, err := repo.DeleteExpired()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, deleted, int64(1))
	})

	t.Run("count and delete by user", func(t *testing.T) {
		fresh := &Session{
			Token:     "sess-" + uuid.NewString(),
			UserId:    user.Id,
			ExpiresAt: time.Now().UTC().Add(time.Hour),
		}
		require.NoError(t, repo.Create(fresh))
		count, err := repo.CountByUserId(user.Id)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, int64(1))

		require.NoError(t, repo.DeleteByUserId(user.Id))
		list, err := repo.GetByUserId(user.Id)
		require.NoError(t, err)
		assert.Empty(t, list)
	})
}
