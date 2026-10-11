// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/clivern/cognit/db"

	"github.com/samber/lo"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitInstance(t *testing.T) {
	t.Run("list returns not found for a missing agent", func(t *testing.T) {
		_, err := newInstanceFixture(&db.Workspace{Id: "ws-1"}, nil).
			instance.
			ListInstances("ws-1", "-bad", 50, 0)
		assert.ErrorIs(t, err, ErrAgentNotFound)
	})

	t.Run("list healthy returns only live instances", func(t *testing.T) {
		f := defaultInstanceFixture()
		f.instances.list = []*db.AgentInstance{
			{Id: "inst-1", InstanceId: "node-1"},
			{Id: "inst-2", InstanceId: "node-2"},
		}
		f.instances.live = []*db.AgentInstance{{Id: "inst-1", InstanceId: "node-1"}}

		all, err := f.instance.ListInstances("ws-1", "support-assistant", 50, 0)
		require.NoError(t, err)
		assert.Len(t, all.Instances, 2)

		live, err := f.instance.ListHealthyInstances("ws-1", "support-assistant", 50, 0)
		require.NoError(t, err)
		require.Len(t, live.Instances, 1)
		assert.Equal(t, "node-1", live.Instances[0].InstanceId)
		assert.Equal(t, int64(1), live.Total)
	})

	t.Run("list returns one page and the total across pages", func(t *testing.T) {
		f := defaultInstanceFixture()
		f.instances.list = []*db.AgentInstance{
			{Id: "inst-1", InstanceId: "node-1"},
			{Id: "inst-2", InstanceId: "node-2"},
			{Id: "inst-3", InstanceId: "node-3"},
		}

		first, err := f.instance.ListInstances("ws-1", "support-assistant", 2, 0)
		require.NoError(t, err)
		require.Len(t, first.Instances, 2)
		assert.Equal(t, int64(3), first.Total)

		last, err := f.instance.ListInstances("ws-1", "support-assistant", 2, 2)
		require.NoError(t, err)
		require.Len(t, last.Instances, 1)
		assert.Equal(t, "node-3", last.Instances[0].InstanceId)
		assert.Equal(t, int64(3), last.Total)
	})

	t.Run("list treats an instance without checks as passing", func(t *testing.T) {
		f := defaultInstanceFixture()
		f.instances.list = []*db.AgentInstance{{Id: "inst-1", InstanceId: "node-1"}}

		list, err := f.instance.ListInstances("ws-1", "support-assistant", 50, 0)
		require.NoError(t, err)
		assert.Equal(t, db.AgentInstanceStatusPassing, list.Instances[0].Status)
		assert.Empty(t, list.Instances[0].Checks)
	})

	t.Run("list makes an instance critical when one check is critical", func(t *testing.T) {
		f := defaultInstanceFixture()
		f.instances.list = []*db.AgentInstance{{Id: "inst-1", InstanceId: "node-1"}}
		f.healthChecks.checks["inst-1"] = []*db.HealthCheck{
			{
				CheckId: "http",
				Name:    "http",
				Type:    db.HealthCheckTypeHTTP,
				Source:  db.HealthCheckSourceAgent,
				Status:  db.HealthCheckStatusCritical,
			},
		}

		list, err := f.instance.ListInstances("ws-1", "support-assistant", 50, 0)
		require.NoError(t, err)
		assert.Equal(t, db.AgentInstanceStatusCritical, list.Instances[0].Status)
	})

	t.Run("list makes an instance warning when one check warns", func(t *testing.T) {
		f := defaultInstanceFixture()
		f.instances.list = []*db.AgentInstance{{Id: "inst-1", InstanceId: "node-1"}}
		f.healthChecks.checks["inst-1"] = []*db.HealthCheck{
			{
				CheckId: "http",
				Name:    "http",
				Type:    db.HealthCheckTypeHTTP,
				Source:  db.HealthCheckSourceAgent,
				Status:  db.HealthCheckStatusPassing,
			},
			{
				CheckId: "tcp",
				Name:    "tcp",
				Type:    db.HealthCheckTypeTCP,
				Source:  db.HealthCheckSourceAgent,
				Status:  db.HealthCheckStatusWarning,
			},
		}

		list, err := f.instance.ListInstances("ws-1", "support-assistant", 50, 0)
		require.NoError(t, err)
		assert.Equal(t, db.AgentInstanceStatusWarning, list.Instances[0].Status)
		require.Len(t, list.Instances[0].Checks, 2)
	})

	t.Run("list ranks unknown between critical and warning", func(t *testing.T) {
		f := defaultInstanceFixture()
		f.instances.list = []*db.AgentInstance{
			{Id: "inst-1", InstanceId: "node-1"},
			{Id: "inst-2", InstanceId: "node-2"},
		}
		f.healthChecks.checks["inst-1"] = []*db.HealthCheck{
			{
				CheckId: "http",
				Name:    "http",
				Type:    db.HealthCheckTypeHTTP,
				Source:  db.HealthCheckSourceAgent,
				Status:  db.HealthCheckStatusWarning,
			},
			{
				CheckId: "tcp",
				Name:    "tcp",
				Type:    db.HealthCheckTypeTCP,
				Source:  db.HealthCheckSourceAgent,
				Status:  db.HealthCheckStatusUnknown,
			},
		}
		f.healthChecks.checks["inst-2"] = []*db.HealthCheck{
			{
				CheckId: "http",
				Name:    "http",
				Type:    db.HealthCheckTypeHTTP,
				Source:  db.HealthCheckSourceAgent,
				Status:  db.HealthCheckStatusUnknown,
			},
			{
				CheckId: "tcp",
				Name:    "tcp",
				Type:    db.HealthCheckTypeTCP,
				Source:  db.HealthCheckSourceAgent,
				Status:  db.HealthCheckStatusCritical,
			},
		}

		list, err := f.instance.ListInstances("ws-1", "support-assistant", 50, 0)
		require.NoError(t, err)
		assert.Equal(t, db.AgentInstanceStatusUnknown, list.Instances[0].Status)
		assert.Equal(t, db.AgentInstanceStatusCritical, list.Instances[1].Status)
	})

	t.Run("list makes an instance critical when a ttl check has expired", func(t *testing.T) {
		f := defaultInstanceFixture()
		f.instances.list = []*db.AgentInstance{{Id: "inst-1", InstanceId: "node-1"}}
		expired := time.Now().UTC().Add(-time.Minute)
		f.healthChecks.checks["inst-1"] = []*db.HealthCheck{
			{
				CheckId:      "heartbeat",
				Name:         "heartbeat",
				Type:         db.HealthCheckTypeTTL,
				Source:       db.HealthCheckSourceAgent,
				Status:       db.HealthCheckStatusPassing,
				TTLExpiresAt: &expired,
			},
		}

		list, err := f.instance.ListInstances("ws-1", "support-assistant", 50, 0)
		require.NoError(t, err)
		assert.Equal(t, db.AgentInstanceStatusCritical, list.Instances[0].Status)
	})

	t.Run("get health returns the worst check status and its checks", func(t *testing.T) {
		f := defaultInstanceFixture()
		f.healthChecks.checks["inst-1"] = []*db.HealthCheck{
			{
				CheckId: "http",
				Name:    "http",
				Type:    db.HealthCheckTypeHTTP,
				Source:  db.HealthCheckSourceAgent,
				Status:  db.HealthCheckStatusWarning,
			},
			{
				CheckId: "tcp",
				Name:    "tcp",
				Type:    db.HealthCheckTypeTCP,
				Source:  db.HealthCheckSourceAgent,
				Status:  db.HealthCheckStatusUnknown,
			},
		}

		status, checks, err := f.instance.GetHealth("inst-1")
		require.NoError(t, err)
		assert.Equal(t, db.AgentInstanceStatusUnknown, status)
		assert.Len(t, checks, 2)

		status, checks, err = defaultInstanceFixture().instance.GetHealth("inst-2")
		require.NoError(t, err)
		assert.Equal(t, db.AgentInstanceStatusPassing, status)
		assert.Empty(t, checks)
	})

	t.Run("get returns the instance with its checks", func(t *testing.T) {
		f := defaultInstanceFixture()
		f.instances.instance = &db.AgentInstance{
			Id:         "inst-1",
			AgentId:    "agent-1",
			InstanceId: "node-1",
			Address:    "10.0.0.12",
			Port:       8080,
		}

		item, err := f.instance.GetInstance("ws-1", "support-assistant", "node-1")
		require.NoError(t, err)
		assert.Equal(t, db.AgentInstanceStatusPassing, item.Status)
		assert.Equal(t, "10.0.0.12", item.Address)
		assert.Equal(t, 8080, item.Port)
		assert.Empty(t, item.Checks)
	})

	t.Run("get returns not found for an unknown agent", func(t *testing.T) {
		_, err := newInstanceFixture(&db.Workspace{Id: "ws-1"}, nil).
			instance.
			GetInstance("ws-1", "-bad", "node-1")
		assert.ErrorIs(t, err, ErrAgentNotFound)
	})

	t.Run("get returns not found for a missing instance", func(t *testing.T) {
		_, err := defaultInstanceFixture().
			instance.
			GetInstance("ws-1", "support-assistant", "node-1")
		assert.ErrorIs(t, err, ErrInstanceNotFound)
	})

	t.Run("register creates the instance without a lease check", func(t *testing.T) {
		f := defaultInstanceFixture()

		item, created, err := f.instance.RegisterInstance(
			"ws-1",
			"support-assistant",
			"node-1",
			&RegisterInstanceRequest{
				Address: "10.0.0.12",
				Port:    8080,
				Meta:    json.RawMessage(`{"region":"eu"}`),
			},
			"en",
		)
		require.NoError(t, err)
		assert.True(t, created)
		assert.Equal(t, "10.0.0.12", item.Address)
		assert.Equal(t, db.AgentInstanceStatusPassing, item.Status)

		require.Len(t, f.instances.created, 1)
		require.NotNil(t, f.instances.created[0].Meta)
		assert.Equal(t, `{"region":"eu"}`, *f.instances.created[0].Meta)

		// Nothing is created for the instance unless the agent defines a check.
		assert.Empty(t, f.healthChecks.created)
	})

	t.Run("register fetches the card from the first instance and points it at the gateway", func(t *testing.T) {
		viper.Set("app.url", "https://cognit.test")
		f := defaultInstanceFixture()
		f.agents.agent.CardPath = "/.well-known/agent-card.json"

		var fetched string
		f.instance.FetchCard = func(ctx context.Context, address string, port int, path string) (map[string]any, error) {
			fetched = fmt.Sprintf("%s:%d%s", address, port, path)
			return map[string]any{
				"name":                "Support",
				"url":                 "http://10.0.0.12:8080",
				"supportedInterfaces": []any{map[string]any{"url": "http://10.0.0.12:8080/a2a", "protocolBinding": "JSONRPC"}},
			}, nil
		}

		_, _, err := f.instance.RegisterInstance(
			"ws-1",
			"support-assistant",
			"node-1",
			&RegisterInstanceRequest{Address: "10.0.0.12", Port: 8080},
			"en",
		)
		require.NoError(t, err)
		assert.Equal(t, "10.0.0.12:8080/.well-known/agent-card.json", fetched)
		require.Len(t, f.agents.cards, 1)
		assert.Equal(
			t,
			`{"name":"Support","supportedInterfaces":[{"protocolBinding":"JSONRPC","url":"https://cognit.test/api/v1/workspaces/ws-1/a2a/agents/support-assistant"}],"url":"https://cognit.test/api/v1/workspaces/ws-1/a2a/agents/support-assistant"}`,
			f.agents.cards[0],
		)
	})

	t.Run("register adds a gateway interface to a fetched card that has none", func(t *testing.T) {
		viper.Set("app.url", "https://cognit.test")
		f := defaultInstanceFixture()
		f.instance.FetchCard = func(ctx context.Context, address string, port int, path string) (map[string]any, error) {
			return map[string]any{"name": "Support"}, nil
		}

		_, _, err := f.instance.RegisterInstance(
			"ws-1",
			"support-assistant",
			"node-1",
			&RegisterInstanceRequest{Address: "10.0.0.12", Port: 8080},
			"en",
		)
		require.NoError(t, err)
		require.Len(t, f.agents.cards, 1)
		assert.Equal(
			t,
			`{"name":"Support","supportedInterfaces":[{"url":"https://cognit.test/api/v1/workspaces/ws-1/a2a/agents/support-assistant"}]}`,
			f.agents.cards[0],
		)
	})

	t.Run("register still succeeds when the card cannot be fetched", func(t *testing.T) {
		f := defaultInstanceFixture()

		_, created, err := f.instance.RegisterInstance(
			"ws-1",
			"support-assistant",
			"node-1",
			&RegisterInstanceRequest{Address: "10.0.0.12", Port: 8080},
			"en",
		)
		require.NoError(t, err)
		assert.True(t, created)
		assert.Empty(t, f.agents.cards)
	})

	t.Run("register does not fetch when the agent already has a card", func(t *testing.T) {
		f := defaultInstanceFixture()
		f.agents.agent.Card = lo.ToPtr(`{"name":"Support"}`)
		f.instance.FetchCard = func(ctx context.Context, address string, port int, path string) (map[string]any, error) {
			t.Fatal("card should not be fetched")
			return nil, nil
		}

		_, _, err := f.instance.RegisterInstance(
			"ws-1",
			"support-assistant",
			"node-1",
			&RegisterInstanceRequest{Address: "10.0.0.12", Port: 8080},
			"en",
		)
		require.NoError(t, err)
		assert.Empty(t, f.agents.cards)
	})

	t.Run(
		"register updates an existing instance without touching checks it does not own",
		func(t *testing.T) {
			f := defaultInstanceFixture()
			f.instances.instance = &db.AgentInstance{
				Id:         "inst-1",
				AgentId:    "agent-1",
				InstanceId: "node-1",
			}

			_, created, err := f.instance.RegisterInstance(
				"ws-1",
				"support-assistant",
				"node-1",
				&RegisterInstanceRequest{
					Address: "10.0.0.13",
					Port:    9090,
				},
				"en",
			)
			require.NoError(t, err)
			assert.False(t, created)
			require.Len(t, f.instances.updated, 1)
			assert.Equal(t, "10.0.0.13", f.instances.updated[0].Address)
			assert.Empty(t, f.instances.created)
			assert.Empty(t, f.healthChecks.created)
		},
	)

	t.Run("register copies agent templates onto the new instance", func(t *testing.T) {
		f := defaultInstanceFixture()
		f.agentChecks.list = []*db.AgentCheck{
			{
				CheckId:    "http",
				Name:       "http",
				Type:       db.HealthCheckTypeHTTP,
				Definition: `{"path":"/healthz","interval":10,"timeout":2}`,
			},
			{
				CheckId:    "heartbeat",
				Name:       "heartbeat",
				Type:       db.HealthCheckTypeTTL,
				Definition: `{"ttl":30}`,
			},
		}

		_, _, err := f.instance.RegisterInstance(
			"ws-1",
			"support-assistant",
			"node-1",
			&RegisterInstanceRequest{
				Address: "10.0.0.12",
				Port:    8080,
			},
			"en",
		)
		require.NoError(t, err)

		require.Len(t, f.healthChecks.created, 2)
		http := f.healthChecks.created[0]
		assert.Equal(t, db.HealthCheckSourceAgent, http.Source)
		assert.Equal(t, db.HealthCheckStatusUnknown, http.Status)
		assert.NotNil(t, http.NextRunAt)

		heartbeat := f.healthChecks.created[1]
		assert.Equal(t, db.HealthCheckStatusUnknown, heartbeat.Status)
		assert.NotNil(t, heartbeat.TTLExpiresAt)
		assert.Nil(t, heartbeat.NextRunAt)
	})

	t.Run("register removes agent copies that no template wants anymore", func(t *testing.T) {
		f := defaultInstanceFixture()
		f.instances.instance = &db.AgentInstance{
			Id:         "inst-1",
			AgentId:    "agent-1",
			InstanceId: "node-1",
		}
		f.healthChecks.checks["inst-1"] = []*db.HealthCheck{
			{
				Id:      "retired",
				CheckId: "retired",
				Type:    db.HealthCheckTypeHTTP,
				Source:  db.HealthCheckSourceAgent,
			},
		}

		_, _, err := f.instance.RegisterInstance(
			"ws-1",
			"support-assistant",
			"node-1",
			&RegisterInstanceRequest{
				Address: "10.0.0.12",
				Port:    8080,
			},
			"en",
		)
		require.NoError(t, err)
		assert.Equal(t, []db.Id{"retired"}, f.healthChecks.deleted)
	})

	t.Run(
		"register rejects invalid ids and returns not found for a missing agent",
		func(t *testing.T) {
			req := &RegisterInstanceRequest{Address: "10.0.0.12", Port: 8080}

			_, _, err := defaultInstanceFixture().
				instance.
				RegisterInstance("ws-1", "support-assistant", "bad id!", req, "en")
			assert.ErrorIs(t, err, ErrInvalidInstanceId)

			_, _, err = newInstanceFixture(
				&db.Workspace{Id: "ws-1"},
				nil,
			).
				instance.RegisterInstance(
				"ws-1",
				"support-assistant",
				"node-1",
				req,
				"en",
			)
			assert.ErrorIs(t, err, ErrAgentNotFound)
		},
	)

	t.Run("deregister returns not found for a missing instance", func(t *testing.T) {
		err := defaultInstanceFixture().
			instance.
			DeregisterInstance("ws-1", "support-assistant", "node-1")
		assert.ErrorIs(t, err, ErrInstanceNotFound)
	})

	t.Run("deregister removes the instance", func(t *testing.T) {
		f := defaultInstanceFixture()
		f.instances.instance = &db.AgentInstance{
			Id:         "inst-1",
			AgentId:    "agent-1",
			InstanceId: "node-1",
		}

		err := f.instance.DeregisterInstance("ws-1", "support-assistant", "node-1")
		require.NoError(t, err)
		assert.Equal(t, []db.Id{"inst-1"}, f.instances.deleted)
	})
}

func TestUnitSyncChecks(t *testing.T) {
	t.Run(
		"creates missing copies with pull checks due now and ttl checks expiring later",
		func(t *testing.T) {
			f := defaultInstanceFixture()
			f.agentChecks.list = []*db.AgentCheck{
				{
					CheckId:    "http",
					Name:       "http",
					Type:       db.HealthCheckTypeHTTP,
					Definition: `{"path":"/healthz"}`,
				},
				{
					CheckId:    "heartbeat",
					Name:       "heartbeat",
					Type:       db.HealthCheckTypeTTL,
					Definition: `{"ttl":30}`,
				},
			}

			require.NoError(t, f.instance.SyncChecks("inst-1", "agent-1", "en"))
			require.Len(t, f.healthChecks.created, 2)

			httpCheck := f.healthChecks.created[0]
			assert.Equal(t, db.HealthCheckStatusUnknown, httpCheck.Status)
			assert.Equal(t, db.HealthCheckSourceAgent, httpCheck.Source)
			assert.NotNil(t, httpCheck.NextRunAt)
			assert.Nil(t, httpCheck.TTLExpiresAt)

			ttlCheck := f.healthChecks.created[1]
			assert.Equal(t, db.HealthCheckStatusUnknown, ttlCheck.Status)
			assert.Nil(t, ttlCheck.NextRunAt)
			require.NotNil(t, ttlCheck.TTLExpiresAt)
			assert.WithinDuration(t, time.Now().Add(30*time.Second), *ttlCheck.TTLExpiresAt, 5*time.Second)
		},
	)

	t.Run("updates an existing copy of the same type and keeps its status", func(t *testing.T) {
		f := defaultInstanceFixture()
		f.agentChecks.list = []*db.AgentCheck{
			{
				CheckId:    "http",
				Name:       "health",
				Type:       db.HealthCheckTypeHTTP,
				Definition: `{"path":"/v2"}`,
			},
		}
		existing := &db.HealthCheck{
			Id:         "check-1",
			InstanceId: "inst-1",
			CheckId:    "http",
			Name:       "http",
			Type:       db.HealthCheckTypeHTTP,
			Source:     db.HealthCheckSourceAgent,
			Status:     db.HealthCheckStatusPassing,
		}
		f.healthChecks.checks["inst-1"] = []*db.HealthCheck{existing}

		require.NoError(t, f.instance.SyncChecks("inst-1", "agent-1", "en"))
		assert.Empty(t, f.healthChecks.created)
		assert.Empty(t, f.healthChecks.deleted)
		require.Len(t, f.healthChecks.updated, 1)
		assert.Equal(t, "health", existing.Name)
		assert.Equal(t, `{"path":"/v2"}`, *existing.Definition)
		assert.Equal(t, db.HealthCheckStatusPassing, existing.Status)
		assert.NotNil(t, existing.NextRunAt)
	})

	t.Run("replaces a copy whose type changed", func(t *testing.T) {
		f := defaultInstanceFixture()
		f.agentChecks.list = []*db.AgentCheck{
			{
				CheckId:    "check",
				Name:       "check",
				Type:       db.HealthCheckTypeTCP,
				Definition: `{"port":5432}`,
			},
		}
		f.healthChecks.checks["inst-1"] = []*db.HealthCheck{
			{
				Id:      "check-1",
				CheckId: "check",
				Type:    db.HealthCheckTypeTTL,
				Source:  db.HealthCheckSourceAgent,
			},
		}

		require.NoError(t, f.instance.SyncChecks("inst-1", "agent-1", "en"))
		assert.Equal(t, []db.Id{"check-1"}, f.healthChecks.deleted)
		require.Len(t, f.healthChecks.created, 1)
		assert.Equal(t, db.HealthCheckTypeTCP, f.healthChecks.created[0].Type)
		assert.Equal(t, db.HealthCheckStatusUnknown, f.healthChecks.created[0].Status)
	})

	t.Run(
		"removes agent copies no template wants and keeps checks added by hand",
		func(t *testing.T) {
			f := defaultInstanceFixture()
			f.agentChecks.list = []*db.AgentCheck{
				{
					CheckId:    "http",
					Name:       "http",
					Type:       db.HealthCheckTypeHTTP,
					Definition: `{"path":"/healthz"}`,
				},
			}
			f.healthChecks.checks["inst-1"] = []*db.HealthCheck{
				{Id: "stale", CheckId: "old", Source: db.HealthCheckSourceAgent},
				{Id: "manual", CheckId: "manual", Source: "manual"},
			}

			require.NoError(t, f.instance.SyncChecks("inst-1", "agent-1", "en"))
			assert.Equal(t, []db.Id{"stale"}, f.healthChecks.deleted)
		},
	)
}
