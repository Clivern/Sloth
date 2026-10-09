// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package api

import (
	"errors"
	"net/http"
	"time"

	"github.com/clivern/cognit/locale"
	"github.com/clivern/cognit/module"
	"github.com/clivern/cognit/pkg/util"

	"github.com/rs/zerolog/log"
)

// GetProfileAction returns the current user's profile.
func (a *API) GetProfileAction(w http.ResponseWriter, r *http.Request) {
	user, ok := a.GetUser(r)
	if !ok || user == nil {
		a.WriteJSON(w, http.StatusUnauthorized, map[string]any{
			"errorMessage": locale.TR(r, "not_authenticated"),
		})
		return
	}

	log.Info().
		Str("userId", user.Id.String()).
		Msg("Getting profile")

	profile, err := a.Profile.GetProfile(user.Id)

	if err != nil {
		log.Error().
			Err(err).
			Str("userId", user.Id.String()).
			Msg("Failed to get profile")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_get_profile"),
		})
		return
	}

	a.WriteJSON(w, http.StatusOK, map[string]any{
		"user": map[string]any{
			"id":              profile.Id,
			"email":           profile.Email,
			"role":            profile.Role,
			"isActive":        profile.IsActive,
			"isEmailVerified": profile.IsEmailVerified,
			"name":            profile.Name,
			"avatar":          util.GravatarURL(profile.Email, 80),
			"language":        profile.Language,
			"theme":           profile.Theme,
			"lastLoginAt":     profile.LastLoginAt.UTC().Format(time.RFC3339),
			"createdAt":       profile.CreatedAt.UTC().Format(time.RFC3339),
			"updatedAt":       profile.UpdatedAt.UTC().Format(time.RFC3339),
		},
	})
}

// UpdateProfileAction updates the current user's profile.
func (a *API) UpdateProfileAction(w http.ResponseWriter, r *http.Request) {
	var req module.UpdateProfileRequest

	user, ok := a.GetUser(r)
	if !ok || user == nil {
		a.WriteJSON(w, http.StatusUnauthorized, map[string]any{
			"errorMessage": locale.TR(r, "not_authenticated"),
		})
		return
	}

	log.Info().
		Str("userId", user.Id.String()).
		Msg("Updating profile")

	err := a.DecodeAndValidate(r, &req)
	if err != nil {
		a.WriteValidationError(w, err)
		return
	}

	updated, err := a.Profile.UpdateProfile(user.Id, &req)

	if err != nil {
		switch {
		case errors.Is(err, module.ErrFailedUpdateProfile), errors.Is(err, module.ErrUserNotFound):
			log.Error().
				Err(err).
				Str("userId", user.Id.String()).
				Msg("Failed to update profile")
			a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
				"errorMessage": locale.TR(r, "failed_update_profile"),
			})
		default:
			log.Error().
				Err(err).
				Str("userId", user.Id.String()).
				Msg("Failed to update profile")
			a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
				"errorMessage": locale.TR(r, "failed_update_profile"),
			})
		}
		return
	}

	log.Info().
		Str("userId", updated.Id.String()).
		Msg("Profile updated")

	a.WriteJSON(w, http.StatusOK, map[string]any{
		"user": map[string]any{
			"id":              updated.Id,
			"email":           updated.Email,
			"role":            updated.Role,
			"isActive":        updated.IsActive,
			"isEmailVerified": updated.IsEmailVerified,
			"name":            updated.Name,
			"avatar":          util.GravatarURL(updated.Email, 80),
			"language":        updated.Language,
			"theme":           updated.Theme,
			"lastLoginAt":     updated.LastLoginAt.UTC().Format(time.RFC3339),
			"createdAt":       updated.CreatedAt.UTC().Format(time.RFC3339),
			"updatedAt":       updated.UpdatedAt.UTC().Format(time.RFC3339),
		},
		"successMessage": locale.TR(r, "profile_updated_successfully"),
	})
}
