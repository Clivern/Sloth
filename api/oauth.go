// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package api

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/clivern/sloth/db"
	"github.com/clivern/sloth/module"
	"github.com/clivern/sloth/pkg/github"
	"github.com/clivern/sloth/pkg/util"

	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

const OauthStateCookie = "_sloth_oauth_state"

// GitHubOAuthStartAction redirects the browser to GitHub's authorize URL.
func (a *API) GitHubOAuthStartAction(w http.ResponseWriter, r *http.Request) {
	state, err := util.GenerateSecureToken(24)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to generate oauth state")
		http.Redirect(w, r, util.AppURL("/login?oauth_error=github"), http.StatusFound)
		return
	}

	authorizeURL := a.GitHubOAuth.AuthorizeURL(state)

	opts := lo.Ternary(
		strings.HasPrefix(util.AppURL(""), "https://"),
		util.SecureCookieOptions(),
		util.DefaultCookieOptions(),
	)
	opts.SameSite = http.SameSiteLaxMode
	opts.MaxAge = int((10 * time.Minute) / time.Second)
	util.SetCookie(w, OauthStateCookie, state, opts)

	http.Redirect(w, r, authorizeURL, http.StatusFound)
}

// GitHubOAuthCallbackAction completes GitHub OAuth and creates a session.
func (a *API) GitHubOAuthCallbackAction(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	expectedState := util.GetCookie(r, OauthStateCookie)

	util.DeleteCookie(w, OauthStateCookie)

	token, err := a.GitHubOAuth.Exchange(r.Context(), code, state, expectedState)
	if err != nil {
		log.Error().
			Err(err).
			Msg("GitHub oauth exchange failed")
		http.Redirect(w, r, util.AppURL("/login?oauth_error=github"), http.StatusFound)
		return
	}

	user, err := a.GitHubOAuth.User(r.Context(), token.AccessToken)
	if err != nil {
		log.Error().
			Err(err).
			Msg("GitHub oauth user fetch failed")
		http.Redirect(w, r, util.AppURL("/login?oauth_error=github"), http.StatusFound)
		return
	}

	emails, err := a.GitHubOAuth.Emails(r.Context(), token.AccessToken)
	if err != nil {
		log.Error().
			Err(err).
			Msg("GitHub oauth emails fetch failed")
		http.Redirect(w, r, util.AppURL("/login?oauth_error=github"), http.StatusFound)
		return
	}

	result, err := a.Auth.LoginWithOAuth(r.Context(), &module.OAuthIdentity{
		Provider:       db.UserProviderGithub,
		ProviderUserID: strconv.FormatInt(user.ID, 10),
		Email:          github.PrimaryEmail(emails, user.Email),
		Name: lo.Ternary(
			lo.IsNotEmpty(user.Name),
			user.Name,
			user.Login,
		),
	})
	if err != nil {
		log.Error().
			Err(err).
			Msg("GitHub oauth login failed")
		http.Redirect(w, r, util.AppURL("/login?oauth_error=github"), http.StatusFound)
		return
	}

	util.SetCookie(w, "_sloth_session", result.Session.Token, result.CookieOptions)

	err = a.Invite.AttachPending(result.User)
	if err != nil {
		log.Error().
			Err(err).
			Str("userId", result.User.Id.String()).
			Msg("Failed to attach pending workspace invites")
	}

	http.Redirect(w, r, util.AppURL("/login?oauth=github"), http.StatusFound)
}

// GoogleOAuthStartAction redirects the browser to Google's authorize URL.
func (a *API) GoogleOAuthStartAction(w http.ResponseWriter, r *http.Request) {
	state, err := util.GenerateSecureToken(24)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to generate oauth state")
		http.Redirect(w, r, util.AppURL("/login?oauth_error=google"), http.StatusFound)
		return
	}

	authorizeURL := a.GoogleOAuth.AuthorizeURL(state)

	opts := lo.Ternary(
		strings.HasPrefix(util.AppURL(""), "https://"),
		util.SecureCookieOptions(),
		util.DefaultCookieOptions(),
	)
	opts.SameSite = http.SameSiteLaxMode
	opts.MaxAge = int((10 * time.Minute) / time.Second)
	util.SetCookie(w, OauthStateCookie, state, opts)

	http.Redirect(w, r, authorizeURL, http.StatusFound)
}

// GoogleOAuthCallbackAction completes Google OAuth and creates a session.
func (a *API) GoogleOAuthCallbackAction(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	expectedState := util.GetCookie(r, OauthStateCookie)

	util.DeleteCookie(w, OauthStateCookie)

	token, err := a.GoogleOAuth.Exchange(r.Context(), code, state, expectedState)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Google oauth exchange failed")
		http.Redirect(w, r, util.AppURL("/login?oauth_error=google"), http.StatusFound)
		return
	}

	user, err := a.GoogleOAuth.User(r.Context(), token.AccessToken)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Google oauth user fetch failed")
		http.Redirect(w, r, util.AppURL("/login?oauth_error=google"), http.StatusFound)
		return
	}

	result, err := a.Auth.LoginWithOAuth(r.Context(), &module.OAuthIdentity{
		Provider:       db.UserProviderGoogle,
		ProviderUserID: user.Sub,
		Email:          user.Email,
		Name: lo.Ternary(
			lo.IsNotEmpty(user.Name),
			user.Name,
			user.Email,
		),
	})
	if err != nil {
		log.Error().
			Err(err).
			Msg("Google oauth login failed")
		http.Redirect(w, r, util.AppURL("/login?oauth_error=google"), http.StatusFound)
		return
	}

	util.SetCookie(w, "_sloth_session", result.Session.Token, result.CookieOptions)

	err = a.Invite.AttachPending(result.User)
	if err != nil {
		log.Error().
			Err(err).
			Str("userId", result.User.Id.String()).
			Msg("Failed to attach pending workspace invites")
	}

	http.Redirect(w, r, util.AppURL("/login?oauth=google"), http.StatusFound)
}
