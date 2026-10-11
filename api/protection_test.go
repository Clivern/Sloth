// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/clivern/cognit/db"
	"github.com/clivern/cognit/module"

	"github.com/stretchr/testify/assert"
)

func TestUnitProtectionActions(t *testing.T) {
	agent := &db.Agent{Id: "agent-1", WorkspaceId: "ws-1", Name: "support-assistant"}
	protection := &db.AgentProtection{
		Id:       "protection-1",
		AgentId:  "agent-1",
		Type:     db.AgentProtectionTypeBasicAuth,
		Config:   `{"username":"cognit","secretEncrypted":"cipher"}`,
		IsActive: true,
	}
	base := "/api/v1/workspaces/ws-1/agents/support-assistant/protection"

	t.Run("get returns the protection without its secret", func(t *testing.T) {
		api := &API{Protection: module.NewProtection(
			&fakeAgentCatalog{agent: agent},
			&fakeAgentProtections{protection: protection},
		)}
		w := httptest.NewRecorder()
		newProtectionRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodGet, base, nil))

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"username":"cognit"`)
		assert.NotContains(t, w.Body.String(), "cipher")
	})

	t.Run("get returns 404 when the agent has no protection", func(t *testing.T) {
		api := &API{Protection: module.NewProtection(&fakeAgentCatalog{agent: agent}, &fakeAgentProtections{})}
		w := httptest.NewRecorder()
		newProtectionRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodGet, base, nil))

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("get returns 404 for unknown agent", func(t *testing.T) {
		api := &API{Protection: module.NewProtection(&fakeAgentCatalog{}, &fakeAgentProtections{})}
		w := httptest.NewRecorder()
		newProtectionRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodGet, base, nil))

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("get returns 500 when the repository fails", func(t *testing.T) {
		api := &API{Protection: module.NewProtection(
			&fakeAgentCatalog{agent: agent},
			&fakeAgentProtections{getErr: errors.New("boom")},
		)}
		w := httptest.NewRecorder()
		newProtectionRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodGet, base, nil))

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("save returns 200 with a generated secret", func(t *testing.T) {
		api := &API{Protection: module.NewProtection(&fakeAgentCatalog{agent: agent}, &fakeAgentProtections{})}
		w := httptest.NewRecorder()
		newProtectionRouter(api).ServeHTTP(w, httptest.NewRequest(
			http.MethodPut,
			base,
			strings.NewReader(`{"type":"api_key"}`),
		))

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"secret":"`)
	})

	t.Run("save does not echo a provided secret", func(t *testing.T) {
		api := &API{Protection: module.NewProtection(&fakeAgentCatalog{agent: agent}, &fakeAgentProtections{})}
		w := httptest.NewRecorder()
		newProtectionRouter(api).ServeHTTP(w, httptest.NewRequest(
			http.MethodPut,
			base,
			strings.NewReader(`{"type":"api_key","secret":"agent-own-secret"}`),
		))

		assert.Equal(t, http.StatusOK, w.Code)
		assert.NotContains(t, w.Body.String(), "agent-own-secret")
	})

	t.Run("save returns 400 for an unknown type", func(t *testing.T) {
		api := &API{Protection: module.NewProtection(&fakeAgentCatalog{agent: agent}, &fakeAgentProtections{})}
		w := httptest.NewRecorder()
		newProtectionRouter(api).ServeHTTP(w, httptest.NewRequest(
			http.MethodPut,
			base,
			strings.NewReader(`{"type":"oauth"}`),
		))

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("save returns 400 for basic auth without a username", func(t *testing.T) {
		api := &API{Protection: module.NewProtection(&fakeAgentCatalog{agent: agent}, &fakeAgentProtections{})}
		w := httptest.NewRecorder()
		newProtectionRouter(api).ServeHTTP(w, httptest.NewRequest(
			http.MethodPut,
			base,
			strings.NewReader(`{"type":"basic_auth"}`),
		))

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("update status returns 200 with the new state", func(t *testing.T) {
		api := &API{Protection: module.NewProtection(
			&fakeAgentCatalog{agent: agent},
			&fakeAgentProtections{protection: &db.AgentProtection{
				Id:       "protection-1",
				AgentId:  "agent-1",
				Type:     db.AgentProtectionTypeBasicAuth,
				Config:   `{"username":"cognit","secretEncrypted":"cipher"}`,
				IsActive: true,
			}},
		)}
		w := httptest.NewRecorder()
		newProtectionRouter(api).ServeHTTP(w, httptest.NewRequest(
			http.MethodPut,
			base+"/status",
			strings.NewReader(`{"isActive":false}`),
		))

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"isActive":false`)
	})

	t.Run("update status returns 404 when the agent has no protection", func(t *testing.T) {
		api := &API{Protection: module.NewProtection(&fakeAgentCatalog{agent: agent}, &fakeAgentProtections{})}
		w := httptest.NewRecorder()
		newProtectionRouter(api).ServeHTTP(w, httptest.NewRequest(
			http.MethodPut,
			base+"/status",
			strings.NewReader(`{"isActive":false}`),
		))

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("delete returns 204", func(t *testing.T) {
		api := &API{Protection: module.NewProtection(
			&fakeAgentCatalog{agent: agent},
			&fakeAgentProtections{protection: protection},
		)}
		w := httptest.NewRecorder()
		newProtectionRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodDelete, base, nil))

		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("delete returns 404 when the agent has no protection", func(t *testing.T) {
		api := &API{Protection: module.NewProtection(&fakeAgentCatalog{agent: agent}, &fakeAgentProtections{})}
		w := httptest.NewRecorder()
		newProtectionRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodDelete, base, nil))

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}
