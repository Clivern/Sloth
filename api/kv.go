// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package api

import (
	"errors"
	"net/http"
	"strings"

	"github.com/clivern/cognit/db"
	"github.com/clivern/cognit/locale"
	"github.com/clivern/cognit/module"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

// ListKeyValueAction lists workspace keys, optionally filtered by prefix.
func (a *API) ListKeyValueAction(w http.ResponseWriter, r *http.Request) {
	wid := chi.URLParam(r, "workspaceId")
	if lo.IsEmpty(wid) {
		a.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"errorMessage": locale.TR(r, "invalid_workspace_id"),
		})
		return
	}

	result, err := a.KeyValue.ListKeyValue(db.Id(wid), r.URL.Query().Get("prefix"))
	if err != nil {
		switch {
		case errors.Is(err, module.ErrWorkspaceNotFound):
			a.WriteJSON(w, http.StatusNotFound, map[string]any{
				"errorMessage": locale.TR(r, "workspace_not_found"),
			})
		case errors.Is(err, module.ErrInvalidKeyValueKey):
			a.WriteJSON(w, http.StatusBadRequest, map[string]any{
				"errorMessage": locale.TR(r, "invalid_kv_key"),
			})
		default:
			log.Error().
				Err(err).
				Str("workspaceId", wid).
				Msg("Failed to list workspace keys")
			a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
				"errorMessage": locale.TR(r, "failed_kv_request"),
			})
		}

		return
	}

	a.WriteJSON(w, http.StatusOK, map[string]any{
		"items": result.Items,
		"_meta": map[string]any{
			"total": result.Total,
		},
	})
}

// GetKeyValueAction returns one workspace key. The key may contain slashes.
func (a *API) GetKeyValueAction(w http.ResponseWriter, r *http.Request) {
	wid := chi.URLParam(r, "workspaceId")
	if lo.IsEmpty(wid) {
		a.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"errorMessage": locale.TR(r, "invalid_workspace_id"),
		})
		return
	}

	key := strings.TrimPrefix(chi.URLParam(r, "*"), "/")
	if lo.IsEmpty(key) {
		a.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"errorMessage": locale.TR(r, "invalid_kv_key"),
		})
		return
	}

	item, err := a.KeyValue.GetKeyValue(db.Id(wid), key)
	if err != nil {
		switch {
		case errors.Is(err, module.ErrWorkspaceNotFound):
			a.WriteJSON(w, http.StatusNotFound, map[string]any{
				"errorMessage": locale.TR(r, "workspace_not_found"),
			})
		case errors.Is(err, module.ErrKeyValueNotFound):
			a.WriteJSON(w, http.StatusNotFound, map[string]any{
				"errorMessage": locale.TR(r, "kv_not_found"),
			})
		case errors.Is(err, module.ErrInvalidKeyValueKey):
			a.WriteJSON(w, http.StatusBadRequest, map[string]any{
				"errorMessage": locale.TR(r, "invalid_kv_key"),
			})
		default:
			log.Error().
				Err(err).
				Str("workspaceId", wid).
				Str("key", key).
				Msg("Failed to get workspace key")
			a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
				"errorMessage": locale.TR(r, "failed_kv_request"),
			})
		}

		return
	}

	a.WriteJSON(w, http.StatusOK, item)
}

// PutKeyValueAction writes a workspace key.
func (a *API) PutKeyValueAction(w http.ResponseWriter, r *http.Request) {
	wid := chi.URLParam(r, "workspaceId")
	if lo.IsEmpty(wid) {
		a.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"errorMessage": locale.TR(r, "invalid_workspace_id"),
		})
		return
	}

	key := strings.TrimPrefix(chi.URLParam(r, "*"), "/")
	if lo.IsEmpty(key) {
		a.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"errorMessage": locale.TR(r, "invalid_kv_key"),
		})
		return
	}

	var req module.PutKeyValueRequest
	err := a.DecodeAndValidate(r, &req)
	if err != nil {
		a.WriteValidationError(w, err)
		return
	}

	log.Info().
		Str("workspaceId", wid).
		Str("key", key).
		Msg("Writing workspace key")

	item, err := a.KeyValue.PutKeyValue(db.Id(wid), key, &req)
	if err != nil {
		switch {
		case errors.Is(err, module.ErrWorkspaceNotFound):
			a.WriteJSON(w, http.StatusNotFound, map[string]any{
				"errorMessage": locale.TR(r, "workspace_not_found"),
			})
		case errors.Is(err, module.ErrInvalidKeyValueKey):
			a.WriteJSON(w, http.StatusBadRequest, map[string]any{
				"errorMessage": locale.TR(r, "invalid_kv_key"),
			})
		case errors.Is(err, module.ErrInvalidExpiresAt):
			a.WriteJSON(w, http.StatusBadRequest, map[string]any{
				"errorMessage": locale.TR(r, "invalid_expires_at_format"),
			})
		default:
			log.Error().
				Err(err).
				Str("workspaceId", wid).
				Str("key", key).
				Msg("Failed to write workspace key")
			a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
				"errorMessage": locale.TR(r, "failed_kv_request"),
			})
		}

		return
	}

	a.WriteJSON(w, http.StatusOK, item)
}

// DeleteKeyValueAction removes a workspace key.
func (a *API) DeleteKeyValueAction(w http.ResponseWriter, r *http.Request) {
	wid := chi.URLParam(r, "workspaceId")
	if lo.IsEmpty(wid) {
		a.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"errorMessage": locale.TR(r, "invalid_workspace_id"),
		})
		return
	}

	key := strings.TrimPrefix(chi.URLParam(r, "*"), "/")
	if lo.IsEmpty(key) {
		a.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"errorMessage": locale.TR(r, "invalid_kv_key"),
		})
		return
	}

	log.Info().
		Str("workspaceId", wid).
		Str("key", key).
		Msg("Deleting workspace key")

	err := a.KeyValue.DeleteKeyValue(db.Id(wid), key)
	if err != nil {
		switch {
		case errors.Is(err, module.ErrWorkspaceNotFound):
			a.WriteJSON(w, http.StatusNotFound, map[string]any{
				"errorMessage": locale.TR(r, "workspace_not_found"),
			})
		case errors.Is(err, module.ErrKeyValueNotFound):
			a.WriteJSON(w, http.StatusNotFound, map[string]any{
				"errorMessage": locale.TR(r, "kv_not_found"),
			})
		case errors.Is(err, module.ErrInvalidKeyValueKey):
			a.WriteJSON(w, http.StatusBadRequest, map[string]any{
				"errorMessage": locale.TR(r, "invalid_kv_key"),
			})
		default:
			log.Error().
				Err(err).
				Str("workspaceId", wid).
				Str("key", key).
				Msg("Failed to delete workspace key")
			a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
				"errorMessage": locale.TR(r, "failed_kv_request"),
			})
		}

		return
	}

	w.WriteHeader(http.StatusNoContent)
}
