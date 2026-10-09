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
)

// CreateAccessKeyAction creates a workspace access key; raw key is returned only once.
func (a *API) CreateAccessKeyAction(w http.ResponseWriter, r *http.Request) {
	wid := chi.URLParam(r, "workspaceId")
	if wid == "" {
		a.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"errorMessage": locale.TR(r, "invalid_workspace_id"),
		})
		return
	}

	log.Info().
		Str("workspaceId", wid).
		Msg("New access key request")

	var req module.CreateAccessKeyRequest
	err := a.DecodeAndValidate(r, &req)
	if err != nil {
		a.WriteValidationError(w, err)
		return
	}

	key, err := a.Access.CreateAccessKey(db.Id(wid), &req)
	if err != nil {
		switch {
		case errors.Is(err, module.ErrWorkspaceNotFound):
			a.WriteJSON(w, http.StatusNotFound, map[string]any{
				"errorMessage": locale.TR(r, "workspace_not_found"),
			})
		case errors.Is(err, module.ErrInvalidExpiresAt):
			a.WriteJSON(w, http.StatusBadRequest, map[string]any{
				"errorMessage": locale.TR(r, "invalid_expires_at_format"),
			})
		case errors.Is(err, module.ErrInvalidAccessKeyPermissions):
			a.WriteJSON(w, http.StatusBadRequest, map[string]any{
				"errorMessage": locale.TR(r, "invalid_access_key_permissions"),
			})
		default:
			log.Error().
				Err(err).
				Str("workspaceId", wid).
				Msg("Failed to create workspace access key")
			a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
				"errorMessage": locale.TR(r, "failed_create_access_key"),
			})
		}

		return
	}

	a.WriteJSON(w, http.StatusCreated, key)
}

// ListAccessKeysAction lists workspace access keys (metadata only, never the secret).
func (a *API) ListAccessKeysAction(w http.ResponseWriter, r *http.Request) {
	wid := chi.URLParam(r, "workspaceId")
	if wid == "" {
		a.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"errorMessage": locale.TR(r, "invalid_workspace_id"),
		})
		return
	}

	log.Info().
		Str("workspaceId", wid).
		Msg("Listing access keys")

	limit, offset := a.ParsePagination(r)

	result, err := a.Access.ListAccessKeys(db.Id(wid), limit, offset)
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
				Msg("Failed to list workspace access keys")
			a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
				"errorMessage": locale.TR(r, "failed_list_access_keys"),
			})
		}

		return
	}

	a.WriteJSON(w, http.StatusOK, map[string]any{
		"keys": result.Keys,
		"_meta": map[string]any{
			"limit":  limit,
			"offset": offset,
			"total":  result.Total,
		},
	})
}

// GetAccessKeyAction returns one workspace access key (never the secret).
func (a *API) GetAccessKeyAction(w http.ResponseWriter, r *http.Request) {
	wid := chi.URLParam(r, "workspaceId")
	if wid == "" {
		a.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"errorMessage": locale.TR(r, "invalid_workspace_id"),
		})
		return
	}

	keyId := chi.URLParam(r, "keyId")
	if keyId == "" {
		a.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"errorMessage": locale.TR(r, "invalid_access_key_id"),
		})
		return
	}

	key, err := a.Access.GetAccessKey(db.Id(wid), db.Id(keyId))
	if err != nil {
		switch {
		case errors.Is(err, module.ErrWorkspaceNotFound):
			a.WriteJSON(w, http.StatusNotFound, map[string]any{
				"errorMessage": locale.TR(r, "workspace_not_found"),
			})
		case errors.Is(err, module.ErrAccessKeyNotFound):
			a.WriteJSON(w, http.StatusNotFound, map[string]any{
				"errorMessage": locale.TR(r, "access_key_not_found"),
			})
		default:
			log.Error().
				Err(err).
				Str("accessKeyId", keyId).
				Msg("Failed to get workspace access key")
			a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
				"errorMessage": locale.TR(r, "failed_get_access_key"),
			})
		}

		return
	}

	a.WriteJSON(w, http.StatusOK, key)
}

// DeleteAccessKeyAction deletes a workspace access key.
func (a *API) DeleteAccessKeyAction(w http.ResponseWriter, r *http.Request) {
	wid := chi.URLParam(r, "workspaceId")
	if wid == "" {
		a.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"errorMessage": locale.TR(r, "invalid_workspace_id"),
		})
		return
	}

	keyId := chi.URLParam(r, "keyId")
	if keyId == "" {
		a.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"errorMessage": locale.TR(r, "invalid_access_key_id"),
		})
		return
	}

	err := a.Access.DeleteAccessKey(db.Id(wid), db.Id(keyId))
	if err != nil {
		switch {
		case errors.Is(err, module.ErrWorkspaceNotFound):
			a.WriteJSON(w, http.StatusNotFound, map[string]any{
				"errorMessage": locale.TR(r, "workspace_not_found"),
			})
		case errors.Is(err, module.ErrAccessKeyNotFound):
			a.WriteJSON(w, http.StatusNotFound, map[string]any{
				"errorMessage": locale.TR(r, "access_key_not_found"),
			})
		default:
			log.Error().
				Err(err).
				Str("accessKeyId", keyId).
				Msg("Failed to delete workspace access key")
			a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
				"errorMessage": locale.TR(r, "failed_delete_access_key"),
			})
		}

		return
	}

	w.WriteHeader(http.StatusNoContent)
}
