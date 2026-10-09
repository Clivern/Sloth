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

func TestUnitInstanceActions(t *testing.T) {
	workspace := &db.Workspace{Id: "ws-1"}
	agent := &db.Agent{Id: "agent-1", WorkspaceId: "ws-1", Name: "user-directory"}
	instance := &db.AgentInstance{
		Id:         "instance-1",
		AgentId:    "agent-1",
		InstanceId: "node-1",
		Address:    "10.0.0.1",
		Port:       8080,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	definition := `{"ttl":30}`
	expiresAt := time.Now().Add(time.Hour)
	lease := &db.HealthCheck{
		Id:              "lease-1",
		AgentInstanceId: "instance-1",
		CheckId:         db.HealthCheckIDTTL,
		Type:            db.HealthCheckTypeTTL,
		Source:          db.HealthCheckSourceLease,
		Status:          db.HealthCheckStatusPassing,
		Definition:      &definition,
		TTLExpiresAt:    &expiresAt,
	}
	registerBody := `{"address":"10.0.0.1","port":8080}`
	base := "/api/v1/workspaces/ws-1/agents/user-directory/instances"

	t.Run("list returns instances", func(t *testing.T) {
		api := newInstanceAPI(workspace, &fakeAgentCatalog{agent: agent}, &fakeInstanceRecords{list: []*db.AgentInstance{instance}}, &fakeLeaseChecks{})
		w := httptest.NewRecorder()
		newInstanceRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodGet, base, nil))

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"instanceId":"node-1"`)
	})

	t.Run("list with passing=true returns only live instances", func(t *testing.T) {
		api := newInstanceAPI(workspace, &fakeAgentCatalog{agent: agent}, &fakeInstanceRecords{list: []*db.AgentInstance{instance}}, &fakeLeaseChecks{})
		w := httptest.NewRecorder()
		newInstanceRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodGet, base+"?passing=true", nil))

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("list returns 404 for unknown workspace", func(t *testing.T) {
		api := newInstanceAPI(nil, &fakeAgentCatalog{agent: agent}, &fakeInstanceRecords{}, &fakeLeaseChecks{})
		w := httptest.NewRecorder()
		newInstanceRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodGet, base, nil))

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("list returns 404 for unknown agent", func(t *testing.T) {
		api := newInstanceAPI(workspace, &fakeAgentCatalog{}, &fakeInstanceRecords{}, &fakeLeaseChecks{})
		w := httptest.NewRecorder()
		newInstanceRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodGet, base, nil))

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "agent_not_found")
	})

	t.Run("list returns 500 when the repository fails", func(t *testing.T) {
		api := newInstanceAPI(workspace, &fakeAgentCatalog{agent: agent}, &fakeInstanceRecords{listErr: errors.New("boom")}, &fakeLeaseChecks{})
		w := httptest.NewRecorder()
		newInstanceRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodGet, base, nil))

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("get returns one instance", func(t *testing.T) {
		api := newInstanceAPI(workspace, &fakeAgentCatalog{agent: agent}, &fakeInstanceRecords{instance: instance}, &fakeLeaseChecks{checks: []*db.HealthCheck{lease}})
		w := httptest.NewRecorder()
		newInstanceRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodGet, base+"/node-1", nil))

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"address":"10.0.0.1"`)
	})

	t.Run("get returns 404 for unknown instance", func(t *testing.T) {
		api := newInstanceAPI(workspace, &fakeAgentCatalog{agent: agent}, &fakeInstanceRecords{}, &fakeLeaseChecks{})
		w := httptest.NewRecorder()
		newInstanceRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodGet, base+"/node-1", nil))

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "instance_not_found")
	})

	t.Run("get returns 500 when the repository fails", func(t *testing.T) {
		api := newInstanceAPI(workspace, &fakeAgentCatalog{agent: agent}, &fakeInstanceRecords{getErr: errors.New("boom")}, &fakeLeaseChecks{})
		w := httptest.NewRecorder()
		newInstanceRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodGet, base+"/node-1", nil))

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("register returns 201 for a new instance", func(t *testing.T) {
		api := newInstanceAPI(workspace, &fakeAgentCatalog{agent: agent}, &fakeInstanceRecords{}, &fakeLeaseChecks{checks: []*db.HealthCheck{lease}})
		w := httptest.NewRecorder()
		newInstanceRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodPut, base+"/node-1", strings.NewReader(registerBody)))

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Body.String(), `"status":"passing"`)
	})

	t.Run("register returns 200 for an existing instance", func(t *testing.T) {
		api := newInstanceAPI(workspace, &fakeAgentCatalog{agent: agent}, &fakeInstanceRecords{instance: instance}, &fakeLeaseChecks{lease: lease})
		w := httptest.NewRecorder()
		newInstanceRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodPut, base+"/node-1", strings.NewReader(registerBody)))

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("register returns 400 for invalid body", func(t *testing.T) {
		api := newInstanceAPI(workspace, &fakeAgentCatalog{agent: agent}, &fakeInstanceRecords{}, &fakeLeaseChecks{})
		w := httptest.NewRecorder()
		newInstanceRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodPut, base+"/node-1", strings.NewReader(`{"address":"10.0.0.1"}`)))

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("register returns 400 for an invalid instance id", func(t *testing.T) {
		api := newInstanceAPI(workspace, &fakeAgentCatalog{agent: agent}, &fakeInstanceRecords{}, &fakeLeaseChecks{})
		w := httptest.NewRecorder()
		newInstanceRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodPut, base+"/-bad", strings.NewReader(registerBody)))

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid_instance_id")
	})

	t.Run("register returns 500 when the repository fails", func(t *testing.T) {
		api := newInstanceAPI(workspace, &fakeAgentCatalog{agent: agent}, &fakeInstanceRecords{createErr: errors.New("boom")}, &fakeLeaseChecks{})
		w := httptest.NewRecorder()
		newInstanceRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodPut, base+"/node-1", strings.NewReader(registerBody)))

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("renew returns 200", func(t *testing.T) {
		api := newInstanceAPI(workspace, &fakeAgentCatalog{agent: agent}, &fakeInstanceRecords{instance: instance}, &fakeLeaseChecks{lease: lease})
		w := httptest.NewRecorder()
		newInstanceRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodPut, base+"/node-1/renew", nil))

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("renew returns 404 when the instance is unknown", func(t *testing.T) {
		api := newInstanceAPI(workspace, &fakeAgentCatalog{agent: agent}, &fakeInstanceRecords{}, &fakeLeaseChecks{lease: lease})
		w := httptest.NewRecorder()
		newInstanceRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodPut, base+"/node-1/renew", nil))

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("renew returns 404 when the lease is missing", func(t *testing.T) {
		api := newInstanceAPI(workspace, &fakeAgentCatalog{agent: agent}, &fakeInstanceRecords{instance: instance}, &fakeLeaseChecks{})
		w := httptest.NewRecorder()
		newInstanceRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodPut, base+"/node-1/renew", nil))

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("renew returns 500 when the repository fails", func(t *testing.T) {
		api := newInstanceAPI(workspace, &fakeAgentCatalog{agent: agent}, &fakeInstanceRecords{instance: instance}, &fakeLeaseChecks{lease: lease, passErr: errors.New("boom")})
		w := httptest.NewRecorder()
		newInstanceRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodPut, base+"/node-1/renew", nil))

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("deregister returns 204", func(t *testing.T) {
		api := newInstanceAPI(workspace, &fakeAgentCatalog{agent: agent}, &fakeInstanceRecords{instance: instance}, &fakeLeaseChecks{})
		w := httptest.NewRecorder()
		newInstanceRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodDelete, base+"/node-1", nil))

		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("deregister returns 404 for unknown instance", func(t *testing.T) {
		api := newInstanceAPI(workspace, &fakeAgentCatalog{agent: agent}, &fakeInstanceRecords{}, &fakeLeaseChecks{})
		w := httptest.NewRecorder()
		newInstanceRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodDelete, base+"/node-1", nil))

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("handlers return 400 when path params are missing", func(t *testing.T) {
		api := newInstanceAPI(workspace, &fakeAgentCatalog{agent: agent}, &fakeInstanceRecords{}, &fakeLeaseChecks{})

		w := httptest.NewRecorder()
		api.GetInstanceAction(w, httptest.NewRequest(http.MethodGet, "/", nil))
		assert.Equal(t, http.StatusBadRequest, w.Code)

		w = httptest.NewRecorder()
		api.DeregisterInstanceAction(w, httptest.NewRequest(http.MethodDelete, "/", nil))
		assert.Equal(t, http.StatusBadRequest, w.Code)

		w = httptest.NewRecorder()
		api.RenewInstanceAction(w, httptest.NewRequest(http.MethodPut, "/", nil))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
