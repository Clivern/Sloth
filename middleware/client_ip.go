// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package middleware

import (
	"net/http"
	"strings"

	cmid "github.com/go-chi/chi/v5/middleware"
	"github.com/samber/lo"
)

// ClientIP resolves the client IP from X-Forwarded-For when trustedProxies is set, otherwise from the TCP peer.
func ClientIP(trustedProxies string) func(http.Handler) http.Handler {
	prefixes := lo.Compact(lo.Map(strings.Split(trustedProxies, ","), func(p string, _ int) string {
		return strings.TrimSpace(p)
	}))

	if len(prefixes) == 0 {
		return cmid.ClientIPFromRemoteAddr
	}

	return cmid.ClientIPFromXFF(prefixes...)
}
