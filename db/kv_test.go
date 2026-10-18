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

func TestIntegrationKeyValueRepository(t *testing.T) {
	database := openTestDB(t)
	repo := NewKeyValueRepository(database)
	key := "k-" + uuid.NewString()[:8]

	t.Run("upsert and get", func(t *testing.T) {
		require.NoError(t, repo.Upsert(&KeyValue{Key: key, Value: "one"}))
		got, err := repo.Get(key)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "one", got.Value)
		assert.Nil(t, got.ExpiresAt)
	})

	t.Run("replace value", func(t *testing.T) {
		require.NoError(t, repo.Upsert(&KeyValue{Key: key, Value: "two"}))
		got, err := repo.Get(key)
		require.NoError(t, err)
		assert.Equal(t, "two", got.Value)
	})

	t.Run("expired key is hidden", func(t *testing.T) {
		expiredKey := "k-" + uuid.NewString()[:8]
		require.NoError(t, repo.Upsert(&KeyValue{
			Key:       expiredKey,
			Value:     "gone",
			ExpiresAt: timePtr(time.Now().UTC().Add(-time.Minute)),
		}))
		got, err := repo.Get(expiredKey)
		require.NoError(t, err)
		assert.Nil(t, got)

		deleted, err := repo.DeleteExpired()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, deleted, int64(1))
	})

	t.Run("delete", func(t *testing.T) {
		require.NoError(t, repo.Delete(key))
		got, err := repo.Get(key)
		require.NoError(t, err)
		assert.Nil(t, got)
	})
}
