// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/clivern/sloth/db"

	"github.com/stretchr/testify/assert"
)

func TestUnitIntegrationActions(t *testing.T) {
	stored := &db.Integration{
		Id:          "int-1",
		WorkspaceId: "ws-1",
		Type:        db.IntegrationTypeWebhook,
		Name:        "audit hook",
		Config:      new(`{"url":"https://example.test/hook","secretEncrypted":"sealed"}`),
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	body := `{"type":"webhook","name":"audit hook","url":"https://example.test/hook","events":["audit.log","agent.health.changed"],"isActive":true}`

	t.Run("list returns integrations with total", func(t *testing.T) {
		api := newIntegrationAPI(&fakeIntegrations{list: []*db.Integration{stored}, total: 1})
		w := httptest.NewRecorder()
		newIntegrationRouter(api).
			ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1/integrations", nil))

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"url":"https://example.test/hook"`)
		assert.Contains(t, w.Body.String(), `"hasSecret":true`)
		assert.NotContains(t, w.Body.String(), "sealed")
		assert.Contains(t, w.Body.String(), `"total":1`)
	})

	t.Run("list returns 500 when the repository fails", func(t *testing.T) {
		api := newIntegrationAPI(&fakeIntegrations{listErr: errors.New("boom")})
		w := httptest.NewRecorder()
		newIntegrationRouter(api).
			ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1/integrations", nil))

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "failed_integration_request")
	})

	t.Run("get returns 404 when the integration belongs to another workspace", func(t *testing.T) {
		other := *stored
		other.WorkspaceId = "ws-2"
		api := newIntegrationAPI(&fakeIntegrations{integration: &other})
		w := httptest.NewRecorder()
		newIntegrationRouter(api).
			ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1/integrations/int-1", nil))

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "integration_not_found")
	})

	t.Run("create returns 201 with the generated secret", func(t *testing.T) {
		api := newIntegrationAPI(&fakeIntegrations{})
		w := httptest.NewRecorder()
		newIntegrationRouter(api).
			ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/ws-1/integrations", strings.NewReader(body)))

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Body.String(), `"secret":"whsec_`)
		assert.Contains(t, w.Body.String(), `"events":["audit.log","agent.health.changed"]`)
	})

	t.Run("create requires at least one known event", func(t *testing.T) {
		for _, events := range []string{`[]`, `["memory.add"]`, `["audit.log","audit.log"]`} {
			api := newIntegrationAPI(&fakeIntegrations{})
			w := httptest.NewRecorder()
			newIntegrationRouter(api).
				ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/ws-1/integrations", strings.NewReader(`{"type":"slack","name":"slack","url":"https://hooks.slack.com/x","events":`+events+`}`)))

			assert.Equal(t, http.StatusBadRequest, w.Code, events)
		}
	})

	t.Run("create rejects an unknown type", func(t *testing.T) {
		api := newIntegrationAPI(&fakeIntegrations{})
		w := httptest.NewRecorder()
		newIntegrationRouter(api).
			ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/ws-1/integrations", strings.NewReader(`{"type":"email","name":"x"}`)))

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("create returns 400 when settings are missing for the type", func(t *testing.T) {
		api := newIntegrationAPI(&fakeIntegrations{})
		w := httptest.NewRecorder()
		newIntegrationRouter(api).
			ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/ws-1/integrations", strings.NewReader(`{"type":"splunk","name":"splunk","url":"https://splunk.test","events":["audit.log"]}`)))

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid_integration_config")
	})

	t.Run("create accepts a pagerduty integration", func(t *testing.T) {
		api := newIntegrationAPI(&fakeIntegrations{})
		w := httptest.NewRecorder()
		newIntegrationRouter(api).
			ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/ws-1/integrations", strings.NewReader(`{"type":"pagerduty","name":"on-call","site":"events.pagerduty.com","secret":"routing-key-1","events":["system.health.degraded"],"isActive":true}`)))

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Body.String(), `"site":"events.pagerduty.com"`)
		assert.NotContains(t, w.Body.String(), "routing-key-1")
	})

	t.Run("update returns the saved integration", func(t *testing.T) {
		current := *stored
		api := newIntegrationAPI(&fakeIntegrations{integration: &current})
		w := httptest.NewRecorder()
		newIntegrationRouter(api).
			ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/api/v1/workspaces/ws-1/integrations/int-1", strings.NewReader(`{"type":"webhook","name":"renamed","url":"https://example.test/new","events":["catalog.changed"],"isActive":false}`)))

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"name":"renamed"`)
		assert.Contains(t, w.Body.String(), `"isActive":false`)
		assert.Contains(t, w.Body.String(), `"events":["catalog.changed"]`)
	})

	t.Run("delete returns 204", func(t *testing.T) {
		api := newIntegrationAPI(&fakeIntegrations{integration: stored})
		w := httptest.NewRecorder()
		newIntegrationRouter(api).
			ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/api/v1/workspaces/ws-1/integrations/int-1", nil))

		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("delete returns 500 when the repository fails", func(t *testing.T) {
		api := newIntegrationAPI(&fakeIntegrations{integration: stored, deleteErr: errors.New("boom")})
		w := httptest.NewRecorder()
		newIntegrationRouter(api).
			ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/api/v1/workspaces/ws-1/integrations/int-1", nil))

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
