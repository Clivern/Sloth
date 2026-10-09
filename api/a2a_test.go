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

// TestUnitA2AStubActions pins A2A handlers to 501 until they forward to agents.
func TestUnitA2AStubActions(t *testing.T) {
	api := &API{}
	router := chi.NewRouter()
	router.Post("/api/v1/workspaces/{workspaceId}/a2a/agents/{agentName}/message:send", api.SendAgentMessageAction)
	router.Post("/api/v1/workspaces/{workspaceId}/a2a/skills/{skill}/message:send", api.SendSkillMessageAction)
	router.Get("/api/v1/workspaces/{workspaceId}/a2a/tasks/{taskId}", api.GetTaskAction)

	cases := []struct {
		name   string
		method string
		path   string
	}{
		{name: "send agent message", method: http.MethodPost, path: "/api/v1/workspaces/ws-1/a2a/agents/user-directory/message:send"},
		{name: "send skill message", method: http.MethodPost, path: "/api/v1/workspaces/ws-1/a2a/skills/lookup/message:send"},
		{name: "get task", method: http.MethodGet, path: "/api/v1/workspaces/ws-1/a2a/tasks/task-1"},
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
