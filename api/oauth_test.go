// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/clivern/cognit/pkg/github"
	"github.com/clivern/cognit/pkg/google"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestUnitOAuthActions(t *testing.T) {
	api := &API{
		GitHubOAuth: github.NewOAuth(github.OAuthConfig{
			ClientID:    "github-client",
			RedirectURL: "http://localhost:8080/api/v1/public/oauth/github/callback",
			Scopes:      []string{"read:user", "user:email"},
		}),
		GoogleOAuth: google.NewOAuth(google.OAuthConfig{
			ClientID:    "google-client",
			RedirectURL: "http://localhost:8080/api/v1/public/oauth/google/callback",
			Scopes:      []string{"openid", "email", "profile"},
		}),
	}
	router := chi.NewRouter()
	router.Get("/api/v1/public/oauth/github", api.GitHubOAuthStartAction)
	router.Get("/api/v1/public/oauth/github/callback", api.GitHubOAuthCallbackAction)
	router.Get("/api/v1/public/oauth/google", api.GoogleOAuthStartAction)
	router.Get("/api/v1/public/oauth/google/callback", api.GoogleOAuthCallbackAction)

	t.Run("github start redirects to the authorize URL and sets a state cookie", func(t *testing.T) {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/public/oauth/github", nil))

		assert.Equal(t, http.StatusFound, w.Code)
		assert.Contains(t, w.Header().Get("Location"), "github-client")
		assert.Contains(t, w.Header().Get("Location"), "state=")
		assert.Contains(t, w.Header().Get("Set-Cookie"), OauthStateCookie+"=")
	})

	t.Run("google start redirects to the authorize URL and sets a state cookie", func(t *testing.T) {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/public/oauth/google", nil))

		assert.Equal(t, http.StatusFound, w.Code)
		assert.Contains(t, w.Header().Get("Location"), "google-client")
		assert.Contains(t, w.Header().Get("Set-Cookie"), OauthStateCookie+"=")
	})

	t.Run("github callback rejects a mismatched state", func(t *testing.T) {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/public/oauth/github/callback?code=abc&state=forged", nil))

		assert.Equal(t, http.StatusFound, w.Code)
		assert.Contains(t, w.Header().Get("Location"), "/login?oauth_error=github")
	})

	t.Run("google callback rejects a mismatched state", func(t *testing.T) {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/public/oauth/google/callback?code=abc&state=forged", nil))

		assert.Equal(t, http.StatusFound, w.Code)
		assert.Contains(t, w.Header().Get("Location"), "/login?oauth_error=google")
	})
}
