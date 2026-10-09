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

func TestUnitKeyValueActions(t *testing.T) {
	workspace := &db.Workspace{Id: "ws-1"}
	stored := &db.WorkspaceKeyValue{
		Id:          "kv-1",
		WorkspaceId: "ws-1",
		Key:         "config/app",
		Value:       "on",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	t.Run("list returns items with total", func(t *testing.T) {
		api := newKeyValueAPI(workspace, &fakeKeyValues{items: []*db.WorkspaceKeyValue{stored}})
		w := httptest.NewRecorder()
		newKeyValueRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1/kv?prefix=config/", nil))

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"key":"config/app"`)
		assert.Contains(t, w.Body.String(), `"total":1`)
	})

	t.Run("list returns 400 for an invalid prefix", func(t *testing.T) {
		api := newKeyValueAPI(workspace, &fakeKeyValues{})
		w := httptest.NewRecorder()
		newKeyValueRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1/kv?prefix=a$", nil))

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid_kv_key")
	})

	t.Run("list returns 404 for unknown workspace", func(t *testing.T) {
		api := newKeyValueAPI(nil, &fakeKeyValues{})
		w := httptest.NewRecorder()
		newKeyValueRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1/kv", nil))

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("get returns one key", func(t *testing.T) {
		api := newKeyValueAPI(workspace, &fakeKeyValues{item: stored})
		w := httptest.NewRecorder()
		newKeyValueRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1/kv/config/app", nil))

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"value":"on"`)
	})

	t.Run("get returns 404 for a missing key", func(t *testing.T) {
		api := newKeyValueAPI(workspace, &fakeKeyValues{})
		w := httptest.NewRecorder()
		newKeyValueRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1/kv/config/app", nil))

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "kv_not_found")
	})

	t.Run("get returns 400 for an invalid key", func(t *testing.T) {
		api := newKeyValueAPI(workspace, &fakeKeyValues{})
		w := httptest.NewRecorder()
		newKeyValueRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1/kv/bad$key", nil))

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("get returns 500 when the repository fails", func(t *testing.T) {
		api := newKeyValueAPI(workspace, &fakeKeyValues{getErr: errors.New("boom")})
		w := httptest.NewRecorder()
		newKeyValueRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1/kv/config/app", nil))

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "failed_kv_request")
	})

	t.Run("put returns 200 with the stored key", func(t *testing.T) {
		api := newKeyValueAPI(workspace, &fakeKeyValues{})
		w := httptest.NewRecorder()
		newKeyValueRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/api/v1/workspaces/ws-1/kv/config/app", strings.NewReader(`{"value":"on"}`)))

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"value":"on"`)
	})

	t.Run("put returns 400 when value is missing", func(t *testing.T) {
		api := newKeyValueAPI(workspace, &fakeKeyValues{})
		w := httptest.NewRecorder()
		newKeyValueRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/api/v1/workspaces/ws-1/kv/config/app", strings.NewReader(`{}`)))

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("put returns 400 for malformed expiresAt", func(t *testing.T) {
		api := newKeyValueAPI(workspace, &fakeKeyValues{})
		w := httptest.NewRecorder()
		newKeyValueRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/api/v1/workspaces/ws-1/kv/config/app", strings.NewReader(`{"value":"on","expiresAt":"later"}`)))

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid_expires_at_format")
	})

	t.Run("put returns 404 for unknown workspace", func(t *testing.T) {
		api := newKeyValueAPI(nil, &fakeKeyValues{})
		w := httptest.NewRecorder()
		newKeyValueRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/api/v1/workspaces/ws-1/kv/config/app", strings.NewReader(`{"value":"on"}`)))

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("put returns 500 when the repository fails", func(t *testing.T) {
		api := newKeyValueAPI(workspace, &fakeKeyValues{upsertErr: errors.New("boom")})
		w := httptest.NewRecorder()
		newKeyValueRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/api/v1/workspaces/ws-1/kv/config/app", strings.NewReader(`{"value":"on"}`)))

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "failed_kv_request")
	})

	t.Run("delete returns 204", func(t *testing.T) {
		api := newKeyValueAPI(workspace, &fakeKeyValues{item: stored})
		w := httptest.NewRecorder()
		newKeyValueRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/api/v1/workspaces/ws-1/kv/config/app", nil))

		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("delete returns 404 for a missing key", func(t *testing.T) {
		api := newKeyValueAPI(workspace, &fakeKeyValues{})
		w := httptest.NewRecorder()
		newKeyValueRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/api/v1/workspaces/ws-1/kv/config/app", nil))

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("delete returns 500 when the repository fails", func(t *testing.T) {
		api := newKeyValueAPI(workspace, &fakeKeyValues{item: stored, deleteErr: errors.New("boom")})
		w := httptest.NewRecorder()
		newKeyValueRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/api/v1/workspaces/ws-1/kv/config/app", nil))

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("handlers return 400 when the key is missing", func(t *testing.T) {
		api := newKeyValueAPI(workspace, &fakeKeyValues{})

		w := httptest.NewRecorder()
		api.GetKeyValueAction(w, httptest.NewRequest(http.MethodGet, "/", nil))
		assert.Equal(t, http.StatusBadRequest, w.Code)

		w = httptest.NewRecorder()
		api.PutKeyValueAction(w, httptest.NewRequest(http.MethodPut, "/", strings.NewReader(`{"value":"on"}`)))
		assert.Equal(t, http.StatusBadRequest, w.Code)

		w = httptest.NewRecorder()
		api.DeleteKeyValueAction(w, httptest.NewRequest(http.MethodDelete, "/", nil))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
