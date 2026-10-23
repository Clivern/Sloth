// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package api

import (
	"io"
	"net/http"

	"github.com/clivern/sloth/db"
	"github.com/clivern/sloth/locale"
	"github.com/clivern/sloth/module"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

// SendAgentMessageAction forwards a message:send from the calling agent to a named agent.
func (a *API) SendAgentMessageAction(w http.ResponseWriter, r *http.Request) {
	wid := chi.URLParam(r, "workspaceId")
	agentName := chi.URLParam(r, "agentName")
	caller, _ := module.AgentFrom(r.Context())

	body, _ := io.ReadAll(r.Body)

	resp, err := a.Agents.SendMessage(r.Context(), db.Id(wid), caller, agentName, &module.AgentsRequest{
		Version: r.Header.Get("A2A-Version"),
		Body:    body,
	})
	if err != nil {
		a.writeAgentsError(w, r, err, "Failed to send message to agent")
		return
	}

	a.writeAgentsResponse(w, resp)
}

// SendSkillMessageAction forwards a message:send from the calling agent to an agent that offers a skill.
func (a *API) SendSkillMessageAction(w http.ResponseWriter, r *http.Request) {
	wid := chi.URLParam(r, "workspaceId")
	skill := chi.URLParam(r, "skill")
	caller, _ := module.AgentFrom(r.Context())

	body, _ := io.ReadAll(r.Body)

	resp, err := a.Agents.SendSkillMessage(r.Context(), db.Id(wid), caller, skill, &module.AgentsRequest{
		Version: r.Header.Get("A2A-Version"),
		Body:    body,
	})
	if err != nil {
		a.writeAgentsError(w, r, err, "Failed to send message to skill")
		return
	}

	a.writeAgentsResponse(w, resp)
}

// GetTaskAction returns a task the calling agent started, from the instance that accepted it.
func (a *API) GetTaskAction(w http.ResponseWriter, r *http.Request) {
	wid := chi.URLParam(r, "workspaceId")
	taskId := chi.URLParam(r, "taskId")
	caller, _ := module.AgentFrom(r.Context())

	resp, err := a.Agents.GetTask(r.Context(), db.Id(wid), caller, taskId, &module.AgentsRequest{
		Version: r.Header.Get("A2A-Version"),
	})
	if err != nil {
		a.writeAgentsError(w, r, err, "Failed to get agent task")
		return
	}

	a.writeAgentsResponse(w, resp)
}

// writeAgentsError maps an agents module error to its response, logging the ones it does not know.
func (a *API) writeAgentsError(w http.ResponseWriter, r *http.Request, err error, message string) {
	if a.WriteError(w, err, []ErrorMapping{
		{module.ErrCallDenied, http.StatusForbidden, locale.TR(r, "a2a_call_denied")},
		{module.ErrAgentNotFound, http.StatusNotFound, locale.TR(r, "agent_not_found")},
		{module.ErrSkillNotFound, http.StatusNotFound, locale.TR(r, "skill_not_found")},
		{module.ErrTaskNotFound, http.StatusNotFound, locale.TR(r, "task_not_found")},
		{module.ErrNoHealthyInstance, http.StatusServiceUnavailable, locale.TR(r, "no_healthy_instance")},
		{module.ErrAgentUnreachable, http.StatusBadGateway, locale.TR(r, "agent_unreachable")},
	}) {
		return
	}

	log.Error().
		Err(err).
		Str("workspaceId", chi.URLParam(r, "workspaceId")).
		Msg(message)
	a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
		"errorMessage": locale.TR(r, "failed_a2a_request"),
	})
}

// writeAgentsResponse relays the instance reply as is.
func (a *API) writeAgentsResponse(w http.ResponseWriter, resp *module.AgentsResponse) {
	w.Header().Set("Content-Type", resp.ContentType)
	w.WriteHeader(resp.StatusCode)
	w.Write(resp.Body)
}
