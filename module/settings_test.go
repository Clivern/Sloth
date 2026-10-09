// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitSettings(t *testing.T) {
	t.Run("update stores trimmed email and maintenance flag", func(t *testing.T) {
		configs := &fakeConfigs{}
		settings := NewSettings(configs)

		require.NoError(t, settings.Update("  ops@example.com ", true))
		assert.Equal(t, "ops@example.com", configs.values["platform_email"])
		assert.Equal(t, "1", configs.values["maintenance_mode"])

		require.NoError(t, settings.Update("ops@example.com", false))
		assert.Equal(t, "0", configs.values["maintenance_mode"])
	})

	t.Run("update wraps the repository error with the key", func(t *testing.T) {
		settings := NewSettings(&fakeConfigs{err: errors.New("boom")})

		err := settings.Update("ops@example.com", false)
		assert.ErrorContains(t, err, "failed to update config")
	})

	t.Run("get returns stored values and empty strings for missing keys", func(t *testing.T) {
		settings := NewSettings(&fakeConfigs{values: map[string]string{"platform_email": "ops@example.com"}})

		values, err := settings.GetSettings()
		require.NoError(t, err)
		assert.Equal(t, "ops@example.com", values["platform_email"])
		assert.Equal(t, "", values["maintenance_mode"])
	})

	t.Run("get returns the repository error", func(t *testing.T) {
		settings := NewSettings(&fakeConfigs{err: errors.New("boom")})

		_, err := settings.GetSettings()
		assert.EqualError(t, err, "boom")
	})
}
