// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/clivern/sloth/db"
	"github.com/clivern/sloth/locale"
	"github.com/clivern/sloth/module"
	"github.com/clivern/sloth/pkg/util"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

// AgentAuth authenticates the calling agent by its identity and stores it on the request context.
func AgentAuth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			wid := chi.URLParam(r, "workspaceId")

			username, password, _ := r.BasicAuth()
			token, bearer := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")

			identity := module.NewIdentity(
				db.NewAgentRepository(db.GetDB()),
				db.NewAgentIdentityRepository(db.GetDB()),
				db.NewAgentIdentityTokenRepository(db.GetDB()),
			)

			agent, err := identity.Authenticate(db.Id(wid), &module.AgentCredential{
				SourceAgent: r.Header.Get("X-Source-Agent"),
				APIKey:      r.Header.Get("X-API-Key"),
				Username:    username,
				Password:    password,
				Token:       lo.Ternary(bearer, token, ""),
			})
			if errors.Is(err, module.ErrInvalidAgentIdentity) {
				util.WriteJSON(w, http.StatusUnauthorized, map[string]any{
					"errorMessage": locale.TR(r, "invalid_agent_identity"),
				})
				return
			}

			if err != nil {
				log.Error().
					Err(err).
					Str("workspaceId", wid).
					Msg("Failed to authenticate agent")
				util.WriteJSON(w, http.StatusInternalServerError, map[string]any{
					"errorMessage": locale.TR(r, "failed_check_access"),
				})
				return
			}

			next.ServeHTTP(w, r.WithContext(module.WithAgent(r.Context(), agent)))
		})
	}
}
