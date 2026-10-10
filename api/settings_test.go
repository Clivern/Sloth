// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/clivern/cognit/db"
	"github.com/clivern/cognit/module"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestUnitSettingsActions(t *testing.T) {
	admin := &db.User{Id: "user-1", Role: db.UserRoleAdmin}
	body := `{"platformEmail":"ops@example.com","maintenanceMode":true}`

	newRouter := func(api *API) *chi.Mux {
		router := chi.NewRouter()
		router.Get("/api/v1/action/settings", api.GetSettingsAction)
		router.Put("/api/v1/action/settings", api.UpdateSettingsAction)
		return router
	}

	t.Run("update returns 200", func(t *testing.T) {
		api := &API{Settings: module.NewSettings(&fakeConfigs{})}
		w := httptest.NewRecorder()
		newRouter(api).
			ServeHTTP(
				w,
				requestWithUser(
					httptest.NewRequest(http.MethodPut, "/api/v1/action/settings", strings.NewReader(body)),
					admin,
				),
			)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "settings_updated_successfully")
	})

	t.Run("update returns 400 for an invalid email", func(t *testing.T) {
		api := &API{Settings: module.NewSettings(&fakeConfigs{})}
		w := httptest.NewRecorder()
		req := requestWithUser(
			httptest.NewRequest(
				http.MethodPut,
				"/api/v1/action/settings",
				strings.NewReader(`{"platformEmail":"nope"}`),
			),
			admin,
		)
		newRouter(api).ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("update returns 401 without a user", func(t *testing.T) {
		api := &API{Settings: module.NewSettings(&fakeConfigs{})}
		w := httptest.NewRecorder()
		newRouter(api).
			ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/api/v1/action/settings", strings.NewReader(body)))

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("update returns 500 when the repository fails", func(t *testing.T) {
		api := &API{Settings: module.NewSettings(&fakeConfigs{updateErr: errors.New("boom")})}
		w := httptest.NewRecorder()
		newRouter(api).
			ServeHTTP(
				w,
				requestWithUser(
					httptest.NewRequest(http.MethodPut, "/api/v1/action/settings", strings.NewReader(body)),
					admin,
				),
			)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "failed_update_settings")
	})

	t.Run("get returns the stored settings", func(t *testing.T) {
		api := &API{Settings: module.NewSettings(&fakeConfigs{})}
		w := httptest.NewRecorder()
		newRouter(api).
			ServeHTTP(w, requestWithUser(httptest.NewRequest(http.MethodGet, "/api/v1/action/settings", nil), admin))

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"platform_email":"value-of-platform_email"`)
	})

	t.Run("get returns 500 when the repository fails", func(t *testing.T) {
		api := &API{Settings: module.NewSettings(&fakeConfigs{getErr: errors.New("boom")})}
		w := httptest.NewRecorder()
		newRouter(api).
			ServeHTTP(w, requestWithUser(httptest.NewRequest(http.MethodGet, "/api/v1/action/settings", nil), admin))

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "failed_get_settings")
	})

	t.Run("get returns 401 without a user", func(t *testing.T) {
		api := &API{Settings: module.NewSettings(&fakeConfigs{})}
		w := httptest.NewRecorder()
		newRouter(api).
			ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/action/settings", nil))

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestUnitLogoutAction(t *testing.T) {
	user := &db.User{Id: "user-1"}

	t.Run("returns 200 for an anonymous request", func(t *testing.T) {
		api := &API{Auth: module.NewAuth(nil, &fakeSessions{}, nil)}
		w := httptest.NewRecorder()
		api.LogoutAction(w, httptest.NewRequest(http.MethodPost, "/api/v1/public/action/logout", nil))

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "logout_successful")
	})

	t.Run("revokes sessions for a signed-in user", func(t *testing.T) {
		api := &API{Auth: module.NewAuth(nil, &fakeSessions{}, nil)}
		w := httptest.NewRecorder()
		api.LogoutAction(
			w,
			requestWithUser(httptest.NewRequest(http.MethodPost, "/api/v1/public/action/logout", nil), user),
		)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Header().Get("Set-Cookie"), "_cognit_session=")
	})

	t.Run("still returns 200 when session revocation fails", func(t *testing.T) {
		api := &API{Auth: module.NewAuth(nil, &fakeSessions{deleteErr: errors.New("boom")}, nil)}
		w := httptest.NewRecorder()
		api.LogoutAction(
			w,
			requestWithUser(httptest.NewRequest(http.MethodPost, "/api/v1/public/action/logout", nil), user),
		)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
