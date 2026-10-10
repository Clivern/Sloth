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
	"github.com/clivern/cognit/module"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestUnitWorkspaceMemberActions(t *testing.T) {
	user := &db.User{Id: "user-1"}
	workspace := &db.Workspace{Id: "ws-1", Name: "Platform", Handle: "platform"}
	member := &db.WorkspaceUser{
		Id:          "member-1",
		WorkspaceId: "ws-1",
		UserId:      "user-2",
		Role:        db.UserRoleRegular,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	profile := &db.User{Id: "user-2", Name: "Bo", Email: "bo@example.com"}

	newAPI := func(workspace *db.Workspace, members *fakeWorkspaceMembers, users *fakeUsers) *API {
		return &API{
			Workspace: module.NewWorkspace(
				&fakeWorkspaceRecords{workspace: workspace},
				members,
				&fakeSubscriptionRecords{},
				users,
			),
		}
	}
	newRouter := func(api *API) *chi.Mux {
		router := chi.NewRouter()
		router.Get("/api/v1/workspaces/{workspaceId}/members", api.ListWorkspaceMembersAction)
		router.Put(
			"/api/v1/workspaces/{workspaceId}/members/{memberUserId}",
			api.UpdateWorkspaceMemberRoleAction,
		)
		router.Delete(
			"/api/v1/workspaces/{workspaceId}/members/{memberUserId}",
			api.DeleteWorkspaceMemberAction,
		)
		return router
	}
	roleBody := `{"role":"admin"}`

	t.Run("list returns members with their names", func(t *testing.T) {
		api := newAPI(
			workspace,
			&fakeWorkspaceMembers{list: []*db.WorkspaceUser{member}, count: 1},
			&fakeUsers{user: profile},
		)
		w := httptest.NewRecorder()
		req := requestWithUser(httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1/members", nil), user)
		newRouter(api).ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"email":"bo@example.com"`)
		assert.Contains(t, w.Body.String(), `"total":1`)
	})

	t.Run("list returns 401 without a user", func(t *testing.T) {
		api := newAPI(workspace, &fakeWorkspaceMembers{}, &fakeUsers{})
		w := httptest.NewRecorder()
		newRouter(api).
			ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1/members", nil))

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("list returns 404 for unknown workspace", func(t *testing.T) {
		api := newAPI(nil, &fakeWorkspaceMembers{}, &fakeUsers{})
		w := httptest.NewRecorder()
		req := requestWithUser(httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1/members", nil), user)
		newRouter(api).ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("list returns 500 when the repository fails", func(t *testing.T) {
		api := newAPI(workspace, &fakeWorkspaceMembers{err: errors.New("boom")}, &fakeUsers{})
		w := httptest.NewRecorder()
		req := requestWithUser(httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1/members", nil), user)
		newRouter(api).ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("update returns 200 with the new role", func(t *testing.T) {
		api := newAPI(workspace, &fakeWorkspaceMembers{member: member}, &fakeUsers{user: profile})
		w := httptest.NewRecorder()
		req := requestWithUser(
			httptest.NewRequest(
				http.MethodPut,
				"/api/v1/workspaces/ws-1/members/user-2",
				strings.NewReader(roleBody),
			),
			user,
		)
		newRouter(api).ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"role":"admin"`)
	})

	t.Run("update returns 400 for an unknown role", func(t *testing.T) {
		api := newAPI(workspace, &fakeWorkspaceMembers{member: member}, &fakeUsers{user: profile})
		w := httptest.NewRecorder()
		req := requestWithUser(
			httptest.NewRequest(
				http.MethodPut,
				"/api/v1/workspaces/ws-1/members/user-2",
				strings.NewReader(`{"role":"owner"}`),
			),
			user,
		)
		newRouter(api).ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("update returns 404 when the user is not a member", func(t *testing.T) {
		api := newAPI(workspace, &fakeWorkspaceMembers{}, &fakeUsers{user: profile})
		w := httptest.NewRecorder()
		req := requestWithUser(
			httptest.NewRequest(
				http.MethodPut,
				"/api/v1/workspaces/ws-1/members/user-2",
				strings.NewReader(roleBody),
			),
			user,
		)
		newRouter(api).ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "workspace_member_not_found")
	})

	t.Run("update returns 500 when the repository fails", func(t *testing.T) {
		api := newAPI(
			workspace,
			&fakeWorkspaceMembers{member: member, updateErr: errors.New("boom")},
			&fakeUsers{user: profile},
		)
		w := httptest.NewRecorder()
		req := requestWithUser(
			httptest.NewRequest(
				http.MethodPut,
				"/api/v1/workspaces/ws-1/members/user-2",
				strings.NewReader(roleBody),
			),
			user,
		)
		newRouter(api).ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("update returns 401 without a user", func(t *testing.T) {
		api := newAPI(workspace, &fakeWorkspaceMembers{member: member}, &fakeUsers{user: profile})
		w := httptest.NewRecorder()
		newRouter(api).
			ServeHTTP(
				w,
				httptest.NewRequest(
					http.MethodPut,
					"/api/v1/workspaces/ws-1/members/user-2",
					strings.NewReader(roleBody),
				),
			)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("delete returns 204", func(t *testing.T) {
		api := newAPI(workspace, &fakeWorkspaceMembers{member: member}, &fakeUsers{user: profile})
		w := httptest.NewRecorder()
		req := requestWithUser(
			httptest.NewRequest(http.MethodDelete, "/api/v1/workspaces/ws-1/members/user-2", nil),
			user,
		)
		newRouter(api).ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("delete returns 404 when the user is not a member", func(t *testing.T) {
		api := newAPI(workspace, &fakeWorkspaceMembers{}, &fakeUsers{user: profile})
		w := httptest.NewRecorder()
		req := requestWithUser(
			httptest.NewRequest(http.MethodDelete, "/api/v1/workspaces/ws-1/members/user-2", nil),
			user,
		)
		newRouter(api).ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("delete returns 500 when the repository fails", func(t *testing.T) {
		api := newAPI(
			workspace,
			&fakeWorkspaceMembers{member: member, deleteErr: errors.New("boom")},
			&fakeUsers{user: profile},
		)
		w := httptest.NewRecorder()
		req := requestWithUser(
			httptest.NewRequest(http.MethodDelete, "/api/v1/workspaces/ws-1/members/user-2", nil),
			user,
		)
		newRouter(api).ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("delete returns 401 without a user", func(t *testing.T) {
		api := newAPI(workspace, &fakeWorkspaceMembers{member: member}, &fakeUsers{user: profile})
		w := httptest.NewRecorder()
		newRouter(api).
			ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/api/v1/workspaces/ws-1/members/user-2", nil))

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
