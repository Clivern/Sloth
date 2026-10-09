// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package middleware

import (
	"context"
	"net"
	"net/http"
	"strings"

	"github.com/clivern/cognit/module"

	"github.com/samber/lo"
)

// AppContext attaches request metadata as module.AppCtx on the request context.
func AppContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ac := module.NewAppCtx(
			GetRequestId(r.Context()),
			r.UserAgent(),
			GetClientIP(r),
			ParseFeatures(r.Header.Get("X-Features")),
		)

		ctx := module.WithAppCtx(r.Context(), ac)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetAppCtx retrieves AppCtx from the request context.
func GetAppCtx(ctx context.Context) (*module.AppCtx, bool) {
	return module.AppCtxFrom(ctx)
}

// GetClientIP retrieves the client IP from the request.
func GetClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); lo.IsNotEmpty(xff) {
		if i := strings.Index(xff, ","); i >= 0 {
			return strings.TrimSpace(xff[:i])
		}
		return strings.TrimSpace(xff)
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); lo.IsNotEmpty(xri) {
		return strings.TrimSpace(xri)
	}

	// Return remote address as fallback
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return host
}

// ParseFeatures parses the features from the header.
func ParseFeatures(header string) map[string]bool {
	features := map[string]bool{}
	if lo.IsEmpty(header) {
		return features
	}

	for part := range strings.SplitSeq(header, ",") {
		name := strings.TrimSpace(part)
		if lo.IsNotEmpty(name) {
			features[name] = true
		}
	}

	return features
}
