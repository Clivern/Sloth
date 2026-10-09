// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/clivern/cognit/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitInstance(t *testing.T) {
	t.Run("list rejects invalid agent names and missing workspaces or agents", func(t *testing.T) {
		_, err := defaultInstanceFixture().instance.ListInstances("ws-1", "-bad", false)
		assert.ErrorIs(t, err, ErrInvalidAgentName)

		_, err = newInstanceFixture(nil, nil).instance.ListInstances("ws-1", "support-assistant", false)
		assert.ErrorIs(t, err, ErrWorkspaceNotFound)

		_, err = newInstanceFixture(&db.Workspace{Id: "ws-1"}, nil).instance.ListInstances("ws-1", "support-assistant", false)
		assert.ErrorIs(t, err, ErrAgentNotFound)
	})

	t.Run("list returns only live instances when passing is set", func(t *testing.T) {
		f := defaultInstanceFixture()
		f.instances.list = []*db.AgentInstance{{Id: "inst-1", InstanceId: "node-1"}, {Id: "inst-2", InstanceId: "node-2"}}
		f.instances.live = []*db.AgentInstance{{Id: "inst-1", InstanceId: "node-1"}}
		f.healthChecks.checks["inst-1"] = []*db.HealthCheck{leaseCheck()}

		all, err := f.instance.ListInstances("ws-1", "support-assistant", false)
		require.NoError(t, err)
		assert.Len(t, all, 2)

		live, err := f.instance.ListInstances("ws-1", "support-assistant", true)
		require.NoError(t, err)
		require.Len(t, live, 1)
		assert.Equal(t, "node-1", live[0].InstanceId)
	})

	t.Run("list marks an instance without a lease as critical", func(t *testing.T) {
		f := defaultInstanceFixture()
		f.instances.list = []*db.AgentInstance{{Id: "inst-1", InstanceId: "node-1"}}

		list, err := f.instance.ListInstances("ws-1", "support-assistant", false)
		require.NoError(t, err)
		assert.Equal(t, db.AgentInstanceStatusCritical, list[0].Status)
		assert.Empty(t, list[0].Checks)
	})

	t.Run("list makes an instance warning when one check warns", func(t *testing.T) {
		f := defaultInstanceFixture()
		f.instances.list = []*db.AgentInstance{{Id: "inst-1", InstanceId: "node-1"}}
		f.healthChecks.checks["inst-1"] = []*db.HealthCheck{
			leaseCheck(),
			{CheckId: "http", Name: "http", Type: db.HealthCheckTypeHTTP, Source: db.HealthCheckSourceAgent, Status: db.HealthCheckStatusWarning},
		}

		list, err := f.instance.ListInstances("ws-1", "support-assistant", false)
		require.NoError(t, err)
		assert.Equal(t, db.AgentInstanceStatusWarning, list[0].Status)
		require.Len(t, list[0].Checks, 2)
		assert.NotNil(t, list[0].LeaseExpiresAt)
	})

	t.Run("get rejects invalid ids and names", func(t *testing.T) {
		_, err := defaultInstanceFixture().instance.GetInstance("ws-1", "support-assistant", "bad id!")
		assert.ErrorIs(t, err, ErrInvalidInstanceId)

		_, err = defaultInstanceFixture().instance.GetInstance("ws-1", "-bad", "node-1")
		assert.ErrorIs(t, err, ErrInvalidAgentName)
	})

	t.Run("get returns not found for a missing instance", func(t *testing.T) {
		_, err := defaultInstanceFixture().instance.GetInstance("ws-1", "support-assistant", "node-1")
		assert.ErrorIs(t, err, ErrInstanceNotFound)
	})

	t.Run("get returns the instance with its lease expiry and checks", func(t *testing.T) {
		f := defaultInstanceFixture()
		f.instances.instance = &db.AgentInstance{Id: "inst-1", AgentId: "agent-1", InstanceId: "node-1", Address: "10.0.0.12", Port: 8080}
		lease := leaseCheck()
		f.healthChecks.checks["inst-1"] = []*db.HealthCheck{lease}

		item, err := f.instance.GetInstance("ws-1", "support-assistant", "node-1")
		require.NoError(t, err)
		assert.Equal(t, db.AgentInstanceStatusPassing, item.Status)
		assert.Equal(t, lease.TTLExpiresAt, item.LeaseExpiresAt)
		assert.Equal(t, "10.0.0.12", item.Address)
		assert.Equal(t, 8080, item.Port)
	})

	t.Run("register creates the instance and its lease", func(t *testing.T) {
		f := defaultInstanceFixture()

		item, created, err := f.instance.RegisterInstance("ws-1", "support-assistant", "node-1", &RegisterInstanceRequest{
			Address: "10.0.0.12",
			Port:    8080,
			Meta:    json.RawMessage(`{"region":"eu"}`),
		})
		require.NoError(t, err)
		assert.True(t, created)
		assert.Equal(t, "10.0.0.12", item.Address)

		require.Len(t, f.instances.created, 1)
		require.NotNil(t, f.instances.created[0].Meta)
		assert.Equal(t, `{"region":"eu"}`, *f.instances.created[0].Meta)

		require.Len(t, f.healthChecks.created, 1)
		lease := f.healthChecks.created[0]
		assert.Equal(t, db.HealthCheckSourceLease, lease.Source)
		assert.Equal(t, db.HealthCheckIDTTL, lease.CheckId)
		require.NotNil(t, lease.Definition)
		assert.Contains(t, *lease.Definition, `"ttl":30`)
		assert.WithinDuration(t, time.Now().Add(DefaultLeaseTTL*time.Second), *lease.TTLExpiresAt, 5*time.Second)
	})

	t.Run("register uses the requested lease ttl", func(t *testing.T) {
		f := defaultInstanceFixture()

		_, _, err := f.instance.RegisterInstance("ws-1", "support-assistant", "node-1", &RegisterInstanceRequest{
			Address:  "10.0.0.12",
			Port:     8080,
			LeaseTTL: 60,
		})
		require.NoError(t, err)
		require.Len(t, f.healthChecks.created, 1)
		assert.Contains(t, *f.healthChecks.created[0].Definition, `"ttl":60`)
	})

	t.Run("register updates an existing instance and renews its lease", func(t *testing.T) {
		f := defaultInstanceFixture()
		f.instances.instance = &db.AgentInstance{Id: "inst-1", AgentId: "agent-1", InstanceId: "node-1"}
		definition := `{"ttl":30}`
		f.healthChecks.byInstance["inst-1"] = &db.HealthCheck{Id: "lease-1", CheckId: db.HealthCheckIDTTL, Type: db.HealthCheckTypeTTL, Source: db.HealthCheckSourceLease, Definition: &definition}

		_, created, err := f.instance.RegisterInstance("ws-1", "support-assistant", "node-1", &RegisterInstanceRequest{Address: "10.0.0.13", Port: 9090})
		require.NoError(t, err)
		assert.False(t, created)
		require.Len(t, f.instances.updated, 1)
		assert.Equal(t, "10.0.0.13", f.instances.updated[0].Address)
		assert.Empty(t, f.instances.created)

		require.Len(t, f.healthChecks.updated, 1)
		require.Len(t, f.healthChecks.passed, 1)
		assert.Equal(t, db.Id("lease-1"), f.healthChecks.passed[0].id)
	})

	t.Run("register copies agent templates onto the new instance", func(t *testing.T) {
		f := defaultInstanceFixture()
		f.agentChecks.list = []*db.AgentCheck{
			{CheckId: "http", Name: "http", Type: db.HealthCheckTypeHTTP, Definition: `{"path":"/healthz","interval":10,"timeout":2}`},
			{CheckId: "heartbeat", Name: "heartbeat", Type: db.HealthCheckTypeTTL, Definition: `{"ttl":30}`},
		}

		_, _, err := f.instance.RegisterInstance("ws-1", "support-assistant", "node-1", &RegisterInstanceRequest{Address: "10.0.0.12", Port: 8080})
		require.NoError(t, err)

		// One lease plus one copy per template.
		require.Len(t, f.healthChecks.created, 3)
		http := f.healthChecks.created[1]
		assert.Equal(t, db.HealthCheckSourceAgent, http.Source)
		assert.Equal(t, db.HealthCheckStatusCritical, http.Status)
		assert.NotNil(t, http.NextRunAt)

		heartbeat := f.healthChecks.created[2]
		assert.Equal(t, db.HealthCheckStatusPassing, heartbeat.Status)
		assert.NotNil(t, heartbeat.TTLExpiresAt)
		assert.Nil(t, heartbeat.NextRunAt)
	})

	t.Run("register removes agent copies that no template wants anymore", func(t *testing.T) {
		f := defaultInstanceFixture()
		f.instances.instance = &db.AgentInstance{Id: "inst-1", AgentId: "agent-1", InstanceId: "node-1"}
		f.healthChecks.checks["inst-1"] = []*db.HealthCheck{
			{Id: "retired", CheckId: "retired", Type: db.HealthCheckTypeHTTP, Source: db.HealthCheckSourceAgent},
			{Id: "lease-1", CheckId: db.HealthCheckIDTTL, Type: db.HealthCheckTypeTTL, Source: db.HealthCheckSourceLease},
		}

		_, _, err := f.instance.RegisterInstance("ws-1", "support-assistant", "node-1", &RegisterInstanceRequest{Address: "10.0.0.12", Port: 8080})
		require.NoError(t, err)
		assert.Equal(t, []db.Id{"retired"}, f.healthChecks.deleted)
	})

	t.Run("register rejects invalid ids and names and missing workspaces or agents", func(t *testing.T) {
		req := &RegisterInstanceRequest{Address: "10.0.0.12", Port: 8080}

		_, _, err := defaultInstanceFixture().instance.RegisterInstance("ws-1", "support-assistant", "bad id!", req)
		assert.ErrorIs(t, err, ErrInvalidInstanceId)

		_, _, err = newInstanceFixture(nil, nil).instance.RegisterInstance("ws-1", "support-assistant", "node-1", req)
		assert.ErrorIs(t, err, ErrWorkspaceNotFound)

		_, _, err = newInstanceFixture(&db.Workspace{Id: "ws-1"}, nil).instance.RegisterInstance("ws-1", "support-assistant", "node-1", req)
		assert.ErrorIs(t, err, ErrAgentNotFound)
	})

	t.Run("renew extends the lease by its registered ttl", func(t *testing.T) {
		f := defaultInstanceFixture()
		f.instances.instance = &db.AgentInstance{Id: "inst-1", AgentId: "agent-1", InstanceId: "node-1"}
		definition := `{"ttl":45}`
		f.healthChecks.byInstance["inst-1"] = &db.HealthCheck{Id: "lease-1", CheckId: db.HealthCheckIDTTL, Type: db.HealthCheckTypeTTL, Source: db.HealthCheckSourceLease, Definition: &definition}

		_, err := f.instance.RenewInstance("ws-1", "support-assistant", "node-1")
		require.NoError(t, err)
		require.Len(t, f.healthChecks.passed, 1)
		assert.Equal(t, db.Id("lease-1"), f.healthChecks.passed[0].id)
		assert.WithinDuration(t, time.Now().Add(45*time.Second), *f.healthChecks.passed[0].ttlExpiresAt, 5*time.Second)
	})

	t.Run("renew returns not found when the instance or its lease is missing", func(t *testing.T) {
		_, err := defaultInstanceFixture().instance.RenewInstance("ws-1", "support-assistant", "node-1")
		assert.ErrorIs(t, err, ErrInstanceNotFound)

		f := defaultInstanceFixture()
		f.instances.instance = &db.AgentInstance{Id: "inst-1", AgentId: "agent-1", InstanceId: "node-1"}
		_, err = f.instance.RenewInstance("ws-1", "support-assistant", "node-1")
		assert.ErrorIs(t, err, ErrInstanceNotFound)
	})

	t.Run("deregister returns not found for a missing instance", func(t *testing.T) {
		err := defaultInstanceFixture().instance.DeregisterInstance("ws-1", "support-assistant", "node-1")
		assert.ErrorIs(t, err, ErrInstanceNotFound)
	})

	t.Run("deregister removes the instance", func(t *testing.T) {
		f := defaultInstanceFixture()
		f.instances.instance = &db.AgentInstance{Id: "inst-1", AgentId: "agent-1", InstanceId: "node-1"}

		err := f.instance.DeregisterInstance("ws-1", "support-assistant", "node-1")
		require.NoError(t, err)
		assert.Equal(t, []db.Id{"inst-1"}, f.instances.deleted)
	})
}
