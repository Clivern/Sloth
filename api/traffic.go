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

// ListTrafficAction lists recorded gateway calls for a workspace.
func (a *API) ListTrafficAction(w http.ResponseWriter, r *http.Request) {
	wid := chi.URLParam(r, "workspaceId")
	if lo.IsEmpty(wid) {
		a.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"errorMessage": locale.TR(r, "invalid_workspace_id"),
		})
		return
	}

	limit, offset := a.ParsePagination(r)

	result, err := a.Traffic.ListTraffic(db.Id(wid), limit, offset)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", wid).
			Msg("Failed to list traffic")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_traffic_request"),
		})

		return
	}

	a.WriteJSON(w, http.StatusOK, map[string]any{
		"calls": result.Calls,
		"_meta": map[string]any{
			"limit":  limit,
			"offset": offset,
			"total":  result.Total,
		},
	})
}

// GetTrafficAction returns one recorded gateway call.
func (a *API) GetTrafficAction(w http.ResponseWriter, r *http.Request) {
	wid := chi.URLParam(r, "workspaceId")
	if lo.IsEmpty(wid) {
		a.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"errorMessage": locale.TR(r, "invalid_workspace_id"),
		})
		return
	}

	trafficId := chi.URLParam(r, "trafficId")
	if lo.IsEmpty(trafficId) {
		a.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"errorMessage": locale.TR(r, "invalid_traffic_id"),
		})
		return
	}

	call, err := a.Traffic.GetTraffic(db.Id(wid), db.Id(trafficId))
	if err != nil {
		if a.WriteError(w, err, []ErrorMapping{
			{module.ErrTrafficNotFound, http.StatusNotFound, locale.TR(r, "traffic_not_found")},
		}) {
			return
		}

		log.Error().
			Err(err).
			Str("workspaceId", wid).
			Str("trafficId", trafficId).
			Msg("Failed to get traffic")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_traffic_request"),
		})

		return
	}

	a.WriteJSON(w, http.StatusOK, call)
}
