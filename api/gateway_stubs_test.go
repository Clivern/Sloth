// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

// TestUnitGatewayStubActions pins catalog, discover and watch handlers to 501 until implemented.
func TestUnitGatewayStubActions(t *testing.T) {
	api := &API{}
	router := chi.NewRouter()
	router.Get("/api/v1/workspaces/{workspaceId}/catalog/skills", api.ListCatalogSkillsAction)
	router.Get("/api/v1/workspaces/{workspaceId}/catalog/skills/{skill}/agents", api.ListCatalogSkillAgentsAction)
	router.Get("/api/v1/workspaces/{workspaceId}/catalog/agents/{agentName}/card", api.GetCatalogAgentCardAction)
	router.Get("/api/v1/workspaces/{workspaceId}/discover", api.DiscoverAgentsAction)
	router.Post("/api/v1/workspaces/{workspaceId}/discover/resolve", api.ResolveDiscoverAction)
	router.Get("/api/v1/workspaces/{workspaceId}/watch/catalog", api.WatchCatalogAction)

	cases := []struct {
		name   string
		method string
		path   string
	}{
		{name: "list catalog skills", method: http.MethodGet, path: "/api/v1/workspaces/ws-1/catalog/skills"},
		{name: "list catalog skill agents", method: http.MethodGet, path: "/api/v1/workspaces/ws-1/catalog/skills/lookup/agents"},
		{name: "get catalog agent card", method: http.MethodGet, path: "/api/v1/workspaces/ws-1/catalog/agents/user-directory/card"},
		{name: "discover agents", method: http.MethodGet, path: "/api/v1/workspaces/ws-1/discover"},
		{name: "resolve discover", method: http.MethodPost, path: "/api/v1/workspaces/ws-1/discover/resolve"},
		{name: "watch catalog", method: http.MethodGet, path: "/api/v1/workspaces/ws-1/watch/catalog"},
	}

	for _, tc := range cases {
		t.Run(tc.name+" returns 501", func(t *testing.T) {
			w := httptest.NewRecorder()
			router.ServeHTTP(w, httptest.NewRequest(tc.method, tc.path, strings.NewReader("{}")))
			assert.Equal(t, http.StatusNotImplemented, w.Code)
			assert.Contains(t, w.Body.String(), "not_implemented")
		})
	}
}
