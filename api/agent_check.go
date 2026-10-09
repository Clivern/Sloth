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

// ListAgentChecksAction lists the health check templates of an agent.
func (a *API) ListAgentChecksAction(w http.ResponseWriter, r *http.Request) {
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

	checks, err := a.Check.ListAgentChecks(db.Id(wid), agentName)
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
				Msg("Failed to list agent checks")
			a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
				"errorMessage": locale.TR(r, "failed_check_request"),
			})
		}

		return
	}

	a.WriteJSON(w, http.StatusOK, map[string]any{
		"checks": checks,
	})
}

// UpsertAgentCheckAction creates or replaces a health check template.
func (a *API) UpsertAgentCheckAction(w http.ResponseWriter, r *http.Request) {
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

	checkId := chi.URLParam(r, "checkId")

	var req module.UpsertAgentCheckRequest
	err := a.DecodeAndValidate(r, &req)
	if err != nil {
		a.WriteValidationError(w, err)
		return
	}

	check, created, err := a.Check.UpsertAgentCheck(db.Id(wid), agentName, checkId, &req)
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
		case errors.Is(err, module.ErrInvalidCheckId):
			a.WriteJSON(w, http.StatusBadRequest, map[string]any{
				"errorMessage": locale.TR(r, "invalid_check_id"),
			})
		case errors.Is(err, module.ErrReservedCheckId):
			a.WriteJSON(w, http.StatusBadRequest, map[string]any{
				"errorMessage": locale.TR(r, "reserved_check_id"),
			})
		case errors.Is(err, module.ErrUnsupportedCheckType):
			a.WriteJSON(w, http.StatusBadRequest, map[string]any{
				"errorMessage": locale.TR(r, "unsupported_check_type"),
			})
		case errors.Is(err, module.ErrInvalidCheckPath):
			a.WriteJSON(w, http.StatusBadRequest, map[string]any{
				"errorMessage": locale.TR(r, "invalid_check_path"),
			})
		case errors.Is(err, module.ErrInvalidCheckScheme):
			a.WriteJSON(w, http.StatusBadRequest, map[string]any{
				"errorMessage": locale.TR(r, "invalid_check_scheme"),
			})
		case errors.Is(err, module.ErrInvalidCheckTiming):
			a.WriteJSON(w, http.StatusBadRequest, map[string]any{
				"errorMessage": locale.TR(r, "invalid_check_timing"),
			})
		default:
			log.Error().
				Err(err).
				Str("workspaceId", wid).
				Str("agent", agentName).
				Msg("Failed to upsert agent check")
			a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
				"errorMessage": locale.TR(r, "failed_check_request"),
			})
		}

		return
	}

	a.WriteJSON(w, lo.Ternary(created, http.StatusCreated, http.StatusOK), check)
}

// DeleteAgentCheckAction removes a health check template from an agent and its instances.
func (a *API) DeleteAgentCheckAction(w http.ResponseWriter, r *http.Request) {
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

	err := a.Check.DeleteAgentCheck(db.Id(wid), agentName, chi.URLParam(r, "checkId"))
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
		case errors.Is(err, module.ErrCheckNotFound):
			a.WriteJSON(w, http.StatusNotFound, map[string]any{
				"errorMessage": locale.TR(r, "check_not_found"),
			})
		case errors.Is(err, module.ErrInvalidAgentName):
			a.WriteJSON(w, http.StatusBadRequest, map[string]any{
				"errorMessage": locale.TR(r, "invalid_agent_name"),
			})
		case errors.Is(err, module.ErrInvalidCheckId):
			a.WriteJSON(w, http.StatusBadRequest, map[string]any{
				"errorMessage": locale.TR(r, "invalid_check_id"),
			})
		default:
			log.Error().
				Err(err).
				Str("workspaceId", wid).
				Str("agent", agentName).
				Msg("Failed to delete agent check")
			a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
				"errorMessage": locale.TR(r, "failed_check_request"),
			})
		}

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// PassCheckAction marks one of an instance's TTL checks passing.
func (a *API) PassCheckAction(w http.ResponseWriter, r *http.Request) {
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

	// The body is optional; an empty one reports without output.
	var req module.ReportCheckRequest
	if r.ContentLength != 0 {
		err := a.DecodeAndValidate(r, &req)
		if err != nil {
			a.WriteValidationError(w, err)
			return
		}
	}

	check, err := a.Check.ReportCheck(
		db.Id(wid),
		agentName,
		chi.URLParam(r, "instanceId"),
		chi.URLParam(r, "checkId"),
		db.HealthCheckStatusPassing,
		&req,
	)
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
		case errors.Is(err, module.ErrCheckNotFound):
			a.WriteJSON(w, http.StatusNotFound, map[string]any{
				"errorMessage": locale.TR(r, "check_not_found"),
			})
		case errors.Is(err, module.ErrInvalidAgentName):
			a.WriteJSON(w, http.StatusBadRequest, map[string]any{
				"errorMessage": locale.TR(r, "invalid_agent_name"),
			})
		case errors.Is(err, module.ErrInvalidInstanceId):
			a.WriteJSON(w, http.StatusBadRequest, map[string]any{
				"errorMessage": locale.TR(r, "invalid_instance_id"),
			})
		case errors.Is(err, module.ErrInvalidCheckId):
			a.WriteJSON(w, http.StatusBadRequest, map[string]any{
				"errorMessage": locale.TR(r, "invalid_check_id"),
			})
		case errors.Is(err, module.ErrInvalidCheckStatus):
			a.WriteJSON(w, http.StatusBadRequest, map[string]any{
				"errorMessage": locale.TR(r, "failed_check_request"),
			})
		case errors.Is(err, module.ErrCheckNotReportable):
			a.WriteJSON(w, http.StatusBadRequest, map[string]any{
				"errorMessage": locale.TR(r, "check_not_reportable"),
			})
		default:
			log.Error().
				Err(err).
				Str("workspaceId", wid).
				Str("agent", agentName).
				Msg("Failed to report check")
			a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
				"errorMessage": locale.TR(r, "failed_check_request"),
			})
		}

		return
	}

	a.WriteJSON(w, http.StatusOK, check)
}

// WarnCheckAction marks one of an instance's TTL checks warning.
func (a *API) WarnCheckAction(w http.ResponseWriter, r *http.Request) {
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

	// The body is optional; an empty one reports without output.
	var req module.ReportCheckRequest
	if r.ContentLength != 0 {
		err := a.DecodeAndValidate(r, &req)
		if err != nil {
			a.WriteValidationError(w, err)
			return
		}
	}

	check, err := a.Check.ReportCheck(
		db.Id(wid),
		agentName,
		chi.URLParam(r, "instanceId"),
		chi.URLParam(r, "checkId"),
		db.HealthCheckStatusWarning,
		&req,
	)
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
		case errors.Is(err, module.ErrCheckNotFound):
			a.WriteJSON(w, http.StatusNotFound, map[string]any{
				"errorMessage": locale.TR(r, "check_not_found"),
			})
		case errors.Is(err, module.ErrInvalidAgentName):
			a.WriteJSON(w, http.StatusBadRequest, map[string]any{
				"errorMessage": locale.TR(r, "invalid_agent_name"),
			})
		case errors.Is(err, module.ErrInvalidInstanceId):
			a.WriteJSON(w, http.StatusBadRequest, map[string]any{
				"errorMessage": locale.TR(r, "invalid_instance_id"),
			})
		case errors.Is(err, module.ErrInvalidCheckId):
			a.WriteJSON(w, http.StatusBadRequest, map[string]any{
				"errorMessage": locale.TR(r, "invalid_check_id"),
			})
		case errors.Is(err, module.ErrInvalidCheckStatus):
			a.WriteJSON(w, http.StatusBadRequest, map[string]any{
				"errorMessage": locale.TR(r, "failed_check_request"),
			})
		case errors.Is(err, module.ErrCheckNotReportable):
			a.WriteJSON(w, http.StatusBadRequest, map[string]any{
				"errorMessage": locale.TR(r, "check_not_reportable"),
			})
		default:
			log.Error().
				Err(err).
				Str("workspaceId", wid).
				Str("agent", agentName).
				Msg("Failed to report check")
			a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
				"errorMessage": locale.TR(r, "failed_check_request"),
			})
		}

		return
	}

	a.WriteJSON(w, http.StatusOK, check)
}

// FailCheckAction marks one of an instance's TTL checks critical.
func (a *API) FailCheckAction(w http.ResponseWriter, r *http.Request) {
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

	// The body is optional; an empty one reports without output.
	var req module.ReportCheckRequest
	if r.ContentLength != 0 {
		err := a.DecodeAndValidate(r, &req)
		if err != nil {
			a.WriteValidationError(w, err)
			return
		}
	}

	check, err := a.Check.ReportCheck(
		db.Id(wid),
		agentName,
		chi.URLParam(r, "instanceId"),
		chi.URLParam(r, "checkId"),
		db.HealthCheckStatusCritical,
		&req,
	)
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
		case errors.Is(err, module.ErrCheckNotFound):
			a.WriteJSON(w, http.StatusNotFound, map[string]any{
				"errorMessage": locale.TR(r, "check_not_found"),
			})
		case errors.Is(err, module.ErrInvalidAgentName):
			a.WriteJSON(w, http.StatusBadRequest, map[string]any{
				"errorMessage": locale.TR(r, "invalid_agent_name"),
			})
		case errors.Is(err, module.ErrInvalidInstanceId):
			a.WriteJSON(w, http.StatusBadRequest, map[string]any{
				"errorMessage": locale.TR(r, "invalid_instance_id"),
			})
		case errors.Is(err, module.ErrInvalidCheckId):
			a.WriteJSON(w, http.StatusBadRequest, map[string]any{
				"errorMessage": locale.TR(r, "invalid_check_id"),
			})
		case errors.Is(err, module.ErrInvalidCheckStatus):
			a.WriteJSON(w, http.StatusBadRequest, map[string]any{
				"errorMessage": locale.TR(r, "failed_check_request"),
			})
		case errors.Is(err, module.ErrCheckNotReportable):
			a.WriteJSON(w, http.StatusBadRequest, map[string]any{
				"errorMessage": locale.TR(r, "check_not_reportable"),
			})
		default:
			log.Error().
				Err(err).
				Str("workspaceId", wid).
				Str("agent", agentName).
				Msg("Failed to report check")
			a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
				"errorMessage": locale.TR(r, "failed_check_request"),
			})
		}

		return
	}

	a.WriteJSON(w, http.StatusOK, check)
}
