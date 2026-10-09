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

// CreateUserAPIKeyAction creates an API key; raw key is returned only once.
func (a *API) CreateUserAPIKeyAction(w http.ResponseWriter, r *http.Request) {
	user, ok := a.GetUser(r)
	if !ok || user == nil {
		a.WriteJSON(w, http.StatusUnauthorized, map[string]any{
			"errorMessage": locale.TR(r, "not_authenticated"),
		})
		return
	}

	log.Info().
		Str("userId", user.Id.String()).
		Msg("New API key request")

	var req module.CreateAPIKeyRequest
	err := a.DecodeAndValidate(r, &req)
	if err != nil {
		a.WriteValidationError(w, err)
		return
	}

	apiKey, err := a.APIKey.CreateAPIKey(&req, user)
	if err != nil {
		switch {
		case errors.Is(err, module.ErrInvalidExpiresAt):
			a.WriteJSON(w, http.StatusBadRequest, map[string]any{
				"errorMessage": locale.TR(r, "invalid_expires_at_format"),
			})
			return
		default:
			log.Error().
				Err(err).
				Str("userId", user.Id.String()).
				Msg("Failed to create API key")
			a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
				"errorMessage": locale.TR(r, "failed_create_api_key"),
			})
			return
		}
	}

	log.Info().
		Str("userId", user.Id.String()).
		Str("apiKeyId", apiKey.Id.String()).
		Msg("API key created")

	a.WriteJSON(w, http.StatusCreated, apiKey)
}

// ListUserAPIKeysAction lists your API keys (metadata only, never the secret).
func (a *API) ListUserAPIKeysAction(w http.ResponseWriter, r *http.Request) {
	user, ok := a.GetUser(r)
	if !ok || user == nil {
		a.WriteJSON(w, http.StatusUnauthorized, map[string]any{
			"errorMessage": locale.TR(r, "not_authenticated"),
		})
		return
	}

	log.Info().
		Str("userId", user.Id.String()).
		Msg("Listing API keys")

	limit, offset := a.ParsePagination(r)

	result, err := a.APIKey.ListAPIKeys(user, limit, offset)
	if err != nil {
		log.Error().
			Err(err).
			Str("userId", user.Id.String()).
			Msg("Failed to list API keys")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_list_api_keys"),
		})
		return
	}

	a.WriteJSON(w, http.StatusOK, map[string]any{
		"keys": result.APIKeys,
		"_meta": map[string]any{
			"limit":  limit,
			"offset": offset,
			"total":  result.Total,
		},
	})
}

// GetUserAPIKeyAction returns one API key's metadata (never the key itself).
func (a *API) GetUserAPIKeyAction(w http.ResponseWriter, r *http.Request) {
	user, ok := a.GetUser(r)
	if !ok || user == nil {
		a.WriteJSON(w, http.StatusUnauthorized, map[string]any{
			"errorMessage": locale.TR(r, "not_authenticated"),
		})
		return
	}

	apiKeyId := chi.URLParam(r, "apiKeyId")
	if lo.IsEmpty(apiKeyId) {
		a.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"errorMessage": locale.TR(r, "invalid_api_key_id"),
		})
		return
	}

	log.Info().
		Str("apiKeyId", apiKeyId).
		Str("userId", user.Id.String()).
		Msg("Getting API key")

	k, err := a.APIKey.GetAPIKey(db.Id(apiKeyId), user)
	if err != nil {
		switch {
		case errors.Is(err, module.ErrAPIKeyNotFound):
			a.WriteJSON(w, http.StatusNotFound, map[string]any{
				"errorMessage": locale.TR(r, "api_key_not_found"),
			})
			return
		default:
			log.Error().
				Err(err).
				Str("apiKeyId", apiKeyId).
				Str("userId", user.Id.String()).
				Msg("Failed to get API key")
			a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
				"errorMessage": locale.TR(r, "failed_get_api_key"),
			})
			return
		}
	}

	a.WriteJSON(w, http.StatusOK, k)
}

// DeleteUserAPIKeyAction deletes one of your API keys.
func (a *API) DeleteUserAPIKeyAction(w http.ResponseWriter, r *http.Request) {
	user, ok := a.GetUser(r)
	if !ok || user == nil {
		a.WriteJSON(w, http.StatusUnauthorized, map[string]any{
			"errorMessage": locale.TR(r, "not_authenticated"),
		})
		return
	}

	apiKeyId := chi.URLParam(r, "apiKeyId")
	if lo.IsEmpty(apiKeyId) {
		a.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"errorMessage": locale.TR(r, "invalid_api_key_id"),
		})
		return
	}

	log.Info().
		Str("apiKeyId", apiKeyId).
		Str("userId", user.Id.String()).
		Msg("Deleting API key")

	err := a.APIKey.DeleteAPIKey(db.Id(apiKeyId), user)
	if err != nil {
		switch {
		case errors.Is(err, module.ErrAPIKeyNotFound):
			a.WriteJSON(w, http.StatusNotFound, map[string]any{
				"errorMessage": locale.TR(r, "api_key_not_found"),
			})
			return
		default:
			log.Error().
				Err(err).
				Str("apiKeyId", apiKeyId).
				Str("userId", user.Id.String()).
				Msg("Failed to delete API key")
			a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
				"errorMessage": locale.TR(r, "failed_delete_api_key"),
			})
			return
		}
	}

	log.Info().
		Str("userId", user.Id.String()).
		Str("apiKeyId", apiKeyId).
		Msg("API key deleted")

	w.WriteHeader(http.StatusNoContent)
}
