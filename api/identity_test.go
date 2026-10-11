// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/clivern/cognit/db"
	"github.com/clivern/cognit/module"
	"github.com/clivern/cognit/pkg/util"

	"github.com/stretchr/testify/assert"
)

func TestUnitIdentityActions(t *testing.T) {
	agent := &db.Agent{Id: "agent-1", WorkspaceId: "ws-1", Name: "support-assistant"}
	identity := &db.AgentIdentity{
		Id:       "identity-1",
		AgentId:  "agent-1",
		Name:     "partner",
		Type:     db.AgentIdentityTypeBasicAuth,
		Config:   `{"username":"support","secretHash":"hash"}`,
		IsActive: true,
	}
	base := "/api/v1/workspaces/ws-1/agents/support-assistant/identities"

	t.Run("list returns identities without secrets", func(t *testing.T) {
		api := &API{Identity: module.NewIdentity(
			&fakeAgentCatalog{agent: agent},
			&fakeAgentIdentities{list: []*db.AgentIdentity{identity}},
			&fakeIdentityTokens{},
		)}
		w := httptest.NewRecorder()
		newIdentityRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodGet, base, nil))

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"username":"support"`)
		assert.NotContains(t, w.Body.String(), "hash")
	})

	t.Run("list returns 404 for unknown agent", func(t *testing.T) {
		api := &API{Identity: module.NewIdentity(&fakeAgentCatalog{}, &fakeAgentIdentities{}, &fakeIdentityTokens{})}
		w := httptest.NewRecorder()
		newIdentityRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodGet, base, nil))

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("list returns 500 when the repository fails", func(t *testing.T) {
		api := &API{Identity: module.NewIdentity(
			&fakeAgentCatalog{agent: agent},
			&fakeAgentIdentities{listErr: errors.New("boom")},
			&fakeIdentityTokens{},
		)}
		w := httptest.NewRecorder()
		newIdentityRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodGet, base, nil))

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("create returns 201 with the secret once", func(t *testing.T) {
		api := &API{Identity: module.NewIdentity(&fakeAgentCatalog{agent: agent}, &fakeAgentIdentities{}, &fakeIdentityTokens{})}
		w := httptest.NewRecorder()
		newIdentityRouter(api).ServeHTTP(w, httptest.NewRequest(
			http.MethodPost,
			base,
			strings.NewReader(`{"name":"default","type":"api_key"}`),
		))

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Body.String(), `"secret":"`)
	})

	t.Run("create returns 400 for an unknown type", func(t *testing.T) {
		api := &API{Identity: module.NewIdentity(&fakeAgentCatalog{agent: agent}, &fakeAgentIdentities{}, &fakeIdentityTokens{})}
		w := httptest.NewRecorder()
		newIdentityRouter(api).ServeHTTP(w, httptest.NewRequest(
			http.MethodPost,
			base,
			strings.NewReader(`{"name":"default","type":"oauth"}`),
		))

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("create returns 400 for basic auth without a username", func(t *testing.T) {
		api := &API{Identity: module.NewIdentity(&fakeAgentCatalog{agent: agent}, &fakeAgentIdentities{}, &fakeIdentityTokens{})}
		w := httptest.NewRecorder()
		newIdentityRouter(api).ServeHTTP(w, httptest.NewRequest(
			http.MethodPost,
			base,
			strings.NewReader(`{"name":"partner","type":"basic_auth"}`),
		))

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("create returns 409 for a duplicate name", func(t *testing.T) {
		api := &API{Identity: module.NewIdentity(
			&fakeAgentCatalog{agent: agent},
			&fakeAgentIdentities{byName: identity},
			&fakeIdentityTokens{},
		)}
		w := httptest.NewRecorder()
		newIdentityRouter(api).ServeHTTP(w, httptest.NewRequest(
			http.MethodPost,
			base,
			strings.NewReader(`{"name":"partner","type":"basic_auth","username":"support"}`),
		))

		assert.Equal(t, http.StatusConflict, w.Code)
	})

	t.Run("update returns 200 with the new state", func(t *testing.T) {
		api := &API{Identity: module.NewIdentity(
			&fakeAgentCatalog{agent: agent},
			&fakeAgentIdentities{identity: &db.AgentIdentity{
				Id:       "identity-1",
				AgentId:  "agent-1",
				Name:     "partner",
				Type:     db.AgentIdentityTypeBasicAuth,
				Config:   `{"username":"support","secretHash":"hash"}`,
				IsActive: true,
			}},
			&fakeIdentityTokens{},
		)}
		w := httptest.NewRecorder()
		newIdentityRouter(api).ServeHTTP(w, httptest.NewRequest(
			http.MethodPut,
			base+"/identity-1",
			strings.NewReader(`{"name":"partner","isActive":false}`),
		))

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"isActive":false`)
	})

	t.Run("update returns 404 for unknown identity", func(t *testing.T) {
		api := &API{Identity: module.NewIdentity(&fakeAgentCatalog{agent: agent}, &fakeAgentIdentities{}, &fakeIdentityTokens{})}
		w := httptest.NewRecorder()
		newIdentityRouter(api).ServeHTTP(w, httptest.NewRequest(
			http.MethodPut,
			base+"/identity-1",
			strings.NewReader(`{"name":"partner","isActive":false}`),
		))

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("delete returns 204", func(t *testing.T) {
		api := &API{Identity: module.NewIdentity(
			&fakeAgentCatalog{agent: agent},
			&fakeAgentIdentities{identity: identity},
			&fakeIdentityTokens{},
		)}
		w := httptest.NewRecorder()
		newIdentityRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodDelete, base+"/identity-1", nil))

		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("delete returns 404 for unknown identity", func(t *testing.T) {
		api := &API{Identity: module.NewIdentity(&fakeAgentCatalog{agent: agent}, &fakeAgentIdentities{}, &fakeIdentityTokens{})}
		w := httptest.NewRecorder()
		newIdentityRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodDelete, base+"/identity-1", nil))

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("token returns 400 for another grant type", func(t *testing.T) {
		api := &API{Identity: module.NewIdentity(&fakeAgentCatalog{}, &fakeAgentIdentities{}, &fakeIdentityTokens{})}
		req := httptest.NewRequest(http.MethodPost, "/api/v1/public/oauth/token", strings.NewReader("grant_type=password"))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		newIdentityRouter(api).ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), `"error":"unsupported_grant_type"`)
	})

	t.Run("token returns 401 for a wrong secret", func(t *testing.T) {
		hash, _ := util.HashPassword("right-secret")
		api := &API{Identity: module.NewIdentity(
			&fakeAgentCatalog{},
			&fakeAgentIdentities{identity: &db.AgentIdentity{
				Id:       "0b9f2a6e-3c1d-4f5a-8e7b-9c0d1e2f3a4b",
				Type:     db.AgentIdentityTypeClientCredentials,
				Config:   `{"secretHash":"` + hash + `","tokenTtl":300}`,
				IsActive: true,
			}},
			&fakeIdentityTokens{},
		)}
		form := url.Values{"grant_type": {"client_credentials"}, "client_id": {"0b9f2a6e-3c1d-4f5a-8e7b-9c0d1e2f3a4b"}, "client_secret": {"wrong"}}
		req := httptest.NewRequest(http.MethodPost, "/api/v1/public/oauth/token", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		newIdentityRouter(api).ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), `"error":"invalid_client"`)
	})

	t.Run("token returns a bearer token for basic auth client credentials", func(t *testing.T) {
		hash, _ := util.HashPassword("right-secret")
		api := &API{Identity: module.NewIdentity(
			&fakeAgentCatalog{},
			&fakeAgentIdentities{identity: &db.AgentIdentity{
				Id:       "0b9f2a6e-3c1d-4f5a-8e7b-9c0d1e2f3a4b",
				Type:     db.AgentIdentityTypeClientCredentials,
				Config:   `{"secretHash":"` + hash + `","tokenTtl":300}`,
				IsActive: true,
			}},
			&fakeIdentityTokens{},
		)}
		req := httptest.NewRequest(http.MethodPost, "/api/v1/public/oauth/token", strings.NewReader("grant_type=client_credentials"))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.SetBasicAuth("0b9f2a6e-3c1d-4f5a-8e7b-9c0d1e2f3a4b", "right-secret")
		w := httptest.NewRecorder()
		newIdentityRouter(api).ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"token_type":"Bearer"`)
		assert.Contains(t, w.Body.String(), `"expires_in":300`)
		assert.Equal(t, "no-store", w.Header().Get("Cache-Control"))
	})
}
