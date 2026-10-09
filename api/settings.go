// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package api

import (
	"net/http"

	"github.com/clivern/cognit/locale"

	"github.com/rs/zerolog/log"
)

// UpdateSettingsRequest is the body for updating app settings.
type UpdateSettingsRequest struct {
	PlatformEmail   string `json:"platformEmail" validate:"required,email,max=255" label:"Platform Email"`
	MaintenanceMode bool   `json:"maintenanceMode" label:"Maintenance Mode"`
}

// UpdateSettingsAction updates app settings (admin only).
func (a *API) UpdateSettingsAction(w http.ResponseWriter, r *http.Request) {
	user, ok := a.GetUser(r)
	if !ok || user == nil {
		a.WriteJSON(w, http.StatusUnauthorized, map[string]any{
			"errorMessage": locale.TR(r, "not_authenticated"),
		})
		return
	}

	log.Info().
		Str("userId", user.Id.String()).
		Msg("Updating settings")

	var req UpdateSettingsRequest
	err := a.DecodeAndValidate(r, &req)
	if err != nil {
		a.WriteValidationError(w, err)
		return
	}

	err = a.Settings.Update(req.PlatformEmail, req.MaintenanceMode)
	if err != nil {
		log.Error().
			Err(err).
			Str("userId", user.Id.String()).
			Msg("Failed to update settings")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_update_settings"),
		})
		return
	}

	log.Info().
		Str("userId", user.Id.String()).
		Msg("Settings updated")

	a.WriteJSON(w, http.StatusOK, map[string]any{
		"successMessage": locale.TR(r, "settings_updated_successfully"),
	})
}

// GetSettingsAction returns current app settings.
func (a *API) GetSettingsAction(w http.ResponseWriter, r *http.Request) {
	user, ok := a.GetUser(r)
	if !ok || user == nil {
		a.WriteJSON(w, http.StatusUnauthorized, map[string]any{
			"errorMessage": locale.TR(r, "not_authenticated"),
		})
		return
	}

	log.Info().
		Str("userId", user.Id.String()).
		Msg("Getting settings")

	settings, err := a.Settings.GetSettings()
	if err != nil {
		log.Error().
			Err(err).
			Str("userId", user.Id.String()).
			Msg("Failed to get settings")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_get_settings"),
		})
		return
	}

	a.WriteJSON(w, http.StatusOK, map[string]any{
		"settings": settings,
	})
}
