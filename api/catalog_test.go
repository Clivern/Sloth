// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/clivern/sloth/db"
	"github.com/clivern/sloth/module"

	"github.com/go-chi/chi/v5"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitListCatalogSkillsAction(t *testing.T) {
	caller := &db.Agent{Id: "agent-0", Name: "support-assistant"}
	directory := &db.Agent{
		Id:   "agent-1",
		Name: "user-directory",
		Card: lo.ToPtr(`{"skills":[{"id":"user.profile.get","name":"Get user profile","tags":["profile"]},{"id":"user.list","tags":["admin"]}]}`),
	}

	newRouter := func(intentions *fakeIntentions) *chi.Mux {
		api := &API{
			Catalog: module.NewCatalog(
				&fakeAgentCatalog{list: []*db.Agent{caller, directory}, total: 2},
				&fakeInstanceRecords{},
				intentions,
			),
		}

		router := chi.NewRouter()
		router.With(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				next.ServeHTTP(w, r.WithContext(module.WithAgent(r.Context(), caller)))
			})
		}).Get("/api/v1/workspaces/{workspaceId}/agents/catalog/skills", api.ListCatalogSkillsAction)

		return router
	}

	t.Run("returns the callable skills filtered by tag with pagination meta", func(t *testing.T) {
		router := newRouter(&fakeIntentions{match: &db.Intention{Action: db.IntentionActionAllow}})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1/agents/catalog/skills?tags=profile,ops", nil))
		require.Equal(t, http.StatusOK, w.Code)

		var body struct {
			Skills []module.CatalogSkill `json:"skills"`
			Meta   map[string]int        `json:"_meta"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		require.Len(t, body.Skills, 1)
		assert.Equal(t, "user.profile.get", body.Skills[0].Id)
		assert.Equal(t, 1, body.Skills[0].Agents)
		assert.Equal(t, 1, body.Meta["total"])
	})

	t.Run("returns an empty list when no intention allows a call", func(t *testing.T) {
		router := newRouter(&fakeIntentions{})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1/agents/catalog/skills", nil))
		require.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, `{"skills":[],"_meta":{"limit":50,"offset":0,"total":0}}`, w.Body.String())
	})

	t.Run("returns 500 when intentions cannot be read", func(t *testing.T) {
		router := newRouter(&fakeIntentions{matchErr: errors.New("db down")})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1/agents/catalog/skills", nil))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "failed_catalog_request")
	})
}
