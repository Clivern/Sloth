// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/clivern/cognit/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitSessionManager(t *testing.T) {
	t.Run("create returns an error for a missing user", func(t *testing.T) {
		manager := NewSessionManager(&fakeSessions{}, &fakeUsers{})

		_, err := manager.CreateSession(context.Background(), "user-1", time.Hour)
		assert.EqualError(t, err, "user not found")
	})

	t.Run("create stores a token with the request ip and user agent", func(t *testing.T) {
		sessions := &fakeSessions{}
		manager := NewSessionManager(sessions, &fakeUsers{user: &db.User{Id: "user-1"}})
		ctx := WithAppCtx(context.Background(), NewAppCtx("req-1", "Mozilla/5.0", "10.0.0.1", nil))

		session, err := manager.CreateSession(ctx, "user-1", time.Hour)
		require.NoError(t, err)
		assert.NotEmpty(t, session.Token)
		assert.WithinDuration(t, time.Now().UTC().Add(time.Hour), session.ExpiresAt, 5*time.Second)
		require.NotNil(t, session.IPAddress)
		assert.Equal(t, "10.0.0.1", *session.IPAddress)
		require.NotNil(t, session.UserAgent)
		assert.Equal(t, "Mozilla/5.0", *session.UserAgent)
		assert.Len(t, sessions.created, 1)
	})

	t.Run("create leaves ip and user agent empty without an app context", func(t *testing.T) {
		manager := NewSessionManager(&fakeSessions{}, &fakeUsers{user: &db.User{Id: "user-1"}})

		session, err := manager.CreateSession(context.Background(), "user-1", time.Hour)
		require.NoError(t, err)
		assert.Nil(t, session.IPAddress)
		assert.Nil(t, session.UserAgent)
	})

	t.Run("validate rejects unknown tokens", func(t *testing.T) {
		manager := NewSessionManager(&fakeSessions{}, &fakeUsers{})

		_, _, err := manager.ValidateSession("unknown")
		assert.EqualError(t, err, "session not found")
	})

	t.Run("validate deletes an expired session", func(t *testing.T) {
		sessions := &fakeSessions{
			session: &db.Session{
				Id:        "sess-1",
				UserId:    "user-1",
				ExpiresAt: time.Now().UTC().Add(-time.Minute),
			},
		}
		manager := NewSessionManager(sessions, &fakeUsers{user: &db.User{Id: "user-1", IsActive: true}})

		_, _, err := manager.ValidateSession("token")
		assert.EqualError(t, err, "session expired")
		assert.Equal(t, []db.Id{"sess-1"}, sessions.deleted)
	})

	t.Run("validate rejects a session whose user is gone or inactive", func(t *testing.T) {
		live := &fakeSessions{
			session: &db.Session{
				Id:        "sess-1",
				UserId:    "user-1",
				ExpiresAt: time.Now().UTC().Add(time.Hour),
			},
		}

		_, _, err := NewSessionManager(live, &fakeUsers{}).ValidateSession("token")
		assert.EqualError(t, err, "user not found")

		_, _, err = NewSessionManager(live, &fakeUsers{user: &db.User{Id: "user-1", IsActive: false}}).
			ValidateSession("token")
		assert.EqualError(t, err, "user is not active")
	})

	t.Run("validate returns the user and session for a live token", func(t *testing.T) {
		session := &db.Session{
			Id:        "sess-1",
			UserId:    "user-1",
			ExpiresAt: time.Now().UTC().Add(time.Hour),
		}
		manager := NewSessionManager(
			&fakeSessions{session: session},
			&fakeUsers{user: &db.User{Id: "user-1", IsActive: true}},
		)

		user, got, err := manager.ValidateSession("token")
		require.NoError(t, err)
		assert.Equal(t, db.Id("user-1"), user.Id)
		assert.Equal(t, db.Id("sess-1"), got.Id)
	})

	t.Run("revoke deletes every session of the user", func(t *testing.T) {
		sessions := &fakeSessions{}

		require.NoError(t, NewSessionManager(sessions, &fakeUsers{}).RevokeUserSessions("user-1"))
		assert.Equal(t, []db.Id{"user-1"}, sessions.deletedUser)
	})

	t.Run("get user sessions skips expired sessions", func(t *testing.T) {
		sessions := &fakeSessions{list: []*db.Session{
			{Id: "live", ExpiresAt: time.Now().UTC().Add(time.Hour)},
			{Id: "expired", ExpiresAt: time.Now().UTC().Add(-time.Hour)},
		}}

		list, err := NewSessionManager(sessions, &fakeUsers{}).GetUserSessions("user-1")
		require.NoError(t, err)
		require.Len(t, list, 1)
		assert.Equal(t, db.Id("live"), list[0].Id)
	})

	t.Run("cleanup returns the removed count and wraps no error", func(t *testing.T) {
		count, err := NewSessionManager(&fakeSessions{expired: 5}, &fakeUsers{}).
			CleanupExpiredSessions()
		require.NoError(t, err)
		assert.Equal(t, int64(5), count)

		_, err = NewSessionManager(&fakeSessions{err: errors.New("boom")}, &fakeUsers{}).
			CleanupExpiredSessions()
		assert.EqualError(t, err, "boom")
	})
}
