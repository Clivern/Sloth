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

func TestUnitTraffic(t *testing.T) {
	t.Run("get returns not found for a missing workspace", func(t *testing.T) {
		traffic := NewTraffic(&fakeTrafficCalls{}, &fakeWorkspaces{})

		_, err := traffic.GetTraffic("ws-1", "call-1")
		assert.ErrorIs(t, err, ErrWorkspaceNotFound)
	})

	t.Run("get hides calls from another workspace", func(t *testing.T) {
		traffic := NewTraffic(
			&fakeTrafficCalls{item: &db.TrafficCall{Id: "call-1", WorkspaceId: "ws-2"}},
			&fakeWorkspaces{workspace: &db.Workspace{Id: "ws-1"}},
		)

		_, err := traffic.GetTraffic("ws-1", "call-1")
		assert.ErrorIs(t, err, ErrTrafficNotFound)
	})

	t.Run("get returns the call with an empty task id when none is set", func(t *testing.T) {
		traffic := NewTraffic(
			&fakeTrafficCalls{item: &db.TrafficCall{Id: "call-1", WorkspaceId: "ws-1", Skill: "invoice.extract"}},
			&fakeWorkspaces{workspace: &db.Workspace{Id: "ws-1"}},
		)

		response, err := traffic.GetTraffic("ws-1", "call-1")
		require.NoError(t, err)
		assert.Equal(t, "invoice.extract", response.Skill)
		assert.Equal(t, "", response.TaskId)
	})

	t.Run("get returns the task id when one is set", func(t *testing.T) {
		taskId := "task-1"
		traffic := NewTraffic(
			&fakeTrafficCalls{item: &db.TrafficCall{Id: "call-1", WorkspaceId: "ws-1", TaskId: &taskId}},
			&fakeWorkspaces{workspace: &db.Workspace{Id: "ws-1"}},
		)

		response, err := traffic.GetTraffic("ws-1", "call-1")
		require.NoError(t, err)
		assert.Equal(t, "task-1", response.TaskId)
	})

	t.Run("get wraps the repository error", func(t *testing.T) {
		traffic := NewTraffic(&fakeTrafficCalls{err: errors.New("boom")}, &fakeWorkspaces{workspace: &db.Workspace{Id: "ws-1"}})

		_, err := traffic.GetTraffic("ws-1", "call-1")
		assert.ErrorIs(t, err, ErrFailedGetTraffic)
	})

	t.Run("list returns the calls and the total", func(t *testing.T) {
		traffic := NewTraffic(
			&fakeTrafficCalls{list: []*db.TrafficCall{{Id: "call-1", WorkspaceId: "ws-1"}, {Id: "call-2", WorkspaceId: "ws-1"}}, total: 9},
			&fakeWorkspaces{workspace: &db.Workspace{Id: "ws-1"}},
		)

		result, err := traffic.ListTraffic("ws-1", 20, 0)
		require.NoError(t, err)
		assert.Equal(t, int64(9), result.Total)
		assert.Len(t, result.Calls, 2)
	})

	t.Run("list wraps the repository error", func(t *testing.T) {
		traffic := NewTraffic(&fakeTrafficCalls{err: errors.New("boom")}, &fakeWorkspaces{workspace: &db.Workspace{Id: "ws-1"}})

		_, err := traffic.ListTraffic("ws-1", 20, 0)
		assert.ErrorIs(t, err, ErrFailedListTraffic)
	})
}
