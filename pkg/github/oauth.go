// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	OauthAuthorizeURL = "https://github.com/login/oauth/authorize"
	OauthTokenURL     = "https://github.com/login/oauth/access_token"
	OauthUserURL      = "https://api.github.com/user"
	OauthEmailsURL    = "https://api.github.com/user/emails"
)

// OAuthConfig holds GitHub OAuth app credentials.
type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
	AllowSignup  bool
}

// OAuth exchanges authorization codes for access tokens.
type OAuth struct {
	cfg OAuthConfig
}

// Token is an OAuth access token response.
type Token struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
}

type OauthTokenRequest struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Code         string `json:"code"`
	State        string `json:"state,omitempty"`
}

// UserInfo is the authenticated GitHub user.
type UserInfo struct {
	ID    int64  `json:"id"`
	Login string `json:"login"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Email is a GitHub account email address.
type Email struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

// PrimaryEmail returns the primary verified email, else the first verified, else fallback.
func PrimaryEmail(emails []Email, fallback string) string {
	var firstVerified string
	for _, e := range emails {
		if !e.Verified {
			continue
		}
		if e.Primary {
			return e.Email
		}
		if firstVerified == "" {
			firstVerified = e.Email
		}
	}
	if firstVerified != "" {
		return firstVerified
	}

	return fallback
}

// NewOAuth returns an OAuth helper for the given app configuration.
func NewOAuth(cfg OAuthConfig) *OAuth {
	return &OAuth{cfg: cfg}
}

// AuthorizeURL builds the GitHub authorization URL for the provided state.
func (o *OAuth) AuthorizeURL(state string) string {
	q := url.Values{
		"client_id":    {o.cfg.ClientID},
		"redirect_uri": {o.cfg.RedirectURL},
		"scope":        {strings.Join(o.cfg.Scopes, ",")},
		"state":        {state},
	}
	if !o.cfg.AllowSignup {
		q.Set("allow_signup", "false")
	}

	return OauthAuthorizeURL + "?" + q.Encode()
}

// Exchange validates state and exchanges an authorization code for an access token.
func (o *OAuth) Exchange(ctx context.Context, code, state, expectedState string) (*Token, error) {
	if state != expectedState {
		return nil, ErrInvalidOAuthState
	}

	payload, err := json.Marshal(OauthTokenRequest{
		ClientID:     o.cfg.ClientID,
		ClientSecret: o.cfg.ClientSecret,
		Code:         code,
		State:        expectedState,
	})
	if err != nil {
		return nil, fmt.Errorf("github oauth encode request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, OauthTokenURL, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("github oauth build request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("github oauth request: %w", err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("github oauth read body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github oauth token: status %d: %s", resp.StatusCode, body)
	}

	var token Token
	if err := json.Unmarshal(body, &token); err != nil {
		return nil, fmt.Errorf("github oauth decode body: %w", err)
	}

	return &token, nil
}

// User fetches the authenticated GitHub user.
func (o *OAuth) User(ctx context.Context, accessToken string) (*UserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, OauthUserURL, nil)
	if err != nil {
		return nil, fmt.Errorf("github oauth user request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("github oauth user: %w", err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("github oauth user body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github oauth user: status %d: %s", resp.StatusCode, body)
	}

	var user UserInfo
	if err := json.Unmarshal(body, &user); err != nil {
		return nil, fmt.Errorf("github oauth user decode: %w", err)
	}

	return &user, nil
}

// Emails fetches the authenticated user's email addresses.
func (o *OAuth) Emails(ctx context.Context, accessToken string) ([]Email, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, OauthEmailsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("github oauth emails request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("github oauth emails: %w", err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("github oauth emails body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github oauth emails: status %d: %s", resp.StatusCode, body)
	}

	var emails []Email
	if err := json.Unmarshal(body, &emails); err != nil {
		return nil, fmt.Errorf("github oauth emails decode: %w", err)
	}

	return emails, nil
}
