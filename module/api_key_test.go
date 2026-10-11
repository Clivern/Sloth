// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"errors"
	"testing"
	"time"

	"github.com/clivern/cognit/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitAPIKey(t *testing.T) {
	user := &db.User{Id: "user-1"}

	t.Run("create rejects an invalid expiry", func(t *testing.T) {
		apiKeys := NewAPIKey(&fakeAPIKeys{})

		_, err := apiKeys.CreateAPIKey(&CreateAPIKeyRequest{Name: "ci", ExpiresAt: "tomorrow"}, user)
		assert.ErrorIs(t, err, ErrInvalidExpiresAt)
	})

	t.Run("create stores the key for the user and returns the secret once", func(t *testing.T) {
		store := &fakeAPIKeys{}
		apiKeys := NewAPIKey(store)

		response, err := apiKeys.CreateAPIKey(&CreateAPIKeyRequest{Name: "ci"}, user)
		require.NoError(t, err)
		require.Len(t, store.created, 1)
		assert.Equal(t, db.Id("user-1"), store.created[0].UserId)
		assert.NotEmpty(t, response.Key)
		assert.Equal(t, response.Key, store.created[0].Key)
		assert.Nil(t, response.ExpiresAt)
	})

	t.Run("create returns the repository error", func(t *testing.T) {
		_, err := NewAPIKey(&fakeAPIKeys{err: errors.New("boom")}).
			CreateAPIKey(&CreateAPIKeyRequest{Name: "ci"}, user)
		assert.EqualError(t, err, "boom")
	})

	t.Run("list returns the user's keys and the total", func(t *testing.T) {
		store := &fakeAPIKeys{list: []*db.APIKey{{Id: "key-1", Name: "ci"}}, total: 3}

		result, err := NewAPIKey(store).ListAPIKeys(user, 20, 0)
		require.NoError(t, err)
		assert.Equal(t, int64(3), result.Total)
		require.Len(t, result.APIKeys, 1)
		assert.Equal(t, "ci", result.APIKeys[0].Name)
	})

	t.Run("get returns not found for another user's key", func(t *testing.T) {
		store := &fakeAPIKeys{item: &db.APIKey{Id: "key-1", UserId: "user-2"}}

		_, err := NewAPIKey(store).GetAPIKey("key-1", user)
		assert.ErrorIs(t, err, ErrAPIKeyNotFound)
	})

	t.Run("get returns the owned key with its expiry", func(t *testing.T) {
		expiresAt := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
		store := &fakeAPIKeys{
			item: &db.APIKey{Id: "key-1", Name: "ci", UserId: "user-1", ExpiresAt: &expiresAt},
		}

		response, err := NewAPIKey(store).GetAPIKey("key-1", user)
		require.NoError(t, err)
		assert.Equal(t, "ci", response.Name)
		require.NotNil(t, response.ExpiresAt)
		assert.Equal(t, "2026-06-01T00:00:00Z", *response.ExpiresAt)
		assert.Empty(t, response.Key)
	})

	t.Run("delete returns not found for another user's key", func(t *testing.T) {
		err := NewAPIKey(&fakeAPIKeys{}).DeleteAPIKey("key-1", user)
		assert.ErrorIs(t, err, ErrAPIKeyNotFound)

		err = NewAPIKey(&fakeAPIKeys{item: &db.APIKey{Id: "key-1", UserId: "user-2"}}).
			DeleteAPIKey("key-1", user)
		assert.ErrorIs(t, err, ErrAPIKeyNotFound)
	})

	t.Run("delete removes the owned key", func(t *testing.T) {
		store := &fakeAPIKeys{item: &db.APIKey{Id: "key-1", UserId: "user-1"}}

		require.NoError(t, NewAPIKey(store).DeleteAPIKey("key-1", user))
		assert.Equal(t, []db.Id{"key-1"}, store.deleted)
	})
}
