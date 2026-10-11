// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"context"
	"errors"
	"testing"

	"github.com/clivern/cognit/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitOAuth(t *testing.T) {
	t.Run("rejects an identity without an email", func(t *testing.T) {
		identity := googleIdentity()
		identity.Email = ""

		_, err := newOAuthAuth(&fakeUsers{}, &fakeSessions{}, &fakeConfigs{}).
			LoginWithOAuth(context.Background(), identity)
		assert.EqualError(t, err, "oauth identity missing email")
	})

	t.Run("wraps the provider lookup error", func(t *testing.T) {
		users := &fakeUsers{err: errors.New("boom")}

		_, err := newOAuthAuth(users, &fakeSessions{}, &fakeConfigs{}).
			LoginWithOAuth(context.Background(), googleIdentity())
		assert.ErrorContains(t, err, "get user by provider")
	})

	t.Run("creates a regular user for a new identity and opens a session", func(t *testing.T) {
		users := &fakeUsers{}
		sessions := &fakeSessions{}

		result, err := newOAuthAuth(users, sessions, &fakeConfigs{}).
			LoginWithOAuth(context.Background(), googleIdentity())
		require.NoError(t, err)
		require.Len(t, users.created, 1)
		assert.Equal(t, "google", users.created[0].Provider)
		assert.Equal(t, db.UserRoleRegular, users.created[0].Role)
		assert.True(t, users.created[0].IsEmailVerified)
		assert.Equal(t, "ada@example.com", result.User.Email)

		require.Len(t, sessions.created, 1)
		assert.NotEmpty(t, sessions.created[0].Token)
		assert.Equal(t, db.Id("user-new"), sessions.created[0].UserId)
		assert.Equal(t, []db.Id{"user-new"}, users.lastLogin)
	})

	t.Run("links an existing local account with the same email", func(t *testing.T) {
		local := &db.User{
			Id:       "user-1",
			Email:    "ada@example.com",
			Provider: db.UserProviderLocal,
			IsActive: true,
			Role:     db.UserRoleRegular,
		}
		users := &fakeUsers{byEmail: local}

		result, err := newOAuthAuth(users, &fakeSessions{}, &fakeConfigs{}).
			LoginWithOAuth(context.Background(), googleIdentity())
		require.NoError(t, err)
		require.Len(t, users.updated, 1)
		assert.Equal(t, "google", users.updated[0].Provider)
		require.NotNil(t, users.updated[0].ProviderUserId)
		assert.Equal(t, "g-123", *users.updated[0].ProviderUserId)
		assert.Equal(t, db.Id("user-1"), result.User.Id)
	})

	t.Run("rejects the bot user", func(t *testing.T) {
		users := &fakeUsers{
			byProvider: &db.User{Id: db.BotUserId, Email: "ada@example.com", IsActive: true},
		}

		_, err := newOAuthAuth(users, &fakeSessions{}, &fakeConfigs{}).
			LoginWithOAuth(context.Background(), googleIdentity())
		assert.ErrorIs(t, err, ErrInvalidCredentials)
	})

	t.Run("rejects an inactive account", func(t *testing.T) {
		users := &fakeUsers{
			byProvider: &db.User{Id: "user-1", Email: "ada@example.com", IsActive: false},
		}

		_, err := newOAuthAuth(users, &fakeSessions{}, &fakeConfigs{}).
			LoginWithOAuth(context.Background(), googleIdentity())
		assert.ErrorIs(t, err, ErrAccountNotActive)
	})

	t.Run("blocks regular users during maintenance but not admins", func(t *testing.T) {
		configs := &fakeConfigs{values: map[string]string{"maintenance_mode": "1"}}

		regular := &fakeUsers{
			byProvider: &db.User{
				Id:       "user-1",
				Email:    "ada@example.com",
				IsActive: true,
				Role:     db.UserRoleRegular,
			},
		}
		_, err := newOAuthAuth(regular, &fakeSessions{}, configs).
			LoginWithOAuth(context.Background(), googleIdentity())
		assert.ErrorIs(t, err, ErrMaintenanceModeEnabled)

		admin := &fakeUsers{
			byProvider: &db.User{
				Id:       "user-1",
				Email:    "ada@example.com",
				IsActive: true,
				Role:     db.UserRoleAdmin,
			},
		}
		_, err = newOAuthAuth(admin, &fakeSessions{}, configs).
			LoginWithOAuth(context.Background(), googleIdentity())
		assert.NoError(t, err)
	})

	t.Run("rejects an email change to an address another user already owns", func(t *testing.T) {
		users := &fakeUsers{
			byProvider: &db.User{Id: "user-1", Email: "old@example.com", IsActive: true},
			byEmail:    &db.User{Id: "user-2", Email: "ada@example.com"},
		}

		_, err := newOAuthAuth(users, &fakeSessions{}, &fakeConfigs{}).
			LoginWithOAuth(context.Background(), googleIdentity())
		assert.ErrorIs(t, err, ErrUserEmailAlreadyExists)
	})

	t.Run("stores the request ip and user agent on the session", func(t *testing.T) {
		users := &fakeUsers{
			byProvider: &db.User{
				Id:       "user-1",
				Email:    "ada@example.com",
				IsActive: true,
				Role:     db.UserRoleRegular,
			},
		}
		sessions := &fakeSessions{}

		ctx := WithAppCtx(context.Background(), NewAppCtx("req-1", "Mozilla/5.0", "10.0.0.1", nil))
		_, err := newOAuthAuth(users, sessions, &fakeConfigs{}).
			LoginWithOAuth(ctx, googleIdentity())
		require.NoError(t, err)
		require.Len(t, sessions.created, 1)
		require.NotNil(t, sessions.created[0].IPAddress)
		assert.Equal(t, "10.0.0.1", *sessions.created[0].IPAddress)
	})
}
