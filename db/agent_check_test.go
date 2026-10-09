// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegrationAgentCheckRepository(t *testing.T) {
	database := openTestDB(t)
	workspace := createTestWorkspace(t, database)
	agent := createTestAgent(t, database, workspace.Id, "checked-agent")
	repo := NewAgentCheckRepository(database)

	check := &AgentCheck{
		AgentId:    agent.Id,
		CheckId:    "http",
		Name:       "Health endpoint",
		Type:       HealthCheckTypeHTTP,
		Definition: `{"path":"/healthz","interval":10,"timeout":2}`,
	}

	t.Run("create and get", func(t *testing.T) {
		require.NoError(t, repo.Create(check))
		assert.False(t, check.CreatedAt.IsZero())

		got, err := repo.GetByAgentAndCheckId(agent.Id, "http")
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, check.Id, got.Id)
		assert.Equal(t, HealthCheckTypeHTTP, got.Type)
		assert.Contains(t, got.Definition, "/healthz")
	})

	t.Run("get missing returns nil", func(t *testing.T) {
		got, err := repo.GetByAgentAndCheckId(agent.Id, "missing")
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("update", func(t *testing.T) {
		check.Name = "Readiness"
		check.Definition = `{"path":"/readyz","interval":30,"timeout":5}`
		require.NoError(t, repo.Update(check))

		got, err := repo.GetByAgentAndCheckId(agent.Id, "http")
		require.NoError(t, err)
		assert.Equal(t, "Readiness", got.Name)
		assert.Contains(t, got.Definition, "/readyz")
	})

	t.Run("list and delete", func(t *testing.T) {
		require.NoError(t, repo.Create(&AgentCheck{
			AgentId:    agent.Id,
			CheckId:    "tcp",
			Name:       "Port open",
			Type:       HealthCheckTypeTCP,
			Definition: `{"interval":15,"timeout":2}`,
		}))

		list, err := repo.ListByAgentId(agent.Id)
		require.NoError(t, err)
		require.Len(t, list, 2)
		assert.Equal(t, "http", list[0].CheckId)
		assert.Equal(t, "tcp", list[1].CheckId)

		require.NoError(t, repo.Delete(check.Id))
		got, err := repo.GetByAgentAndCheckId(agent.Id, "http")
		require.NoError(t, err)
		assert.Nil(t, got)
	})
}
