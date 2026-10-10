// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/clivern/cognit/module"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestUnitSetupActions(t *testing.T) {
	newRouter := func(api *API) *chi.Mux {
		router := chi.NewRouter()
		router.Post("/api/v1/public/action/setup", api.SetupAction)
		router.Get("/api/v1/public/action/setup/status", api.SetupStatusAction)
		return router
	}
	body := `{"platformEmail":"ops@example.com"}`

	t.Run("setup returns 200 on a fresh install", func(t *testing.T) {
		api := &API{Setup: module.NewSetup(&fakeSetupConfigs{}, &fakeSetupUsers{})}
		w := httptest.NewRecorder()
		newRouter(api).
			ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/v1/public/action/setup", strings.NewReader(body)))

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "setup_completed_successfully")
	})

	t.Run("setup returns 400 when the platform is already installed", func(t *testing.T) {
		api := &API{Setup: module.NewSetup(&fakeSetupConfigs{installed: true}, &fakeSetupUsers{})}
		w := httptest.NewRecorder()
		newRouter(api).
			ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/v1/public/action/setup", strings.NewReader(body)))

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "platform_already_installed")
	})

	t.Run("setup returns 400 for an invalid email", func(t *testing.T) {
		api := &API{Setup: module.NewSetup(&fakeSetupConfigs{}, &fakeSetupUsers{})}
		w := httptest.NewRecorder()
		newRouter(api).
			ServeHTTP(
				w,
				httptest.NewRequest(
					http.MethodPost,
					"/api/v1/public/action/setup",
					strings.NewReader(`{"platformEmail":"nope"}`),
				),
			)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("setup returns 500 when the bot user cannot be created", func(t *testing.T) {
		api := &API{
			Setup: module.NewSetup(&fakeSetupConfigs{}, &fakeSetupUsers{createErr: errors.New("boom")}),
		}
		w := httptest.NewRecorder()
		newRouter(api).
			ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/v1/public/action/setup", strings.NewReader(body)))

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "failed_complete_setup")
	})

	t.Run("status reports not installed on a fresh database", func(t *testing.T) {
		api := &API{Setup: module.NewSetup(&fakeSetupConfigs{}, &fakeSetupUsers{})}
		w := httptest.NewRecorder()
		newRouter(api).
			ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/public/action/setup/status", nil))

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"installed":false`)
	})

	t.Run("status reports installed after setup", func(t *testing.T) {
		api := &API{Setup: module.NewSetup(&fakeSetupConfigs{installed: true}, &fakeSetupUsers{})}
		w := httptest.NewRecorder()
		newRouter(api).
			ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/public/action/setup/status", nil))

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"installed":true`)
	})
}
