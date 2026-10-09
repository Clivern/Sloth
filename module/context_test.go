// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"context"
	"testing"

	"github.com/clivern/cognit/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitContext(t *testing.T) {
	t.Run("new app ctx defaults features to an empty map", func(t *testing.T) {
		ac := NewAppCtx("req-1", "agent", "10.0.0.1", nil)
		assert.NotNil(t, ac.Features)
		assert.False(t, ac.HasFeature("beta"))
	})

	t.Run("has feature reads enabled flags and is nil safe", func(t *testing.T) {
		ac := NewAppCtx("req-1", "agent", "10.0.0.1", map[string]bool{"beta": true})
		assert.True(t, ac.HasFeature("beta"))
		assert.False(t, ac.HasFeature("other"))

		var nilCtx *AppCtx
		assert.False(t, nilCtx.HasFeature("beta"))
	})

	t.Run("app ctx round trips through a context", func(t *testing.T) {
		ac := NewAppCtx("req-1", "agent", "10.0.0.1", nil)

		got, ok := AppCtxFrom(WithAppCtx(context.Background(), ac))
		require.True(t, ok)
		assert.Equal(t, "req-1", got.RequestID)

		_, ok = AppCtxFrom(context.Background())
		assert.False(t, ok)
	})

	t.Run("principal reports whether it is a user or an access key", func(t *testing.T) {
		user := &Principal{User: &db.User{Id: "user-1"}}
		assert.True(t, user.IsUser())
		assert.False(t, user.IsAccessKey())

		access := &Principal{AccessKey: &db.AccessKey{Id: "access-1"}}
		assert.False(t, access.IsUser())
		assert.True(t, access.IsAccessKey())

		var nilPrincipal *Principal
		assert.False(t, nilPrincipal.IsUser())
		assert.False(t, nilPrincipal.IsAccessKey())
	})

	t.Run("user and access key are read from the principal in context", func(t *testing.T) {
		ctx := WithPrincipal(context.Background(), &Principal{User: &db.User{Id: "user-1"}})

		user, ok := UserFrom(ctx)
		require.True(t, ok)
		assert.Equal(t, db.Id("user-1"), user.Id)

		_, ok = AccessKeyFrom(ctx)
		assert.False(t, ok)

		_, ok = UserFrom(context.Background())
		assert.False(t, ok)
	})

	t.Run("access key is read from the principal in context", func(t *testing.T) {
		ctx := WithPrincipal(context.Background(), &Principal{AccessKey: &db.AccessKey{Id: "access-1"}})

		key, ok := AccessKeyFrom(ctx)
		require.True(t, ok)
		assert.Equal(t, db.Id("access-1"), key.Id)

		_, ok = UserFrom(ctx)
		assert.False(t, ok)
	})
}
