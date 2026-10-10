// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/clivern/cognit/db"

	"github.com/stretchr/testify/assert"
)

func TestUnitIntentionActions(t *testing.T) {
	workspace := &db.Workspace{Id: "ws-1"}
	skill := "lookup"
	stored := &db.Intention{
		Id:               "intent-1",
		WorkspaceId:      "ws-1",
		SourceAgent:      "frontend",
		DestinationAgent: "users",
		Skill:            &skill,
		Action:           db.IntentionActionAllow,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	body := `{"source":{"agent":"frontend"},"destination":{"agent":"users","skill":"lookup"},"action":"allow"}`

	t.Run("list returns intentions with total", func(t *testing.T) {
		api := newIntentionAPI(workspace, &fakeIntentions{list: []*db.Intention{stored}, total: 1})
		w := httptest.NewRecorder()
		newIntentionRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1/intentions", nil))

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"source":"frontend"`)
		assert.Contains(t, w.Body.String(), `"total":1`)
	})

	t.Run("list returns an empty page for unknown workspace", func(t *testing.T) {
		api := newIntentionAPI(nil, &fakeIntentions{})
		w := httptest.NewRecorder()
		newIntentionRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1/intentions", nil))

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("list returns 500 when the repository fails", func(t *testing.T) {
		api := newIntentionAPI(workspace, &fakeIntentions{listErr: errors.New("boom")})
		w := httptest.NewRecorder()
		newIntentionRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1/intentions", nil))

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "failed_intention_request")
	})

	t.Run("get returns one intention", func(t *testing.T) {
		api := newIntentionAPI(workspace, &fakeIntentions{intention: stored})
		w := httptest.NewRecorder()
		newIntentionRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1/intentions/intent-1", nil))

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"skill":"lookup"`)
	})

	t.Run("get returns 404 when the intention belongs to another workspace", func(t *testing.T) {
		other := *stored
		other.WorkspaceId = "ws-2"
		api := newIntentionAPI(workspace, &fakeIntentions{intention: &other})
		w := httptest.NewRecorder()
		newIntentionRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1/intentions/intent-1", nil))

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "intention_not_found")
	})

	t.Run("get returns 500 when the repository fails", func(t *testing.T) {
		api := newIntentionAPI(workspace, &fakeIntentions{getErr: errors.New("boom")})
		w := httptest.NewRecorder()
		newIntentionRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1/intentions/intent-1", nil))

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("create returns 201", func(t *testing.T) {
		api := newIntentionAPI(workspace, &fakeIntentions{})
		w := httptest.NewRecorder()
		newIntentionRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/api/v1/workspaces/ws-1/intentions", strings.NewReader(body)))

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Body.String(), `"destination":"users"`)
	})

	t.Run("create returns 400 for an unknown action", func(t *testing.T) {
		api := newIntentionAPI(workspace, &fakeIntentions{})
		w := httptest.NewRecorder()
		bad := strings.Replace(body, `"allow"`, `"maybe"`, 1)
		newIntentionRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/api/v1/workspaces/ws-1/intentions", strings.NewReader(bad)))

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("create returns 500 when the repository fails", func(t *testing.T) {
		api := newIntentionAPI(workspace, &fakeIntentions{createErr: errors.New("boom")})
		w := httptest.NewRecorder()
		newIntentionRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/api/v1/workspaces/ws-1/intentions", strings.NewReader(body)))

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("update returns 200", func(t *testing.T) {
		api := newIntentionAPI(workspace, &fakeIntentions{intention: stored})
		w := httptest.NewRecorder()
		newIntentionRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/api/v1/workspaces/ws-1/intentions/intent-1", strings.NewReader(body)))

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("update returns 404 for unknown intention", func(t *testing.T) {
		api := newIntentionAPI(workspace, &fakeIntentions{})
		w := httptest.NewRecorder()
		newIntentionRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/api/v1/workspaces/ws-1/intentions/intent-1", strings.NewReader(body)))

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "intention_not_found")
	})

	t.Run("update returns 500 when the repository fails", func(t *testing.T) {
		api := newIntentionAPI(workspace, &fakeIntentions{intention: stored, updateErr: errors.New("boom")})
		w := httptest.NewRecorder()
		newIntentionRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/api/v1/workspaces/ws-1/intentions/intent-1", strings.NewReader(body)))

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("delete returns 204", func(t *testing.T) {
		api := newIntentionAPI(workspace, &fakeIntentions{intention: stored})
		w := httptest.NewRecorder()
		newIntentionRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/api/v1/workspaces/ws-1/intentions/intent-1", nil))

		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("delete returns 404 for unknown intention", func(t *testing.T) {
		api := newIntentionAPI(workspace, &fakeIntentions{})
		w := httptest.NewRecorder()
		newIntentionRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/api/v1/workspaces/ws-1/intentions/intent-1", nil))

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("delete returns 500 when the repository fails", func(t *testing.T) {
		api := newIntentionAPI(workspace, &fakeIntentions{intention: stored, deleteErr: errors.New("boom")})
		w := httptest.NewRecorder()
		newIntentionRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/api/v1/workspaces/ws-1/intentions/intent-1", nil))

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("check returns the matching rule", func(t *testing.T) {
		api := newIntentionAPI(workspace, &fakeIntentions{match: stored})
		checkBody := `{"source":{"agent":"frontend"},"destination":{"agent":"users","skill":"lookup"}}`
		w := httptest.NewRecorder()
		newIntentionRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/ws-1/intentions/check", strings.NewReader(checkBody)))

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"action":"allow"`)
		assert.Contains(t, w.Body.String(), `"intention":{`)
	})

	t.Run("check denies when no rule matches", func(t *testing.T) {
		api := newIntentionAPI(workspace, &fakeIntentions{})
		checkBody := `{"source":{"agent":"frontend"},"destination":{"agent":"users"}}`
		w := httptest.NewRecorder()
		newIntentionRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/ws-1/intentions/check", strings.NewReader(checkBody)))

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"action":"deny"`)
		assert.NotContains(t, w.Body.String(), `"intention"`)
	})

	t.Run("check does not look up the workspace", func(t *testing.T) {
		api := newIntentionAPI(nil, &fakeIntentions{})
		checkBody := `{"source":{"agent":"frontend"},"destination":{"agent":"users"}}`
		w := httptest.NewRecorder()
		newIntentionRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/ws-1/intentions/check", strings.NewReader(checkBody)))

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("check returns 500 when the repository fails", func(t *testing.T) {
		api := newIntentionAPI(workspace, &fakeIntentions{matchErr: errors.New("boom")})
		checkBody := `{"source":{"agent":"frontend"},"destination":{"agent":"users"}}`
		w := httptest.NewRecorder()
		newIntentionRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/ws-1/intentions/check", strings.NewReader(checkBody)))

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("handlers return 400 when path params are missing", func(t *testing.T) {
		api := newIntentionAPI(workspace, &fakeIntentions{})

		w := httptest.NewRecorder()
		api.GetIntentionAction(w, httptest.NewRequest(http.MethodGet, "/", nil))
		assert.Equal(t, http.StatusBadRequest, w.Code)

		w = httptest.NewRecorder()
		api.DeleteIntentionAction(w, httptest.NewRequest(http.MethodDelete, "/", nil))
		assert.Equal(t, http.StatusBadRequest, w.Code)

		w = httptest.NewRecorder()
		api.CreateIntentionAction(w, httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body)))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
