// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package middleware

import (
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"
)

// RequestSizeLimit creates a middleware that limits the size of request bodies
func RequestSizeLimit(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.HasPrefix(r.URL.Path, "/api/v1/") {
				if !IsAssetPath(r.URL.Path) {
					log.Info().
						Str("path", r.URL.Path).
						Msg("Skipping request size limit for non-API route")
				}

				next.ServeHTTP(w, r)
				return
			}

			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)

			next.ServeHTTP(w, r)
		})
	}
}
