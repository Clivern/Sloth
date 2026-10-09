// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package util

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUnitCookies(t *testing.T) {
	t.Run("Cookie options", func(t *testing.T) {
		defaultOpts := DefaultCookieOptions()
		assert.Equal(t, "/", defaultOpts.Path)
		assert.True(t, defaultOpts.HTTPOnly)
		assert.False(t, defaultOpts.Secure)
		assert.Equal(t, http.SameSiteLaxMode, defaultOpts.SameSite)

		secureOpts := SecureCookieOptions()
		assert.Equal(t, "/", secureOpts.Path)
		assert.True(t, secureOpts.HTTPOnly)
		assert.True(t, secureOpts.Secure)
		assert.Equal(t, http.SameSiteStrictMode, secureOpts.SameSite)
	})

	t.Run("Complete cookie workflow", func(t *testing.T) {
		w := httptest.NewRecorder()
		opts := DefaultCookieOptions()
		opts.MaxAge = 3600

		SetCookie(w, "test_cookie", "test_value", opts)

		cookies := w.Result().Cookies()
		assert.Len(t, cookies, 1)

		cookie := cookies[0]
		assert.Equal(t, "test_cookie", cookie.Name)
		assert.Equal(t, "test_value", cookie.Value)
		assert.Equal(t, 3600, cookie.MaxAge)
		assert.Equal(t, "/", cookie.Path)
		assert.True(t, cookie.HttpOnly)
		assert.False(t, cookie.Expires.IsZero())

		expectedExpires := time.Now().UTC().Add(time.Duration(opts.MaxAge) * time.Second)
		timeDiff := cookie.Expires.Sub(expectedExpires)
		if timeDiff < 0 {
			timeDiff = -timeDiff
		}

		assert.LessOrEqual(t, timeDiff, 5*time.Second)

		wNil := httptest.NewRecorder()
		SetCookie(wNil, "nil_opts", "value", nil)
		nilCookie := wNil.Result().Cookies()[0]
		assert.Equal(t, "nil_opts", nilCookie.Name)
		assert.True(t, nilCookie.HttpOnly)

		wSecure := httptest.NewRecorder()
		secureOpts := SecureCookieOptions()
		secureOpts.MaxAge = 86400
		SetCookie(wSecure, "secure_cookie", "secure_value", secureOpts)
		secureCookie := wSecure.Result().Cookies()[0]
		assert.True(t, secureCookie.Secure)
		assert.Equal(t, http.SameSiteStrictMode, secureCookie.SameSite)

		wCustom := httptest.NewRecorder()
		SetCookie(wCustom, "custom_cookie", "custom_value", &CookieOptions{
			MaxAge:   7200,
			Path:     "/api",
			Domain:   "example.com",
			Secure:   true,
			HTTPOnly: true,
			SameSite: http.SameSiteNoneMode,
		})
		customCookie := wCustom.Result().Cookies()[0]
		assert.Equal(t, 7200, customCookie.MaxAge)
		assert.Equal(t, "/api", customCookie.Path)
		assert.Equal(t, "example.com", customCookie.Domain)
		assert.Equal(t, http.SameSiteNoneMode, customCookie.SameSite)

		req := httptest.NewRequest("GET", "/", nil)
		req.AddCookie(&http.Cookie{Name: "test_cookie", Value: "test_value"})

		assert.Equal(t, "test_value", GetCookie(req, "test_cookie"))
		assert.Empty(t, GetCookie(req, "nonexistent_cookie"))
		assert.True(t, HasCookie(req, "test_cookie"))
		assert.False(t, HasCookie(req, "nonexistent_cookie"))

		wDelete := httptest.NewRecorder()
		DeleteCookie(wDelete, "test_cookie")
		deleteCookie := wDelete.Result().Cookies()[0]
		assert.Equal(t, "test_cookie", deleteCookie.Name)
		assert.Equal(t, -1, deleteCookie.MaxAge)
		assert.Empty(t, deleteCookie.Value)
	})
}
