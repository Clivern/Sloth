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

// GetProtectionAction returns the protection of an agent.
func (a *API) GetProtectionAction(w http.ResponseWriter, r *http.Request) {
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

	protection, err := a.Protection.GetProtection(db.Id(wid), agentName)
	if err != nil {
		if a.WriteError(w, err, []ErrorMapping{
			{module.ErrAgentNotFound, http.StatusNotFound, locale.TR(r, "agent_not_found")},
			{module.ErrProtectionNotFound, http.StatusNotFound, locale.TR(r, "protection_not_found")},
		}) {
			return
		}

		log.Error().
			Err(err).
			Str("workspaceId", wid).
			Str("agent", agentName).
			Msg("Failed to get agent protection")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_protection_request"),
		})

		return
	}

	a.WriteJSON(w, http.StatusOK, protection)
}

// SaveProtectionAction sets or replaces the protection of an agent; a generated secret is returned only once.
func (a *API) SaveProtectionAction(w http.ResponseWriter, r *http.Request) {
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

	var req module.SaveProtectionRequest
	err := a.DecodeAndValidate(r, &req)
	if err != nil {
		a.WriteValidationError(w, err)
		return
	}

	protection, err := a.Protection.SaveProtection(db.Id(wid), agentName, &req)
	if err != nil {
		if a.WriteError(w, err, []ErrorMapping{
			{module.ErrAgentNotFound, http.StatusNotFound, locale.TR(r, "agent_not_found")},
			{module.ErrInvalidProtectionConfig, http.StatusBadRequest, locale.TR(r, "invalid_protection_config")},
		}) {
			return
		}

		log.Error().
			Err(err).
			Str("workspaceId", wid).
			Str("agent", agentName).
			Msg("Failed to save agent protection")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_protection_request"),
		})

		return
	}

	log.Info().
		Str("workspaceId", wid).
		Str("agent", agentName).
		Str("protectionId", protection.Id.String()).
		Msg("Agent protection saved")

	a.WriteJSON(w, http.StatusOK, protection)
}

// UpdateProtectionStatusAction turns the protection of an agent on or off.
func (a *API) UpdateProtectionStatusAction(w http.ResponseWriter, r *http.Request) {
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

	var req module.UpdateProtectionStatusRequest
	err := a.DecodeAndValidate(r, &req)
	if err != nil {
		a.WriteValidationError(w, err)
		return
	}

	protection, err := a.Protection.UpdateProtectionStatus(db.Id(wid), agentName, &req)
	if err != nil {
		if a.WriteError(w, err, []ErrorMapping{
			{module.ErrAgentNotFound, http.StatusNotFound, locale.TR(r, "agent_not_found")},
			{module.ErrProtectionNotFound, http.StatusNotFound, locale.TR(r, "protection_not_found")},
		}) {
			return
		}

		log.Error().
			Err(err).
			Str("workspaceId", wid).
			Str("agent", agentName).
			Msg("Failed to update agent protection status")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_protection_request"),
		})

		return
	}

	a.WriteJSON(w, http.StatusOK, protection)
}

// DeleteProtectionAction removes the protection of an agent.
func (a *API) DeleteProtectionAction(w http.ResponseWriter, r *http.Request) {
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

	err := a.Protection.DeleteProtection(db.Id(wid), agentName)
	if err != nil {
		if a.WriteError(w, err, []ErrorMapping{
			{module.ErrAgentNotFound, http.StatusNotFound, locale.TR(r, "agent_not_found")},
			{module.ErrProtectionNotFound, http.StatusNotFound, locale.TR(r, "protection_not_found")},
		}) {
			return
		}

		log.Error().
			Err(err).
			Str("workspaceId", wid).
			Str("agent", agentName).
			Msg("Failed to delete agent protection")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_protection_request"),
		})

		return
	}

	log.Info().
		Str("workspaceId", wid).
		Str("agent", agentName).
		Msg("Agent protection deleted")

	w.WriteHeader(http.StatusNoContent)
}
