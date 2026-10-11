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

func TestUnitStats(t *testing.T) {
	t.Run("returns not found for a missing workspace", func(t *testing.T) {
		stats := NewStats(&fakeWorkspaces{}, &fakeWorkspaceStats{})

		_, err := stats.GetWorkspaceStats("ws-1")
		assert.ErrorIs(t, err, ErrWorkspaceNotFound)
	})

	t.Run("returns stats for an existing workspace", func(t *testing.T) {
		stats := NewStats(&fakeWorkspaces{workspace: &db.Workspace{Id: "ws-1"}}, &fakeWorkspaceStats{})

		response, err := stats.GetWorkspaceStats("ws-1")
		require.NoError(t, err)
		assert.NotNil(t, response)
	})

	t.Run("returns the stats repository error", func(t *testing.T) {
		stats := NewStats(
			&fakeWorkspaces{workspace: &db.Workspace{Id: "ws-1"}},
			&fakeWorkspaceStats{err: errors.New("boom")},
		)

		_, err := stats.GetWorkspaceStats("ws-1")
		assert.EqualError(t, err, "boom")
	})
}
