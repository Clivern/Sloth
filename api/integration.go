// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package api

import (
	"net/http"

	"github.com/clivern/sloth/db"
	"github.com/clivern/sloth/locale"
	"github.com/clivern/sloth/module"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

// ListIntegrationsAction lists integrations in a workspace.
func (a *API) ListIntegrationsAction(w http.ResponseWriter, r *http.Request) {
	wid := chi.URLParam(r, "workspaceId")
	if lo.IsEmpty(wid) {
		a.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"errorMessage": locale.TR(r, "invalid_workspace_id"),
		})
		return
	}

	limit, offset := a.ParsePagination(r)

	result, err := a.Integration.ListIntegrations(db.Id(wid), limit, offset)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", wid).
			Msg("Failed to list integrations")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_integration_request"),
		})

		return
	}

	a.WriteJSON(w, http.StatusOK, map[string]any{
		"integrations": result.Integrations,
		"_meta": map[string]any{
			"limit":  limit,
			"offset": offset,
			"total":  result.Total,
		},
	})
}

// GetIntegrationAction returns one integration by id.
func (a *API) GetIntegrationAction(w http.ResponseWriter, r *http.Request) {
	wid := chi.URLParam(r, "workspaceId")
	if lo.IsEmpty(wid) {
		a.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"errorMessage": locale.TR(r, "invalid_workspace_id"),
		})
		return
	}

	integrationId := chi.URLParam(r, "integrationId")
	if lo.IsEmpty(integrationId) {
		a.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"errorMessage": locale.TR(r, "invalid_integration_id"),
		})
		return
	}

	integration, err := a.Integration.GetIntegration(db.Id(wid), db.Id(integrationId))
	if err != nil {
		if a.WriteError(w, err, []ErrorMapping{
			{module.ErrIntegrationNotFound, http.StatusNotFound, locale.TR(r, "integration_not_found")},
		}) {
			return
		}

		log.Error().
			Err(err).
			Str("workspaceId", wid).
			Str("integrationId", integrationId).
			Msg("Failed to get integration")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_integration_request"),
		})

		return
	}

	a.WriteJSON(w, http.StatusOK, integration)
}

// CreateIntegrationAction stores an event destination.
func (a *API) CreateIntegrationAction(w http.ResponseWriter, r *http.Request) {
	wid := chi.URLParam(r, "workspaceId")
	if lo.IsEmpty(wid) {
		a.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"errorMessage": locale.TR(r, "invalid_workspace_id"),
		})
		return
	}

	var req module.SaveIntegrationRequest
	err := a.DecodeAndValidate(r, &req)
	if err != nil {
		a.WriteValidationError(w, err)
		return
	}

	log.Info().
		Str("workspaceId", wid).
		Str("type", req.Type).
		Msg("Creating integration")

	integration, err := a.Integration.CreateIntegration(db.Id(wid), &req)
	if err != nil {
		if a.WriteError(w, err, []ErrorMapping{
			{module.ErrInvalidIntegrationConfig, http.StatusBadRequest, locale.TR(r, "invalid_integration_config")},
		}) {
			return
		}

		log.Error().
			Err(err).
			Str("workspaceId", wid).
			Msg("Failed to create integration")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_integration_request"),
		})

		return
	}

	a.WriteJSON(w, http.StatusCreated, integration)
}

// UpdateIntegrationAction replaces an event destination.
func (a *API) UpdateIntegrationAction(w http.ResponseWriter, r *http.Request) {
	wid := chi.URLParam(r, "workspaceId")
	if lo.IsEmpty(wid) {
		a.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"errorMessage": locale.TR(r, "invalid_workspace_id"),
		})
		return
	}

	integrationId := chi.URLParam(r, "integrationId")
	if lo.IsEmpty(integrationId) {
		a.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"errorMessage": locale.TR(r, "invalid_integration_id"),
		})
		return
	}

	var req module.SaveIntegrationRequest
	err := a.DecodeAndValidate(r, &req)
	if err != nil {
		a.WriteValidationError(w, err)
		return
	}

	log.Info().
		Str("workspaceId", wid).
		Str("integrationId", integrationId).
		Str("type", req.Type).
		Msg("Updating integration")

	integration, err := a.Integration.UpdateIntegration(db.Id(wid), db.Id(integrationId), &req)
	if err != nil {
		if a.WriteError(w, err, []ErrorMapping{
			{module.ErrIntegrationNotFound, http.StatusNotFound, locale.TR(r, "integration_not_found")},
			{module.ErrInvalidIntegrationConfig, http.StatusBadRequest, locale.TR(r, "invalid_integration_config")},
		}) {
			return
		}

		log.Error().
			Err(err).
			Str("workspaceId", wid).
			Str("integrationId", integrationId).
			Msg("Failed to update integration")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_integration_request"),
		})

		return
	}

	a.WriteJSON(w, http.StatusOK, integration)
}

// DeleteIntegrationAction removes an integration.
func (a *API) DeleteIntegrationAction(w http.ResponseWriter, r *http.Request) {
	wid := chi.URLParam(r, "workspaceId")
	if lo.IsEmpty(wid) {
		a.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"errorMessage": locale.TR(r, "invalid_workspace_id"),
		})
		return
	}

	integrationId := chi.URLParam(r, "integrationId")
	if lo.IsEmpty(integrationId) {
		a.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"errorMessage": locale.TR(r, "invalid_integration_id"),
		})
		return
	}

	log.Info().
		Str("workspaceId", wid).
		Str("integrationId", integrationId).
		Msg("Deleting integration")

	err := a.Integration.DeleteIntegration(db.Id(wid), db.Id(integrationId))
	if err != nil {
		if a.WriteError(w, err, []ErrorMapping{
			{module.ErrIntegrationNotFound, http.StatusNotFound, locale.TR(r, "integration_not_found")},
		}) {
			return
		}

		log.Error().
			Err(err).
			Str("workspaceId", wid).
			Str("integrationId", integrationId).
			Msg("Failed to delete integration")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_integration_request"),
		})

		return
	}

	w.WriteHeader(http.StatusNoContent)
}
