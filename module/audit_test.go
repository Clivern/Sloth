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

func TestUnitAudit(t *testing.T) {
	t.Run("get returns not found for a missing workspace", func(t *testing.T) {
		audit := NewAudit(&fakeAuditEvents{}, &fakeWorkspaces{})

		_, err := audit.GetAuditEvent("ws-1", "evt-1")
		assert.ErrorIs(t, err, ErrWorkspaceNotFound)
	})

	t.Run("get hides events from another workspace", func(t *testing.T) {
		audit := NewAudit(
			&fakeAuditEvents{item: &db.AuditEvent{Id: "evt-1", WorkspaceId: "ws-2"}},
			&fakeWorkspaces{workspace: &db.Workspace{Id: "ws-1"}},
		)

		_, err := audit.GetAuditEvent("ws-1", "evt-1")
		assert.ErrorIs(t, err, ErrAuditEventNotFound)
	})

	t.Run("get returns the event", func(t *testing.T) {
		createdAt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
		audit := NewAudit(
			&fakeAuditEvents{
				item: &db.AuditEvent{
					Id:          "evt-1",
					WorkspaceId: "ws-1",
					Action:      "workspace.update",
					CreatedAt:   createdAt,
				},
			},
			&fakeWorkspaces{workspace: &db.Workspace{Id: "ws-1"}},
		)

		response, err := audit.GetAuditEvent("ws-1", "evt-1")
		require.NoError(t, err)
		assert.Equal(t, "workspace.update", response.Action)
		assert.Equal(t, "2026-01-02T03:04:05Z", response.CreatedAt)
	})

	t.Run("get wraps the repository error", func(t *testing.T) {
		audit := NewAudit(
			&fakeAuditEvents{err: errors.New("boom")},
			&fakeWorkspaces{workspace: &db.Workspace{Id: "ws-1"}},
		)

		_, err := audit.GetAuditEvent("ws-1", "evt-1")
		assert.ErrorIs(t, err, ErrFailedGetAuditEvent)
	})

	t.Run("list returns the events and the total", func(t *testing.T) {
		audit := NewAudit(
			&fakeAuditEvents{
				list:  []*db.AuditEvent{{Id: "evt-1", WorkspaceId: "ws-1", Action: "a"}},
				total: 7,
			},
			&fakeWorkspaces{workspace: &db.Workspace{Id: "ws-1"}},
		)

		result, err := audit.ListAuditEvents("ws-1", 20, 0)
		require.NoError(t, err)
		assert.Equal(t, int64(7), result.Total)
		require.Len(t, result.Events, 1)
	})

	t.Run("list wraps the repository error", func(t *testing.T) {
		audit := NewAudit(
			&fakeAuditEvents{err: errors.New("boom")},
			&fakeWorkspaces{workspace: &db.Workspace{Id: "ws-1"}},
		)

		_, err := audit.ListAuditEvents("ws-1", 20, 0)
		assert.ErrorIs(t, err, ErrFailedListAuditEvents)
	})
}
