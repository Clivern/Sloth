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

func TestIntegrationAPIKeyRepository(t *testing.T) {
	database := openTestDB(t)
	user := createTestUser(t, database)
	repo := NewAPIKeyRepository(database)
	token := "ak-" + uuid.NewString()
	key := &APIKey{UserId: user.Id, Name: "cli", Key: token}

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
		list, err := repo.ListByUserId(user.Id, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(list), 1)

		count, err := repo.CountByUserId(user.Id)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, int64(1))
	})

	t.Run("delete expired", func(t *testing.T) {
		expired := &APIKey{
			UserId:    user.Id,
			Name:      "old",
			Key:       "ak-" + uuid.NewString(),
			ExpiresAt: timePtr(time.Now().UTC().Add(-time.Minute)),
		}
		require.NoError(t, repo.Create(expired))
		deleted, err := repo.DeleteExpired()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, deleted, int64(1))
	})

	t.Run("delete", func(t *testing.T) {
		require.NoError(t, repo.Delete(key.Id))
		got, err := repo.GetById(key.Id)
		require.NoError(t, err)
		assert.Nil(t, got)
	})
}
