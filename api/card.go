// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package api

import (
	"fmt"
	"net/http"

	"github.com/clivern/cognit/db"
	"github.com/clivern/cognit/locale"
	"github.com/clivern/cognit/module"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

// GetAgentCardAction returns the agent card at the well-known path.
func (a *API) GetAgentCardAction(w http.ResponseWriter, r *http.Request) {
	wid := chi.URLParam(r, "workspaceId")
	if lo.IsEmpty(wid) {
		a.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"errorMessage": locale.TR(r, "invalid_workspace_id"),
		})
		return
	}

	agentName := chi.URLParam(r, "agentName")
	if lo.IsEmpty(agentName) {
		a.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"errorMessage": locale.TR(r, "invalid_agent_name"),
		})
		return
	}

	agent, err := a.Agent.GetAgent(db.Id(wid), agentName)
	if err != nil {
		if a.WriteError(w, err, []ErrorMapping{
			{module.ErrAgentNotFound, http.StatusNotFound, locale.TR(r, "agent_not_found")},
		}) {
			return
		}

		log.Error().
			Err(err).
			Str("workspaceId", wid).
			Str("agent", agentName).
			Msg("Failed to get agent card")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_agent_request"),
		})

		return
	}

	if agent.Card == nil {
		a.WriteJSON(w, http.StatusNotFound, map[string]any{
			"errorMessage": locale.TR(r, "agent_card_pending"),
		})
		return
	}

	// The checksum is a SHA-256 of the compacted card, so it works as a strong ETag.
	etag := fmt.Sprintf(`"%s"`, agent.CardChecksum)

	w.Header().Set("ETag", etag)
	w.Header().Set("Cache-Control", "no-cache")

	if match := r.Header.Get("If-None-Match"); match == etag || match == "W/"+etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(agent.Card)
}
