// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/clivern/cognit/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitAgent(t *testing.T) {
	t.Run("list returns not found for a missing workspace", func(t *testing.T) {
		f := newAgentFixture(nil, nil)

		_, err := f.agent.ListAgents("ws-1", 20, 0)
		assert.ErrorIs(t, err, ErrWorkspaceNotFound)
	})

	t.Run("list reports critical health when the agent has no instances", func(t *testing.T) {
		f := defaultAgentFixture()
		f.agents.list = []*db.Agent{f.agents.agent}
		f.agents.total = 1

		result, err := f.agent.ListAgents("ws-1", 20, 0)
		require.NoError(t, err)
		assert.Equal(t, int64(1), result.Total)
		require.Len(t, result.Agents, 1)
		assert.Equal(t, db.AgentInstanceStatusCritical, result.Agents[0].Health)
		assert.Empty(t, result.Agents[0].Instances)
	})

	t.Run("list reports passing health when an instance has no failing checks", func(t *testing.T) {
		f := defaultAgentFixture()
		f.agents.list = []*db.Agent{f.agents.agent}
		f.instances.list = []*db.AgentInstance{{Id: "inst-1", InstanceId: "node-1"}}

		result, err := f.agent.ListAgents("ws-1", 20, 0)
		require.NoError(t, err)
		assert.Equal(t, db.AgentInstanceStatusPassing, result.Agents[0].Health)
		assert.Equal(t, db.AgentInstanceStatusPassing, result.Agents[0].Instances[0].Status)
	})

	t.Run("list marks an instance with a critical check as critical", func(t *testing.T) {
		f := defaultAgentFixture()
		f.agents.list = []*db.Agent{f.agents.agent}
		f.instances.list = []*db.AgentInstance{{Id: "inst-1", InstanceId: "node-1"}}
		f.healthChecks.checks["inst-1"] = []*db.HealthCheck{{Type: db.HealthCheckTypeHTTP, Source: db.HealthCheckSourceAgent, Status: db.HealthCheckStatusCritical}}

		result, err := f.agent.ListAgents("ws-1", 20, 0)
		require.NoError(t, err)
		assert.Equal(t, db.AgentInstanceStatusCritical, result.Agents[0].Instances[0].Status)
		assert.Equal(t, db.AgentInstanceStatusCritical, result.Agents[0].Health)
	})

	t.Run("list makes the agent warning when its worst instance is warning", func(t *testing.T) {
		f := defaultAgentFixture()
		f.agents.list = []*db.Agent{f.agents.agent}
		f.instances.list = []*db.AgentInstance{{Id: "inst-1"}, {Id: "inst-2"}}
		f.healthChecks.checks["inst-2"] = []*db.HealthCheck{{Type: db.HealthCheckTypeHTTP, Source: db.HealthCheckSourceAgent, Status: db.HealthCheckStatusWarning}}

		result, err := f.agent.ListAgents("ws-1", 20, 0)
		require.NoError(t, err)
		assert.Equal(t, db.AgentInstanceStatusPassing, result.Agents[0].Instances[0].Status)
		assert.Equal(t, db.AgentInstanceStatusWarning, result.Agents[0].Instances[1].Status)
		assert.Equal(t, db.AgentInstanceStatusWarning, result.Agents[0].Health)
	})

	t.Run("list treats an expired ttl check as critical", func(t *testing.T) {
		f := defaultAgentFixture()
		f.agents.list = []*db.Agent{f.agents.agent}
		expired := time.Now().Add(-time.Minute)
		f.instances.list = []*db.AgentInstance{{Id: "inst-1"}}
		f.healthChecks.checks["inst-1"] = []*db.HealthCheck{{Type: db.HealthCheckTypeTTL, Source: db.HealthCheckSourceAgent, Status: db.HealthCheckStatusPassing, TTLExpiresAt: &expired}}

		result, err := f.agent.ListAgents("ws-1", 20, 0)
		require.NoError(t, err)
		assert.Equal(t, db.AgentInstanceStatusCritical, result.Agents[0].Instances[0].Status)
		assert.Equal(t, db.AgentInstanceStatusCritical, result.Agents[0].Health)
	})

	t.Run("get returns invalid name, not found and missing workspace errors", func(t *testing.T) {
		_, err := defaultAgentFixture().agent.GetAgent("ws-1", "-bad")
		assert.ErrorIs(t, err, ErrInvalidAgentName)

		noWorkspace := newAgentFixture(nil, nil)
		_, err = noWorkspace.agent.GetAgent("ws-1", "support-assistant")
		assert.ErrorIs(t, err, ErrWorkspaceNotFound)

		noAgent := newAgentFixture(&db.Workspace{Id: "ws-1"}, nil)
		_, err = noAgent.agent.GetAgent("ws-1", "support-assistant")
		assert.ErrorIs(t, err, ErrAgentNotFound)
	})

	t.Run("get returns the agent with its instance health", func(t *testing.T) {
		f := defaultAgentFixture()
		f.instances.list = []*db.AgentInstance{{Id: "inst-1", InstanceId: "node-1", Address: "10.0.0.12", Port: 8080}}

		response, err := f.agent.GetAgent("ws-1", "support-assistant")
		require.NoError(t, err)
		assert.Equal(t, db.Id("agent-1"), response.Id)
		assert.Equal(t, "1.0.0", response.Version)
		assert.Equal(t, db.AgentInstanceStatusPassing, response.Health)
		require.Len(t, response.Instances, 1)
		assert.Equal(t, "10.0.0.12", response.Instances[0].Address)
		assert.Equal(t, 8080, response.Instances[0].Port)
	})

	t.Run("upsert rejects invalid names and missing workspaces", func(t *testing.T) {
		req := &UpsertAgentRequest{Card: json.RawMessage(`{"version":"1.0.0"}`)}

		_, _, err := defaultAgentFixture().agent.UpsertAgent("ws-1", "-bad", req)
		assert.ErrorIs(t, err, ErrInvalidAgentName)

		noWorkspace := newAgentFixture(nil, nil)
		_, _, err = noWorkspace.agent.UpsertAgent("ws-1", "support-assistant", req)
		assert.ErrorIs(t, err, ErrWorkspaceNotFound)
	})

	t.Run("upsert rejects cards that are missing or not a JSON object", func(t *testing.T) {
		f := defaultAgentFixture()

		_, _, err := f.agent.UpsertAgent("ws-1", "support-assistant", nil)
		assert.ErrorIs(t, err, ErrInvalidAgentCard)

		_, _, err = f.agent.UpsertAgent("ws-1", "support-assistant", &UpsertAgentRequest{Card: json.RawMessage(`[1,2]`)})
		assert.ErrorIs(t, err, ErrInvalidAgentCard)

		_, _, err = f.agent.UpsertAgent("ws-1", "support-assistant", &UpsertAgentRequest{Card: json.RawMessage(`{"version":`)})
		assert.ErrorIs(t, err, ErrInvalidAgentCard)
	})

	t.Run("upsert requires a version in the request or the card", func(t *testing.T) {
		f := defaultAgentFixture()

		_, _, err := f.agent.UpsertAgent("ws-1", "support-assistant", &UpsertAgentRequest{Card: json.RawMessage(`{"name":"Support"}`)})
		assert.ErrorIs(t, err, ErrAgentVersionRequired)

		_, _, err = f.agent.UpsertAgent("ws-1", "support-assistant", &UpsertAgentRequest{Card: json.RawMessage(`{}`), Version: strings.Repeat("v", 41)})
		assert.ErrorIs(t, err, ErrAgentVersionRequired)
	})

	t.Run("upsert creates a new agent with a card checksum", func(t *testing.T) {
		f := newAgentFixture(&db.Workspace{Id: "ws-1"}, nil)

		response, created, err := f.agent.UpsertAgent("ws-1", "support-assistant", &UpsertAgentRequest{
			Card: json.RawMessage(`{"name":  "Support", "version":"2.1.0"}`),
		})
		require.NoError(t, err)
		assert.True(t, created)
		require.Len(t, f.agents.created, 1)
		assert.Empty(t, f.agents.updated)

		assert.Equal(t, "2.1.0", response.Version)
		assert.Equal(t, `{"name":"Support","version":"2.1.0"}`, string(response.Card))
		assert.Len(t, response.CardChecksum, 64)
	})

	t.Run("upsert updates an existing agent and prefers the request version", func(t *testing.T) {
		f := defaultAgentFixture()

		_, created, err := f.agent.UpsertAgent("ws-1", "support-assistant", &UpsertAgentRequest{
			Card:    json.RawMessage(`{"version":"1.0.0"}`),
			Version: "3.0.0",
		})
		require.NoError(t, err)
		assert.False(t, created)
		require.Len(t, f.agents.updated, 1)
		assert.Empty(t, f.agents.created)
		assert.Equal(t, "3.0.0", f.agents.updated[0].Version)
	})

	t.Run("delete returns not found for a missing agent", func(t *testing.T) {
		f := newAgentFixture(&db.Workspace{Id: "ws-1"}, nil)

		err := f.agent.DeleteAgent("ws-1", "support-assistant")
		assert.ErrorIs(t, err, ErrAgentNotFound)
	})

	t.Run("delete removes the agent", func(t *testing.T) {
		f := defaultAgentFixture()

		err := f.agent.DeleteAgent("ws-1", "support-assistant")
		require.NoError(t, err)
		assert.Equal(t, []db.Id{"agent-1"}, f.agents.deleted)
	})
}
