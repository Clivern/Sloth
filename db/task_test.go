// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegrationAgentTaskRepository(t *testing.T) {
	database := openTestDB(t)
	workspace := createTestWorkspace(t, database)
	agent := createTestAgent(t, database, workspace.Id, "user-directory")
	first := createTestInstance(t, database, agent.Id, "directory-1")
	second := createTestInstance(t, database, agent.Id, "directory-2")
	repo := NewAgentTaskRepository(database)

	t.Run("save and read back a task only for its caller", func(t *testing.T) {
		task := &AgentTask{
			WorkspaceId: workspace.Id,
			TaskId:      "task-7f3a",
			SourceAgent: "support-assistant",
			AgentId:     agent.Id,
			InstanceId:  &first.Id,
			Skill:       "user.profile.get",
		}
		require.NoError(t, repo.Save(task))
		assert.NotEmpty(t, task.Id)

		got, err := repo.GetBySourceAndTaskId(workspace.Id, "support-assistant", "task-7f3a")
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, agent.Id, got.AgentId)
		require.NotNil(t, got.InstanceId)
		assert.Equal(t, first.Id, *got.InstanceId)
		assert.Equal(t, "user.profile.get", got.Skill)

		got, err = repo.GetBySourceAndTaskId(workspace.Id, "other-agent", "task-7f3a")
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("saving a reused task id moves it", func(t *testing.T) {
		require.NoError(t, repo.Save(&AgentTask{WorkspaceId: workspace.Id, TaskId: "task-reused", SourceAgent: "support-assistant", AgentId: agent.Id, InstanceId: &first.Id}))
		require.NoError(t, repo.Save(&AgentTask{WorkspaceId: workspace.Id, TaskId: "task-reused", SourceAgent: "support-assistant", AgentId: agent.Id, InstanceId: &second.Id}))

		got, err := repo.GetBySourceAndTaskId(workspace.Id, "support-assistant", "task-reused")
		require.NoError(t, err)
		require.NotNil(t, got.InstanceId)
		assert.Equal(t, second.Id, *got.InstanceId)
	})

	t.Run("deleting the instance keeps the task without an instance", func(t *testing.T) {
		require.NoError(t, NewAgentInstanceRepository(database).Delete(first.Id))

		got, err := repo.GetBySourceAndTaskId(workspace.Id, "support-assistant", "task-7f3a")
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Nil(t, got.InstanceId)
	})
}
