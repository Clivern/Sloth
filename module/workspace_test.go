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

func TestUnitWorkspace(t *testing.T) {
	owner := &db.User{Id: "user-1"}

	t.Run("create makes the user owner and adds a subscription", func(t *testing.T) {
		workspaces := &fakeWorkspaces{}
		members := &fakeWorkspaceUsers{total: 1}
		subscriptions := &fakeSubscriptions{}

		response, err := newWorkspaceModule(workspaces, members, subscriptions).
			CreateWorkspace(&CreateWorkspaceRequest{Name: "Acme"}, owner)
		require.NoError(t, err)
		require.Len(t, workspaces.created, 1)
		assert.NotEmpty(t, response.Handle)
		assert.Equal(t, db.UserRoleOwner, response.Role)
		assert.Equal(t, int64(1), response.MembersCount)

		require.Len(t, members.created, 1)
		assert.Equal(t, db.UserRoleOwner, members.created[0].Role)
		assert.Equal(t, db.Id("user-1"), members.created[0].UserId)

		require.Len(t, subscriptions.created, 1)
	})

	t.Run("create returns the repository error", func(t *testing.T) {
		workspaces := &fakeWorkspaces{}
		members := &fakeWorkspaceUsers{err: errors.New("boom")}

		_, err := newWorkspaceModule(workspaces, members, &fakeSubscriptions{}).
			CreateWorkspace(&CreateWorkspaceRequest{Name: "Acme"}, owner)
		assert.EqualError(t, err, "boom")
	})

	t.Run("list returns the workspaces with the caller's role", func(t *testing.T) {
		workspaces := &fakeWorkspaces{
			list:       []*db.Workspace{{Id: "ws-1", Name: "Acme"}},
			total:      1,
			membership: &db.WorkspaceUser{Role: db.UserRoleAdmin},
		}

		result, err := newWorkspaceModule(workspaces, &fakeWorkspaceUsers{total: 2}, &fakeSubscriptions{}).
			ListWorkspaces(owner, 20, 0)
		require.NoError(t, err)
		assert.Equal(t, int64(1), result.Total)
		require.Len(t, result.Workspaces, 1)
		assert.Equal(t, db.UserRoleAdmin, result.Workspaces[0].Role)
		assert.Equal(t, int64(2), result.Workspaces[0].MembersCount)
	})

	t.Run("get returns not found for a missing workspace", func(t *testing.T) {
		_, err := newWorkspaceModule(&fakeWorkspaces{}, &fakeWorkspaceUsers{}, &fakeSubscriptions{}).
			GetWorkspace("ws-1", owner)
		assert.ErrorIs(t, err, ErrWorkspaceNotFound)
	})

	t.Run("get returns an empty role for a non-member", func(t *testing.T) {
		workspaces := &fakeWorkspaces{workspace: &db.Workspace{Id: "ws-1", Name: "Acme"}}

		response, err := newWorkspaceModule(workspaces, &fakeWorkspaceUsers{}, &fakeSubscriptions{}).
			GetWorkspace("ws-1", owner)
		require.NoError(t, err)
		assert.Equal(t, "", response.Role)
		assert.Equal(t, "Acme", response.Name)
	})

	t.Run("update returns not found for a missing workspace", func(t *testing.T) {
		_, err := newWorkspaceModule(&fakeWorkspaces{}, &fakeWorkspaceUsers{}, &fakeSubscriptions{}).
			UpdateWorkspace("ws-1", &UpdateWorkspaceRequest{Name: "Renamed"}, owner)
		assert.ErrorIs(t, err, ErrWorkspaceNotFound)
	})

	t.Run("update renames the workspace", func(t *testing.T) {
		workspaces := &fakeWorkspaces{
			workspace:  &db.Workspace{Id: "ws-1", Name: "Acme"},
			membership: &db.WorkspaceUser{Role: db.UserRoleOwner},
		}

		response, err := newWorkspaceModule(workspaces, &fakeWorkspaceUsers{}, &fakeSubscriptions{}).
			UpdateWorkspace("ws-1", &UpdateWorkspaceRequest{Name: "Renamed"}, owner)
		require.NoError(t, err)
		require.Len(t, workspaces.updated, 1)
		assert.Equal(t, "Renamed", response.Name)
		assert.Equal(t, db.UserRoleOwner, response.Role)
	})

	t.Run("delete returns not found for a missing workspace", func(t *testing.T) {
		err := newWorkspaceModule(&fakeWorkspaces{}, &fakeWorkspaceUsers{}, &fakeSubscriptions{}).
			DeleteWorkspace("ws-1")
		assert.ErrorIs(t, err, ErrWorkspaceNotFound)
	})

	t.Run("delete removes the workspace", func(t *testing.T) {
		workspaces := &fakeWorkspaces{workspace: &db.Workspace{Id: "ws-1"}}

		require.NoError(
			t,
			newWorkspaceModule(workspaces, &fakeWorkspaceUsers{}, &fakeSubscriptions{}).
				DeleteWorkspace("ws-1"),
		)
		assert.Equal(t, []db.Id{"ws-1"}, workspaces.deleted)
	})
}
