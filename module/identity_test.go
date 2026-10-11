// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"encoding/json"
	"testing"

	"github.com/clivern/cognit/db"
	"github.com/clivern/cognit/pkg/util"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitIdentity(t *testing.T) {
	t.Run("create issues an api key and stores only its hash", func(t *testing.T) {
		f := defaultIdentityFixture()

		got, err := f.identity.CreateIdentity(
			"ws-1",
			"support-assistant",
			&CreateIdentityRequest{Name: "default", Type: db.AgentIdentityTypeAPIKey},
		)
		require.NoError(t, err)
		require.Len(t, f.identities.created, 1)

		stored := f.identities.created[0]
		assert.Equal(t, db.Id("agent-1"), stored.AgentId)
		assert.True(t, stored.IsActive)
		assert.NotEmpty(t, got.Secret)
		assert.NotContains(t, stored.Config, got.Secret)

		var config IdentityConfig
		require.NoError(t, json.Unmarshal([]byte(stored.Config), &config))
		assert.True(t, util.ComparePassword(config.SecretHash, got.Secret))
		assert.Empty(t, config.Username)
	})

	t.Run("create stores the username of a basic auth identity", func(t *testing.T) {
		f := defaultIdentityFixture()

		got, err := f.identity.CreateIdentity(
			"ws-1",
			"support-assistant",
			&CreateIdentityRequest{Name: "partner", Type: db.AgentIdentityTypeBasicAuth, Username: "support"},
		)
		require.NoError(t, err)
		assert.Equal(t, "support", got.Username)
		assert.Contains(t, f.identities.created[0].Config, `"username":"support"`)
	})

	t.Run("create rejects basic auth without a username", func(t *testing.T) {
		f := defaultIdentityFixture()

		_, err := f.identity.CreateIdentity(
			"ws-1",
			"support-assistant",
			&CreateIdentityRequest{Name: "partner", Type: db.AgentIdentityTypeBasicAuth},
		)
		assert.ErrorIs(t, err, ErrInvalidIdentityConfig)
	})

	t.Run("create rejects an api key with a username", func(t *testing.T) {
		f := defaultIdentityFixture()

		_, err := f.identity.CreateIdentity(
			"ws-1",
			"support-assistant",
			&CreateIdentityRequest{Name: "default", Type: db.AgentIdentityTypeAPIKey, Username: "support"},
		)
		assert.ErrorIs(t, err, ErrInvalidIdentityConfig)
	})

	t.Run("create rejects a duplicate name", func(t *testing.T) {
		f := defaultIdentityFixture()
		f.identities.byName = &db.AgentIdentity{Id: "identity-0", AgentId: "agent-1", Name: "default"}

		_, err := f.identity.CreateIdentity(
			"ws-1",
			"support-assistant",
			&CreateIdentityRequest{Name: "default", Type: db.AgentIdentityTypeAPIKey},
		)
		assert.ErrorIs(t, err, ErrIdentityAlreadyExists)
	})

	t.Run("create returns not found for a missing agent", func(t *testing.T) {
		f := newIdentityFixture(nil)

		_, err := f.identity.CreateIdentity(
			"ws-1",
			"support-assistant",
			&CreateIdentityRequest{Name: "default", Type: db.AgentIdentityTypeAPIKey},
		)
		assert.ErrorIs(t, err, ErrAgentNotFound)
	})

	t.Run("list never returns secrets", func(t *testing.T) {
		f := defaultIdentityFixture()
		f.identities.list = []*db.AgentIdentity{{
			Id:       "identity-1",
			AgentId:  "agent-1",
			Name:     "partner",
			Type:     db.AgentIdentityTypeBasicAuth,
			Config:   `{"username":"support","secretHash":"hash"}`,
			IsActive: true,
		}}

		got, err := f.identity.ListIdentities("ws-1", "support-assistant")
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, "support", got[0].Username)
		assert.Empty(t, got[0].Secret)
	})

	t.Run("update toggles an identity off", func(t *testing.T) {
		f := defaultIdentityFixture()
		f.identities.identity = &db.AgentIdentity{
			Id:       "identity-1",
			AgentId:  "agent-1",
			Name:     "default",
			Type:     db.AgentIdentityTypeAPIKey,
			Config:   `{"secretHash":"hash"}`,
			IsActive: true,
		}

		got, err := f.identity.UpdateIdentity(
			"ws-1",
			"support-assistant",
			"identity-1",
			&UpdateIdentityRequest{Name: "default", IsActive: false},
		)
		require.NoError(t, err)
		assert.False(t, got.IsActive)
		require.Len(t, f.identities.updated, 1)
		assert.False(t, f.identities.updated[0].IsActive)
	})

	t.Run("update rejects a rename onto an existing name", func(t *testing.T) {
		f := defaultIdentityFixture()
		f.identities.identity = &db.AgentIdentity{Id: "identity-1", AgentId: "agent-1", Name: "default", Config: `{}`}
		f.identities.byName = &db.AgentIdentity{Id: "identity-2", AgentId: "agent-1", Name: "partner"}

		_, err := f.identity.UpdateIdentity(
			"ws-1",
			"support-assistant",
			"identity-1",
			&UpdateIdentityRequest{Name: "partner", IsActive: true},
		)
		assert.ErrorIs(t, err, ErrIdentityAlreadyExists)
	})

	t.Run("delete refuses an identity of another agent", func(t *testing.T) {
		f := defaultIdentityFixture()
		f.identities.identity = &db.AgentIdentity{Id: "identity-1", AgentId: "agent-2", Name: "default"}

		err := f.identity.DeleteIdentity("ws-1", "support-assistant", "identity-1")
		assert.ErrorIs(t, err, ErrIdentityNotFound)
		assert.Empty(t, f.identities.deleted)
	})

	t.Run("delete removes the identity", func(t *testing.T) {
		f := defaultIdentityFixture()
		f.identities.identity = &db.AgentIdentity{Id: "identity-1", AgentId: "agent-1", Name: "default"}

		err := f.identity.DeleteIdentity("ws-1", "support-assistant", "identity-1")
		require.NoError(t, err)
		assert.Equal(t, []db.Id{"identity-1"}, f.identities.deleted)
	})
}
