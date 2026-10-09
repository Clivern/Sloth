// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package api

import (
	"net/http"

	"github.com/clivern/cognit/locale"
)

// WatchCatalogAction streams catalog changes since an index.
func (a *API) WatchCatalogAction(w http.ResponseWriter, r *http.Request) {
	// TODO: stream catalog changes after the given index (GET watch/catalog).
	a.WriteJSON(w, http.StatusNotImplemented, map[string]any{
		"errorMessage": locale.TR(r, "not_implemented"),
	})
}
