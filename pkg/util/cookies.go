// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package util

import (
	"net/http"
	"time"
)

// CookieOptions holds configuration for cookie creation.
type CookieOptions struct {
	// MaxAge is the maximum age of the cookie in seconds.
	// If zero, the cookie is deleted when the browser closes (session cookie).
	// If negative, the cookie is deleted immediately.
	MaxAge int

	// Path specifies the cookie path. Defaults to "/".
	Path string

	// Domain specifies the cookie domain.
	Domain string

	// Secure indicates if the cookie should only be sent over HTTPS.
	Secure bool

	// HTTPOnly makes the cookie inaccessible to JavaScript.
	HTTPOnly bool

	// SameSite controls the cookie's SameSite attribute.
	// Options: SameSiteDefaultMode, SameSiteLaxMode, SameSiteStrictMode, SameSiteNoneMode.
	SameSite http.SameSite
}

// DefaultCookieOptions returns sensible default options for cookies.
func DefaultCookieOptions() *CookieOptions {
	return &CookieOptions{
		Path:     "/",
		HTTPOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
	}
}

// SecureCookieOptions returns secure cookie options for production use.
func SecureCookieOptions() *CookieOptions {
	return &CookieOptions{
		Path:     "/",
		HTTPOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}
}

// SetCookie sets a cookie with the given name, value, and options.
func SetCookie(w http.ResponseWriter, name, value string, options *CookieOptions) {
	if options == nil {
		options = DefaultCookieOptions()
	}

	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     options.Path,
		Domain:   options.Domain,
		MaxAge:   options.MaxAge,
		Secure:   options.Secure,
		HttpOnly: options.HTTPOnly,
		SameSite: options.SameSite,
	}
	if options.MaxAge > 0 {
		cookie.Expires = time.Now().UTC().Add(time.Duration(options.MaxAge) * time.Second)
	}

	http.SetCookie(w, cookie)
}

// GetCookie retrieves a cookie value by name.
func GetCookie(r *http.Request, name string) string {
	cookie, err := r.Cookie(name)
	if err != nil {
		return ""
	}

	return cookie.Value
}

// HasCookie checks if a cookie with the given name exists.
func HasCookie(r *http.Request, name string) bool {
	_, err := r.Cookie(name)

	return err == nil
}

// DeleteCookie deletes a cookie by setting its MaxAge to -1.
func DeleteCookie(w http.ResponseWriter, name string) {
	cookie := &http.Cookie{
		Name:   name,
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(w, cookie)
}
