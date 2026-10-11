// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUnitRateLimit(t *testing.T) {
	ok := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	t.Run("RateLimit should reject requests over the limit with 429", func(t *testing.T) {
		handler := ClientIP("")(RateLimit(2, time.Minute)(ok))

		codes := []int{}
		for range 3 {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/public/oauth/token", nil)
			req.RemoteAddr = "10.0.0.1:1234"
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			codes = append(codes, w.Code)
		}

		assert.Equal(t, []int{http.StatusOK, http.StatusOK, http.StatusTooManyRequests}, codes)
	})

	t.Run("RateLimit should track each client IP separately", func(t *testing.T) {
		handler := ClientIP("")(RateLimit(1, time.Minute)(ok))

		for _, addr := range []string{"10.0.0.1:1234", "10.0.0.2:1234"} {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/public/oauth/token", nil)
			req.RemoteAddr = addr
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code)
		}
	})

	t.Run("RateLimit should ignore spoofed X-Forwarded-For without trusted proxies", func(t *testing.T) {
		handler := ClientIP("")(RateLimit(1, time.Minute)(ok))

		codes := []int{}
		for _, spoofed := range []string{"1.1.1.1", "2.2.2.2"} {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/public/oauth/token", nil)
			req.RemoteAddr = "10.0.0.1:1234"
			req.Header.Set("X-Forwarded-For", spoofed)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			codes = append(codes, w.Code)
		}

		assert.Equal(t, []int{http.StatusOK, http.StatusTooManyRequests}, codes)
	})

	t.Run("RateLimit should key by X-Forwarded-For client behind a trusted proxy", func(t *testing.T) {
		handler := ClientIP("10.0.0.0/8")(RateLimit(1, time.Minute)(ok))

		for _, client := range []string{"1.1.1.1", "2.2.2.2"} {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/public/oauth/token", nil)
			req.RemoteAddr = "10.0.0.1:1234"
			req.Header.Set("X-Forwarded-For", client)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code)
		}
	})
}
