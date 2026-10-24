// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package worker

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/clivern/sloth/db"
	"github.com/clivern/sloth/module"
	"github.com/clivern/sloth/pkg/broker"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// recordingChecks captures the Record call the handler makes.
type recordingChecks struct {
	db.HealthCheckRepository
	id      db.Id
	status  string
	nextRun time.Time
	called  bool
}

func (f *recordingChecks) Record(id db.Id, status, output string, nextRunAt time.Time) error {
	f.called = true
	f.id = id
	f.status = status
	f.nextRun = nextRunAt
	return nil
}

// checkInstances returns the instance a changed check belongs to.
type checkInstances struct {
	db.AgentInstanceRepository
}

func (f *checkInstances) GetById(id db.Id) (*db.AgentInstance, error) {
	return &db.AgentInstance{Id: id, AgentId: "agent-1", InstanceId: "support-1"}, nil
}

// checkAgents returns the agent that owns the instance.
type checkAgents struct {
	db.AgentRepository
}

func (f *checkAgents) GetById(id db.Id) (*db.Agent, error) {
	return &db.Agent{Id: id, WorkspaceId: "ws-1", Name: "support-assistant"}, nil
}

func healthCheckMsg(t *testing.T, job module.HealthCheckJob) *broker.Msg {
	t.Helper()
	raw, err := json.Marshal(job)
	require.NoError(t, err)
	return &broker.Msg{Subject: module.HealthCheckRunSubject, Data: raw}
}

func TestUnitHandleHealthCheck(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	host, portText, err := net.SplitHostPort(server.Listener.Addr().String())
	require.NoError(t, err)
	port, err := strconv.Atoi(portText)
	require.NoError(t, err)

	t.Run("http probe passing records passing with next run", func(t *testing.T) {
		checks := &recordingChecks{}
		h := &handlers{checks: checks}
		definition := json.RawMessage(`{"path":"/healthz","scheme":"http","interval":30}`)

		before := time.Now().UTC()
		err := h.HandleHealthCheck(context.Background(), healthCheckMsg(t, module.HealthCheckJob{
			CheckId:    "check-1",
			Type:       db.HealthCheckTypeHTTP,
			Status:     db.HealthCheckStatusPassing,
			Address:    host,
			Port:       port,
			Definition: definition,
		}))

		require.NoError(t, err)
		assert.Equal(t, db.Id("check-1"), checks.id)
		assert.Equal(t, db.HealthCheckStatusPassing, checks.status)
		assert.WithinDuration(t, before.Add(30*time.Second), checks.nextRun, 5*time.Second)
	})

	t.Run("http probe failing records critical", func(t *testing.T) {
		checks := &recordingChecks{}
		h := &handlers{checks: checks}

		err := h.HandleHealthCheck(context.Background(), healthCheckMsg(t, module.HealthCheckJob{
			CheckId:    "check-2",
			Type:       db.HealthCheckTypeHTTP,
			Status:     db.HealthCheckStatusCritical,
			Address:    host,
			Port:       port,
			Definition: json.RawMessage(`{"path":"/broken","scheme":"http","timeout":3,"interval":30}`),
		}))

		require.NoError(t, err)
		assert.Equal(t, db.HealthCheckStatusCritical, checks.status)
	})

	t.Run("tcp probe to closed port records critical", func(t *testing.T) {
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)
		closedPort := listener.Addr().(*net.TCPAddr).Port
		require.NoError(t, listener.Close())

		checks := &recordingChecks{}
		h := &handlers{checks: checks}

		err = h.HandleHealthCheck(context.Background(), healthCheckMsg(t, module.HealthCheckJob{
			CheckId: "check-3",
			Type:    db.HealthCheckTypeTCP,
			Status:  db.HealthCheckStatusCritical,
			Address: "127.0.0.1",
			Port:    closedPort,
		}))

		require.NoError(t, err)
		assert.Equal(t, db.HealthCheckStatusCritical, checks.status)
	})

	t.Run("status change triggers agent health changed", func(t *testing.T) {
		var event string
		var request module.EventRequest
		var workspaceId db.Id
		h := &handlers{
			checks:    &recordingChecks{},
			instances: &checkInstances{},
			agents:    &checkAgents{},
			trigger: func(wid db.Id, name string, req module.EventRequest) {
				workspaceId = wid
				event = name
				request = req
			},
		}

		err := h.HandleHealthCheck(context.Background(), healthCheckMsg(t, module.HealthCheckJob{
			CheckId:    "check-5",
			CheckKey:   "http-health",
			InstanceId: "instance-1",
			Type:       db.HealthCheckTypeHTTP,
			Status:     db.HealthCheckStatusPassing,
			Address:    host,
			Port:       port,
			Definition: json.RawMessage(`{"path":"/broken","scheme":"http","timeout":3,"interval":30}`),
		}))

		require.NoError(t, err)
		assert.Equal(t, db.Id("ws-1"), workspaceId)
		assert.Equal(t, module.EventAgentHealthChanged, event)
		assert.Equal(t, db.Id("check-5"), request.ResourceId)
		assert.Equal(t, map[string]string{
			"agentName":  "support-assistant",
			"instanceId": "support-1",
			"checkId":    "http-health",
			"from":       db.HealthCheckStatusPassing,
			"to":         db.HealthCheckStatusCritical,
		}, request.Data)
	})

	t.Run("first passing result triggers nothing", func(t *testing.T) {
		triggered := false
		h := &handlers{
			checks: &recordingChecks{},
			trigger: func(wid db.Id, name string, req module.EventRequest) {
				triggered = true
			},
		}

		err := h.HandleHealthCheck(context.Background(), healthCheckMsg(t, module.HealthCheckJob{
			CheckId:    "check-6",
			Type:       db.HealthCheckTypeHTTP,
			Status:     db.HealthCheckStatusUnknown,
			Address:    host,
			Port:       port,
			Definition: json.RawMessage(`{"path":"/healthz","scheme":"http","interval":30}`),
		}))

		require.NoError(t, err)
		assert.False(t, triggered)
	})

	t.Run("unsupported type returns error without recording", func(t *testing.T) {
		checks := &recordingChecks{}
		h := &handlers{checks: checks}

		err := h.HandleHealthCheck(context.Background(), healthCheckMsg(t, module.HealthCheckJob{
			CheckId: "check-4",
			Type:    db.HealthCheckTypeA2A,
		}))

		assert.Error(t, err)
		assert.False(t, checks.called)
	})

	t.Run("invalid payload returns ErrInvalidPayload", func(t *testing.T) {
		h := &handlers{checks: &recordingChecks{}}

		err := h.HandleHealthCheck(context.Background(), &broker.Msg{Data: []byte("not json")})
		assert.ErrorIs(t, err, ErrInvalidPayload)
	})
}
