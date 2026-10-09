// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package api

import (
	"errors"
	"net/http"

	"github.com/clivern/cognit/db"
	"github.com/clivern/cognit/locale"
	"github.com/clivern/cognit/module"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

// ListAgentsAction lists agents in a workspace.
func (a *API) ListAgentsAction(w http.ResponseWriter, r *http.Request) {
	wid := chi.URLParam(r, "workspaceId")
	if lo.IsEmpty(wid) {
		a.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"errorMessage": locale.TR(r, "invalid_workspace_id"),
		})
		return
	}

	limit, offset := a.ParsePagination(r)

	result, err := a.Agent.ListAgents(db.Id(wid), limit, offset)
	if err != nil {
		switch {
		case errors.Is(err, module.ErrWorkspaceNotFound):
			a.WriteJSON(w, http.StatusNotFound, map[string]any{
				"errorMessage": locale.TR(r, "workspace_not_found"),
			})
		default:
			log.Error().
				Err(err).
				Str("workspaceId", wid).
				Msg("Failed to list agents")
			a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
				"errorMessage": locale.TR(r, "failed_agent_request"),
			})
		}

		return
	}

	a.WriteJSON(w, http.StatusOK, map[string]any{
		"agents": result.Agents,
		"_meta": map[string]any{
			"limit":  limit,
			"offset": offset,
			"total":  result.Total,
		},
	})
}

// GetAgentAction returns one agent by name.
func (a *API) GetAgentAction(w http.ResponseWriter, r *http.Request) {
	wid := chi.URLParam(r, "workspaceId")
	if lo.IsEmpty(wid) {
		a.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"errorMessage": locale.TR(r, "invalid_workspace_id"),
		})
		return
	}

	agentName := chi.URLParam(r, "agentName")
	if lo.IsEmpty(agentName) {
		a.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"errorMessage": locale.TR(r, "invalid_agent_name"),
		})
		return
	}

	agent, err := a.Agent.GetAgent(db.Id(wid), agentName)
	if err != nil {
		switch {
		case errors.Is(err, module.ErrWorkspaceNotFound):
			a.WriteJSON(w, http.StatusNotFound, map[string]any{
				"errorMessage": locale.TR(r, "workspace_not_found"),
			})
		case errors.Is(err, module.ErrAgentNotFound):
			a.WriteJSON(w, http.StatusNotFound, map[string]any{
				"errorMessage": locale.TR(r, "agent_not_found"),
			})
		case errors.Is(err, module.ErrInvalidAgentName):
			a.WriteJSON(w, http.StatusBadRequest, map[string]any{
				"errorMessage": locale.TR(r, "invalid_agent_name"),
			})
		default:
			log.Error().
				Err(err).
				Str("workspaceId", wid).
				Str("agent", agentName).
				Msg("Failed to get agent")
			a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
				"errorMessage": locale.TR(r, "failed_agent_request"),
			})
		}

		return
	}

	a.WriteJSON(w, http.StatusOK, agent)
}

// UpsertAgentAction registers or replaces an agent card.
func (a *API) UpsertAgentAction(w http.ResponseWriter, r *http.Request) {
	wid := chi.URLParam(r, "workspaceId")
	if lo.IsEmpty(wid) {
		a.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"errorMessage": locale.TR(r, "invalid_workspace_id"),
		})
		return
	}

	agentName := chi.URLParam(r, "agentName")
	if lo.IsEmpty(agentName) {
		a.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"errorMessage": locale.TR(r, "invalid_agent_name"),
		})
		return
	}

	var req module.UpsertAgentRequest
	err := a.DecodeAndValidate(r, &req)
	if err != nil {
		a.WriteValidationError(w, err)
		return
	}

	log.Info().
		Str("workspaceId", wid).
		Str("agent", agentName).
		Msg("Upserting agent")

	agent, created, err := a.Agent.UpsertAgent(db.Id(wid), agentName, &req)
	if err != nil {
		switch {
		case errors.Is(err, module.ErrWorkspaceNotFound):
			a.WriteJSON(w, http.StatusNotFound, map[string]any{
				"errorMessage": locale.TR(r, "workspace_not_found"),
			})
		case errors.Is(err, module.ErrInvalidAgentName):
			a.WriteJSON(w, http.StatusBadRequest, map[string]any{
				"errorMessage": locale.TR(r, "invalid_agent_name"),
			})
		case errors.Is(err, module.ErrInvalidAgentCard):
			a.WriteJSON(w, http.StatusBadRequest, map[string]any{
				"errorMessage": locale.TR(r, "invalid_agent_card"),
			})
		case errors.Is(err, module.ErrAgentVersionRequired):
			a.WriteJSON(w, http.StatusBadRequest, map[string]any{
				"errorMessage": locale.TR(r, "invalid_agent_version"),
			})
		default:
			log.Error().
				Err(err).
				Str("workspaceId", wid).
				Str("agent", agentName).
				Msg("Failed to upsert agent")
			a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
				"errorMessage": locale.TR(r, "failed_agent_request"),
			})
		}

		return
	}

	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}

	a.WriteJSON(w, status, agent)
}

// DeleteAgentAction removes an agent by name.
func (a *API) DeleteAgentAction(w http.ResponseWriter, r *http.Request) {
	wid := chi.URLParam(r, "workspaceId")
	if lo.IsEmpty(wid) {
		a.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"errorMessage": locale.TR(r, "invalid_workspace_id"),
		})
		return
	}

	agentName := chi.URLParam(r, "agentName")
	if lo.IsEmpty(agentName) {
		a.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"errorMessage": locale.TR(r, "invalid_agent_name"),
		})
		return
	}

	log.Info().
		Str("workspaceId", wid).
		Str("agent", agentName).
		Msg("Deleting agent")

	err := a.Agent.DeleteAgent(db.Id(wid), agentName)
	if err != nil {
		switch {
		case errors.Is(err, module.ErrWorkspaceNotFound):
			a.WriteJSON(w, http.StatusNotFound, map[string]any{
				"errorMessage": locale.TR(r, "workspace_not_found"),
			})
		case errors.Is(err, module.ErrAgentNotFound):
			a.WriteJSON(w, http.StatusNotFound, map[string]any{
				"errorMessage": locale.TR(r, "agent_not_found"),
			})
		case errors.Is(err, module.ErrInvalidAgentName):
			a.WriteJSON(w, http.StatusBadRequest, map[string]any{
				"errorMessage": locale.TR(r, "invalid_agent_name"),
			})
		default:
			log.Error().
				Err(err).
				Str("workspaceId", wid).
				Str("agent", agentName).
				Msg("Failed to delete agent")
			a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
				"errorMessage": locale.TR(r, "failed_agent_request"),
			})
		}

		return
	}

	w.WriteHeader(http.StatusNoContent)
}
