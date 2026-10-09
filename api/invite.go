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

// CreateInviteAction creates a new user invite.
func (a *API) CreateInviteAction(w http.ResponseWriter, r *http.Request) {
	user, ok := a.GetUser(r)
	if !ok || user == nil {
		a.WriteJSON(w, http.StatusUnauthorized, map[string]any{
			"errorMessage": locale.TR(r, "not_authenticated"),
		})
		return
	}

	workspaceId := chi.URLParam(r, "workspaceId")
	if lo.IsEmpty(workspaceId) {
		a.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"errorMessage": locale.TR(r, "invalid_workspace_id"),
		})
		return
	}

	log.Info().
		Str("userId", user.Id.String()).
		Str("workspaceId", workspaceId).
		Msg("New invite request")

	var req module.CreateInviteRequest
	err := a.DecodeAndValidate(r, &req)
	if err != nil {
		a.WriteValidationError(w, err)
		return
	}

	invite, err := a.Invite.CreateInvite(db.Id(workspaceId), &req, user)
	if err != nil {
		switch {
		case errors.Is(err, module.ErrWorkspaceNotFound):
			a.WriteJSON(w, http.StatusNotFound, map[string]any{
				"errorMessage": locale.TR(r, "workspace_not_found"),
			})
		case errors.Is(err, module.ErrUserAlreadyInWorkspace):
			a.WriteJSON(w, http.StatusConflict, map[string]any{
				"errorMessage": locale.TR(r, "user_already_workspace_member"),
			})
		case errors.Is(err, module.ErrPendingInviteExists):
			a.WriteJSON(w, http.StatusConflict, map[string]any{
				"errorMessage": locale.TR(r, "pending_invite_already_exists"),
			})
		default:
			log.Error().
				Err(err).
				Str("userId", user.Id.String()).
				Str("workspaceId", workspaceId).
				Msg("Failed to create invite")
			a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
				"errorMessage": locale.TR(r, "failed_create_invite"),
			})
		}
		return
	}

	log.Info().
		Str("inviteId", invite.Id.String()).
		Str("userId", user.Id.String()).
		Str("workspaceId", workspaceId).
		Str("email", invite.Email).
		Msg("Invite created")

	a.WriteJSON(w, http.StatusCreated, invite)
}

// ListInvitesAction returns invites for a workspace.
func (a *API) ListInvitesAction(w http.ResponseWriter, r *http.Request) {
	workspaceId := chi.URLParam(r, "workspaceId")
	if lo.IsEmpty(workspaceId) {
		a.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"errorMessage": locale.TR(r, "invalid_workspace_id"),
		})
		return
	}

	log.Info().
		Str("workspaceId", workspaceId).
		Msg("Listing invites")

	limit, offset := a.ParsePagination(r)

	result, err := a.Invite.ListInvites(db.Id(workspaceId), limit, offset)
	if err != nil {
		switch {
		case errors.Is(err, module.ErrWorkspaceNotFound):
			a.WriteJSON(w, http.StatusNotFound, map[string]any{
				"errorMessage": locale.TR(r, "workspace_not_found"),
			})
		default:
			log.Error().
				Err(err).
				Str("workspaceId", workspaceId).
				Msg("Failed to list invites")
			a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
				"errorMessage": locale.TR(r, "failed_list_invites"),
			})
		}
		return
	}

	a.WriteJSON(w, http.StatusOK, map[string]any{
		"invites": result.Invites,
		"_meta": map[string]any{
			"limit":  limit,
			"offset": offset,
			"total":  result.Total,
		},
	})
}

// GetInviteAction returns one invite by Id.
func (a *API) GetInviteAction(w http.ResponseWriter, r *http.Request) {
	workspaceId := chi.URLParam(r, "workspaceId")
	if lo.IsEmpty(workspaceId) {
		a.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"errorMessage": locale.TR(r, "invalid_workspace_id"),
		})
		return
	}

	inviteId := chi.URLParam(r, "inviteId")
	if lo.IsEmpty(inviteId) {
		a.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"errorMessage": locale.TR(r, "invalid_invite_id"),
		})
		return
	}

	log.Info().
		Str("inviteId", inviteId).
		Str("workspaceId", workspaceId).
		Msg("Getting invite")

	invite, err := a.Invite.GetInvite(db.Id(workspaceId), db.Id(inviteId))

	if err != nil {
		switch {
		case errors.Is(err, module.ErrWorkspaceNotFound):
			a.WriteJSON(w, http.StatusNotFound, map[string]any{
				"errorMessage": locale.TR(r, "workspace_not_found"),
			})
		case errors.Is(err, module.ErrInviteNotFound):
			a.WriteJSON(w, http.StatusNotFound, map[string]any{
				"errorMessage": locale.TR(r, "invite_not_found"),
			})
		default:
			log.Error().
				Err(err).
				Str("inviteId", inviteId).
				Str("workspaceId", workspaceId).
				Msg("Failed to get invite")
			a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
				"errorMessage": locale.TR(r, "failed_get_invite"),
			})
		}
		return
	}

	a.WriteJSON(w, http.StatusOK, invite)
}

// DeleteInviteAction deletes an invite by Id.
func (a *API) DeleteInviteAction(w http.ResponseWriter, r *http.Request) {
	workspaceId := chi.URLParam(r, "workspaceId")
	if lo.IsEmpty(workspaceId) {
		a.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"errorMessage": locale.TR(r, "invalid_workspace_id"),
		})
		return
	}

	inviteId := chi.URLParam(r, "inviteId")
	if lo.IsEmpty(inviteId) {
		a.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"errorMessage": locale.TR(r, "invalid_invite_id"),
		})
		return
	}

	log.Info().
		Str("inviteId", inviteId).
		Str("workspaceId", workspaceId).
		Msg("Deleting invite")

	err := a.Invite.DeleteInvite(db.Id(workspaceId), db.Id(inviteId))

	if err != nil {
		switch {
		case errors.Is(err, module.ErrWorkspaceNotFound):
			a.WriteJSON(w, http.StatusNotFound, map[string]any{
				"errorMessage": locale.TR(r, "workspace_not_found"),
			})
		case errors.Is(err, module.ErrInviteNotFound):
			a.WriteJSON(w, http.StatusNotFound, map[string]any{
				"errorMessage": locale.TR(r, "invite_not_found"),
			})
		default:
			log.Error().
				Err(err).
				Str("inviteId", inviteId).
				Str("workspaceId", workspaceId).
				Msg("Failed to delete invite")
			a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
				"errorMessage": locale.TR(r, "failed_delete_invite"),
			})
		}
		return
	}

	log.Info().
		Str("inviteId", inviteId).
		Str("workspaceId", workspaceId).
		Msg("Invite deleted")

	w.WriteHeader(http.StatusNoContent)
}
