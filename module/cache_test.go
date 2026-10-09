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

func TestUnitCache(t *testing.T) {
	t.Run("get returns empty for a missing key", func(t *testing.T) {
		cache := NewCache(&fakeKeyValues{})

		value, expiresAt, err := cache.Get("missing")
		require.NoError(t, err)
		assert.Empty(t, value)
		assert.Nil(t, expiresAt)
	})

	t.Run("get returns the value and its expiry", func(t *testing.T) {
		expiresAt := time.Now().Add(time.Minute)
		cache := NewCache(&fakeKeyValues{item: &db.KeyValue{Key: "k", Value: "v", ExpiresAt: &expiresAt}})

		value, got, err := cache.Get("k")
		require.NoError(t, err)
		assert.Equal(t, "v", value)
		assert.Equal(t, &expiresAt, got)
	})

	t.Run("get returns the repository error", func(t *testing.T) {
		cache := NewCache(&fakeKeyValues{err: errors.New("boom")})

		_, _, err := cache.Get("k")
		assert.EqualError(t, err, "boom")
	})

	t.Run("set upserts the key with its expiry", func(t *testing.T) {
		store := &fakeKeyValues{}
		cache := NewCache(store)
		expiresAt := time.Now().Add(time.Minute)

		require.NoError(t, cache.Set("k", "v", &expiresAt))
		require.Len(t, store.upserted, 1)
		assert.Equal(t, "k", store.upserted[0].Key)
		assert.Equal(t, "v", store.upserted[0].Value)
		assert.Equal(t, &expiresAt, store.upserted[0].ExpiresAt)
	})

	t.Run("delete expired returns the removed count", func(t *testing.T) {
		cache := NewCache(&fakeKeyValues{expired: 4})

		count, err := cache.DeleteExpired()
		require.NoError(t, err)
		assert.Equal(t, int64(4), count)
	})
}
