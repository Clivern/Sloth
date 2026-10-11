// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package middleware

import (
	"net"
	"net/http"
	"time"

	"github.com/clivern/cognit/locale"
	"github.com/clivern/cognit/pkg/util"

	cmid "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"github.com/rs/zerolog/log"
)

// RateLimit returns middleware that allows at most limit requests per client IP and endpoint within window.
func RateLimit(limit int, window time.Duration) func(http.Handler) http.Handler {
	return httprate.LimitBy(
		limit,
		window,
		httprate.JoinKeys(
			func(r *http.Request) (string, error) {
				ip := cmid.GetClientIP(r.Context())
				if ip == "" {
					// Fall back to the TCP peer so an unresolved IP never collapses into one shared bucket
					host, _, err := net.SplitHostPort(r.RemoteAddr)
					if err != nil {
						host = r.RemoteAddr
					}
					ip = host
				}
				return httprate.CanonicalizeIP(ip), nil
			},
			httprate.KeyByEndpoint,
		),
		httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
			log.Warn().
				Str("path", r.URL.Path).
				Str("clientIp", cmid.GetClientIP(r.Context())).
				Msg("Request rejected by rate limit")
			util.WriteJSON(w, http.StatusTooManyRequests, map[string]interface{}{
				"errorMessage": locale.TR(r, "too_many_requests"),
			})
		}),
	)
}
