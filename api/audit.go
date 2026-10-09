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

// ListWorkspaceAuditsAction lists audit events for a workspace.
func (a *API) ListWorkspaceAuditsAction(w http.ResponseWriter, r *http.Request) {
	wid := chi.URLParam(r, "workspaceId")
	if lo.IsEmpty(wid) {
		a.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"errorMessage": locale.TR(r, "invalid_workspace_id"),
		})
		return
	}

	workspaceId := db.Id(wid)

	limit, offset := a.ParsePagination(r)

	result, err := a.Audit.ListAuditEvents(workspaceId, limit, offset)
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
				Msg("Failed to list audit events")
			a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
				"errorMessage": locale.TR(r, "failed_list_audits"),
			})
		}

		return
	}

	a.WriteJSON(w, http.StatusOK, map[string]any{
		"audits": result.Events,
		"_meta": map[string]any{
			"limit":  limit,
			"offset": offset,
			"total":  result.Total,
		},
	})
}

// GetWorkspaceAuditAction returns one audit event by id.
func (a *API) GetWorkspaceAuditAction(w http.ResponseWriter, r *http.Request) {
	wid := chi.URLParam(r, "workspaceId")
	if lo.IsEmpty(wid) {
		a.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"errorMessage": locale.TR(r, "invalid_workspace_id"),
		})
		return
	}

	workspaceId := db.Id(wid)

	auditId := chi.URLParam(r, "auditId")
	if lo.IsEmpty(auditId) {
		a.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"errorMessage": locale.TR(r, "invalid_audit_id"),
		})
		return
	}

	event, err := a.Audit.GetAuditEvent(workspaceId, db.Id(auditId))
	if err != nil {
		switch {
		case errors.Is(err, module.ErrWorkspaceNotFound):
			a.WriteJSON(w, http.StatusNotFound, map[string]any{
				"errorMessage": locale.TR(r, "workspace_not_found"),
			})
		case errors.Is(err, module.ErrAuditEventNotFound):
			a.WriteJSON(w, http.StatusNotFound, map[string]any{
				"errorMessage": locale.TR(r, "audit_event_not_found"),
			})
		default:
			log.Error().
				Err(err).
				Str("auditId", auditId).
				Msg("Failed to get audit event")
			a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
				"errorMessage": locale.TR(r, "failed_get_audit"),
			})
		}

		return
	}

	a.WriteJSON(w, http.StatusOK, event)
}
