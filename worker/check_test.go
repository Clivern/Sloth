// Copyright 2026 Cognit. All rights reserved.
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

	"github.com/clivern/cognit/db"
	"github.com/clivern/cognit/module"
	"github.com/clivern/cognit/pkg/broker"

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
			Address: "127.0.0.1",
			Port:    closedPort,
		}))

		require.NoError(t, err)
		assert.Equal(t, db.HealthCheckStatusCritical, checks.status)
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
