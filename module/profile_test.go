// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"errors"
	"testing"

	"github.com/clivern/cognit/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitProfile(t *testing.T) {
	t.Run("get returns the user", func(t *testing.T) {
		profile := NewProfile(&fakeUsers{user: &db.User{Id: "user-1", Name: "Ada"}})

		user, err := profile.GetProfile("user-1")
		require.NoError(t, err)
		assert.Equal(t, "Ada", user.Name)
	})

	t.Run("get returns not found for a missing user", func(t *testing.T) {
		profile := NewProfile(&fakeUsers{})

		_, err := profile.GetProfile("user-1")
		assert.ErrorIs(t, err, ErrUserNotFound)
	})

	t.Run("get wraps the repository error", func(t *testing.T) {
		profile := NewProfile(&fakeUsers{err: errors.New("boom")})

		_, err := profile.GetProfile("user-1")
		assert.ErrorIs(t, err, ErrFailedGetProfile)
	})

	t.Run("update keeps the name when the request leaves it empty", func(t *testing.T) {
		users := &fakeUsers{
			user: &db.User{Id: "user-1", Name: "Ada", Language: "en", Theme: "default"},
		}
		profile := NewProfile(users)

		user, err := profile.UpdateProfile("user-1", &UpdateProfileRequest{Language: "fr", Theme: "dark"})
		require.NoError(t, err)
		assert.Equal(t, "Ada", user.Name)
		assert.Equal(t, "fr", user.Language)
		assert.Equal(t, "dark", user.Theme)
		require.Len(t, users.updated, 1)
	})

	t.Run("update changes the name when one is given", func(t *testing.T) {
		profile := NewProfile(&fakeUsers{user: &db.User{Id: "user-1", Name: "Ada"}})

		user, err := profile.UpdateProfile(
			"user-1",
			&UpdateProfileRequest{Name: "Grace", Language: "en", Theme: "blue"},
		)
		require.NoError(t, err)
		assert.Equal(t, "Grace", user.Name)
	})

	t.Run("update returns not found and wraps repository errors", func(t *testing.T) {
		_, err := NewProfile(&fakeUsers{}).
			UpdateProfile("user-1", &UpdateProfileRequest{Language: "en", Theme: "blue"})
		assert.ErrorIs(t, err, ErrUserNotFound)

		_, err = NewProfile(&fakeUsers{user: &db.User{Id: "user-1"}, err: errors.New("boom")}).
			UpdateProfile("user-1", &UpdateProfileRequest{Language: "en", Theme: "blue"})
		assert.ErrorIs(t, err, ErrFailedUpdateProfile)
	})
}
