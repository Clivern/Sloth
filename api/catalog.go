// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package api

import (
	"net/http"

	"github.com/clivern/cognit/locale"
)

// ListCatalogSkillsAction lists the skills advertised in the workspace catalog.
func (a *API) ListCatalogSkillsAction(w http.ResponseWriter, r *http.Request) {
	// TODO: list skills across registered agent cards (GET /catalog/skills).
	a.WriteJSON(w, http.StatusNotImplemented, map[string]any{
		"errorMessage": locale.TR(r, "not_implemented"),
	})
}

// ListCatalogSkillAgentsAction lists the agents that advertise a skill.
func (a *API) ListCatalogSkillAgentsAction(w http.ResponseWriter, r *http.Request) {
	// TODO: list agents that expose a skill (GET /catalog/skills/{skill}/agents).
	a.WriteJSON(w, http.StatusNotImplemented, map[string]any{
		"errorMessage": locale.TR(r, "not_implemented"),
	})
}

// GetCatalogAgentCardAction returns an agent card from the catalog.
func (a *API) GetCatalogAgentCardAction(w http.ResponseWriter, r *http.Request) {
	// TODO: return the card of an agent from the catalog (GET /catalog/agents/{agentName}/card).
	a.WriteJSON(w, http.StatusNotImplemented, map[string]any{
		"errorMessage": locale.TR(r, "not_implemented"),
	})
}
