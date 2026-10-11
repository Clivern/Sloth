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

func TestUnitWorkspaceActions(t *testing.T) {
	user := &db.User{Id: "user-1"}
	stored := &db.Workspace{
		Id:        "ws-1",
		Name:      "Platform",
		Handle:    "platform",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	owner := &db.WorkspaceUser{WorkspaceId: "ws-1", UserId: "user-1", Role: db.UserRoleOwner}
	body := `{"name":"Platform"}`

	t.Run("create returns 201 with the owner role", func(t *testing.T) {
		api := newWorkspaceAPI(&fakeWorkspaceRecords{})
		w := httptest.NewRecorder()
		req := requestWithUser(
			httptest.NewRequest(http.MethodPost, "/api/v1/workspaces", strings.NewReader(body)),
			user,
		)
		newWorkspaceRouter(api).ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Body.String(), `"role":"`+db.UserRoleOwner+`"`)
	})

	t.Run("create returns 400 for a name that is too short", func(t *testing.T) {
		api := newWorkspaceAPI(&fakeWorkspaceRecords{})
		w := httptest.NewRecorder()
		req := requestWithUser(
			httptest.NewRequest(http.MethodPost, "/api/v1/workspaces", strings.NewReader(`{"name":"ab"}`)),
			user,
		)
		newWorkspaceRouter(api).ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("create returns 401 without a user", func(t *testing.T) {
		api := newWorkspaceAPI(&fakeWorkspaceRecords{})
		w := httptest.NewRecorder()
		newWorkspaceRouter(api).
			ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/v1/workspaces", strings.NewReader(body)))

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("create returns 500 when the repository fails", func(t *testing.T) {
		api := newWorkspaceAPI(&fakeWorkspaceRecords{createErr: errors.New("boom")})
		w := httptest.NewRecorder()
		req := requestWithUser(
			httptest.NewRequest(http.MethodPost, "/api/v1/workspaces", strings.NewReader(body)),
			user,
		)
		newWorkspaceRouter(api).ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("list returns the user's workspaces with pagination meta", func(t *testing.T) {
		api := newWorkspaceAPI(&fakeWorkspaceRecords{list: []*db.Workspace{stored}, total: 1, membership: owner})
		w := httptest.NewRecorder()
		req := requestWithUser(httptest.NewRequest(http.MethodGet, "/api/v1/workspaces", nil), user)
		newWorkspaceRouter(api).ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"handle":"platform"`)
		assert.Contains(t, w.Body.String(), `"total":1`)
	})

	t.Run("list returns 401 without a user", func(t *testing.T) {
		api := newWorkspaceAPI(&fakeWorkspaceRecords{})
		w := httptest.NewRecorder()
		newWorkspaceRouter(api).
			ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces", nil))

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("list returns 500 when the repository fails", func(t *testing.T) {
		api := newWorkspaceAPI(&fakeWorkspaceRecords{listErr: errors.New("boom")})
		w := httptest.NewRecorder()
		req := requestWithUser(httptest.NewRequest(http.MethodGet, "/api/v1/workspaces", nil), user)
		newWorkspaceRouter(api).ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("get returns the workspace with the caller's role", func(t *testing.T) {
		api := newWorkspaceAPI(&fakeWorkspaceRecords{workspace: stored, membership: owner})
		w := httptest.NewRecorder()
		req := requestWithUser(httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1", nil), user)
		newWorkspaceRouter(api).ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"name":"Platform"`)
		assert.Contains(t, w.Body.String(), `"count":2`)
	})

	t.Run("get returns 404 for unknown workspace", func(t *testing.T) {
		api := newWorkspaceAPI(&fakeWorkspaceRecords{})
		w := httptest.NewRecorder()
		req := requestWithUser(httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1", nil), user)
		newWorkspaceRouter(api).ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "workspace_not_found")
	})

	t.Run("get works for an access key with no user and gives no role", func(t *testing.T) {
		api := newWorkspaceAPI(&fakeWorkspaceRecords{workspace: stored})
		w := httptest.NewRecorder()
		newWorkspaceRouter(api).
			ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1", nil))

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"role":""`)
	})

	t.Run("get returns 500 when the repository fails", func(t *testing.T) {
		api := newWorkspaceAPI(&fakeWorkspaceRecords{getErr: errors.New("boom")})
		w := httptest.NewRecorder()
		req := requestWithUser(httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1", nil), user)
		newWorkspaceRouter(api).ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("update returns 200 with the new name", func(t *testing.T) {
		api := newWorkspaceAPI(&fakeWorkspaceRecords{workspace: stored, membership: owner})
		w := httptest.NewRecorder()
		req := requestWithUser(
			httptest.NewRequest(
				http.MethodPut,
				"/api/v1/workspaces/ws-1",
				strings.NewReader(`{"name":"Platform Team"}`),
			),
			user,
		)
		newWorkspaceRouter(api).ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("update returns 400 for a name that is too short", func(t *testing.T) {
		api := newWorkspaceAPI(&fakeWorkspaceRecords{workspace: stored})
		w := httptest.NewRecorder()
		req := requestWithUser(
			httptest.NewRequest(http.MethodPut, "/api/v1/workspaces/ws-1", strings.NewReader(`{"name":"ab"}`)),
			user,
		)
		newWorkspaceRouter(api).ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("update returns 404 for unknown workspace", func(t *testing.T) {
		api := newWorkspaceAPI(&fakeWorkspaceRecords{})
		w := httptest.NewRecorder()
		req := requestWithUser(
			httptest.NewRequest(http.MethodPut, "/api/v1/workspaces/ws-1", strings.NewReader(body)),
			user,
		)
		newWorkspaceRouter(api).ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("update returns 500 when the repository fails", func(t *testing.T) {
		api := newWorkspaceAPI(&fakeWorkspaceRecords{workspace: stored, updateErr: errors.New("boom")})
		w := httptest.NewRecorder()
		req := requestWithUser(
			httptest.NewRequest(http.MethodPut, "/api/v1/workspaces/ws-1", strings.NewReader(body)),
			user,
		)
		newWorkspaceRouter(api).ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("update works for an access key with no user", func(t *testing.T) {
		api := newWorkspaceAPI(&fakeWorkspaceRecords{workspace: stored})
		w := httptest.NewRecorder()
		newWorkspaceRouter(api).
			ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/api/v1/workspaces/ws-1", strings.NewReader(body)))

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("delete returns 204", func(t *testing.T) {
		api := newWorkspaceAPI(&fakeWorkspaceRecords{workspace: stored})
		w := httptest.NewRecorder()
		newWorkspaceRouter(api).
			ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/api/v1/workspaces/ws-1", nil))

		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("delete returns 500 when the repository fails", func(t *testing.T) {
		api := newWorkspaceAPI(&fakeWorkspaceRecords{workspace: stored, deleteErr: errors.New("boom")})
		w := httptest.NewRecorder()
		newWorkspaceRouter(api).
			ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/api/v1/workspaces/ws-1", nil))

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("handlers return 400 when the workspace id is missing", func(t *testing.T) {
		api := newWorkspaceAPI(&fakeWorkspaceRecords{})
		req := requestWithUser(httptest.NewRequest(http.MethodGet, "/", nil), user)

		w := httptest.NewRecorder()
		api.GetWorkspaceAction(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)

		w = httptest.NewRecorder()
		api.DeleteWorkspaceAction(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
