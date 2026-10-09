// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/clivern/cognit/db"
	"github.com/clivern/cognit/migration"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func openModuleTestDB(t *testing.T) *sql.DB {
	t.Helper()

	envOr := func(key, fallback string) string {
		if value := os.Getenv(key); value != "" {
			return value
		}
		return fallback
	}
	port, err := strconv.Atoi(envOr("COGNIT_DATABASE_PORT", "5432"))
	require.NoError(t, err)

	conn, err := db.NewConnection(db.DatabaseConfig{
		Driver:   envOr("COGNIT_DATABASE_DRIVER", "postgres"),
		Host:     envOr("COGNIT_DATABASE_HOST", "localhost"),
		Port:     port,
		Username: envOr("COGNIT_DATABASE_USERNAME", "postgres"),
		Password: envOr("COGNIT_DATABASE_PASSWORD", "postgres"),
		Database: envOr("COGNIT_DATABASE_NAME", "cognit"),
	})
	if err != nil {
		t.Fatalf("postgres unavailable: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	mgr := migration.NewManager(conn.DB)
	for _, item := range migration.GetAll() {
		mgr.Register(item)
	}
	require.NoError(t, mgr.Up())
	return conn.DB
}

func TestIntegrationAgentChecks(t *testing.T) {
	database := openModuleTestDB(t)

	workspaces := db.NewWorkspaceRepository(database)
	agents := db.NewAgentRepository(database)
	instances := db.NewAgentInstanceRepository(database)
	healthChecks := db.NewHealthCheckRepository(database)
	agentChecks := db.NewAgentCheckRepository(database)

	agentModule := NewAgent(agents, instances, healthChecks, workspaces)
	instanceModule := NewInstance(agents, instances, healthChecks, agentChecks, workspaces)
	checkModule := NewCheck(agents, agentChecks, instances, healthChecks, workspaces)

	workspace := &db.Workspace{Name: "checks", Handle: fmt.Sprintf("t-%s", uuid.NewString())}
	require.NoError(t, workspaces.Create(workspace))
	t.Cleanup(func() { _ = workspaces.Delete(workspace.Id) })

	_, _, err := agentModule.UpsertAgent(workspace.Id, "support", &UpsertAgentRequest{
		Card: []byte(`{"name":"support","version":"1.0.0"}`),
	})
	require.NoError(t, err)

	register := func(instanceId string) *InstanceResponse {
		item, _, err := instanceModule.RegisterInstance(workspace.Id, "support", instanceId, &RegisterInstanceRequest{
			Address: "10.0.12.20",
			Port:    8080,
		})
		require.NoError(t, err)
		return item
	}
	checkIds := func(item *InstanceResponse) []string {
		ids := make([]string, 0, len(item.Checks))
		for _, check := range item.Checks {
			ids = append(ids, check.CheckId)
		}
		return ids
	}

	first := register("support-1")
	assert.Equal(t, db.AgentInstanceStatusPassing, first.Status)
	assert.Equal(t, []string{db.HealthCheckIDTTL}, checkIds(first))

	t.Run("rejects the reserved lease id", func(t *testing.T) {
		_, _, err := checkModule.UpsertAgentCheck(workspace.Id, "support", db.HealthCheckIDTTL, &UpsertAgentCheckRequest{Type: db.HealthCheckTypeTTL, TTL: 30})
		assert.ErrorIs(t, err, ErrReservedCheckId)
	})

	t.Run("new template reaches existing instances", func(t *testing.T) {
		check, created, err := checkModule.UpsertAgentCheck(workspace.Id, "support", "http", &UpsertAgentCheckRequest{
			Type: db.HealthCheckTypeHTTP,
			Path: "/healthz",
		})
		require.NoError(t, err)
		assert.True(t, created)
		assert.Equal(t, "http", check.Name)
		assert.Equal(t, DefaultCheckInterval, check.Interval)

		item, err := instanceModule.GetInstance(workspace.Id, "support", "support-1")
		require.NoError(t, err)
		assert.ElementsMatch(t, []string{"http", db.HealthCheckIDTTL}, checkIds(item))
		// A pull check is critical until the runner first probes it.
		assert.Equal(t, db.AgentInstanceStatusCritical, item.Status)

		agent, err := agentModule.GetAgent(workspace.Id, "support")
		require.NoError(t, err)
		assert.Equal(t, db.AgentInstanceStatusCritical, agent.Health)
	})

	t.Run("a passing probe makes the instance passing", func(t *testing.T) {
		instance, err := instances.GetByAgentAndInstanceId(first.AgentId, "support-1")
		require.NoError(t, err)
		check, err := healthChecks.GetByInstanceAndCheckId(instance.Id, "http")
		require.NoError(t, err)
		require.NotNil(t, check.NextRunAt)

		require.NoError(t, healthChecks.Record(check.Id, db.HealthCheckStatusPassing, "HTTP 200", time.Now().UTC().Add(time.Minute)))

		item, err := instanceModule.GetInstance(workspace.Id, "support", "support-1")
		require.NoError(t, err)
		assert.Equal(t, db.AgentInstanceStatusPassing, item.Status)

		live, err := instanceModule.ListInstances(workspace.Id, "support", true)
		require.NoError(t, err)
		require.Len(t, live, 1)
	})

	t.Run("instances registering later get the templates", func(t *testing.T) {
		second := register("support-2")
		assert.ElementsMatch(t, []string{"http", db.HealthCheckIDTTL}, checkIds(second))
	})

	t.Run("ttl template gets a grace window and can be reported", func(t *testing.T) {
		_, _, err := checkModule.UpsertAgentCheck(workspace.Id, "support", "heartbeat", &UpsertAgentCheckRequest{
			Type: db.HealthCheckTypeTTL,
			TTL:  60,
		})
		require.NoError(t, err)

		item, err := instanceModule.GetInstance(workspace.Id, "support", "support-1")
		require.NoError(t, err)
		for _, check := range item.Checks {
			if check.CheckId == "heartbeat" {
				assert.Equal(t, db.HealthCheckStatusPassing, check.Status)
				require.NotNil(t, check.TTLExpiresAt)
			}
		}

		reported, err := checkModule.ReportCheck(workspace.Id, "support", "support-1", "heartbeat", db.HealthCheckStatusWarning, &ReportCheckRequest{Output: "queue is backing up"})
		require.NoError(t, err)
		assert.Equal(t, db.HealthCheckStatusWarning, reported.Status)
		require.NotNil(t, reported.Output)
		assert.Equal(t, "queue is backing up", *reported.Output)

		item, err = instanceModule.GetInstance(workspace.Id, "support", "support-1")
		require.NoError(t, err)
		assert.Equal(t, db.AgentInstanceStatusWarning, item.Status)

		// Warning instances stay in discovery.
		live, err := instanceModule.ListInstances(workspace.Id, "support", true)
		require.NoError(t, err)
		assert.Len(t, live, 1)
	})

	t.Run("lease and pull checks cannot be reported", func(t *testing.T) {
		_, err := checkModule.ReportCheck(workspace.Id, "support", "support-1", db.HealthCheckIDTTL, db.HealthCheckStatusPassing, &ReportCheckRequest{})
		assert.ErrorIs(t, err, ErrCheckNotReportable)

		_, err = checkModule.ReportCheck(workspace.Id, "support", "support-1", "http", db.HealthCheckStatusPassing, &ReportCheckRequest{})
		assert.ErrorIs(t, err, ErrCheckNotReportable)

		_, err = checkModule.ReportCheck(workspace.Id, "support", "support-1", "missing", db.HealthCheckStatusPassing, &ReportCheckRequest{})
		assert.ErrorIs(t, err, ErrCheckNotFound)
	})

	t.Run("changing a template type resets the instance check", func(t *testing.T) {
		_, created, err := checkModule.UpsertAgentCheck(workspace.Id, "support", "http", &UpsertAgentCheckRequest{
			Type: db.HealthCheckTypeTCP,
		})
		require.NoError(t, err)
		assert.False(t, created)

		instance, err := instances.GetByAgentAndInstanceId(first.AgentId, "support-1")
		require.NoError(t, err)
		check, err := healthChecks.GetByInstanceAndCheckId(instance.Id, "http")
		require.NoError(t, err)
		assert.Equal(t, db.HealthCheckTypeTCP, check.Type)
		assert.Equal(t, db.HealthCheckStatusCritical, check.Status)
	})

	t.Run("deleting a template removes it from every instance", func(t *testing.T) {
		require.NoError(t, checkModule.DeleteAgentCheck(workspace.Id, "support", "http"))
		require.NoError(t, checkModule.DeleteAgentCheck(workspace.Id, "support", "heartbeat"))

		for _, id := range []string{"support-1", "support-2"} {
			item, err := instanceModule.GetInstance(workspace.Id, "support", id)
			require.NoError(t, err)
			assert.Equal(t, []string{db.HealthCheckIDTTL}, checkIds(item))
		}

		err := checkModule.DeleteAgentCheck(workspace.Id, "support", "http")
		assert.ErrorIs(t, err, ErrCheckNotFound)

		templates, err := checkModule.ListAgentChecks(workspace.Id, "support")
		require.NoError(t, err)
		assert.Empty(t, templates)
	})
}
