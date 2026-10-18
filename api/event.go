// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package api

import (
	"net/http"

	"github.com/clivern/sloth/db"
	"github.com/clivern/sloth/locale"
	"github.com/clivern/sloth/module"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

// ListEventsAction lists events in a workspace.
func (a *API) ListEventsAction(w http.ResponseWriter, r *http.Request) {
	wid := chi.URLParam(r, "workspaceId")
	if lo.IsEmpty(wid) {
		a.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"errorMessage": locale.TR(r, "invalid_workspace_id"),
		})
		return
	}

	limit, offset := a.ParsePagination(r)

	result, err := a.Event.ListEvents(db.Id(wid), limit, offset)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", wid).
			Msg("Failed to list events")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_list_events"),
		})

		return
	}

	a.WriteJSON(w, http.StatusOK, map[string]any{
		"events": result.Events,
		"_meta": map[string]any{
			"limit":  limit,
			"offset": offset,
			"total":  result.Total,
		},
	})
}

// GetEventAction returns one event by id.
func (a *API) GetEventAction(w http.ResponseWriter, r *http.Request) {
	wid := chi.URLParam(r, "workspaceId")
	if lo.IsEmpty(wid) {
		a.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"errorMessage": locale.TR(r, "invalid_workspace_id"),
		})
		return
	}

	eventId := chi.URLParam(r, "eventId")
	if lo.IsEmpty(eventId) {
		a.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"errorMessage": locale.TR(r, "invalid_event_id"),
		})
		return
	}

	event, err := a.Event.GetEvent(db.Id(wid), db.Id(eventId))
	if err != nil {
		if a.WriteError(w, err, []ErrorMapping{
			{module.ErrEventNotFound, http.StatusNotFound, locale.TR(r, "event_not_found")},
		}) {
			return
		}

		log.Error().
			Err(err).
			Str("workspaceId", wid).
			Str("eventId", eventId).
			Msg("Failed to get event")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_get_event"),
		})

		return
	}

	a.WriteJSON(w, http.StatusOK, event)
}
