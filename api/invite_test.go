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

func TestUnitInviteActions(t *testing.T) {
	user := &db.User{Id: "user-1"}
	workspace := &db.Workspace{Id: "ws-1"}
	stored := &db.UserInvite{
		Id:            "invite-1",
		Email:         "bo@example.com",
		Role:          db.UserRoleRegular,
		Status:        module.InviteStatusPending,
		InviterUserId: "user-1",
		WorkspaceId:   "ws-1",
		ExpiresAt:     time.Now().Add(time.Hour),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	body := `{"email":"bo@example.com","role":"regular"}`

	newAPI := func(workspace *db.Workspace, invites *fakeInvites, users *fakeUsers, mailer *fakeInviteMailer) *API {
		return &API{
			Invite: module.NewInvite(
				invites,
				users,
				&fakeConfigs{},
				&fakeWorkspaceRecords{workspace: workspace},
				&fakeWorkspaceMembers{},
				mailer,
			),
		}
	}
	newRouter := func(api *API) *chi.Mux {
		router := chi.NewRouter()
		router.Post("/api/v1/workspaces/{workspaceId}/invites", api.CreateInviteAction)
		router.Get("/api/v1/workspaces/{workspaceId}/invites", api.ListInvitesAction)
		router.Get("/api/v1/workspaces/{workspaceId}/invites/{inviteId}", api.GetInviteAction)
		router.Delete("/api/v1/workspaces/{workspaceId}/invites/{inviteId}", api.DeleteInviteAction)
		return router
	}

	t.Run("create returns 201 and sends the invite email", func(t *testing.T) {
		api := newAPI(workspace, &fakeInvites{}, &fakeUsers{}, &fakeInviteMailer{})
		w := httptest.NewRecorder()
		req := requestWithUser(
			httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/ws-1/invites", strings.NewReader(body)),
			user,
		)
		newRouter(api).ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Body.String(), `"status":"pending"`)
	})

	t.Run("create still returns 201 when the email fails to send", func(t *testing.T) {
		api := newAPI(workspace, &fakeInvites{}, &fakeUsers{}, &fakeInviteMailer{err: errors.New("smtp down")})
		w := httptest.NewRecorder()
		req := requestWithUser(
			httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/ws-1/invites", strings.NewReader(body)),
			user,
		)
		newRouter(api).ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("create returns 401 without a user", func(t *testing.T) {
		api := newAPI(workspace, &fakeInvites{}, &fakeUsers{}, &fakeInviteMailer{})
		w := httptest.NewRecorder()
		newRouter(api).
			ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/ws-1/invites", strings.NewReader(body)))

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("create returns 400 for an invalid email", func(t *testing.T) {
		api := newAPI(workspace, &fakeInvites{}, &fakeUsers{}, &fakeInviteMailer{})
		w := httptest.NewRecorder()
		req := requestWithUser(
			httptest.NewRequest(
				http.MethodPost,
				"/api/v1/workspaces/ws-1/invites",
				strings.NewReader(`{"email":"nope","role":"regular"}`),
			),
			user,
		)
		newRouter(api).ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("create returns 404 for unknown workspace", func(t *testing.T) {
		api := newAPI(nil, &fakeInvites{}, &fakeUsers{}, &fakeInviteMailer{})
		w := httptest.NewRecorder()
		req := requestWithUser(
			httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/ws-1/invites", strings.NewReader(body)),
			user,
		)
		newRouter(api).ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("create returns 409 when a pending invite already exists", func(t *testing.T) {
		api := newAPI(workspace, &fakeInvites{pending: 1}, &fakeUsers{}, &fakeInviteMailer{})
		w := httptest.NewRecorder()
		req := requestWithUser(
			httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/ws-1/invites", strings.NewReader(body)),
			user,
		)
		newRouter(api).ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
	})

	t.Run("create returns 409 when the user is already a member", func(t *testing.T) {
		existing := &db.User{Id: "user-2", Email: "bo@example.com"}
		api := &API{
			Invite: module.NewInvite(
				&fakeInvites{},
				&fakeUsers{user: existing},
				&fakeConfigs{},
				&fakeWorkspaceRecords{workspace: workspace},
				&fakeWorkspaceMembers{member: &db.WorkspaceUser{Id: "m-1"}},
				&fakeInviteMailer{},
			),
		}
		w := httptest.NewRecorder()
		req := requestWithUser(
			httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/ws-1/invites", strings.NewReader(body)),
			user,
		)
		newRouter(api).ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
		assert.Contains(t, w.Body.String(), "user_already_workspace_member")
	})

	t.Run("create returns 500 when the repository fails", func(t *testing.T) {
		api := newAPI(workspace, &fakeInvites{createErr: errors.New("boom")}, &fakeUsers{}, &fakeInviteMailer{})
		w := httptest.NewRecorder()
		req := requestWithUser(
			httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/ws-1/invites", strings.NewReader(body)),
			user,
		)
		newRouter(api).ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("list returns invites", func(t *testing.T) {
		api := newAPI(workspace, &fakeInvites{list: []*db.UserInvite{stored}}, &fakeUsers{}, &fakeInviteMailer{})
		w := httptest.NewRecorder()
		newRouter(api).
			ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1/invites", nil))

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"email":"bo@example.com"`)
	})

	t.Run("list returns an empty page for unknown workspace", func(t *testing.T) {
		api := newAPI(nil, &fakeInvites{}, &fakeUsers{}, &fakeInviteMailer{})
		w := httptest.NewRecorder()
		newRouter(api).
			ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1/invites", nil))

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("get returns one invite", func(t *testing.T) {
		api := newAPI(workspace, &fakeInvites{invite: stored}, &fakeUsers{}, &fakeInviteMailer{})
		w := httptest.NewRecorder()
		newRouter(api).
			ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1/invites/invite-1", nil))

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"role":"regular"`)
	})

	t.Run("get returns 404 for unknown invite", func(t *testing.T) {
		api := newAPI(workspace, &fakeInvites{}, &fakeUsers{}, &fakeInviteMailer{})
		w := httptest.NewRecorder()
		newRouter(api).
			ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1/invites/invite-1", nil))

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "invite_not_found")
	})

	t.Run("get returns 500 when the repository fails", func(t *testing.T) {
		api := newAPI(workspace, &fakeInvites{getErr: errors.New("boom")}, &fakeUsers{}, &fakeInviteMailer{})
		w := httptest.NewRecorder()
		newRouter(api).
			ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1/invites/invite-1", nil))

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("delete returns 204", func(t *testing.T) {
		api := newAPI(workspace, &fakeInvites{invite: stored}, &fakeUsers{}, &fakeInviteMailer{})
		w := httptest.NewRecorder()
		newRouter(api).
			ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/api/v1/workspaces/ws-1/invites/invite-1", nil))

		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("delete returns 404 for unknown invite", func(t *testing.T) {
		api := newAPI(workspace, &fakeInvites{}, &fakeUsers{}, &fakeInviteMailer{})
		w := httptest.NewRecorder()
		newRouter(api).
			ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/api/v1/workspaces/ws-1/invites/invite-1", nil))

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("delete returns 500 when the repository fails", func(t *testing.T) {
		api := newAPI(
			workspace,
			&fakeInvites{invite: stored, deleteErr: errors.New("boom")},
			&fakeUsers{},
			&fakeInviteMailer{},
		)
		w := httptest.NewRecorder()
		newRouter(api).
			ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/api/v1/workspaces/ws-1/invites/invite-1", nil))

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
