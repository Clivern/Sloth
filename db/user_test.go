// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package db

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegrationUserRepository(t *testing.T) {
	database := openTestDB(t)
	repo := NewUserRepository(database)
	user := createTestUser(t, database)

	t.Run("get by id and email", func(t *testing.T) {
		got, err := repo.GetById(user.Id)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, user.Email, got.Email)

		byEmail, err := repo.GetByEmail(user.Email)
		require.NoError(t, err)
		require.NotNil(t, byEmail)
		assert.Equal(t, user.Id, byEmail.Id)
	})

	t.Run("update and get by provider", func(t *testing.T) {
		providerId := "gh-" + user.Id.String()[:8]
		user.Provider = UserProviderGithub
		user.ProviderUserId = &providerId
		user.Name = "updated"
		require.NoError(t, repo.Update(user))

		byProvider, err := repo.GetByProvider(UserProviderGithub, providerId)
		require.NoError(t, err)
		require.NotNil(t, byProvider)
		assert.Equal(t, "updated", byProvider.Name)
	})

	t.Run("update last login", func(t *testing.T) {
		require.NoError(t, repo.UpdateLastLogin(user.Id))
		afterLogin, err := repo.GetById(user.Id)
		require.NoError(t, err)
		assert.True(t, afterLogin.LastLoginAt.After(time.Now().UTC().Add(-time.Minute)))
	})

	t.Run("list and count", func(t *testing.T) {
		list, err := repo.List(50, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(list), 1)

		count, err := repo.Count()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, int64(1))
	})

	t.Run("missing email", func(t *testing.T) {
		missing, err := repo.GetByEmail("missing@t.test")
		require.NoError(t, err)
		assert.Nil(t, missing)
	})
}

func TestIntegrationUserGetByAPIKey(t *testing.T) {
	database := openTestDB(t)
	user := createTestUser(t, database)
	key := &APIKey{UserId: user.Id, Name: "cli", Key: "tok-" + user.Id.String()[:8]}
	require.NoError(t, NewAPIKeyRepository(database).Create(key))
	repo := NewUserRepository(database)

	t.Run("matching key", func(t *testing.T) {
		got, err := repo.GetByAPIKey(key.Key)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, user.Id, got.Id)
	})

	t.Run("empty key", func(t *testing.T) {
		got, err := repo.GetByAPIKey("")
		require.NoError(t, err)
		assert.Nil(t, got)
	})
}

func TestIntegrationUserGetBot(t *testing.T) {
	database := openTestDB(t)
	repo := NewUserRepository(database)

	bot := &User{
		Id:          BotUserId,
		Name:        BotUserName,
		Email:       "bot-" + BotUserId.String()[:8] + "@t.test",
		Password:    "hash",
		Provider:    UserProviderLocal,
		Role:        UserRoleBot,
		IsActive:    true,
		Language:    UserLanguageEN,
		Theme:       UserThemeDefault,
		LastLoginAt: time.Now().UTC(),
	}
	require.NoError(t, repo.Create(bot))
	t.Cleanup(func() { _ = repo.Delete(BotUserId) })

	t.Run("returns bot user", func(t *testing.T) {
		got, err := repo.GetBot()
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, BotUserName, got.Name)
	})
}

func TestIntegrationUserMetaRepository(t *testing.T) {
	database := openTestDB(t)
	user := createTestUser(t, database)
	repo := NewUserMetaRepository(database)

	t.Run("create and get", func(t *testing.T) {
		require.NoError(t, repo.Create(user.Id, "plan", `{"v":"pro"}`))
		got, err := repo.Get(user.Id, "plan")
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Contains(t, got.Value, "pro")
	})

	t.Run("upsert", func(t *testing.T) {
		require.NoError(t, repo.Upsert(user.Id, "plan", `{"v":"team"}`))
		got, err := repo.Get(user.Id, "plan")
		require.NoError(t, err)
		assert.Contains(t, got.Value, "team")
	})

	t.Run("list and delete", func(t *testing.T) {
		list, err := repo.ListByUser(user.Id)
		require.NoError(t, err)
		assert.Len(t, list, 1)

		require.NoError(t, repo.Delete(user.Id, "plan"))
		got, err := repo.Get(user.Id, "plan")
		require.NoError(t, err)
		assert.Nil(t, got)
	})
}
