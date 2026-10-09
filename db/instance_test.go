// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegrationAgentInstanceRepository(t *testing.T) {
	database := openTestDB(t)
	workspace := createTestWorkspace(t, database)
	agent := createTestAgent(t, database, workspace.Id, "invoice-extractor")
	repo := NewAgentInstanceRepository(database)

	t.Run("create and get", func(t *testing.T) {
		instance := createTestInstance(t, database, agent.Id, "invoice-extractor-3")
		assert.Equal(t, AgentInstanceStatusPassing, instance.Status)

		got, err := repo.GetById(instance.Id)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "10.0.12.41", got.Address)
		assert.Equal(t, 8080, got.Port)

		byKey, err := repo.GetByAgentAndInstanceId(agent.Id, "invoice-extractor-3")
		require.NoError(t, err)
		require.NotNil(t, byKey)
		assert.Equal(t, instance.Id, byKey.Id)
	})

	t.Run("missing row", func(t *testing.T) {
		id, err := NewId()
		require.NoError(t, err)

		got, err := repo.GetById(id)
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("update list and delete", func(t *testing.T) {
		instance := createTestInstance(t, database, agent.Id, "invoice-extractor-4")
		createTestInstance(t, database, agent.Id, "invoice-extractor-5")

		instance.Address = "10.0.12.42"
		instance.Port = 9090
		instance.Status = AgentInstanceStatusWarning
		dc := "eu-west-1"
		instance.Datacenter = &dc
		require.NoError(t, repo.Update(instance))

		got, err := repo.GetById(instance.Id)
		require.NoError(t, err)
		assert.Equal(t, "10.0.12.42", got.Address)
		assert.Equal(t, 9090, got.Port)
		assert.Equal(t, AgentInstanceStatusWarning, got.Status)
		require.NotNil(t, got.Datacenter)
		assert.Equal(t, "eu-west-1", *got.Datacenter)

		list, err := repo.ListByAgentId(agent.Id)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(list), 2)

		require.NoError(t, repo.Delete(instance.Id))
		got, err = repo.GetById(instance.Id)
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("unique instance id per agent", func(t *testing.T) {
		createTestInstance(t, database, agent.Id, "dup-instance")
		assert.Error(t, repo.Create(&AgentInstance{
			AgentId:    agent.Id,
			InstanceId: "dup-instance",
			Address:    "10.0.0.1",
			Port:       80,
		}))
	})
}

func TestIntegrationAgentInstanceMetaRepository(t *testing.T) {
	database := openTestDB(t)
	workspace := createTestWorkspace(t, database)
	agent := createTestAgent(t, database, workspace.Id, "meta-instance-agent")
	instance := createTestInstance(t, database, agent.Id, "meta-instance")
	repo := NewAgentInstanceMetaRepository(database)

	t.Run("create and get", func(t *testing.T) {
		require.NoError(t, repo.Create(instance.Id, "model", "gpt-5"))
		got, err := repo.Get(instance.Id, "model")
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "gpt-5", got.Value)
	})

	t.Run("upsert", func(t *testing.T) {
		require.NoError(t, repo.Upsert(instance.Id, "model", "gpt-4"))
		got, err := repo.Get(instance.Id, "model")
		require.NoError(t, err)
		assert.Equal(t, "gpt-4", got.Value)
	})

	t.Run("list and delete", func(t *testing.T) {
		list, err := repo.ListByAgentInstanceId(instance.Id)
		require.NoError(t, err)
		assert.Len(t, list, 1)

		require.NoError(t, repo.Delete(instance.Id, "model"))
		got, err := repo.Get(instance.Id, "model")
		require.NoError(t, err)
		assert.Nil(t, got)
	})
}
