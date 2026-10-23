// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/clivern/sloth/db"
	"github.com/clivern/sloth/module"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitListCatalogSkillsAction(t *testing.T) {
	caller := &db.Agent{Id: "agent-0", Name: "support-assistant"}

	newRouter := func(skills *fakeAgentSkills) *chi.Mux {
		api := &API{Catalog: module.NewCatalog(skills)}

		router := chi.NewRouter()
		router.With(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				next.ServeHTTP(w, r.WithContext(module.WithAgent(r.Context(), caller)))
			})
		}).Get("/api/v1/workspaces/{workspaceId}/agents/catalog/skills", api.ListCatalogSkillsAction)

		return router
	}

	t.Run("returns the callable skills with pagination meta", func(t *testing.T) {
		router := newRouter(&fakeAgentSkills{
			skills: []*db.CallableSkill{
				{
					SkillId:     "user.profile.get",
					Name:        "Get user profile",
					Tags:        []string{"profile"},
					InputModes:  []string{"application/json"},
					OutputModes: []string{"application/json"},
					Agents:      1,
				},
			},
			total: 1,
		})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(
			http.MethodGet,
			"/api/v1/workspaces/ws-1/agents/catalog/skills?tags=profile&limit=20",
			nil,
		).WithContext(module.WithAgent(context.Background(), caller)))
		require.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, `{
			"skills": [{
				"id": "user.profile.get",
				"name": "Get user profile",
				"description": "",
				"tags": ["profile"],
				"inputModes": ["application/json"],
				"outputModes": ["application/json"],
				"agents": 1
			}],
			"_meta": {"limit": 20, "offset": 0, "total": 1}
		}`, w.Body.String())
	})

	t.Run("returns an empty list when nothing is callable", func(t *testing.T) {
		router := newRouter(&fakeAgentSkills{skills: []*db.CallableSkill{}})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1/agents/catalog/skills", nil))
		require.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, `{"skills":[],"_meta":{"limit":50,"offset":0,"total":0}}`, w.Body.String())
	})

	t.Run("returns 500 when the catalog cannot be read", func(t *testing.T) {
		router := newRouter(&fakeAgentSkills{err: errors.New("db down")})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1/agents/catalog/skills", nil))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "failed_catalog_request")
	})
}
