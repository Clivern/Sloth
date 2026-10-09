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

	"github.com/stretchr/testify/assert"
)

func TestUnitGetMeAction(t *testing.T) {
	user := &db.User{Id: "user-1", Email: "ann@example.com", Name: "Ann", Role: db.UserRoleRegular}
	owner := &db.User{Id: "user-1", Email: "ann@example.com", Name: "Ann", Role: db.UserRoleRegular}
	apiKey := &db.APIKey{Id: "key-1", UserId: "user-1", Name: "laptop", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	meta := `{"permissions":["` + module.CanRegisterInstance + `"]}`
	accessKey := &db.AccessKey{Id: "access-1", WorkspaceId: "ws-1", Name: "ci", Meta: &meta, CreatedAt: time.Now(), UpdatedAt: time.Now()}

	newAPI := func(keys *fakeAPIKeys, access *fakeAccessKeys, users *fakeUsers) *API {
		return &API{Me: module.NewMe(keys, users, access)}
	}

	t.Run("returns the session user when no key header is sent", func(t *testing.T) {
		api := newAPI(&fakeAPIKeys{}, &fakeAccessKeys{}, &fakeUsers{})
		w := httptest.NewRecorder()
		api.GetMeAction(w, requestWithUser(httptest.NewRequest(http.MethodGet, "/api/v1/me", nil), user))

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"principalType":"session"`)
	})

	t.Run("returns 403 with no key and no session", func(t *testing.T) {
		api := newAPI(&fakeAPIKeys{}, &fakeAccessKeys{}, &fakeUsers{})
		w := httptest.NewRecorder()
		api.GetMeAction(w, httptest.NewRequest(http.MethodGet, "/api/v1/me", nil))

		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Contains(t, w.Body.String(), "me_requires_key_header")
	})

	t.Run("returns the API key principal for X-API-Key", func(t *testing.T) {
		api := newAPI(&fakeAPIKeys{apiKey: apiKey}, &fakeAccessKeys{}, &fakeUsers{user: owner})
		req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
		req.Header.Set("X-API-Key", "raw-key")
		w := httptest.NewRecorder()
		api.GetMeAction(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"principalType":"api_key"`)
		assert.Contains(t, w.Body.String(), `"name":"laptop"`)
	})

	t.Run("returns 401 for an unknown X-API-Key", func(t *testing.T) {
		api := newAPI(&fakeAPIKeys{}, &fakeAccessKeys{}, &fakeUsers{})
		req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
		req.Header.Set("X-API-Key", "missing")
		w := httptest.NewRecorder()
		api.GetMeAction(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "invalid_api_key")
	})

	t.Run("returns 500 when the API key lookup fails", func(t *testing.T) {
		keys := &fakeAPIKeys{getErr: errors.New("boom")}
		api := newAPI(keys, &fakeAccessKeys{}, &fakeUsers{})
		req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
		req.Header.Set("X-API-Key", "raw-key")
		w := httptest.NewRecorder()
		api.GetMeAction(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "failed_get_me")
	})

	t.Run("returns the access key principal for X-Access-Key", func(t *testing.T) {
		api := newAPI(&fakeAPIKeys{}, &fakeAccessKeys{key: accessKey}, &fakeUsers{})
		req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
		req.Header.Set("X-Access-Key", "raw-key")
		w := httptest.NewRecorder()
		api.GetMeAction(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"principalType":"access_key"`)
		assert.Contains(t, w.Body.String(), `"permissions":["CAN_REGISTER_INSTANCE"]`)
	})

	t.Run("returns 401 for an unknown X-Access-Key", func(t *testing.T) {
		api := newAPI(&fakeAPIKeys{}, &fakeAccessKeys{}, &fakeUsers{})
		req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
		req.Header.Set("X-Access-Key", "missing")
		w := httptest.NewRecorder()
		api.GetMeAction(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "invalid_access_key")
	})
}
