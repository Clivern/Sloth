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

	"github.com/stretchr/testify/assert"
)

func TestUnitAccessKeyActions(t *testing.T) {
	workspace := &db.Workspace{Id: "ws-1"}
	permissions := `["` + module.CanRegisterInstance + `"]`
	meta := `{"permissions":` + permissions + `}`
	storedKey := &db.AccessKey{
		Id:          "key-1",
		WorkspaceId: "ws-1",
		Name:        "ci",
		Meta:        &meta,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	t.Run("create returns 201 with the raw key", func(t *testing.T) {
		api := newAccessKeyAPI(workspace, &fakeAccessKeys{})
		body := `{"name":"ci","permissions":` + permissions + `}`
		w := httptest.NewRecorder()
		newAccessKeyRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/ws-1/access-keys", strings.NewReader(body)))

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Body.String(), `"key":"`)
		assert.Contains(t, w.Body.String(), `"name":"ci"`)
	})

	t.Run("create returns 400 for invalid body", func(t *testing.T) {
		api := newAccessKeyAPI(workspace, &fakeAccessKeys{})
		w := httptest.NewRecorder()
		newAccessKeyRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/ws-1/access-keys", strings.NewReader(`{}`)))

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("create returns 400 for unknown permission", func(t *testing.T) {
		api := newAccessKeyAPI(workspace, &fakeAccessKeys{})
		body := `{"name":"ci","permissions":["NOPE"]}`
		w := httptest.NewRecorder()
		newAccessKeyRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/ws-1/access-keys", strings.NewReader(body)))

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid_access_key_permissions")
	})

	t.Run("create returns 400 for malformed expiresAt", func(t *testing.T) {
		api := newAccessKeyAPI(workspace, &fakeAccessKeys{})
		body := `{"name":"ci","expiresAt":"tomorrow","permissions":` + permissions + `}`
		w := httptest.NewRecorder()
		newAccessKeyRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/ws-1/access-keys", strings.NewReader(body)))

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid_expires_at_format")
	})

	t.Run("create returns 404 for unknown workspace", func(t *testing.T) {
		api := newAccessKeyAPI(nil, &fakeAccessKeys{})
		body := `{"name":"ci","permissions":` + permissions + `}`
		w := httptest.NewRecorder()
		newAccessKeyRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/ws-1/access-keys", strings.NewReader(body)))

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "workspace_not_found")
	})

	t.Run("create returns 500 when the repository fails", func(t *testing.T) {
		api := newAccessKeyAPI(workspace, &fakeAccessKeys{createErr: errors.New("boom")})
		body := `{"name":"ci","permissions":` + permissions + `}`
		w := httptest.NewRecorder()
		newAccessKeyRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/ws-1/access-keys", strings.NewReader(body)))

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "failed_create_access_key")
	})

	t.Run("list returns keys with pagination meta", func(t *testing.T) {
		api := newAccessKeyAPI(workspace, &fakeAccessKeys{keys: []*db.AccessKey{storedKey}, total: 1})
		w := httptest.NewRecorder()
		newAccessKeyRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1/access-keys?limit=10&offset=0", nil))

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"keys":[`)
		assert.Contains(t, w.Body.String(), `"total":1`)
		assert.NotContains(t, w.Body.String(), `"key":`)
	})

	t.Run("list returns 404 for unknown workspace", func(t *testing.T) {
		api := newAccessKeyAPI(nil, &fakeAccessKeys{})
		w := httptest.NewRecorder()
		newAccessKeyRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1/access-keys", nil))

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("list returns 500 when the repository fails", func(t *testing.T) {
		api := newAccessKeyAPI(workspace, &fakeAccessKeys{listErr: errors.New("boom")})
		w := httptest.NewRecorder()
		newAccessKeyRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1/access-keys", nil))

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "failed_list_access_keys")
	})

	t.Run("get returns one key", func(t *testing.T) {
		api := newAccessKeyAPI(workspace, &fakeAccessKeys{key: storedKey})
		w := httptest.NewRecorder()
		newAccessKeyRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1/access-keys/key-1", nil))

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"name":"ci"`)
	})

	t.Run("get returns 404 when the key belongs to another workspace", func(t *testing.T) {
		other := *storedKey
		other.WorkspaceId = "ws-2"
		api := newAccessKeyAPI(workspace, &fakeAccessKeys{key: &other})
		w := httptest.NewRecorder()
		newAccessKeyRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1/access-keys/key-1", nil))

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "access_key_not_found")
	})

	t.Run("get returns 404 for unknown workspace", func(t *testing.T) {
		api := newAccessKeyAPI(nil, &fakeAccessKeys{key: storedKey})
		w := httptest.NewRecorder()
		newAccessKeyRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1/access-keys/key-1", nil))

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "workspace_not_found")
	})

	t.Run("delete returns 204", func(t *testing.T) {
		api := newAccessKeyAPI(workspace, &fakeAccessKeys{key: storedKey})
		w := httptest.NewRecorder()
		newAccessKeyRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/api/v1/workspaces/ws-1/access-keys/key-1", nil))

		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("delete returns 404 for unknown key", func(t *testing.T) {
		api := newAccessKeyAPI(workspace, &fakeAccessKeys{})
		w := httptest.NewRecorder()
		newAccessKeyRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/api/v1/workspaces/ws-1/access-keys/key-1", nil))

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "access_key_not_found")
	})

	t.Run("delete returns 500 when the repository fails", func(t *testing.T) {
		api := newAccessKeyAPI(workspace, &fakeAccessKeys{key: storedKey, deleteErr: errors.New("boom")})
		w := httptest.NewRecorder()
		newAccessKeyRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/api/v1/workspaces/ws-1/access-keys/key-1", nil))

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "failed_delete_access_key")
	})

	t.Run("handlers return 400 when path params are missing", func(t *testing.T) {
		api := newAccessKeyAPI(workspace, &fakeAccessKeys{})

		w := httptest.NewRecorder()
		api.ListAccessKeysAction(w, httptest.NewRequest(http.MethodGet, "/", nil))
		assert.Equal(t, http.StatusBadRequest, w.Code)

		w = httptest.NewRecorder()
		api.GetAccessKeyAction(w, httptest.NewRequest(http.MethodGet, "/", nil))
		assert.Equal(t, http.StatusBadRequest, w.Code)

		w = httptest.NewRecorder()
		api.DeleteAccessKeyAction(w, httptest.NewRequest(http.MethodDelete, "/", nil))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
