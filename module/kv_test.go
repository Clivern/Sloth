// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"testing"
	"time"

	"github.com/clivern/cognit/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitKeyValue(t *testing.T) {
	t.Run("list returns an empty result for an invalid prefix and a missing workspace", func(t *testing.T) {
		kv := NewKeyValue(&fakeWorkspaceKeyValues{}, &fakeWorkspaces{})

		response, err := kv.ListKeyValue("ws-1", "a//b")
		require.NoError(t, err)
		assert.Empty(t, response.Items)
	})

	t.Run(
		"list escapes the LIKE underscore in the prefix and returns the items",
		func(t *testing.T) {
			items := &fakeWorkspaceKeyValues{
				items: []*db.WorkspaceKeyValue{{Id: "kv-1", Key: "app/a", Value: "1"}},
			}
			kv := NewKeyValue(items, &fakeWorkspaces{workspace: &db.Workspace{Id: "ws-1"}})

			result, err := kv.ListKeyValue("ws-1", "a_b")
			require.NoError(t, err)
			assert.Equal(t, []string{`a\_b`}, items.prefixes)
			assert.Equal(t, int64(1), result.Total)
			assert.Equal(t, "app/a", result.Items[0].Key)
		},
	)

	t.Run("list wraps the repository error", func(t *testing.T) {
		kv := NewKeyValue(
			&fakeWorkspaceKeyValues{err: errTest},
			&fakeWorkspaces{workspace: &db.Workspace{Id: "ws-1"}},
		)

		_, err := kv.ListKeyValue("ws-1", "app/")
		assert.ErrorIs(t, err, ErrFailedListKeyValue)
	})

	t.Run("get returns not found for an invalid key and for missing ones", func(t *testing.T) {
		workspace := &fakeWorkspaces{workspace: &db.Workspace{Id: "ws-1"}}

		_, err := NewKeyValue(&fakeWorkspaceKeyValues{}, workspace).GetKeyValue("ws-1", "a/../b")
		assert.ErrorIs(t, err, ErrKeyValueNotFound)

		_, err = NewKeyValue(&fakeWorkspaceKeyValues{}, workspace).GetKeyValue("ws-1", "app/theme")
		assert.ErrorIs(t, err, ErrKeyValueNotFound)
	})

	t.Run("get returns the item with its expiry", func(t *testing.T) {
		expiresAt := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
		item := &db.WorkspaceKeyValue{
			Id:          "kv-1",
			WorkspaceId: "ws-1",
			Key:         "app/theme",
			Value:       "dark",
			ExpiresAt:   &expiresAt,
		}
		kv := NewKeyValue(
			&fakeWorkspaceKeyValues{item: item},
			&fakeWorkspaces{workspace: &db.Workspace{Id: "ws-1"}},
		)

		response, err := kv.GetKeyValue("ws-1", "app/theme")
		require.NoError(t, err)
		assert.Equal(t, "dark", response.Value)
		require.NotNil(t, response.ExpiresAt)
		assert.Equal(t, "2026-06-01T00:00:00Z", *response.ExpiresAt)
	})

	t.Run("put rejects invalid keys and invalid expiry", func(t *testing.T) {
		workspace := &fakeWorkspaces{workspace: &db.Workspace{Id: "ws-1"}}

		_, err := NewKeyValue(&fakeWorkspaceKeyValues{}, workspace).
			PutKeyValue("ws-1", "app/../x", &PutKeyValueRequest{Value: "v"})
		assert.ErrorIs(t, err, ErrInvalidKeyValueKey)

		_, err = NewKeyValue(&fakeWorkspaceKeyValues{}, workspace).
			PutKeyValue("ws-1", "app/x", &PutKeyValueRequest{Value: "v", ExpiresAt: "tomorrow"})
		assert.ErrorIs(t, err, ErrInvalidExpiresAt)
	})

	t.Run("put upserts the item with a parsed expiry", func(t *testing.T) {
		store := &fakeWorkspaceKeyValues{}
		kv := NewKeyValue(store, &fakeWorkspaces{workspace: &db.Workspace{Id: "ws-1"}})

		response, err := kv.PutKeyValue(
			"ws-1",
			"app/theme",
			&PutKeyValueRequest{Value: "dark", ExpiresAt: "2026-06-01T00:00:00Z"},
		)
		require.NoError(t, err)
		require.Len(t, store.upserted, 1)
		assert.Equal(t, "app/theme", store.upserted[0].Key)
		require.NotNil(t, store.upserted[0].ExpiresAt)
		assert.Equal(t, "dark", response.Value)
		require.NotNil(t, response.ExpiresAt)
	})

	t.Run("put wraps the repository error", func(t *testing.T) {
		kv := NewKeyValue(
			&fakeWorkspaceKeyValues{err: errTest},
			&fakeWorkspaces{workspace: &db.Workspace{Id: "ws-1"}},
		)

		_, err := kv.PutKeyValue("ws-1", "app/theme", &PutKeyValueRequest{Value: "dark"})
		assert.ErrorIs(t, err, ErrFailedPutKeyValue)
	})

	t.Run("delete returns not found for a missing key", func(t *testing.T) {
		kv := NewKeyValue(&fakeWorkspaceKeyValues{}, &fakeWorkspaces{workspace: &db.Workspace{Id: "ws-1"}})

		err := kv.DeleteKeyValue("ws-1", "app/theme")
		assert.ErrorIs(t, err, ErrKeyValueNotFound)
	})

	t.Run("delete removes an existing key", func(t *testing.T) {
		store := &fakeWorkspaceKeyValues{item: &db.WorkspaceKeyValue{Key: "app/theme"}}
		kv := NewKeyValue(store, &fakeWorkspaces{workspace: &db.Workspace{Id: "ws-1"}})

		require.NoError(t, kv.DeleteKeyValue("ws-1", "app/theme"))
		assert.Equal(t, []string{"app/theme"}, store.deleted)
	})
}
