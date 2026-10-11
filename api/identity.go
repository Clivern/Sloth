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

// ListIdentitiesAction lists the identities of an agent.
func (a *API) ListIdentitiesAction(w http.ResponseWriter, r *http.Request) {
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

	identities, err := a.Identity.ListIdentities(db.Id(wid), agentName)
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
			Msg("Failed to list agent identities")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_identity_request"),
		})

		return
	}

	a.WriteJSON(w, http.StatusOK, map[string]any{
		"identities": identities,
	})
}

// CreateIdentityAction issues an identity for an agent; the secret is returned only once.
func (a *API) CreateIdentityAction(w http.ResponseWriter, r *http.Request) {
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

	var req module.CreateIdentityRequest
	err := a.DecodeAndValidate(r, &req)
	if err != nil {
		a.WriteValidationError(w, err)
		return
	}

	identity, err := a.Identity.CreateIdentity(db.Id(wid), agentName, &req)
	if err != nil {
		if a.WriteError(w, err, []ErrorMapping{
			{module.ErrAgentNotFound, http.StatusNotFound, locale.TR(r, "agent_not_found")},
			{module.ErrIdentityAlreadyExists, http.StatusConflict, locale.TR(r, "identity_already_exists")},
			{module.ErrInvalidIdentityConfig, http.StatusBadRequest, locale.TR(r, "invalid_identity_config")},
		}) {
			return
		}

		log.Error().
			Err(err).
			Str("workspaceId", wid).
			Str("agent", agentName).
			Msg("Failed to create agent identity")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_identity_request"),
		})

		return
	}

	log.Info().
		Str("workspaceId", wid).
		Str("agent", agentName).
		Str("identityId", identity.Id.String()).
		Msg("Agent identity created")

	a.WriteJSON(w, http.StatusCreated, identity)
}

// UpdateIdentityAction renames an agent identity or turns it on or off.
func (a *API) UpdateIdentityAction(w http.ResponseWriter, r *http.Request) {
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

	identityId := chi.URLParam(r, "identityId")

	var req module.UpdateIdentityRequest
	err := a.DecodeAndValidate(r, &req)
	if err != nil {
		a.WriteValidationError(w, err)
		return
	}

	identity, err := a.Identity.UpdateIdentity(db.Id(wid), agentName, db.Id(identityId), &req)
	if err != nil {
		if a.WriteError(w, err, []ErrorMapping{
			{module.ErrAgentNotFound, http.StatusNotFound, locale.TR(r, "agent_not_found")},
			{module.ErrIdentityNotFound, http.StatusNotFound, locale.TR(r, "identity_not_found")},
			{module.ErrIdentityAlreadyExists, http.StatusConflict, locale.TR(r, "identity_already_exists")},
		}) {
			return
		}

		log.Error().
			Err(err).
			Str("workspaceId", wid).
			Str("agent", agentName).
			Str("identityId", identityId).
			Msg("Failed to update agent identity")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_identity_request"),
		})

		return
	}

	log.Info().
		Str("workspaceId", wid).
		Str("agent", agentName).
		Str("identityId", identityId).
		Bool("isActive", identity.IsActive).
		Msg("Agent identity updated")

	a.WriteJSON(w, http.StatusOK, identity)
}

// DeleteIdentityAction revokes an agent identity.
func (a *API) DeleteIdentityAction(w http.ResponseWriter, r *http.Request) {
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

	identityId := chi.URLParam(r, "identityId")

	err := a.Identity.DeleteIdentity(db.Id(wid), agentName, db.Id(identityId))
	if err != nil {
		if a.WriteError(w, err, []ErrorMapping{
			{module.ErrAgentNotFound, http.StatusNotFound, locale.TR(r, "agent_not_found")},
			{module.ErrIdentityNotFound, http.StatusNotFound, locale.TR(r, "identity_not_found")},
		}) {
			return
		}

		log.Error().
			Err(err).
			Str("workspaceId", wid).
			Str("agent", agentName).
			Str("identityId", identityId).
			Msg("Failed to delete agent identity")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_identity_request"),
		})

		return
	}

	log.Info().
		Str("workspaceId", wid).
		Str("agent", agentName).
		Str("identityId", identityId).
		Msg("Agent identity deleted")

	w.WriteHeader(http.StatusNoContent)
}

// IssueIdentityTokenAction exchanges client credentials for an access token (OAuth 2.0 client credentials grant).
func (a *API) IssueIdentityTokenAction(w http.ResponseWriter, r *http.Request) {
	if r.PostFormValue("grant_type") != "client_credentials" {
		a.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"error":        "unsupported_grant_type",
			"errorMessage": locale.TR(r, "unsupported_grant_type"),
		})
		return
	}

	clientId, clientSecret, ok := r.BasicAuth()
	if !ok {
		clientId = r.PostFormValue("client_id")
		clientSecret = r.PostFormValue("client_secret")
	}

	token, err := a.Identity.IssueToken(clientId, clientSecret)
	if err != nil {
		if errors.Is(err, module.ErrInvalidClient) {
			log.Warn().
				Str("clientId", clientId).
				Msg("Rejected identity token request with invalid client")
			a.WriteJSON(w, http.StatusUnauthorized, map[string]any{
				"error":        "invalid_client",
				"errorMessage": locale.TR(r, "invalid_client"),
			})
			return
		}

		log.Error().
			Err(err).
			Str("clientId", clientId).
			Msg("Failed to issue identity token")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_identity_request"),
		})

		return
	}

	log.Info().
		Str("clientId", clientId).
		Int("expiresIn", token.ExpiresIn).
		Msg("Identity token issued")

	w.Header().Set("Cache-Control", "no-store")
	a.WriteJSON(w, http.StatusOK, token)
}
