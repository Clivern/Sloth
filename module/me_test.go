// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"errors"
	"testing"
	"time"

	"github.com/clivern/cognit/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitMe(t *testing.T) {
	t.Run("session returns the user", func(t *testing.T) {
		me := NewMe(&fakeAPIKeys{}, &fakeUsers{}, &fakeAccessKeys{})

		response := me.GetBySession(&db.User{Id: "user-1", Email: "ada@example.com", Name: "Ada", Role: "regular"})
		assert.Equal(t, "session", response.PrincipalType)
		assert.Equal(t, db.Id("user-1"), response.User.Id)
	})

	t.Run("api key returns not found for an unknown key", func(t *testing.T) {
		me := NewMe(&fakeAPIKeys{}, &fakeUsers{}, &fakeAccessKeys{})

		_, err := me.GetByAPIKey("unknown")
		assert.ErrorIs(t, err, ErrAPIKeyNotFound)
	})

	t.Run("api key returns not found when the owner is gone", func(t *testing.T) {
		me := NewMe(&fakeAPIKeys{item: &db.APIKey{Id: "key-1", UserId: "user-1"}}, &fakeUsers{}, &fakeAccessKeys{})

		_, err := me.GetByAPIKey("key")
		assert.ErrorIs(t, err, ErrAPIKeyNotFound)
	})

	t.Run("api key returns the owner and key metadata", func(t *testing.T) {
		expiresAt := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
		me := NewMe(
			&fakeAPIKeys{item: &db.APIKey{Id: "key-1", Name: "ci", UserId: "user-1", ExpiresAt: &expiresAt}},
			&fakeUsers{user: &db.User{Id: "user-1", Email: "ada@example.com", Name: "Ada", Role: "regular"}},
			&fakeAccessKeys{},
		)

		response, err := me.GetByAPIKey("key")
		require.NoError(t, err)
		assert.Equal(t, "api_key", response.PrincipalType)
		assert.Equal(t, "ada@example.com", response.Owner.Email)
		assert.Equal(t, "ci", response.APIKey.Name)
		require.NotNil(t, response.APIKey.ExpiresAt)
		assert.Equal(t, "2026-06-01T00:00:00Z", *response.APIKey.ExpiresAt)
	})

	t.Run("api key lookup returns the repository error", func(t *testing.T) {
		me := NewMe(&fakeAPIKeys{err: errors.New("boom")}, &fakeUsers{}, &fakeAccessKeys{})

		_, err := me.GetByAPIKey("key")
		assert.EqualError(t, err, "boom")
	})

	t.Run("access key returns not found for an unknown key", func(t *testing.T) {
		me := NewMe(&fakeAPIKeys{}, &fakeUsers{}, &fakeAccessKeys{})

		_, err := me.GetByAccessKey("unknown")
		assert.ErrorIs(t, err, ErrAccessKeyNotFound)
	})

	t.Run("access key returns its permissions", func(t *testing.T) {
		me := NewMe(&fakeAPIKeys{}, &fakeUsers{}, &fakeAccessKeys{item: &db.AccessKey{
			Id:          "access-1",
			WorkspaceId: "ws-1",
			Name:        "agent",
			Meta:        new(`{"permissions":["` + CanRegisterInstance + `"]}`),
		}})

		response, err := me.GetByAccessKey("key")
		require.NoError(t, err)
		assert.Equal(t, "access_key", response.PrincipalType)
		assert.Equal(t, []string{CanRegisterInstance}, response.AccessKey.Permissions)
		assert.Nil(t, response.AccessKey.ExpiresAt)
	})

	t.Run("access key returns an error for corrupt metadata", func(t *testing.T) {
		me := NewMe(&fakeAPIKeys{}, &fakeUsers{}, &fakeAccessKeys{item: &db.AccessKey{Id: "access-1", Meta: new(`not json`)}})

		_, err := me.GetByAccessKey("key")
		assert.Error(t, err)
	})
}
