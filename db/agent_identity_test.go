// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegrationAgentIdentityRepository(t *testing.T) {
	database := openTestDB(t)
	workspace := createTestWorkspace(t, database)
	agent := createTestAgent(t, database, workspace.Id, "identity-agent")
	repo := NewAgentIdentityRepository(database)

	t.Run("create and get", func(t *testing.T) {
		identity := &AgentIdentity{
			AgentId:  agent.Id,
			Name:     "gateway-key",
			Type:     AgentIdentityTypeAPIKey,
			Config:   `{"hash":"7a1c"}`,
			IsActive: true,
		}
		require.NoError(t, repo.Create(identity))

		got, err := repo.GetById(identity.Id)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, agent.Id, got.AgentId)
		assert.Equal(t, "gateway-key", got.Name)
		assert.Equal(t, AgentIdentityTypeAPIKey, got.Type)
		assert.True(t, got.IsActive)
		assert.Contains(t, got.Config, "7a1c")

		byName, err := repo.GetByAgentAndName(agent.Id, "gateway-key")
		require.NoError(t, err)
		require.NotNil(t, byName)
		assert.Equal(t, identity.Id, byName.Id)
	})

	t.Run("missing row", func(t *testing.T) {
		id, err := NewId()
		require.NoError(t, err)

		got, err := repo.GetById(id)
		require.NoError(t, err)
		assert.Nil(t, got)

		got, err = repo.GetByAgentAndName(agent.Id, "missing")
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("update list and delete", func(t *testing.T) {
		basic := &AgentIdentity{
			AgentId:  agent.Id,
			Name:     "basic",
			Type:     AgentIdentityTypeBasicAuth,
			Config:   `{"hash":"b91d"}`,
			IsActive: true,
		}
		require.NoError(t, repo.Create(basic))

		partner := &AgentIdentity{
			AgentId:  agent.Id,
			Name:     "partner-key",
			Type:     AgentIdentityTypeAPIKey,
			Config:   `{"hash":"c02e"}`,
			IsActive: true,
		}
		require.NoError(t, repo.Create(partner))

		basic.Config = `{"hash":"rotated"}`
		basic.IsActive = false
		require.NoError(t, repo.Update(basic))

		got, err := repo.GetById(basic.Id)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.False(t, got.IsActive)
		assert.Contains(t, got.Config, "rotated")

		list, err := repo.ListByAgentId(agent.Id)
		require.NoError(t, err)
		assert.Len(t, list, 3)

		active, err := repo.ListActiveByAgentId(agent.Id)
		require.NoError(t, err)
		assert.Len(t, active, 2)
		for _, item := range active {
			assert.True(t, item.IsActive)
		}

		require.NoError(t, repo.Delete(basic.Id))
		got, err = repo.GetById(basic.Id)
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("unique name per agent", func(t *testing.T) {
		duplicate := &AgentIdentity{
			AgentId:  agent.Id,
			Name:     "gateway-key",
			Type:     AgentIdentityTypeAPIKey,
			Config:   `{"hash":"d73f"}`,
			IsActive: true,
		}
		assert.Error(t, repo.Create(duplicate))
	})

	t.Run("cascade delete with agent", func(t *testing.T) {
		other := createTestAgent(t, database, workspace.Id, "cascade-identity-agent")
		identity := &AgentIdentity{
			AgentId:  other.Id,
			Name:     "gateway-key",
			Type:     AgentIdentityTypeAPIKey,
			Config:   `{"hash":"e58a"}`,
			IsActive: true,
		}
		require.NoError(t, repo.Create(identity))
		require.NoError(t, NewAgentRepository(database).Delete(other.Id))

		got, err := repo.GetById(identity.Id)
		require.NoError(t, err)
		assert.Nil(t, got)
	})
}
