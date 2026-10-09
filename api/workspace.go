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

// CreateWorkspaceAction creates a new workspace.
func (a *API) CreateWorkspaceAction(w http.ResponseWriter, r *http.Request) {
	var req module.CreateWorkspaceRequest
	err := a.DecodeAndValidate(r, &req)
	if err != nil {
		a.WriteValidationError(w, err)
		return
	}

	user, ok := a.GetUser(r)
	if !ok || user == nil {
		a.WriteJSON(w, http.StatusUnauthorized, map[string]any{
			"errorMessage": locale.TR(r, "not_authenticated"),
		})
		return
	}

	log.Info().
		Str("userId", user.Id.String()).
		Msg("New workspace request")

	workspace, err := a.Workspace.CreateWorkspace(&req, user)
	if err != nil {
		log.Error().
			Err(err).
			Str("userId", user.Id.String()).
			Msg("Failed to create workspace")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_create_workspace"),
		})
		return
	}

	log.Info().
		Str("workspaceId", workspace.Id.String()).
		Str("userId", user.Id.String()).
		Msg("Workspace created")

	a.WriteJSON(w, http.StatusCreated, workspace)
}

// ListWorkspacesAction returns workspaces the user is a member of (paginated).
func (a *API) ListWorkspacesAction(w http.ResponseWriter, r *http.Request) {
	user, ok := a.GetUser(r)
	if !ok || user == nil {
		a.WriteJSON(w, http.StatusUnauthorized, map[string]any{
			"errorMessage": locale.TR(r, "not_authenticated"),
		})
		return
	}

	log.Info().
		Str("userId", user.Id.String()).
		Msg("Listing workspaces")

	limit, offset := a.ParsePagination(r)

	result, err := a.Workspace.ListWorkspaces(user, limit, offset)
	if err != nil {
		log.Error().
			Err(err).
			Str("userId", user.Id.String()).
			Msg("Failed to list workspaces")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_list_workspaces"),
		})
		return
	}

	a.WriteJSON(w, http.StatusOK, map[string]any{
		"workspaces": result.Workspaces,
		"_meta": map[string]any{
			"limit":  limit,
			"offset": offset,
			"total":  result.Total,
		},
	})
}

// GetWorkspaceAction returns one workspace by Id.
func (a *API) GetWorkspaceAction(w http.ResponseWriter, r *http.Request) {
	user, ok := a.GetUser(r)
	if !ok || user == nil {
		a.WriteJSON(w, http.StatusUnauthorized, map[string]any{
			"errorMessage": locale.TR(r, "not_authenticated"),
		})
		return
	}

	wid := chi.URLParam(r, "workspaceId")
	if lo.IsEmpty(wid) {
		a.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"errorMessage": locale.TR(r, "invalid_workspace_id"),
		})
		return
	}

	log.Info().
		Str("workspaceId", wid).
		Str("userId", user.Id.String()).
		Msg("Getting workspace")

	workspace, err := a.Workspace.GetWorkspace(db.Id(wid), user)
	if err != nil {
		switch {
		case errors.Is(err, module.ErrWorkspaceNotFound):
			a.WriteJSON(w, http.StatusNotFound, map[string]any{
				"errorMessage": locale.TR(r, "workspace_not_found"),
			})
			return
		default:
			log.Error().
				Err(err).
				Str("workspaceId", wid).
				Str("userId", user.Id.String()).
				Msg("Failed to get workspace")
			a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
				"errorMessage": locale.TR(r, "failed_get_workspace"),
			})
			return
		}
	}

	a.WriteJSON(w, http.StatusOK, workspace)
}

// UpdateWorkspaceAction updates a workspace.
func (a *API) UpdateWorkspaceAction(w http.ResponseWriter, r *http.Request) {
	user, ok := a.GetUser(r)
	if !ok || user == nil {
		a.WriteJSON(w, http.StatusUnauthorized, map[string]any{
			"errorMessage": locale.TR(r, "not_authenticated"),
		})
		return
	}

	wid := chi.URLParam(r, "workspaceId")
	if lo.IsEmpty(wid) {
		a.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"errorMessage": locale.TR(r, "invalid_workspace_id"),
		})
		return
	}

	log.Info().
		Str("workspaceId", wid).
		Str("userId", user.Id.String()).
		Msg("Updating workspace")

	var req module.UpdateWorkspaceRequest
	err := a.DecodeAndValidate(r, &req)
	if err != nil {
		a.WriteValidationError(w, err)
		return
	}

	workspace, err := a.Workspace.UpdateWorkspace(db.Id(wid), &req, user)
	if err != nil {
		switch {
		case errors.Is(err, module.ErrWorkspaceNotFound):
			a.WriteJSON(w, http.StatusNotFound, map[string]any{
				"errorMessage": locale.TR(r, "workspace_not_found"),
			})
			return
		default:
			log.Error().
				Err(err).
				Str("workspaceId", wid).
				Str("userId", user.Id.String()).
				Msg("Failed to update workspace")
			a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
				"errorMessage": locale.TR(r, "failed_update_workspace"),
			})
			return
		}
	}

	log.Info().
		Str("workspaceId", workspace.Id.String()).
		Str("userId", user.Id.String()).
		Msg("Workspace updated")

	a.WriteJSON(w, http.StatusOK, workspace)
}

// DeleteWorkspaceAction deletes a workspace.
func (a *API) DeleteWorkspaceAction(w http.ResponseWriter, r *http.Request) {
	wid := chi.URLParam(r, "workspaceId")
	if lo.IsEmpty(wid) {
		a.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"errorMessage": locale.TR(r, "invalid_workspace_id"),
		})
		return
	}

	log.Info().
		Str("workspaceId", wid).
		Msg("Deleting workspace")

	err := a.Workspace.DeleteWorkspace(db.Id(wid))
	if err != nil {
		switch {
		case errors.Is(err, module.ErrWorkspaceNotFound):
			a.WriteJSON(w, http.StatusNotFound, map[string]any{
				"errorMessage": locale.TR(r, "workspace_not_found"),
			})
			return
		default:
			log.Error().
				Err(err).
				Str("workspaceId", wid).
				Msg("Failed to delete workspace")
			a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
				"errorMessage": locale.TR(r, "failed_delete_workspace"),
			})
			return
		}
	}

	log.Info().
		Str("workspaceId", wid).
		Msg("Workspace deleted")

	a.WriteJSON(w, http.StatusNoContent, map[string]any{})
}
