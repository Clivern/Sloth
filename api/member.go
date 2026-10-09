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

// ListWorkspaceMembersAction returns workspace members for managers.
func (a *API) ListWorkspaceMembersAction(w http.ResponseWriter, r *http.Request) {
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
		Msg("Listing workspace members")

	limit, offset := a.ParsePagination(r)

	result, err := a.Workspace.ListWorkspaceMembers(
		db.Id(wid),
		limit,
		offset,
	)
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
				Str("userId", user.Id.String()).
				Msg("Failed to list workspace members")
			a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
				"errorMessage": locale.TR(r, "failed_list_workspace_members"),
			})
		}

		return
	}

	a.WriteJSON(w, http.StatusOK, map[string]any{
		"members": result.Members,
		"_meta": map[string]any{
			"limit":  limit,
			"offset": offset,
			"total":  result.Total,
		},
	})
}

// UpdateWorkspaceMemberRoleAction updates a workspace member role.
func (a *API) UpdateWorkspaceMemberRoleAction(w http.ResponseWriter, r *http.Request) {
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

	memberUserId := chi.URLParam(r, "memberUserId")
	if lo.IsEmpty(memberUserId) {
		a.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"errorMessage": locale.TR(r, "invalid_user_id"),
		})
		return
	}

	var req module.UpdateWorkspaceMemberRoleRequest
	err := a.DecodeAndValidate(r, &req)
	if err != nil {
		a.WriteValidationError(w, err)
		return
	}

	log.Info().
		Str("workspaceId", wid).
		Str("memberUserId", memberUserId).
		Str("userId", user.Id.String()).
		Str("role", req.Role).
		Msg("Updating workspace member role")

	member, err := a.Workspace.UpdateWorkspaceMemberRole(
		db.Id(wid),
		db.Id(memberUserId),
		req.Role,
	)
	if err != nil {
		switch {
		case errors.Is(err, module.ErrWorkspaceNotFound):
			a.WriteJSON(w, http.StatusNotFound, map[string]any{
				"errorMessage": locale.TR(r, "workspace_not_found"),
			})
		case errors.Is(err, module.ErrWorkspaceUserNotFound):
			a.WriteJSON(w, http.StatusNotFound, map[string]any{
				"errorMessage": locale.TR(r, "workspace_member_not_found"),
			})
		default:
			log.Error().
				Err(err).
				Str("workspaceId", wid).
				Str("memberUserId", memberUserId).
				Str("userId", user.Id.String()).
				Msg("Failed to update workspace member role")
			a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
				"errorMessage": locale.TR(r, "failed_update_workspace_member"),
			})
		}

		return
	}

	a.WriteJSON(w, http.StatusOK, member)
}

// DeleteWorkspaceMemberAction removes a user from a workspace.
func (a *API) DeleteWorkspaceMemberAction(w http.ResponseWriter, r *http.Request) {
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

	memberUserId := chi.URLParam(r, "memberUserId")
	if lo.IsEmpty(memberUserId) {
		a.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"errorMessage": locale.TR(r, "invalid_user_id"),
		})
		return
	}

	log.Info().
		Str("workspaceId", wid).
		Str("memberUserId", memberUserId).
		Str("userId", user.Id.String()).
		Msg("Removing workspace member")

	err := a.Workspace.DeleteWorkspaceMember(
		db.Id(wid),
		db.Id(memberUserId),
	)
	if err != nil {
		switch {
		case errors.Is(err, module.ErrWorkspaceNotFound):
			a.WriteJSON(w, http.StatusNotFound, map[string]any{
				"errorMessage": locale.TR(r, "workspace_not_found"),
			})
		case errors.Is(err, module.ErrWorkspaceUserNotFound):
			a.WriteJSON(w, http.StatusNotFound, map[string]any{
				"errorMessage": locale.TR(r, "workspace_member_not_found"),
			})
		default:
			log.Error().
				Err(err).
				Str("workspaceId", wid).
				Str("memberUserId", memberUserId).
				Str("userId", user.Id.String()).
				Msg("Failed to delete workspace member")
			a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
				"errorMessage": locale.TR(r, "failed_delete_workspace_member"),
			})
		}

		return
	}

	w.WriteHeader(http.StatusNoContent)
}
