// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package api

import (
	"net/http"

	"github.com/clivern/cognit/conf"
	"github.com/clivern/cognit/locale"
	"github.com/clivern/cognit/module"

	"github.com/rs/zerolog/log"
)

// SetupAction runs the initial platform setup.
func (a *API) SetupAction(w http.ResponseWriter, r *http.Request) {
	var req module.SetupRequest
	err := a.DecodeAndValidate(r, &req)
	if err != nil {
		a.WriteValidationError(w, err)
		return
	}

	log.Info().
		Str("platformEmail", req.PlatformEmail).
		Msg("New setup request")

	err = a.Setup.Install(&req)
	if err != nil {
		if a.WriteError(w, err, []ErrorMapping{
			{module.ErrPlatformAlreadyInstalled, http.StatusBadRequest, locale.TR(r, "platform_already_installed")},
		}) {
			return
		}

		log.Error().
			Err(err).
			Str("platformEmail", req.PlatformEmail).
			Msg("Setup failed")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_complete_setup"),
		})

		return
	}

	log.Info().
		Str("platformEmail", req.PlatformEmail).
		Msg("Platform setup completed")

	a.WriteJSON(w, http.StatusOK, map[string]any{
		"successMessage": locale.TR(r, "setup_completed_successfully"),
	})
}

// SetupStatusAction returns whether the platform is already installed.
func (a *API) SetupStatusAction(w http.ResponseWriter, _ *http.Request) {
	log.Info().
		Msg("Setup status request")

	a.WriteJSON(w, http.StatusOK, map[string]any{
		"installed": a.Setup.IsInstalled(),
		"edition":   conf.Edition(),
	})
}
