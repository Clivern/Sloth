// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package google

import (
	"context"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnitOAuth(t *testing.T) {
	t.Run("AuthorizeURL", func(t *testing.T) {
		oauth := NewOAuth(OAuthConfig{
			ClientID:    "cid",
			RedirectURL: "https://app.example/callback",
			Scopes:      []string{"openid", "email", "profile"},
		})

		raw := oauth.AuthorizeURL("state-1")
		u, err := url.Parse(raw)
		assert.NoError(t, err)
		assert.Equal(t, OauthAuthorizeURL, u.Scheme+"://"+u.Host+u.Path)
		q := u.Query()
		assert.Equal(t, "cid", q.Get("client_id"))
		assert.Equal(t, "https://app.example/callback", q.Get("redirect_uri"))
		assert.Equal(t, "code", q.Get("response_type"))
		assert.Equal(t, "openid email profile", q.Get("scope"))
		assert.Equal(t, "state-1", q.Get("state"))
		assert.Equal(t, "online", q.Get("access_type"))
		assert.Equal(t, "select_account", q.Get("prompt"))
	})

	t.Run("Exchange invalid state", func(t *testing.T) {
		oauth := NewOAuth(OAuthConfig{})
		token, err := oauth.Exchange(context.Background(), "code", "a", "b")
		assert.Nil(t, token)
		assert.ErrorIs(t, err, ErrInvalidOAuthState)
	})
}
