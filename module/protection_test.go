// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"encoding/json"
	"testing"

	"github.com/clivern/cognit/db"
	"github.com/clivern/cognit/pkg/encrypt"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitProtection(t *testing.T) {
	t.Run("save generates an api key and stores it encrypted", func(t *testing.T) {
		f := defaultProtectionFixture()

		got, err := f.protection.SaveProtection(
			"ws-1",
			"support-assistant",
			&SaveProtectionRequest{Type: db.AgentProtectionTypeAPIKey},
		)
		require.NoError(t, err)
		require.Len(t, f.protections.created, 1)

		stored := f.protections.created[0]
		assert.Equal(t, db.Id("agent-1"), stored.AgentId)
		assert.True(t, stored.IsActive)
		assert.NotEmpty(t, got.Secret)
		assert.Equal(t, "X-API-Key", got.Header)
		assert.NotContains(t, stored.Config, got.Secret)

		var config ProtectionConfig
		require.NoError(t, json.Unmarshal([]byte(stored.Config), &config))

		secret, err := encrypt.Decrypt(config.SecretEncrypted)
		require.NoError(t, err)
		assert.Equal(t, got.Secret, secret)
	})

	t.Run("save keeps a provided secret and does not echo it", func(t *testing.T) {
		f := defaultProtectionFixture()

		got, err := f.protection.SaveProtection(
			"ws-1",
			"support-assistant",
			&SaveProtectionRequest{Type: db.AgentProtectionTypeAPIKey, Header: "Authorization", Secret: "agent-own-secret"},
		)
		require.NoError(t, err)
		assert.Empty(t, got.Secret)
		assert.Equal(t, "Authorization", got.Header)

		var config ProtectionConfig
		require.NoError(t, json.Unmarshal([]byte(f.protections.created[0].Config), &config))

		secret, err := encrypt.Decrypt(config.SecretEncrypted)
		require.NoError(t, err)
		assert.Equal(t, "agent-own-secret", secret)
	})

	t.Run("save stores the username of a basic auth protection", func(t *testing.T) {
		f := defaultProtectionFixture()

		got, err := f.protection.SaveProtection(
			"ws-1",
			"support-assistant",
			&SaveProtectionRequest{Type: db.AgentProtectionTypeBasicAuth, Username: "cognit"},
		)
		require.NoError(t, err)
		assert.Equal(t, "cognit", got.Username)
		assert.Contains(t, f.protections.created[0].Config, `"username":"cognit"`)
	})

	t.Run("save replaces the existing protection and turns it back on", func(t *testing.T) {
		f := defaultProtectionFixture()
		f.protections.protection = &db.AgentProtection{
			Id:       "protection-1",
			AgentId:  "agent-1",
			Type:     db.AgentProtectionTypeBasicAuth,
			Config:   `{"username":"cognit","secretEncrypted":"cipher"}`,
			IsActive: false,
		}

		got, err := f.protection.SaveProtection(
			"ws-1",
			"support-assistant",
			&SaveProtectionRequest{Type: db.AgentProtectionTypeAPIKey},
		)
		require.NoError(t, err)
		assert.Empty(t, f.protections.created)
		require.Len(t, f.protections.updated, 1)
		assert.Equal(t, db.Id("protection-1"), got.Id)
		assert.Equal(t, db.AgentProtectionTypeAPIKey, f.protections.updated[0].Type)
		assert.True(t, f.protections.updated[0].IsActive)
		assert.NotContains(t, f.protections.updated[0].Config, "cipher")
	})

	t.Run("save rejects basic auth without a username", func(t *testing.T) {
		f := defaultProtectionFixture()

		_, err := f.protection.SaveProtection(
			"ws-1",
			"support-assistant",
			&SaveProtectionRequest{Type: db.AgentProtectionTypeBasicAuth},
		)
		assert.ErrorIs(t, err, ErrInvalidProtectionConfig)
	})

	t.Run("save rejects an api key with a username", func(t *testing.T) {
		f := defaultProtectionFixture()

		_, err := f.protection.SaveProtection(
			"ws-1",
			"support-assistant",
			&SaveProtectionRequest{Type: db.AgentProtectionTypeAPIKey, Username: "cognit"},
		)
		assert.ErrorIs(t, err, ErrInvalidProtectionConfig)
	})

	t.Run("save returns not found for a missing agent", func(t *testing.T) {
		f := newProtectionFixture(nil)

		_, err := f.protection.SaveProtection(
			"ws-1",
			"support-assistant",
			&SaveProtectionRequest{Type: db.AgentProtectionTypeAPIKey},
		)
		assert.ErrorIs(t, err, ErrAgentNotFound)
	})

	t.Run("get never returns the secret", func(t *testing.T) {
		f := defaultProtectionFixture()
		f.protections.protection = &db.AgentProtection{
			Id:       "protection-1",
			AgentId:  "agent-1",
			Type:     db.AgentProtectionTypeAPIKey,
			Config:   `{"header":"X-API-Key","secretEncrypted":"cipher"}`,
			IsActive: true,
		}

		got, err := f.protection.GetProtection("ws-1", "support-assistant")
		require.NoError(t, err)
		assert.Equal(t, "X-API-Key", got.Header)
		assert.Empty(t, got.Secret)
	})

	t.Run("get returns not found when the agent has no protection", func(t *testing.T) {
		f := defaultProtectionFixture()

		_, err := f.protection.GetProtection("ws-1", "support-assistant")
		assert.ErrorIs(t, err, ErrProtectionNotFound)
	})

	t.Run("update status turns the protection off", func(t *testing.T) {
		f := defaultProtectionFixture()
		f.protections.protection = &db.AgentProtection{
			Id:       "protection-1",
			AgentId:  "agent-1",
			Type:     db.AgentProtectionTypeAPIKey,
			Config:   `{"header":"X-API-Key","secretEncrypted":"cipher"}`,
			IsActive: true,
		}

		got, err := f.protection.UpdateProtectionStatus(
			"ws-1",
			"support-assistant",
			&UpdateProtectionStatusRequest{IsActive: false},
		)
		require.NoError(t, err)
		assert.False(t, got.IsActive)
		require.Len(t, f.protections.updated, 1)
		assert.Equal(t, `{"header":"X-API-Key","secretEncrypted":"cipher"}`, f.protections.updated[0].Config)
	})

	t.Run("delete returns not found when the agent has no protection", func(t *testing.T) {
		f := defaultProtectionFixture()

		err := f.protection.DeleteProtection("ws-1", "support-assistant")
		assert.ErrorIs(t, err, ErrProtectionNotFound)
		assert.Empty(t, f.protections.deleted)
	})

	t.Run("delete removes the protection", func(t *testing.T) {
		f := defaultProtectionFixture()
		f.protections.protection = &db.AgentProtection{Id: "protection-1", AgentId: "agent-1"}

		err := f.protection.DeleteProtection("ws-1", "support-assistant")
		require.NoError(t, err)
		assert.Equal(t, []db.Id{"protection-1"}, f.protections.deleted)
	})
}
