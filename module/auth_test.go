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

func TestUnitAuth(t *testing.T) {
	t.Run("logout revokes every session of the user", func(t *testing.T) {
		sessions := &fakeSessions{expired: 2}
		auth := NewAuth(&fakeUsers{}, sessions, &fakeConfigs{})

		require.NoError(t, auth.Logout("user-1"))
		assert.Equal(t, []db.Id{"user-1"}, sessions.deletedUser)
	})

	t.Run("logout wraps the revoke error", func(t *testing.T) {
		auth := NewAuth(&fakeUsers{}, &fakeSessions{err: errors.New("boom")}, &fakeConfigs{})

		err := auth.Logout("user-1")
		assert.ErrorIs(t, err, ErrFailedRevokeSessions)
	})
}
