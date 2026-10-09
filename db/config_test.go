// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package db

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegrationConfigRepository(t *testing.T) {
	database := openTestDB(t)
	repo := NewConfigRepository(database)
	key := "c-" + uuid.NewString()[:8]

	t.Run("create and get", func(t *testing.T) {
		require.NoError(t, repo.Create(key, "alpha"))
		got, err := repo.Get(key)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "alpha", got.Value)
	})

	t.Run("update", func(t *testing.T) {
		require.NoError(t, repo.Update(key, "beta"))
		got, err := repo.Get(key)
		require.NoError(t, err)
		assert.Equal(t, "beta", got.Value)
	})

	t.Run("list", func(t *testing.T) {
		list, err := repo.List()
		require.NoError(t, err)
		found := false
		for _, item := range list {
			if item.Key == key {
				found = true
				break
			}
		}

		assert.True(t, found)
	})

	t.Run("delete", func(t *testing.T) {
		require.NoError(t, repo.Delete(key))
		got, err := repo.Get(key)
		require.NoError(t, err)
		assert.Nil(t, got)
	})
}
