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

func TestUnitUserAPIKeyActions(t *testing.T) {
	user := &db.User{Id: "user-1"}
	stored := &db.APIKey{
		Id:        "key-1",
		UserId:    "user-1",
		Name:      "laptop",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	t.Run("create returns 201 with the raw key", func(t *testing.T) {
		api := &API{APIKey: module.NewAPIKey(&fakeAPIKeys{})}
		req := requestWithUser(httptest.NewRequest(http.MethodPost, "/api/v1/apiKeys", strings.NewReader(`{"name":"laptop"}`)), user)
		w := httptest.NewRecorder()
		newAPIKeyRouter(api).ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Body.String(), `"key":"`)
		assert.Contains(t, w.Body.String(), `"name":"laptop"`)
	})

	t.Run("create returns 401 without a user", func(t *testing.T) {
		api := &API{APIKey: module.NewAPIKey(&fakeAPIKeys{})}
		w := httptest.NewRecorder()
		newAPIKeyRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/v1/apiKeys", strings.NewReader(`{"name":"laptop"}`)))

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "not_authenticated")
	})

	t.Run("create returns 400 for invalid body", func(t *testing.T) {
		api := &API{APIKey: module.NewAPIKey(&fakeAPIKeys{})}
		req := requestWithUser(httptest.NewRequest(http.MethodPost, "/api/v1/apiKeys", strings.NewReader(`{}`)), user)
		w := httptest.NewRecorder()
		newAPIKeyRouter(api).ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("create returns 400 for malformed expiresAt", func(t *testing.T) {
		api := &API{APIKey: module.NewAPIKey(&fakeAPIKeys{})}
		req := requestWithUser(httptest.NewRequest(http.MethodPost, "/api/v1/apiKeys", strings.NewReader(`{"name":"laptop","expiresAt":"soon"}`)), user)
		w := httptest.NewRecorder()
		newAPIKeyRouter(api).ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid_expires_at_format")
	})

	t.Run("create returns 500 when the repository fails", func(t *testing.T) {
		api := &API{APIKey: module.NewAPIKey(&fakeAPIKeys{createErr: errors.New("boom")})}
		req := requestWithUser(httptest.NewRequest(http.MethodPost, "/api/v1/apiKeys", strings.NewReader(`{"name":"laptop"}`)), user)
		w := httptest.NewRecorder()
		newAPIKeyRouter(api).ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "failed_create_api_key")
	})

	t.Run("list returns the user's keys with pagination meta", func(t *testing.T) {
		api := &API{APIKey: module.NewAPIKey(&fakeAPIKeys{keys: []*db.APIKey{stored}, total: 1})}
		req := requestWithUser(httptest.NewRequest(http.MethodGet, "/api/v1/apiKeys", nil), user)
		w := httptest.NewRecorder()
		newAPIKeyRouter(api).ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"total":1`)
		assert.NotContains(t, w.Body.String(), `"key":`)
	})

	t.Run("list returns 401 without a user", func(t *testing.T) {
		api := &API{APIKey: module.NewAPIKey(&fakeAPIKeys{})}
		w := httptest.NewRecorder()
		newAPIKeyRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/apiKeys", nil))

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("list returns 500 when the repository fails", func(t *testing.T) {
		api := &API{APIKey: module.NewAPIKey(&fakeAPIKeys{listErr: errors.New("boom")})}
		req := requestWithUser(httptest.NewRequest(http.MethodGet, "/api/v1/apiKeys", nil), user)
		w := httptest.NewRecorder()
		newAPIKeyRouter(api).ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "failed_list_api_keys")
	})

	t.Run("get returns one of the user's keys", func(t *testing.T) {
		api := &API{APIKey: module.NewAPIKey(&fakeAPIKeys{apiKey: stored})}
		req := requestWithUser(httptest.NewRequest(http.MethodGet, "/api/v1/apiKeys/key-1", nil), user)
		w := httptest.NewRecorder()
		newAPIKeyRouter(api).ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"name":"laptop"`)
	})

	t.Run("get returns 404 for another user's key", func(t *testing.T) {
		other := *stored
		other.UserId = "user-2"
		api := &API{APIKey: module.NewAPIKey(&fakeAPIKeys{apiKey: &other})}
		req := requestWithUser(httptest.NewRequest(http.MethodGet, "/api/v1/apiKeys/key-1", nil), user)
		w := httptest.NewRecorder()
		newAPIKeyRouter(api).ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "api_key_not_found")
	})

	t.Run("delete returns 204 for the user's key", func(t *testing.T) {
		api := &API{APIKey: module.NewAPIKey(&fakeAPIKeys{apiKey: stored})}
		req := requestWithUser(httptest.NewRequest(http.MethodDelete, "/api/v1/apiKeys/key-1", nil), user)
		w := httptest.NewRecorder()
		newAPIKeyRouter(api).ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("delete returns 404 for another user's key", func(t *testing.T) {
		other := *stored
		other.UserId = "user-2"
		api := &API{APIKey: module.NewAPIKey(&fakeAPIKeys{apiKey: &other})}
		req := requestWithUser(httptest.NewRequest(http.MethodDelete, "/api/v1/apiKeys/key-1", nil), user)
		w := httptest.NewRecorder()
		newAPIKeyRouter(api).ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("delete returns 500 when the repository fails", func(t *testing.T) {
		api := &API{APIKey: module.NewAPIKey(&fakeAPIKeys{apiKey: stored, deleteErr: errors.New("boom")})}
		req := requestWithUser(httptest.NewRequest(http.MethodDelete, "/api/v1/apiKeys/key-1", nil), user)
		w := httptest.NewRecorder()
		newAPIKeyRouter(api).ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "failed_delete_api_key")
	})

	t.Run("get and delete return 401 without a user", func(t *testing.T) {
		api := &API{APIKey: module.NewAPIKey(&fakeAPIKeys{apiKey: stored})}

		w := httptest.NewRecorder()
		newAPIKeyRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/apiKeys/key-1", nil))
		assert.Equal(t, http.StatusUnauthorized, w.Code)

		w = httptest.NewRecorder()
		newAPIKeyRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/api/v1/apiKeys/key-1", nil))
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
