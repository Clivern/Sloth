// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package api

import (
	"errors"
	"net/http"

	"github.com/clivern/cognit/locale"
	"github.com/clivern/cognit/module"

	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

// GetMeAction returns the authenticated session, API key or access key principal.
func (a *API) GetMeAction(w http.ResponseWriter, r *http.Request) {
	apiKey := r.Header.Get("X-API-Key")
	accessKey := r.Header.Get("X-Access-Key")

	// Fall back to the session user resolved by the auth middleware
	if lo.IsEmpty(apiKey) && lo.IsEmpty(accessKey) {
		if user, ok := a.GetUser(r); ok && user != nil {
			a.WriteJSON(w, http.StatusOK, a.Me.GetBySession(user))
			return
		}

		a.WriteJSON(w, http.StatusForbidden, map[string]any{
			"errorMessage": locale.TR(r, "me_requires_key_header"),
		})
		return
	}

	// Check if the request is for an API key
	if lo.IsNotEmpty(apiKey) {
		me, err := a.Me.GetByAPIKey(apiKey)
		if err != nil {
			switch {
			case errors.Is(err, module.ErrAPIKeyNotFound):
				a.WriteJSON(w, http.StatusUnauthorized, map[string]any{
					"errorMessage": locale.TR(r, "invalid_api_key"),
				})
			default:
				log.Error().
					Err(err).
					Msg("Failed to resolve API key for /me")
				a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
					"errorMessage": locale.TR(r, "failed_get_me"),
				})
			}

			return
		}

		a.WriteJSON(w, http.StatusOK, me)
		return
	}

	// Check if the request is for an access key
	me, err := a.Me.GetByAccessKey(accessKey)
	if err != nil {
		switch {
		case errors.Is(err, module.ErrAccessKeyNotFound):
			a.WriteJSON(w, http.StatusUnauthorized, map[string]any{
				"errorMessage": locale.TR(r, "invalid_access_key"),
			})
		default:
			log.Error().
				Err(err).
				Msg("Failed to resolve access key for /me")
			a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
				"errorMessage": locale.TR(r, "failed_get_me"),
			})
		}

		return
	}

	a.WriteJSON(w, http.StatusOK, me)
}
