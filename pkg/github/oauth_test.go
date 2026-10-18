// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package github

import (
	"context"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnitOAuth(t *testing.T) {
	t.Run("PrimaryEmail", func(t *testing.T) {
		assert.Equal(t, "fallback@example.com", PrimaryEmail(nil, "fallback@example.com"))

		assert.Equal(t, "b@example.com", PrimaryEmail([]Email{
			{Email: "a@example.com", Verified: true},
			{Email: "b@example.com", Primary: true, Verified: true},
		}, "fallback"))

		assert.Equal(t, "first@example.com", PrimaryEmail([]Email{
			{Email: "unverified@example.com", Primary: true},
			{Email: "first@example.com", Verified: true},
			{Email: "second@example.com", Verified: true},
		}, "fallback"))
	})

	t.Run("AuthorizeURL", func(t *testing.T) {
		oauth := NewOAuth(OAuthConfig{
			ClientID:    "cid",
			RedirectURL: "https://app.example/callback",
			Scopes:      []string{"read:user", "user:email"},
			AllowSignup: true,
		})

		raw := oauth.AuthorizeURL("state-1")
		u, err := url.Parse(raw)
		assert.NoError(t, err)
		assert.Equal(t, OauthAuthorizeURL, u.Scheme+"://"+u.Host+u.Path)
		q := u.Query()
		assert.Equal(t, "cid", q.Get("client_id"))
		assert.Equal(t, "https://app.example/callback", q.Get("redirect_uri"))
		assert.Equal(t, "read:user,user:email", q.Get("scope"))
		assert.Equal(t, "state-1", q.Get("state"))
		assert.Empty(t, q.Get("allow_signup"))

		blocked := NewOAuth(OAuthConfig{AllowSignup: false})
		assert.Contains(t, blocked.AuthorizeURL("s"), "allow_signup=false")
	})

	t.Run("Exchange invalid state", func(t *testing.T) {
		oauth := NewOAuth(OAuthConfig{})
		token, err := oauth.Exchange(context.Background(), "code", "a", "b")
		assert.Nil(t, token)
		assert.ErrorIs(t, err, ErrInvalidOAuthState)
	})
}
