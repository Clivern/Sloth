// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package api

import (
	"net/http"

	"github.com/clivern/cognit/db"
	"github.com/clivern/cognit/locale"
	"github.com/clivern/cognit/module"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

// GetWorkspaceStatsAction returns dashboard metrics for a workspace.
func (a *API) GetWorkspaceStatsAction(w http.ResponseWriter, r *http.Request) {
	wid := chi.URLParam(r, "workspaceId")
	if lo.IsEmpty(wid) {
		a.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"errorMessage": locale.TR(r, "invalid_workspace_id"),
		})
		return
	}

	stats, err := a.Stats.GetWorkspaceStats(db.Id(wid))
	if err != nil {
		if a.WriteError(w, r, err, []ErrorMapping{
			{module.ErrWorkspaceNotFound, http.StatusNotFound, "workspace_not_found"},
		}) {
			return
		}

		log.Error().
			Err(err).
			Str("workspaceId", wid).
			Msg("Failed to get workspace stats")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_get_workspace_stats"),
		})

		return
	}

	a.WriteJSON(w, http.StatusOK, stats)
}
