// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package api

import (
	"net/http"

	"github.com/clivern/cognit/locale"
)

// GetAgentCardAction returns the agent card at the well-known path.
func (a *API) GetAgentCardAction(w http.ResponseWriter, r *http.Request) {
	// TODO: serve the agent card at /.well-known/agent-card.json with ETag support.
	a.WriteJSON(w, http.StatusNotImplemented, map[string]any{
		"errorMessage": locale.TR(r, "not_implemented"),
	})
}
