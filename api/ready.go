// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package api

import (
	"net/http"

	"github.com/clivern/sloth/db"

	"github.com/rs/zerolog/log"
)

// ReadyAction returns whether the app is ready (e.g. DB reachable).
func (a *API) ReadyAction(w http.ResponseWriter, _ *http.Request) {
	log.Debug().
		Msg("Readiness check")

	err := db.GetDB().Ping()
	if err != nil {
		log.Error().
			Err(err).
			Msg("Database ping failed during readiness check")

		a.WriteJSON(w, http.StatusServiceUnavailable, map[string]any{
			"status": "not_ok",
		})
		return
	}

	log.Debug().
		Msg("Readiness check passed")

	a.WriteJSON(w, http.StatusOK, map[string]any{
		"status": "ok",
	})
}
