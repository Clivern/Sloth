// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/clivern/cognit/db"
	"github.com/clivern/cognit/module"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestUnitWorkspaceAuditActions(t *testing.T) {
	workspace := &db.Workspace{Id: "ws-1"}
	action := "agent.upsert"
	stored := &db.AuditEvent{
		Id:          "event-1",
		WorkspaceId: "ws-1",
		Action:      action,
		CreatedAt:   time.Now(),
	}

	newAPI := func(workspace *db.Workspace, events *fakeAuditEvents) *API {
		return &API{
			Audit: module.NewAudit(events, &fakeWorkspaceRecords{workspace: workspace}),
		}
	}
	newRouter := func(api *API) *chi.Mux {
		router := chi.NewRouter()
		router.Get("/api/v1/workspaces/{workspaceId}/audits", api.ListWorkspaceAuditsAction)
		router.Get("/api/v1/workspaces/{workspaceId}/audits/{auditId}", api.GetWorkspaceAuditAction)
		return router
	}

	t.Run("list returns events with pagination meta", func(t *testing.T) {
		api := newAPI(workspace, &fakeAuditEvents{events: []*db.AuditEvent{stored}, total: 1})
		w := httptest.NewRecorder()
		newRouter(api).
			ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1/audits?limit=5&offset=0", nil))

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"action":"agent.upsert"`)
		assert.Contains(t, w.Body.String(), `"total":1`)
	})

	t.Run("list returns 404 for unknown workspace", func(t *testing.T) {
		api := newAPI(nil, &fakeAuditEvents{})
		w := httptest.NewRecorder()
		newRouter(api).
			ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1/audits", nil))

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "workspace_not_found")
	})

	t.Run("list returns 500 when the repository fails", func(t *testing.T) {
		api := newAPI(workspace, &fakeAuditEvents{listErr: errors.New("boom")})
		w := httptest.NewRecorder()
		newRouter(api).
			ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1/audits", nil))

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "failed_list_audits")
	})

	t.Run("list returns 400 when the workspace id is missing", func(t *testing.T) {
		api := newAPI(workspace, &fakeAuditEvents{})
		w := httptest.NewRecorder()
		api.ListWorkspaceAuditsAction(w, httptest.NewRequest(http.MethodGet, "/", nil))

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("get returns one event", func(t *testing.T) {
		api := newAPI(workspace, &fakeAuditEvents{event: stored})
		w := httptest.NewRecorder()
		newRouter(api).
			ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1/audits/event-1", nil))

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"id":"event-1"`)
	})

	t.Run("get returns 404 for an event in another workspace", func(t *testing.T) {
		other := *stored
		other.WorkspaceId = "ws-2"
		api := newAPI(workspace, &fakeAuditEvents{event: &other})
		w := httptest.NewRecorder()
		newRouter(api).
			ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1/audits/event-1", nil))

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "audit_event_not_found")
	})

	t.Run("get returns 404 for unknown workspace", func(t *testing.T) {
		api := newAPI(nil, &fakeAuditEvents{event: stored})
		w := httptest.NewRecorder()
		newRouter(api).
			ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1/audits/event-1", nil))

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("get returns 500 when the repository fails", func(t *testing.T) {
		api := newAPI(workspace, &fakeAuditEvents{getErr: errors.New("boom")})
		w := httptest.NewRecorder()
		newRouter(api).
			ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1/audits/event-1", nil))

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "failed_get_audit")
	})

	t.Run("get returns 400 when ids are missing", func(t *testing.T) {
		api := newAPI(workspace, &fakeAuditEvents{})
		w := httptest.NewRecorder()
		api.GetWorkspaceAuditAction(w, httptest.NewRequest(http.MethodGet, "/", nil))

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
