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

// ListInstancesAction lists the instances of an agent.
func (a *API) ListInstancesAction(w http.ResponseWriter, r *http.Request) {
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

	passing := r.URL.Query().Get("passing") == "true"

	instances, err := a.Instance.ListInstances(db.Id(wid), agentName, passing)
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
				Msg("Failed to list instances")
			a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
				"errorMessage": locale.TR(r, "failed_instance_request"),
			})
		}

		return
	}

	a.WriteJSON(w, http.StatusOK, map[string]any{
		"instances": instances,
	})
}

// GetInstanceAction returns one instance of an agent.
func (a *API) GetInstanceAction(w http.ResponseWriter, r *http.Request) {
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

	instanceId := chi.URLParam(r, "instanceId")
	if lo.IsEmpty(instanceId) {
		a.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"errorMessage": locale.TR(r, "invalid_instance_id"),
		})
		return
	}

	instance, err := a.Instance.GetInstance(db.Id(wid), agentName, instanceId)
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
		case errors.Is(err, module.ErrInstanceNotFound):
			a.WriteJSON(w, http.StatusNotFound, map[string]any{
				"errorMessage": locale.TR(r, "instance_not_found"),
			})
		case errors.Is(err, module.ErrInvalidAgentName):
			a.WriteJSON(w, http.StatusBadRequest, map[string]any{
				"errorMessage": locale.TR(r, "invalid_agent_name"),
			})
		case errors.Is(err, module.ErrInvalidInstanceId):
			a.WriteJSON(w, http.StatusBadRequest, map[string]any{
				"errorMessage": locale.TR(r, "invalid_instance_id"),
			})
		default:
			log.Error().
				Err(err).
				Str("workspaceId", wid).
				Str("agent", agentName).
				Str("instance", instanceId).
				Msg("Failed to get instance")
			a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
				"errorMessage": locale.TR(r, "failed_instance_request"),
			})
		}

		return
	}

	a.WriteJSON(w, http.StatusOK, instance)
}

// RegisterInstanceAction registers or updates an instance and starts its lease.
func (a *API) RegisterInstanceAction(w http.ResponseWriter, r *http.Request) {
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

	instanceId := chi.URLParam(r, "instanceId")
	if lo.IsEmpty(instanceId) {
		a.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"errorMessage": locale.TR(r, "invalid_instance_id"),
		})
		return
	}

	var req module.RegisterInstanceRequest
	err := a.DecodeAndValidate(r, &req)
	if err != nil {
		a.WriteValidationError(w, err)
		return
	}

	log.Info().
		Str("workspaceId", wid).
		Str("agent", agentName).
		Str("instance", instanceId).
		Msg("Registering instance")

	instance, created, err := a.Instance.RegisterInstance(db.Id(wid), agentName, instanceId, &req)
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
		case errors.Is(err, module.ErrInvalidInstanceId):
			a.WriteJSON(w, http.StatusBadRequest, map[string]any{
				"errorMessage": locale.TR(r, "invalid_instance_id"),
			})
		default:
			log.Error().
				Err(err).
				Str("workspaceId", wid).
				Str("agent", agentName).
				Str("instance", instanceId).
				Msg("Failed to register instance")
			a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
				"errorMessage": locale.TR(r, "failed_instance_request"),
			})
		}

		return
	}

	a.WriteJSON(w, lo.Ternary(created, http.StatusCreated, http.StatusOK), instance)
}

// RenewInstanceAction extends an instance lease.
func (a *API) RenewInstanceAction(w http.ResponseWriter, r *http.Request) {
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

	instanceId := chi.URLParam(r, "instanceId")
	if lo.IsEmpty(instanceId) {
		a.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"errorMessage": locale.TR(r, "invalid_instance_id"),
		})
		return
	}

	instance, err := a.Instance.RenewInstance(db.Id(wid), agentName, instanceId)
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
		case errors.Is(err, module.ErrInstanceNotFound):
			a.WriteJSON(w, http.StatusNotFound, map[string]any{
				"errorMessage": locale.TR(r, "instance_not_found"),
			})
		case errors.Is(err, module.ErrInvalidAgentName):
			a.WriteJSON(w, http.StatusBadRequest, map[string]any{
				"errorMessage": locale.TR(r, "invalid_agent_name"),
			})
		case errors.Is(err, module.ErrInvalidInstanceId):
			a.WriteJSON(w, http.StatusBadRequest, map[string]any{
				"errorMessage": locale.TR(r, "invalid_instance_id"),
			})
		default:
			log.Error().
				Err(err).
				Str("workspaceId", wid).
				Str("agent", agentName).
				Str("instance", instanceId).
				Msg("Failed to renew instance")
			a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
				"errorMessage": locale.TR(r, "failed_instance_request"),
			})
		}

		return
	}

	a.WriteJSON(w, http.StatusOK, instance)
}

// DeregisterInstanceAction removes an instance.
func (a *API) DeregisterInstanceAction(w http.ResponseWriter, r *http.Request) {
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

	instanceId := chi.URLParam(r, "instanceId")
	if lo.IsEmpty(instanceId) {
		a.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"errorMessage": locale.TR(r, "invalid_instance_id"),
		})
		return
	}

	log.Info().
		Str("workspaceId", wid).
		Str("agent", agentName).
		Str("instance", instanceId).
		Msg("Deregistering instance")

	err := a.Instance.DeregisterInstance(db.Id(wid), agentName, instanceId)
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
		case errors.Is(err, module.ErrInstanceNotFound):
			a.WriteJSON(w, http.StatusNotFound, map[string]any{
				"errorMessage": locale.TR(r, "instance_not_found"),
			})
		case errors.Is(err, module.ErrInvalidAgentName):
			a.WriteJSON(w, http.StatusBadRequest, map[string]any{
				"errorMessage": locale.TR(r, "invalid_agent_name"),
			})
		case errors.Is(err, module.ErrInvalidInstanceId):
			a.WriteJSON(w, http.StatusBadRequest, map[string]any{
				"errorMessage": locale.TR(r, "invalid_instance_id"),
			})
		default:
			log.Error().
				Err(err).
				Str("workspaceId", wid).
				Str("agent", agentName).
				Str("instance", instanceId).
				Msg("Failed to deregister instance")
			a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
				"errorMessage": locale.TR(r, "failed_instance_request"),
			})
		}

		return
	}

	w.WriteHeader(http.StatusNoContent)
}
