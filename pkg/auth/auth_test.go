// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitAuth(t *testing.T) {
	t.Run("api key", func(t *testing.T) {
		auth := New(TypeAPIKey, map[string]string{"value": "secret"}, "enc-key")
		require.NoError(t, auth.Encrypt())
		assert.NotContains(t, auth.Config["value"], "secret")
		require.NoError(t, auth.Validate())
		require.NoError(t, auth.Match(map[string]string{"value": "secret"}))
		assert.ErrorIs(t, auth.Match(map[string]string{"value": "wrong"}), ErrCredentials)
	})

	t.Run("basic auth", func(t *testing.T) {
		auth := New(
			TypeBasicAuth,
			map[string]string{"username": "agent", "password": "pass"},
			"enc-key",
		)
		require.NoError(t, auth.Encrypt())
		require.NoError(t, auth.Validate())
		require.NoError(t, auth.Match(map[string]string{"username": "agent", "password": "pass"}))
		assert.ErrorIs(
			t,
			auth.Match(map[string]string{"username": "agent", "password": "wrong"}),
			ErrCredentials,
		)
	})

	t.Run("rejects incomplete config", func(t *testing.T) {
		empty := New(TypeAPIKey, map[string]string{}, "enc-key")
		require.NoError(t, empty.Encrypt())
		assert.ErrorIs(t, empty.Validate(), ErrConfig)

		basic := New(TypeBasicAuth, map[string]string{"username": "agent"}, "enc-key")
		require.NoError(t, basic.Encrypt())
		assert.ErrorIs(t, basic.Validate(), ErrConfig)
	})

	t.Run("rejects an unknown type", func(t *testing.T) {
		auth := New("client_credentials", map[string]string{"value": "x"}, "enc-key")
		require.NoError(t, auth.Encrypt())
		assert.ErrorIs(t, auth.Validate(), ErrType)
	})

	t.Run("rejects a bad encryption key", func(t *testing.T) {
		auth := New(TypeAPIKey, map[string]string{"value": "secret"}, "enc-key")
		require.NoError(t, auth.Encrypt())
		auth.EncryptionKey = "wrong"
		require.Error(t, auth.Validate())
	})
}
