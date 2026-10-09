// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package api

import (
	"net/http"

	"github.com/clivern/cognit/locale"
)

// SendAgentMessageAction sends a message to an agent through the gateway.
func (a *API) SendAgentMessageAction(w http.ResponseWriter, r *http.Request) {
	// TODO: forward an A2A message to a named agent (POST a2a/agents/{agentName}/message:send).
	a.WriteJSON(w, http.StatusNotImplemented, map[string]any{
		"errorMessage": locale.TR(r, "not_implemented"),
	})
}

// SendSkillMessageAction sends a message to the agent that serves a skill.
func (a *API) SendSkillMessageAction(w http.ResponseWriter, r *http.Request) {
	// TODO: resolve a skill to an agent and forward the A2A message (POST a2a/skills/{skill}/message:send).
	a.WriteJSON(w, http.StatusNotImplemented, map[string]any{
		"errorMessage": locale.TR(r, "not_implemented"),
	})
}

// GetTaskAction returns the status of an A2A task.
func (a *API) GetTaskAction(w http.ResponseWriter, r *http.Request) {
	// TODO: return the task state and result (GET a2a/tasks/{taskId}).
	a.WriteJSON(w, http.StatusNotImplemented, map[string]any{
		"errorMessage": locale.TR(r, "not_implemented"),
	})
}
