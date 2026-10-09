// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package api

import (
	"net/http"

	"github.com/clivern/cognit/locale"
)

// DiscoverAgentsAction finds agents by skill, tags, and passing status.
func (a *API) DiscoverAgentsAction(w http.ResponseWriter, r *http.Request) {
	// TODO: query agents by skill, tags, passing, and limit (GET /discover).
	a.WriteJSON(w, http.StatusNotImplemented, map[string]any{
		"errorMessage": locale.TR(r, "not_implemented"),
	})
}

// ResolveDiscoverAction resolves a discovery request to matching agents.
func (a *API) ResolveDiscoverAction(w http.ResponseWriter, r *http.Request) {
	// TODO: resolve a skill or tag request to ranked agent cards (POST /discover/resolve).
	a.WriteJSON(w, http.StatusNotImplemented, map[string]any{
		"errorMessage": locale.TR(r, "not_implemented"),
	})
}
