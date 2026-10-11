// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package api

import (
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

	limit, offset := a.ParsePagination(r)

	result, err := a.Check.ListAgentChecks(db.Id(wid), agentName, limit, offset)
	if err != nil {
		if a.WriteError(w, err, []ErrorMapping{
			{module.ErrAgentNotFound, http.StatusNotFound, locale.TR(r, "agent_not_found")},
		}) {
			return
		}

		log.Error().
			Err(err).
			Str("workspaceId", wid).
			Str("agent", agentName).
			Msg("Failed to list agent checks")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_check_request"),
		})

		return
	}

	a.WriteJSON(w, http.StatusOK, map[string]any{
		"checks": result.Checks,
		"_meta": map[string]any{
			"limit":  limit,
			"offset": offset,
			"total":  result.Total,
		},
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

	check, created, err := a.Check.UpsertAgentCheck(
		db.Id(wid),
		agentName,
		checkId,
		&req,
		locale.GetLangFromRequest(r),
	)
	if err != nil {
		if a.WriteError(w, err, []ErrorMapping{
			{module.ErrAgentNotFound, http.StatusNotFound, locale.TR(r, "agent_not_found")},
			{module.ErrInvalidCheckId, http.StatusBadRequest, locale.TR(r, "invalid_check_id")},
			{module.ErrInvalidCheckPath, http.StatusBadRequest, locale.TR(r, "invalid_check_path")},
			{module.ErrInvalidCheckTiming, http.StatusBadRequest, locale.TR(r, "invalid_check_timing")},
		}) {
			return
		}

		log.Error().
			Err(err).
			Str("workspaceId", wid).
			Str("agent", agentName).
			Msg("Failed to upsert agent check")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_check_request"),
		})

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
		if a.WriteError(w, err, []ErrorMapping{
			{module.ErrAgentNotFound, http.StatusNotFound, locale.TR(r, "agent_not_found")},
			{module.ErrCheckNotFound, http.StatusNotFound, locale.TR(r, "check_not_found")},
		}) {
			return
		}

		log.Error().
			Err(err).
			Str("workspaceId", wid).
			Str("agent", agentName).
			Msg("Failed to delete agent check")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_check_request"),
		})

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
		if a.WriteError(w, err, []ErrorMapping{
			{module.ErrAgentNotFound, http.StatusNotFound, locale.TR(r, "agent_not_found")},
			{module.ErrInstanceNotFound, http.StatusNotFound, locale.TR(r, "instance_not_found")},
			{module.ErrCheckNotFound, http.StatusNotFound, locale.TR(r, "check_not_found")},
			{module.ErrCheckNotReportable, http.StatusBadRequest, locale.TR(r, "check_not_reportable")},
		}) {
			return
		}

		log.Error().
			Err(err).
			Str("workspaceId", wid).
			Str("agent", agentName).
			Msg("Failed to report check")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_check_request"),
		})

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
		if a.WriteError(w, err, []ErrorMapping{
			{module.ErrAgentNotFound, http.StatusNotFound, locale.TR(r, "agent_not_found")},
			{module.ErrInstanceNotFound, http.StatusNotFound, locale.TR(r, "instance_not_found")},
			{module.ErrCheckNotFound, http.StatusNotFound, locale.TR(r, "check_not_found")},
			{module.ErrCheckNotReportable, http.StatusBadRequest, locale.TR(r, "check_not_reportable")},
		}) {
			return
		}

		log.Error().
			Err(err).
			Str("workspaceId", wid).
			Str("agent", agentName).
			Msg("Failed to report check")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_check_request"),
		})

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
		if a.WriteError(w, err, []ErrorMapping{
			{module.ErrAgentNotFound, http.StatusNotFound, locale.TR(r, "agent_not_found")},
			{module.ErrInstanceNotFound, http.StatusNotFound, locale.TR(r, "instance_not_found")},
			{module.ErrCheckNotFound, http.StatusNotFound, locale.TR(r, "check_not_found")},
			{module.ErrCheckNotReportable, http.StatusBadRequest, locale.TR(r, "check_not_reportable")},
		}) {
			return
		}

		log.Error().
			Err(err).
			Str("workspaceId", wid).
			Str("agent", agentName).
			Msg("Failed to report check")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_check_request"),
		})

		return
	}

	a.WriteJSON(w, http.StatusOK, check)
}
