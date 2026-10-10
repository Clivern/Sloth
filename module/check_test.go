// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"strings"
	"testing"
	"time"

	"github.com/clivern/cognit/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitCheck(t *testing.T) {
	t.Run("upsert rejects malformed check ids", func(t *testing.T) {
		f := defaultCheckFixture()

		_, _, err := f.check.UpsertAgentCheck("ws-1", "support-assistant", "bad id!", &UpsertAgentCheckRequest{Type: db.HealthCheckTypeTTL, TTL: 30})
		assert.ErrorIs(t, err, ErrInvalidCheckId)
	})

	t.Run("upsert rejects a ttl check without a ttl", func(t *testing.T) {
		f := defaultCheckFixture()

		_, _, err := f.check.UpsertAgentCheck("ws-1", "support-assistant", "heartbeat", &UpsertAgentCheckRequest{Type: db.HealthCheckTypeTTL})
		assert.ErrorIs(t, err, ErrInvalidCheckTiming)
	})

	t.Run("upsert rejects http paths without a leading slash", func(t *testing.T) {
		f := defaultCheckFixture()

		_, _, err := f.check.UpsertAgentCheck("ws-1", "support-assistant", "http", &UpsertAgentCheckRequest{Type: db.HealthCheckTypeHTTP, Path: "healthz"})
		assert.ErrorIs(t, err, ErrInvalidCheckPath)
	})

	t.Run("upsert rejects a timeout that is not below the interval", func(t *testing.T) {
		f := defaultCheckFixture()

		_, _, err := f.check.UpsertAgentCheck("ws-1", "support-assistant", "tcp", &UpsertAgentCheckRequest{Type: db.HealthCheckTypeTCP, Port: 5432, Interval: 10, Timeout: 10})
		assert.ErrorIs(t, err, ErrInvalidCheckTiming)
	})

	t.Run("upsert returns not found for a missing agent", func(t *testing.T) {
		f := newCheckFixture(nil)

		_, _, err := f.check.UpsertAgentCheck("ws-1", "support-assistant", "heartbeat", &UpsertAgentCheckRequest{Type: db.HealthCheckTypeTTL, TTL: 30})
		assert.ErrorIs(t, err, ErrAgentNotFound)
	})

	t.Run("upsert creates an http template and copies it to every instance", func(t *testing.T) {
		f := defaultCheckFixture()
		f.instances.list = []*db.AgentInstance{{Id: "inst-1"}, {Id: "inst-2"}}

		response, created, err := f.check.UpsertAgentCheck("ws-1", "support-assistant", "http", &UpsertAgentCheckRequest{
			Type: db.HealthCheckTypeHTTP,
			Path: "/healthz",
		})
		require.NoError(t, err)
		assert.True(t, created)
		assert.Equal(t, 10, response.Interval)
		assert.Equal(t, 2, response.Timeout)
		assert.Equal(t, "http", response.Scheme)

		require.Len(t, f.agentChecks.created, 1)
		assert.Contains(t, f.agentChecks.created[0].Definition, `"interval":10`)

		require.Len(t, f.healthChecks.created, 2)
		for _, check := range f.healthChecks.created {
			assert.Equal(t, db.HealthCheckStatusUnknown, check.Status)
			assert.Equal(t, db.HealthCheckSourceAgent, check.Source)
			assert.NotNil(t, check.NextRunAt)
			require.NotNil(t, check.Output)
			assert.Equal(t, "Waiting for the first check", *check.Output)
		}
		assert.Equal(t, db.Id("inst-1"), f.healthChecks.created[0].InstanceId)
		assert.Equal(t, db.Id("inst-2"), f.healthChecks.created[1].InstanceId)
	})

	t.Run("upsert creates a ttl template as unknown with an expiry", func(t *testing.T) {
		f := defaultCheckFixture()
		f.instances.list = []*db.AgentInstance{{Id: "inst-1"}}

		_, _, err := f.check.UpsertAgentCheck("ws-1", "support-assistant", "heartbeat", &UpsertAgentCheckRequest{
			Type: db.HealthCheckTypeTTL,
			TTL:  30,
		})
		require.NoError(t, err)

		require.Len(t, f.healthChecks.created, 1)
		check := f.healthChecks.created[0]
		assert.Equal(t, db.HealthCheckStatusUnknown, check.Status)
		assert.NotNil(t, check.TTLExpiresAt)
		assert.Nil(t, check.NextRunAt)
	})

	t.Run("upsert resets a copied check when its type changes", func(t *testing.T) {
		f := defaultCheckFixture()
		f.instances.list = []*db.AgentInstance{{Id: "inst-1"}}
		f.healthChecks.byInstance["inst-1"] = &db.HealthCheck{Id: "check-1", Type: db.HealthCheckTypeTTL, Source: db.HealthCheckSourceAgent}

		_, _, err := f.check.UpsertAgentCheck("ws-1", "support-assistant", "heartbeat", &UpsertAgentCheckRequest{Type: db.HealthCheckTypeHTTP, Path: "/healthz"})
		require.NoError(t, err)

		assert.Equal(t, []db.Id{"check-1"}, f.healthChecks.deleted)
		require.Len(t, f.healthChecks.created, 1)
		assert.Equal(t, db.HealthCheckStatusUnknown, f.healthChecks.created[0].Status)
	})

	t.Run("upsert updates an existing template in place", func(t *testing.T) {
		f := defaultCheckFixture()
		f.instances.list = []*db.AgentInstance{{Id: "inst-1"}}
		f.agentChecks.template = &db.AgentCheck{Id: "tpl-1", AgentId: "agent-1", CheckId: "http", Name: "http", Type: db.HealthCheckTypeHTTP}
		f.healthChecks.byInstance["inst-1"] = &db.HealthCheck{Id: "check-1", Type: db.HealthCheckTypeHTTP, Source: db.HealthCheckSourceAgent}

		_, created, err := f.check.UpsertAgentCheck("ws-1", "support-assistant", "http", &UpsertAgentCheckRequest{
			Type: db.HealthCheckTypeHTTP,
			Path: "/ready",
			Name: "  Readiness  ",
		})
		require.NoError(t, err)
		assert.False(t, created)
		assert.Len(t, f.agentChecks.updated, 1)
		assert.Empty(t, f.agentChecks.created)

		require.Len(t, f.healthChecks.updated, 1)
		assert.Equal(t, "Readiness", f.healthChecks.updated[0].Name)
		assert.NotNil(t, f.healthChecks.updated[0].NextRunAt)
	})

	t.Run("delete returns not found when the template is missing", func(t *testing.T) {
		f := defaultCheckFixture()

		err := f.check.DeleteAgentCheck("ws-1", "support-assistant", "http")
		assert.ErrorIs(t, err, ErrCheckNotFound)
	})

	t.Run("delete removes the copies and the template", func(t *testing.T) {
		f := defaultCheckFixture()
		f.agentChecks.template = &db.AgentCheck{Id: "tpl-1", AgentId: "agent-1", CheckId: "http"}

		err := f.check.DeleteAgentCheck("ws-1", "support-assistant", "http")
		require.NoError(t, err)
		assert.Equal(t, []db.Id{"agent-1"}, f.healthChecks.deletedByAgent)
		assert.Equal(t, []db.Id{"tpl-1"}, f.agentChecks.deleted)
	})

	t.Run("list parses template definitions and returns the total", func(t *testing.T) {
		f := defaultCheckFixture()
		createdAt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
		f.agentChecks.total = 3
		f.agentChecks.list = []*db.AgentCheck{{
			CheckId:    "http",
			Name:       "http",
			Type:       db.HealthCheckTypeHTTP,
			Definition: `{"path":"/healthz","scheme":"http","port":8080,"interval":10,"timeout":2}`,
			CreatedAt:  createdAt,
			UpdatedAt:  createdAt,
		}}

		result, err := f.check.ListAgentChecks("ws-1", "support-assistant", 20, 0)
		require.NoError(t, err)
		assert.Equal(t, int64(3), result.Total)
		require.Len(t, result.Checks, 1)
		assert.Equal(t, "/healthz", result.Checks[0].Path)
		assert.Equal(t, 8080, result.Checks[0].Port)
		assert.Equal(t, "2026-01-02T03:04:05Z", result.Checks[0].CreatedAt)
	})

	t.Run("report returns not found for a missing instance or check", func(t *testing.T) {
		f := defaultCheckFixture()

		_, err := f.check.ReportCheck("ws-1", "support-assistant", "inst-1", "heartbeat", db.HealthCheckStatusPassing, &ReportCheckRequest{})
		assert.ErrorIs(t, err, ErrInstanceNotFound)

		f.instances.instance = &db.AgentInstance{Id: "inst-1", AgentId: "agent-1"}
		_, err = f.check.ReportCheck("ws-1", "support-assistant", "inst-1", "heartbeat", db.HealthCheckStatusPassing, &ReportCheckRequest{})
		assert.ErrorIs(t, err, ErrCheckNotFound)
	})

	t.Run("report rejects http checks", func(t *testing.T) {
		f := defaultCheckFixture()
		f.instances.instance = &db.AgentInstance{Id: "inst-1", AgentId: "agent-1"}

		f.healthChecks.byInstance["inst-1"] = &db.HealthCheck{Id: "check-1", Type: db.HealthCheckTypeHTTP, Source: db.HealthCheckSourceAgent}
		_, err := f.check.ReportCheck("ws-1", "support-assistant", "inst-1", "heartbeat", db.HealthCheckStatusPassing, &ReportCheckRequest{})
		assert.ErrorIs(t, err, ErrCheckNotReportable)

	})

	t.Run("report stores the status and restarts the ttl", func(t *testing.T) {
		f := defaultCheckFixture()
		f.instances.instance = &db.AgentInstance{Id: "inst-1", AgentId: "agent-1"}
		definition := `{"ttl":30}`
		f.healthChecks.byInstance["inst-1"] = &db.HealthCheck{
			Id:         "check-1",
			CheckId:    "heartbeat",
			Name:       "heartbeat",
			Type:       db.HealthCheckTypeTTL,
			Source:     db.HealthCheckSourceAgent,
			Definition: &definition,
		}

		response, err := f.check.ReportCheck("ws-1", "support-assistant", "inst-1", "heartbeat", db.HealthCheckStatusWarning, &ReportCheckRequest{Output: "queue is slow"})
		require.NoError(t, err)
		assert.Equal(t, db.HealthCheckStatusWarning, response.Status)
		require.NotNil(t, response.Output)
		assert.Equal(t, "queue is slow", *response.Output)

		require.Len(t, f.healthChecks.reports, 1)
		report := f.healthChecks.reports[0]
		assert.Equal(t, db.Id("check-1"), report.id)
		assert.Equal(t, db.HealthCheckStatusWarning, report.status)
		require.NotNil(t, report.ttlExpiresAt)
		assert.WithinDuration(t, time.Now().Add(30*time.Second), *report.ttlExpiresAt, 5*time.Second)
	})

	t.Run("report truncates output to the maximum size", func(t *testing.T) {
		f := defaultCheckFixture()
		f.instances.instance = &db.AgentInstance{Id: "inst-1", AgentId: "agent-1"}
		definition := `{"ttl":30}`
		f.healthChecks.byInstance["inst-1"] = &db.HealthCheck{Id: "check-1", Type: db.HealthCheckTypeTTL, Source: db.HealthCheckSourceAgent, Definition: &definition}

		_, err := f.check.ReportCheck("ws-1", "support-assistant", "inst-1", "heartbeat", db.HealthCheckStatusPassing, &ReportCheckRequest{
			Output: strings.Repeat("a", MaxCheckOutput+10),
		})
		require.NoError(t, err)
		assert.Len(t, f.healthChecks.reports[0].output, MaxCheckOutput)
	})
}
