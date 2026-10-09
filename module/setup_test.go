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

func TestUnitSetup(t *testing.T) {
	t.Run("is installed is false before setup and true after", func(t *testing.T) {
		configs := &fakeConfigs{values: map[string]string{}}
		setup := NewSetup(configs, &fakeUsers{})

		assert.False(t, setup.IsInstalled())

		configs.values["is_installed"] = "1"
		assert.True(t, setup.IsInstalled())
	})

	t.Run("is installed is false when the config read fails", func(t *testing.T) {
		setup := NewSetup(&fakeConfigs{err: errors.New("boom")}, &fakeUsers{})

		assert.False(t, setup.IsInstalled())
	})

	t.Run("install rejects an already installed platform", func(t *testing.T) {
		setup := NewSetup(&fakeConfigs{values: map[string]string{"is_installed": "1"}}, &fakeUsers{})

		err := setup.Install(&SetupRequest{PlatformEmail: "ops@example.com"})
		assert.ErrorIs(t, err, ErrPlatformAlreadyInstalled)
	})

	t.Run("install creates the bot user and default config", func(t *testing.T) {
		configs := &fakeConfigs{values: map[string]string{}}
		users := &fakeUsers{}
		setup := NewSetup(configs, users)

		require.NoError(t, setup.Install(&SetupRequest{PlatformEmail: "ops@example.com"}))

		require.Len(t, users.created, 1)
		bot := users.created[0]
		assert.Equal(t, db.BotUserId, bot.Id)
		assert.Equal(t, db.UserRoleBot, bot.Role)
		assert.Equal(t, "ops@example.com", bot.Email)
		assert.NotEmpty(t, bot.Password)
		assert.NotEqual(t, "cognit-bot-user-no-login", bot.Password)

		assert.Equal(t, "1", configs.values["is_installed"])
		assert.Equal(t, "ops@example.com", configs.values["platform_email"])
		assert.Equal(t, "0", configs.values["maintenance_mode"])
		assert.NotEmpty(t, configs.values["platform_description"])
	})

	t.Run("install wraps the bot user create error", func(t *testing.T) {
		setup := NewSetup(&fakeConfigs{values: map[string]string{}}, &fakeUsers{err: errors.New("boom")})

		err := setup.Install(&SetupRequest{PlatformEmail: "ops@example.com"})
		assert.ErrorIs(t, err, ErrFailedCompleteSetup)
		assert.ErrorContains(t, err, "create bot user")
	})
}
