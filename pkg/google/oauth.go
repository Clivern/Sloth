// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package google

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	OauthAuthorizeURL = "https://accounts.google.com/o/oauth2/v2/auth"
	OauthTokenURL     = "https://oauth2.googleapis.com/token"
	OauthUserURL      = "https://www.googleapis.com/oauth2/v3/userinfo"
)

// OAuthConfig holds Google OAuth app credentials.
type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
}

// OAuth exchanges authorization codes for access tokens.
type OAuth struct {
	cfg OAuthConfig
}

// Token is an OAuth access token response.
type Token struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
	IDToken     string `json:"id_token"`
}

// UserInfo is the authenticated Google user.
type UserInfo struct {
	Sub           string `json:"sub"`
	Name          string `json:"name"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Picture       string `json:"picture"`
}

// NewOAuth returns an OAuth helper for the given app configuration.
func NewOAuth(cfg OAuthConfig) *OAuth {
	return &OAuth{cfg: cfg}
}

// AuthorizeURL builds the Google authorization URL for the provided state.
func (o *OAuth) AuthorizeURL(state string) string {
	q := url.Values{
		"client_id":     {o.cfg.ClientID},
		"redirect_uri":  {o.cfg.RedirectURL},
		"response_type": {"code"},
		"scope":         {strings.Join(o.cfg.Scopes, " ")},
		"state":         {state},
		"access_type":   {"online"},
		"prompt":        {"select_account"},
	}

	return OauthAuthorizeURL + "?" + q.Encode()
}

// Exchange validates state and exchanges an authorization code for an access token.
func (o *OAuth) Exchange(ctx context.Context, code, state, expectedState string) (*Token, error) {
	if state != expectedState {
		return nil, ErrInvalidOAuthState
	}

	form := url.Values{
		"client_id":     {o.cfg.ClientID},
		"client_secret": {o.cfg.ClientSecret},
		"code":          {code},
		"grant_type":    {"authorization_code"},
		"redirect_uri":  {o.cfg.RedirectURL},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, OauthTokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("google oauth build request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("google oauth request: %w", err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("google oauth read body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google oauth token: status %d: %s", resp.StatusCode, body)
	}

	var token Token
	if err := json.Unmarshal(body, &token); err != nil {
		return nil, fmt.Errorf("google oauth decode body: %w", err)
	}

	return &token, nil
}

// User fetches the authenticated Google user.
func (o *OAuth) User(ctx context.Context, accessToken string) (*UserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, OauthUserURL, nil)
	if err != nil {
		return nil, fmt.Errorf("google oauth user request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("google oauth user: %w", err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("google oauth user body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google oauth user: status %d: %s", resp.StatusCode, body)
	}

	var user UserInfo
	if err := json.Unmarshal(body, &user); err != nil {
		return nil, fmt.Errorf("google oauth user decode: %w", err)
	}

	return &user, nil
}
