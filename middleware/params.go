// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package middleware

import (
	"net/http"

	"github.com/clivern/cognit/locale"
	"github.com/clivern/cognit/pkg/util"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/samber/lo"
)

// ValidateParams answers 400 for a malformed id route param before it reaches the database.
func ValidateParams(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		if rctx == nil || rctx.Routes == nil {
			next.ServeHTTP(w, r)
			return
		}

		// Params are only filled in as routing runs, so match the route up front to read them
		match := chi.NewRouteContext()
		if !rctx.Routes.Match(match, r.Method, r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// Only these params hold UUIDs; instanceId and checkId are names
		ids := []string{
			"workspaceId",
			"apiKeyId",
			"inviteId",
			"memberUserId",
			"keyId",
			"auditId",
			"identityId",
			"intentionId",
			"trafficId",
		}

		for i, key := range match.URLParams.Keys {
			if lo.Contains(ids, key) && uuid.Validate(match.URLParams.Values[i]) != nil {
				util.WriteJSON(w, http.StatusBadRequest, map[string]any{
					"errorMessage": locale.TR(r, "invalid_request"),
				})
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
