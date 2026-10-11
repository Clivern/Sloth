// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/clivern/cognit/db"
	"github.com/clivern/cognit/module"

	"github.com/go-chi/chi/v5"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestUnitAgentCardEndpoint(t *testing.T) {
	workspaceId := db.Id("")
	card := `{"name":"User Directory","version":"1.0.0"}`
	agent := &db.Agent{
		Name:         "user-directory",
		Card:         lo.ToPtr(card),
		CardChecksum: lo.ToPtr("9f2c"),
		Version:      "1.0.0",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	api := &API{
		Agent: module.NewAgent(
			&fakeAgentCatalog{agent: agent},
			&fakeInstanceRecords{},
			nil,
			&fakeWorkspaceRecords{workspace: &db.Workspace{Id: workspaceId}},
		),
	}

	router := chi.NewRouter()
	router.Get(
		"/api/v1/workspaces/{workspaceId}/agents/{agentName}/.well-known/agent-card.json",
		api.GetAgentCardAction,
	)

	path := "/api/v1/workspaces/ws-1/agents/user-directory/.well-known/agent-card.json"

	t.Run("returns card with ETag", func(t *testing.T) {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, `"9f2c"`, w.Header().Get("ETag"))
		assert.Equal(t, "no-cache", w.Header().Get("Cache-Control"))
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		assert.Equal(t, card, w.Body.String())
	})

	t.Run("returns 304 when If-None-Match matches", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.Header.Set("If-None-Match", `"9f2c"`)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotModified, w.Code)
		assert.Empty(t, w.Body.String())
		assert.Equal(t, `"9f2c"`, w.Header().Get("ETag"))
	})

	t.Run("returns 200 when If-None-Match is stale", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.Header.Set("If-None-Match", `"old"`)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, card, w.Body.String())
	})

	t.Run("returns 404 for unknown agent", func(t *testing.T) {
		w := httptest.NewRecorder()
		router.ServeHTTP(
			w,
			httptest.NewRequest(
				http.MethodGet,
				"/api/v1/workspaces/ws-1/agents/missing/.well-known/agent-card.json",
				nil,
			),
		)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("returns 404 while the card has not been fetched", func(t *testing.T) {
		pending := &API{
			Agent: module.NewAgent(
				&fakeAgentCatalog{agent: &db.Agent{Name: "user-directory", Version: "1.0.0"}},
				&fakeInstanceRecords{},
				nil,
				&fakeWorkspaceRecords{workspace: &db.Workspace{Id: workspaceId}},
			),
		}
		pendingRouter := chi.NewRouter()
		pendingRouter.Get(
			"/api/v1/workspaces/{workspaceId}/agents/{agentName}/.well-known/agent-card.json",
			pending.GetAgentCardAction,
		)

		w := httptest.NewRecorder()
		pendingRouter.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "agent_card_pending")
		assert.Empty(t, w.Header().Get("ETag"))
	})
}
