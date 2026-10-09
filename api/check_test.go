// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/clivern/cognit/db"

	"github.com/stretchr/testify/assert"
)

func TestUnitAgentCheckActions(t *testing.T) {
	workspace := &db.Workspace{Id: "ws-1"}
	agent := &db.Agent{Id: "agent-1", WorkspaceId: "ws-1", Name: "user-directory"}
	template := &db.AgentCheck{
		Id:         "check-1",
		AgentId:    "agent-1",
		CheckId:    "heartbeat",
		Name:       "heartbeat",
		Type:       db.HealthCheckTypeTTL,
		Definition: `{"ttl":30}`,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	instance := &db.AgentInstance{Id: "instance-1", AgentId: "agent-1", InstanceId: "node-1"}
	ttlCheck := &db.HealthCheck{
		Id:         "health-1",
		InstanceId: "instance-1",
		CheckId:    "heartbeat",
		Name:       "heartbeat",
		Type:       db.HealthCheckTypeTTL,
		Source:     db.HealthCheckSourceAgent,
		Status:     db.HealthCheckStatusPassing,
		Definition: new(`{"ttl":30}`),
	}
	ttlBody := `{"type":"ttl","ttl":30}`

	t.Run("list returns templates with total", func(t *testing.T) {
		api := newCheckAPI(workspace, &fakeAgentCatalog{agent: agent}, &fakeAgentChecks{list: []*db.AgentCheck{template}, total: 1}, &fakeCheckInstances{}, &fakeHealthChecks{})
		w := httptest.NewRecorder()
		newCheckRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1/agents/user-directory/checks", nil))

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"checkId":"heartbeat"`)
		assert.Contains(t, w.Body.String(), `"total":1`)
	})

	t.Run("list returns 404 for unknown workspace", func(t *testing.T) {
		api := newCheckAPI(nil, &fakeAgentCatalog{agent: agent}, &fakeAgentChecks{}, &fakeCheckInstances{}, &fakeHealthChecks{})
		w := httptest.NewRecorder()
		newCheckRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1/agents/user-directory/checks", nil))

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("list returns 404 for unknown agent", func(t *testing.T) {
		api := newCheckAPI(workspace, &fakeAgentCatalog{}, &fakeAgentChecks{}, &fakeCheckInstances{}, &fakeHealthChecks{})
		w := httptest.NewRecorder()
		newCheckRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1/agents/user-directory/checks", nil))

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("list returns 500 when the repository fails", func(t *testing.T) {
		api := newCheckAPI(workspace, &fakeAgentCatalog{agent: agent}, &fakeAgentChecks{listErr: errors.New("boom")}, &fakeCheckInstances{}, &fakeHealthChecks{})
		w := httptest.NewRecorder()
		newCheckRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1/agents/user-directory/checks", nil))

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("upsert returns 201 for a new ttl check", func(t *testing.T) {
		api := newCheckAPI(workspace, &fakeAgentCatalog{agent: agent}, &fakeAgentChecks{}, &fakeCheckInstances{}, &fakeHealthChecks{})
		w := httptest.NewRecorder()
		newCheckRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/api/v1/workspaces/ws-1/agents/user-directory/checks/heartbeat", strings.NewReader(ttlBody)))

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Body.String(), `"ttl":30`)
	})

	t.Run("upsert returns 200 when the template already exists", func(t *testing.T) {
		api := newCheckAPI(workspace, &fakeAgentCatalog{agent: agent}, &fakeAgentChecks{template: template}, &fakeCheckInstances{}, &fakeHealthChecks{})
		w := httptest.NewRecorder()
		newCheckRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/api/v1/workspaces/ws-1/agents/user-directory/checks/heartbeat", strings.NewReader(ttlBody)))

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("upsert accepts a valid http check", func(t *testing.T) {
		api := newCheckAPI(workspace, &fakeAgentCatalog{agent: agent}, &fakeAgentChecks{}, &fakeCheckInstances{}, &fakeHealthChecks{})
		w := httptest.NewRecorder()
		body := `{"type":"http","path":"/health","port":8080,"interval":10,"timeout":2}`
		newCheckRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/api/v1/workspaces/ws-1/agents/user-directory/checks/probe", strings.NewReader(body)))

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Body.String(), `"scheme":"http"`)
	})

	t.Run("upsert returns 400 for a ttl check without ttl", func(t *testing.T) {
		api := newCheckAPI(workspace, &fakeAgentCatalog{agent: agent}, &fakeAgentChecks{}, &fakeCheckInstances{}, &fakeHealthChecks{})
		w := httptest.NewRecorder()
		newCheckRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/api/v1/workspaces/ws-1/agents/user-directory/checks/heartbeat", strings.NewReader(`{"type":"ttl"}`)))

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid_check_timing")
	})

	t.Run("upsert returns 400 for the reserved ttl check id", func(t *testing.T) {
		api := newCheckAPI(workspace, &fakeAgentCatalog{agent: agent}, &fakeAgentChecks{}, &fakeCheckInstances{}, &fakeHealthChecks{})
		w := httptest.NewRecorder()
		newCheckRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/api/v1/workspaces/ws-1/agents/user-directory/checks/ttl", strings.NewReader(ttlBody)))

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "reserved_check_id")
	})

	t.Run("upsert returns 400 for an http path without a leading slash", func(t *testing.T) {
		api := newCheckAPI(workspace, &fakeAgentCatalog{agent: agent}, &fakeAgentChecks{}, &fakeCheckInstances{}, &fakeHealthChecks{})
		w := httptest.NewRecorder()
		body := `{"type":"http","path":"health"}`
		newCheckRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/api/v1/workspaces/ws-1/agents/user-directory/checks/probe", strings.NewReader(body)))

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid_check_path")
	})

	t.Run("upsert returns 400 when timeout is not below interval", func(t *testing.T) {
		api := newCheckAPI(workspace, &fakeAgentCatalog{agent: agent}, &fakeAgentChecks{}, &fakeCheckInstances{}, &fakeHealthChecks{})
		w := httptest.NewRecorder()
		body := `{"type":"tcp","port":8080,"interval":5,"timeout":5}`
		newCheckRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/api/v1/workspaces/ws-1/agents/user-directory/checks/probe", strings.NewReader(body)))

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid_check_timing")
	})

	t.Run("upsert returns 400 for an unsupported type", func(t *testing.T) {
		api := newCheckAPI(workspace, &fakeAgentCatalog{agent: agent}, &fakeAgentChecks{}, &fakeCheckInstances{}, &fakeHealthChecks{})
		w := httptest.NewRecorder()
		newCheckRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/api/v1/workspaces/ws-1/agents/user-directory/checks/probe", strings.NewReader(`{"type":"grpc"}`)))

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("upsert returns 404 for unknown agent", func(t *testing.T) {
		api := newCheckAPI(workspace, &fakeAgentCatalog{}, &fakeAgentChecks{}, &fakeCheckInstances{}, &fakeHealthChecks{})
		w := httptest.NewRecorder()
		newCheckRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/api/v1/workspaces/ws-1/agents/user-directory/checks/heartbeat", strings.NewReader(ttlBody)))

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "agent_not_found")
	})

	t.Run("upsert returns 500 when the repository fails", func(t *testing.T) {
		api := newCheckAPI(workspace, &fakeAgentCatalog{agent: agent}, &fakeAgentChecks{getErr: errors.New("boom")}, &fakeCheckInstances{}, &fakeHealthChecks{})
		w := httptest.NewRecorder()
		newCheckRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/api/v1/workspaces/ws-1/agents/user-directory/checks/heartbeat", strings.NewReader(ttlBody)))

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("delete returns 204", func(t *testing.T) {
		api := newCheckAPI(workspace, &fakeAgentCatalog{agent: agent}, &fakeAgentChecks{template: template}, &fakeCheckInstances{}, &fakeHealthChecks{})
		w := httptest.NewRecorder()
		newCheckRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/api/v1/workspaces/ws-1/agents/user-directory/checks/heartbeat", nil))

		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("delete returns 404 for unknown check", func(t *testing.T) {
		api := newCheckAPI(workspace, &fakeAgentCatalog{agent: agent}, &fakeAgentChecks{}, &fakeCheckInstances{}, &fakeHealthChecks{})
		w := httptest.NewRecorder()
		newCheckRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/api/v1/workspaces/ws-1/agents/user-directory/checks/heartbeat", nil))

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "check_not_found")
	})

	t.Run("delete returns 500 when the repository fails", func(t *testing.T) {
		api := newCheckAPI(workspace, &fakeAgentCatalog{agent: agent}, &fakeAgentChecks{template: template, deleteErr: errors.New("boom")}, &fakeCheckInstances{}, &fakeHealthChecks{})
		w := httptest.NewRecorder()
		newCheckRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/api/v1/workspaces/ws-1/agents/user-directory/checks/heartbeat", nil))

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	reportCases := []struct {
		name   string
		action string
		status string
	}{
		{name: "pass", action: "pass", status: db.HealthCheckStatusPassing},
		{name: "warn", action: "warn", status: db.HealthCheckStatusWarning},
		{name: "fail", action: "fail", status: db.HealthCheckStatusCritical},
	}
	for _, tc := range reportCases {
		t.Run(tc.name+" reports a ttl check", func(t *testing.T) {
			api := newCheckAPI(workspace, &fakeAgentCatalog{agent: agent}, &fakeAgentChecks{}, &fakeCheckInstances{instance: instance}, &fakeHealthChecks{check: ttlCheck})
			w := httptest.NewRecorder()
			path := "/api/v1/workspaces/ws-1/agents/user-directory/instances/node-1/checks/heartbeat/" + tc.action
			newCheckRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodPost, path, strings.NewReader(`{"output":"ok"}`)))

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Contains(t, w.Body.String(), `"status":"`+tc.status+`"`)
		})
	}

	t.Run("report returns 404 for unknown instance", func(t *testing.T) {
		api := newCheckAPI(workspace, &fakeAgentCatalog{agent: agent}, &fakeAgentChecks{}, &fakeCheckInstances{}, &fakeHealthChecks{check: ttlCheck})
		w := httptest.NewRecorder()
		newCheckRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/ws-1/agents/user-directory/instances/node-1/checks/heartbeat/pass", nil))

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "instance_not_found")
	})

	t.Run("report returns 404 for unknown check", func(t *testing.T) {
		api := newCheckAPI(workspace, &fakeAgentCatalog{agent: agent}, &fakeAgentChecks{}, &fakeCheckInstances{instance: instance}, &fakeHealthChecks{})
		w := httptest.NewRecorder()
		newCheckRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/ws-1/agents/user-directory/instances/node-1/checks/heartbeat/pass", nil))

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "check_not_found")
	})

	t.Run("report returns 400 for a check that is not a ttl check", func(t *testing.T) {
		httpCheck := *ttlCheck
		httpCheck.Type = db.HealthCheckTypeHTTP
		api := newCheckAPI(workspace, &fakeAgentCatalog{agent: agent}, &fakeAgentChecks{}, &fakeCheckInstances{instance: instance}, &fakeHealthChecks{check: &httpCheck})
		w := httptest.NewRecorder()
		newCheckRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/ws-1/agents/user-directory/instances/node-1/checks/heartbeat/pass", nil))

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "check_not_reportable")
	})

	t.Run("report returns 500 when the repository fails", func(t *testing.T) {
		api := newCheckAPI(workspace, &fakeAgentCatalog{agent: agent}, &fakeAgentChecks{}, &fakeCheckInstances{instance: instance}, &fakeHealthChecks{check: ttlCheck, reportErr: errors.New("boom")})
		w := httptest.NewRecorder()
		newCheckRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/ws-1/agents/user-directory/instances/node-1/checks/heartbeat/pass", nil))

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("handlers return 400 when path params are missing", func(t *testing.T) {
		api := newCheckAPI(workspace, &fakeAgentCatalog{agent: agent}, &fakeAgentChecks{}, &fakeCheckInstances{}, &fakeHealthChecks{})

		w := httptest.NewRecorder()
		api.ListAgentChecksAction(w, httptest.NewRequest(http.MethodGet, "/", nil))
		assert.Equal(t, http.StatusBadRequest, w.Code)

		w = httptest.NewRecorder()
		api.UpsertAgentCheckAction(w, httptest.NewRequest(http.MethodPut, "/", strings.NewReader(ttlBody)))
		assert.Equal(t, http.StatusBadRequest, w.Code)

		w = httptest.NewRecorder()
		api.DeleteAgentCheckAction(w, httptest.NewRequest(http.MethodDelete, "/", nil))
		assert.Equal(t, http.StatusBadRequest, w.Code)

		w = httptest.NewRecorder()
		api.PassCheckAction(w, httptest.NewRequest(http.MethodPost, "/", nil))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
