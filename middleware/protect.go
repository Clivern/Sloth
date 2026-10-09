// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/clivern/cognit/db"
	"github.com/clivern/cognit/locale"
	"github.com/clivern/cognit/module"
	"github.com/clivern/cognit/pkg/util"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

// Config configures Protect middleware for a route or route group.
type Config struct {
	User      bool     // require user principal (reject access keys)
	Roles     []string // platform roles; implies User
	Workspace bool     // load {workspaceId} into context
	Perm      string   // workspace permission check
}

// Protect returns middleware that applies user auth, role, workspace load, and permission checks.
func Protect(cfg Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// User / roles — Roles implies User so access keys are rejected
			requireUser := cfg.User || len(cfg.Roles) > 0

			if requireUser {
				user, ok := GetUserFromContext(r.Context())
				if !ok || user == nil {
					log.Info().
						Str("path", r.URL.Path).
						Msg("User required: principal not in context")
					util.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
						"errorMessage": locale.TR(r, "not_authenticated"),
					})
					return
				}

				if !user.IsActive {
					log.Info().
						Str("path", r.URL.Path).
						Str("userId", user.Id.String()).
						Msg("Inactive user blocked from protected route")
					util.WriteJSON(w, http.StatusForbidden, map[string]interface{}{
						"errorMessage": locale.TR(r, "account_not_active"),
					})
					return
				}

				if len(cfg.Roles) > 0 && !lo.Contains(cfg.Roles, user.Role) {
					log.Info().
						Str("path", r.URL.Path).
						Str("userId", user.Id.String()).
						Str("userRole", user.Role).
						Strs("allowedRoles", cfg.Roles).
						Msg("Insufficient role for route")
					util.WriteJSON(w, http.StatusForbidden, map[string]interface{}{
						"errorMessage": locale.TR(r, "insufficient_permissions"),
					})
					return
				}
			}

			// Workspace — load {workspaceId} into context for handlers / perm check
			if cfg.Workspace {
				wid := chi.URLParam(r, "workspaceId")
				if lo.IsEmpty(wid) {
					log.Error().
						Str("path", r.URL.Path).
						Msg("Protect Workspace missing workspace id param")
					util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
						"errorMessage": locale.TR(r, "invalid_route_configuration"),
					})
					return
				}

				workspace, err := db.NewWorkspaceRepository(db.GetDB()).GetById(db.Id(wid))
				if err != nil {
					log.Error().
						Err(err).
						Str("workspaceId", wid).
						Msg("Failed to load workspace")
					util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
						"errorMessage": locale.TR(r, "failed_load_workspace"),
					})
					return
				}
				if workspace == nil {
					util.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
						"errorMessage": locale.TR(r, "workspace_not_found"),
					})
					return
				}

				r = r.WithContext(context.WithValue(r.Context(), ContextKeyWorkspace, workspace))
			}

			// Perm — workspace permission; allows access keys unless User/Roles ran above
			if lo.IsNotEmpty(cfg.Perm) {
				perm := module.NewPerm(db.GetDB())

				// Prefer workspace already loaded by Workspace step or a group Protect
				if workspace, ok := GetWorkspaceFromContext(r.Context()); ok {
					perm = perm.WithWorkspace(workspace)
				} else {
					wid := chi.URLParam(r, "workspaceId")
					if lo.IsEmpty(wid) {
						log.Error().
							Str("path", r.URL.Path).
							Str("permission", cfg.Perm).
							Msg("Workspace permission check missing workspace id param")
						util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
							"errorMessage": locale.TR(r, "invalid_route_configuration"),
						})
						return
					}

					perm = perm.WithWorkspaceId(db.Id(wid))
				}

				if user, ok := GetUserFromContext(r.Context()); ok && user != nil {
					if !user.IsActive {
						log.Info().
							Str("path", r.URL.Path).
							Str("userId", user.Id.String()).
							Msg("Inactive user blocked from workspace route")
						util.WriteJSON(w, http.StatusForbidden, map[string]interface{}{
							"errorMessage": locale.TR(r, "account_not_active"),
						})
						return
					}

					perm = perm.WithUser(user)
				} else if key, ok := GetAccessKeyFromContext(r.Context()); ok && key != nil {
					perm = perm.WithAccessKey(key)
				} else {
					log.Info().
						Str("path", r.URL.Path).
						Msg("Workspace permission check failed: no principal in context")
					util.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
						"errorMessage": locale.TR(r, "not_authenticated"),
					})
					return
				}

				allowed, err := perm.Can(cfg.Perm)
				if err != nil {
					if errors.Is(err, module.ErrPermissionWorkspaceMissing) {
						util.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
							"errorMessage": locale.TR(r, "workspace_not_found"),
						})
						return
					}

					log.Error().
						Err(err).
						Str("path", r.URL.Path).
						Str("permission", cfg.Perm).
						Msg("Failed to check workspace permission")
					util.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
						"errorMessage": locale.TR(r, "failed_check_access"),
					})
					return
				}
				if !allowed {
					log.Info().
						Str("path", r.URL.Path).
						Str("permission", cfg.Perm).
						Msg("Workspace permission denied")
					util.WriteJSON(w, http.StatusForbidden, map[string]interface{}{
						"errorMessage": locale.TR(r, "insufficient_permissions"),
					})
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
