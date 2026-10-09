// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package db

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegrationHealthCheckRepository(t *testing.T) {
	database := openTestDB(t)
	workspace := createTestWorkspace(t, database)
	agent := createTestAgent(t, database, workspace.Id, "invoice-extractor")
	instance := createTestInstance(t, database, agent.Id, "invoice-extractor-3")
	repo := NewHealthCheckRepository(database)

	t.Run("create starts critical", func(t *testing.T) {
		check := &HealthCheck{
			AgentInstanceId: instance.Id,
			CheckId:         HealthCheckIDTTL,
			Name:            "lease",
			Type:            HealthCheckTypeTTL,
			Definition:      stringPtr(`{"ttl":"30s"}`),
		}
		require.NoError(t, repo.Create(check))
		assert.Equal(t, HealthCheckStatusCritical, check.Status)

		got, err := repo.GetByInstanceAndCheckId(instance.Id, HealthCheckIDTTL)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, check.Id, got.Id)
		assert.Contains(t, *got.Definition, "30s")
	})

	t.Run("pass warn fail", func(t *testing.T) {
		expires := time.Now().UTC().Add(30 * time.Second)
		require.NoError(t, repo.Pass(
			mustCheck(t, repo, instance.Id, HealthCheckIDTTL).Id,
			"renewed",
			&expires,
		))

		got, err := repo.GetByInstanceAndCheckId(instance.Id, HealthCheckIDTTL)
		require.NoError(t, err)
		assert.Equal(t, HealthCheckStatusPassing, got.Status)
		require.NotNil(t, got.TTLExpiresAt)
		assert.WithinDuration(t, expires, *got.TTLExpiresAt, time.Second)
		require.NotNil(t, got.Output)
		assert.Equal(t, "renewed", *got.Output)

		httpCheck := &HealthCheck{
			AgentInstanceId: instance.Id,
			CheckId:         "http",
			Name:            "/health",
			Type:            HealthCheckTypeHTTP,
			Definition:      stringPtr(`{"http":"http://10.0.12.41:8080/health","interval":"10s"}`),
			Status:          HealthCheckStatusPassing,
		}
		require.NoError(t, repo.Create(httpCheck))

		require.NoError(t, repo.Warn(httpCheck.Id, "429"))
		got, err = repo.GetById(httpCheck.Id)
		require.NoError(t, err)
		assert.Equal(t, HealthCheckStatusWarning, got.Status)

		require.NoError(t, repo.Fail(httpCheck.Id, "connection refused"))
		got, err = repo.GetById(httpCheck.Id)
		require.NoError(t, err)
		assert.Equal(t, HealthCheckStatusCritical, got.Status)

		list, err := repo.ListByAgentInstanceId(instance.Id)
		require.NoError(t, err)
		assert.Len(t, list, 2)

		require.NoError(t, repo.Delete(httpCheck.Id))
		got, err = repo.GetById(httpCheck.Id)
		require.NoError(t, err)
		assert.Nil(t, got)
	})
}

func TestIntegrationListLiveByAgentId(t *testing.T) {
	database := openTestDB(t)
	workspace := createTestWorkspace(t, database)
	agent := createTestAgent(t, database, workspace.Id, "live-agent")
	instances := NewAgentInstanceRepository(database)
	checks := NewHealthCheckRepository(database)
	now := time.Now().UTC()

	live := createTestInstance(t, database, agent.Id, "live-1")
	expired := createTestInstance(t, database, agent.Id, "expired-1")
	critical := createTestInstance(t, database, agent.Id, "critical-1")
	warning := createTestInstance(t, database, agent.Id, "warning-1")
	createTestInstance(t, database, agent.Id, "no-ttl-1")

	require.NoError(t, checks.Create(&HealthCheck{
		AgentInstanceId: live.Id,
		CheckId:         HealthCheckIDTTL,
		Name:            "lease",
		Type:            HealthCheckTypeTTL,
		Status:          HealthCheckStatusPassing,
		Definition:      stringPtr(`{"ttl":"30s"}`),
		TTLExpiresAt:    timePtr(now.Add(time.Minute)),
	}))
	require.NoError(t, checks.Create(&HealthCheck{
		AgentInstanceId: expired.Id,
		CheckId:         HealthCheckIDTTL,
		Name:            "lease",
		Type:            HealthCheckTypeTTL,
		Status:          HealthCheckStatusPassing,
		Definition:      stringPtr(`{"ttl":"30s"}`),
		TTLExpiresAt:    timePtr(now.Add(-time.Second)),
	}))
	require.NoError(t, checks.Create(&HealthCheck{
		AgentInstanceId: critical.Id,
		CheckId:         HealthCheckIDTTL,
		Name:            "lease",
		Type:            HealthCheckTypeTTL,
		Status:          HealthCheckStatusPassing,
		Definition:      stringPtr(`{"ttl":"30s"}`),
		TTLExpiresAt:    timePtr(now.Add(time.Minute)),
	}))
	require.NoError(t, checks.Create(&HealthCheck{
		AgentInstanceId: critical.Id,
		CheckId:         "http",
		Name:            "/health",
		Type:            HealthCheckTypeHTTP,
		Status:          HealthCheckStatusCritical,
		Definition:      stringPtr(`{"http":"http://10.0.12.41:8080/health"}`),
	}))
	require.NoError(t, checks.Create(&HealthCheck{
		AgentInstanceId: warning.Id,
		CheckId:         HealthCheckIDTTL,
		Name:            "lease",
		Type:            HealthCheckTypeTTL,
		Status:          HealthCheckStatusPassing,
		Definition:      stringPtr(`{"ttl":"30s"}`),
		TTLExpiresAt:    timePtr(now.Add(time.Minute)),
	}))
	require.NoError(t, checks.Create(&HealthCheck{
		AgentInstanceId: warning.Id,
		CheckId:         "card",
		Name:            "card freshness",
		Type:            HealthCheckTypeCard,
		Status:          HealthCheckStatusWarning,
		Definition:      stringPtr(`{"path":"/.well-known/agent-card.json"}`),
	}))

	list, err := instances.ListLiveByAgentId(agent.Id, now)
	require.NoError(t, err)

	ids := map[string]bool{}
	for _, item := range list {
		ids[item.InstanceId] = true
	}

	t.Run("includes live ttl", func(t *testing.T) {
		assert.True(t, ids["live-1"])
	})
	t.Run("includes warning", func(t *testing.T) {
		assert.True(t, ids["warning-1"])
	})
	t.Run("drops expired ttl", func(t *testing.T) {
		assert.False(t, ids["expired-1"])
	})
	t.Run("drops critical check", func(t *testing.T) {
		assert.False(t, ids["critical-1"])
	})
	t.Run("drops missing ttl", func(t *testing.T) {
		assert.False(t, ids["no-ttl-1"])
	})
}

func TestIntegrationHealthCheckMetaRepository(t *testing.T) {
	database := openTestDB(t)
	workspace := createTestWorkspace(t, database)
	agent := createTestAgent(t, database, workspace.Id, "health-meta-agent")
	instance := createTestInstance(t, database, agent.Id, "health-meta-instance")
	check := &HealthCheck{
		AgentInstanceId: instance.Id,
		CheckId:         "http",
		Name:            "/health",
		Type:            HealthCheckTypeHTTP,
		Status:          HealthCheckStatusPassing,
		Definition:      stringPtr(`{"http":"http://127.0.0.1/health"}`),
	}
	require.NoError(t, NewHealthCheckRepository(database).Create(check))
	repo := NewHealthCheckMetaRepository(database)

	t.Run("create and get", func(t *testing.T) {
		require.NoError(t, repo.Create(check.Id, "owner", "sre"))
		got, err := repo.Get(check.Id, "owner")
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "sre", got.Value)
	})

	t.Run("upsert", func(t *testing.T) {
		require.NoError(t, repo.Upsert(check.Id, "owner", "platform"))
		got, err := repo.Get(check.Id, "owner")
		require.NoError(t, err)
		assert.Equal(t, "platform", got.Value)
	})

	t.Run("list and delete", func(t *testing.T) {
		list, err := repo.ListByHealthCheckId(check.Id)
		require.NoError(t, err)
		assert.Len(t, list, 1)

		require.NoError(t, repo.Delete(check.Id, "owner"))
		got, err := repo.Get(check.Id, "owner")
		require.NoError(t, err)
		assert.Nil(t, got)
	})
}

func mustCheck(t *testing.T, repo HealthCheckRepository, instanceId Id, checkId string) *HealthCheck {
	t.Helper()
	got, err := repo.GetByInstanceAndCheckId(instanceId, checkId)
	require.NoError(t, err)
	require.NotNil(t, got)
	return got
}
