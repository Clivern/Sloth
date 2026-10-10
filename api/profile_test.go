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

func TestUnitProfileActions(t *testing.T) {
	stored := &db.User{
		Id:          "user-1",
		Name:        "Ann",
		Email:       "ann@example.com",
		Role:        db.UserRoleRegular,
		Language:    "en",
		Theme:       "default",
		LastLoginAt: time.Now(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	user := &db.User{Id: "user-1"}
	body := `{"name":"Ann","language":"fr","theme":"dark"}`

	t.Run("get returns the profile", func(t *testing.T) {
		api := &API{Profile: module.NewProfile(&fakeUsers{user: stored})}
		w := httptest.NewRecorder()
		newProfileRouter(api).
			ServeHTTP(w, requestWithUser(httptest.NewRequest(http.MethodGet, "/api/v1/action/profile", nil), user))

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"email":"ann@example.com"`)
		assert.Contains(t, w.Body.String(), `"avatar":"https://`)
	})

	t.Run("get returns 401 without a user", func(t *testing.T) {
		api := &API{Profile: module.NewProfile(&fakeUsers{user: stored})}
		w := httptest.NewRecorder()
		newProfileRouter(api).
			ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/action/profile", nil))

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("get returns 500 when the user is missing", func(t *testing.T) {
		api := &API{Profile: module.NewProfile(&fakeUsers{})}
		w := httptest.NewRecorder()
		newProfileRouter(api).
			ServeHTTP(w, requestWithUser(httptest.NewRequest(http.MethodGet, "/api/v1/action/profile", nil), user))

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "failed_get_profile")
	})

	t.Run("update returns 200 with the saved profile", func(t *testing.T) {
		api := &API{Profile: module.NewProfile(&fakeUsers{user: stored})}
		w := httptest.NewRecorder()
		newProfileRouter(api).
			ServeHTTP(
				w,
				requestWithUser(
					httptest.NewRequest(http.MethodPut, "/api/v1/action/profile", strings.NewReader(body)),
					user,
				),
			)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"theme":"dark"`)
		assert.Contains(t, w.Body.String(), `"successMessage":"profile_updated_successfully"`)
	})

	t.Run("update returns 400 for an unsupported language", func(t *testing.T) {
		api := &API{Profile: module.NewProfile(&fakeUsers{user: stored})}
		w := httptest.NewRecorder()
		req := requestWithUser(
			httptest.NewRequest(
				http.MethodPut,
				"/api/v1/action/profile",
				strings.NewReader(`{"language":"de","theme":"dark"}`),
			),
			user,
		)
		newProfileRouter(api).ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("update returns 500 when the repository fails", func(t *testing.T) {
		api := &API{
			Profile: module.NewProfile(&fakeUsers{user: stored, updateErr: errors.New("boom")}),
		}
		w := httptest.NewRecorder()
		newProfileRouter(api).
			ServeHTTP(
				w,
				requestWithUser(
					httptest.NewRequest(http.MethodPut, "/api/v1/action/profile", strings.NewReader(body)),
					user,
				),
			)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "failed_update_profile")
	})

	t.Run("update returns 401 without a user", func(t *testing.T) {
		api := &API{Profile: module.NewProfile(&fakeUsers{user: stored})}
		w := httptest.NewRecorder()
		newProfileRouter(api).
			ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/api/v1/action/profile", strings.NewReader(body)))

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
