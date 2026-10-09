// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/clivern/cognit/db"
	"github.com/clivern/cognit/module"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestUnitWorkspaceStatsAction(t *testing.T) {
	newAPI := func(workspace *db.Workspace, stats *fakeWorkspaceStats) *API {
		return &API{
			Stats: module.NewStats(&fakeWorkspaceRecords{workspace: workspace}, stats),
		}
	}
	newRouter := func(api *API) *chi.Mux {
		router := chi.NewRouter()
		router.Get("/api/v1/workspaces/{workspaceId}/stats", api.GetWorkspaceStatsAction)
		return router
	}

	t.Run("returns 200 for a known workspace", func(t *testing.T) {
		api := newAPI(&db.Workspace{Id: "ws-1"}, &fakeWorkspaceStats{})
		w := httptest.NewRecorder()
		newRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1/stats", nil))

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("returns 404 for unknown workspace", func(t *testing.T) {
		api := newAPI(nil, &fakeWorkspaceStats{})
		w := httptest.NewRecorder()
		newRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1/stats", nil))

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "workspace_not_found")
	})

	t.Run("returns 500 when the repository fails", func(t *testing.T) {
		api := newAPI(&db.Workspace{Id: "ws-1"}, &fakeWorkspaceStats{err: errors.New("boom")})
		w := httptest.NewRecorder()
		newRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1/stats", nil))

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "failed_get_workspace_stats")
	})

	t.Run("returns 400 when the workspace id is missing", func(t *testing.T) {
		api := newAPI(&db.Workspace{Id: "ws-1"}, &fakeWorkspaceStats{})
		w := httptest.NewRecorder()
		api.GetWorkspaceStatsAction(w, httptest.NewRequest(http.MethodGet, "/", nil))

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
