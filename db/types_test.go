// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitId(t *testing.T) {
	t.Run("NewId", func(t *testing.T) {
		id, err := NewId()
		require.NoError(t, err)
		assert.Len(t, id.String(), 36)
	})
	t.Run("Scan", func(t *testing.T) {
		var id Id
		require.NoError(t, id.Scan([]byte("11111111-1111-1111-1111-111111111111")))
		assert.Equal(t, "11111111-1111-1111-1111-111111111111", id.String())

		require.NoError(t, id.Scan("22222222-2222-2222-2222-222222222222"))
		assert.Equal(t, "22222222-2222-2222-2222-222222222222", id.String())

		require.NoError(t, id.Scan(nil))
		assert.Equal(t, "", id.String())

		assert.Error(t, id.Scan(42))
	})
	t.Run("Value", func(t *testing.T) {
		value, err := Id("").Value()
		require.NoError(t, err)
		assert.Nil(t, value)

		value, err = Id("11111111-1111-1111-1111-111111111111").Value()
		require.NoError(t, err)
		assert.Equal(t, "11111111-1111-1111-1111-111111111111", value)
	})
	t.Run("IsBotUser", func(t *testing.T) {
		assert.True(t, IsBotUser(BotUserId))
		assert.False(t, IsBotUser("11111111-1111-1111-1111-111111111111"))
	})
}
