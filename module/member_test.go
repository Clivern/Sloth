// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"testing"

	"github.com/clivern/cognit/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitWorkspaceMember(t *testing.T) {
	workspace := &fakeWorkspaces{workspace: &db.Workspace{Id: "ws-1"}}

	t.Run("list returns not found for a missing workspace", func(t *testing.T) {
		_, err := newWorkspaceModule(&fakeWorkspaces{}, &fakeWorkspaceUsers{}, &fakeSubscriptions{}).
			ListWorkspaceMembers("ws-1", 20, 0)
		assert.ErrorIs(t, err, ErrWorkspaceNotFound)
	})

	t.Run("list skips members whose user no longer exists", func(t *testing.T) {
		members := &fakeWorkspaceUsers{
			list:  []*db.WorkspaceUser{{Id: "m-1", UserId: "user-1", Role: db.UserRoleAdmin}},
			total: 2,
		}
		module := NewWorkspace(workspace, members, &fakeSubscriptions{}, &fakeUsers{})

		result, err := module.ListWorkspaceMembers("ws-1", 20, 0)
		require.NoError(t, err)
		assert.Empty(t, result.Members)
		assert.Equal(t, int64(2), result.Total)
	})

	t.Run("list returns members with their user details", func(t *testing.T) {
		members := &fakeWorkspaceUsers{
			list:  []*db.WorkspaceUser{{Id: "m-1", UserId: "user-1", Role: db.UserRoleAdmin}},
			total: 1,
		}
		module := NewWorkspace(
			workspace,
			members,
			&fakeSubscriptions{},
			&fakeUsers{user: &db.User{Id: "user-1", Name: "Ada", Email: "ada@example.com"}},
		)

		result, err := module.ListWorkspaceMembers("ws-1", 20, 0)
		require.NoError(t, err)
		require.Len(t, result.Members, 1)
		assert.Equal(t, "Ada", result.Members[0].Name)
		assert.Equal(t, db.UserRoleAdmin, result.Members[0].Role)
	})

	t.Run("get returns the membership or nil", func(t *testing.T) {
		members := &fakeWorkspaceUsers{
			membership: &db.WorkspaceUser{Id: "m-1", Role: db.UserRoleRegular},
		}

		membership, err := newWorkspaceModule(workspace, members, &fakeSubscriptions{}).
			GetWorkspaceMember("ws-1", "user-1")
		require.NoError(t, err)
		assert.Equal(t, db.UserRoleRegular, membership.Role)

		none, err := newWorkspaceModule(workspace, &fakeWorkspaceUsers{}, &fakeSubscriptions{}).
			GetWorkspaceMember("ws-1", "user-1")
		require.NoError(t, err)
		assert.Nil(t, none)
	})

	t.Run("update role returns not found for a missing member", func(t *testing.T) {
		_, err := newWorkspaceModule(workspace, &fakeWorkspaceUsers{}, &fakeSubscriptions{}).
			UpdateWorkspaceMemberRole("ws-1", "user-1", db.UserRoleAdmin)
		assert.ErrorIs(t, err, ErrWorkspaceUserNotFound)
	})

	t.Run("update role changes the member role", func(t *testing.T) {
		members := &fakeWorkspaceUsers{
			membership: &db.WorkspaceUser{Id: "m-1", UserId: "user-1", Role: db.UserRoleRegular},
		}
		module := NewWorkspace(
			workspace,
			members,
			&fakeSubscriptions{},
			&fakeUsers{user: &db.User{Id: "user-1", Name: "Ada"}},
		)

		response, err := module.UpdateWorkspaceMemberRole("ws-1", "user-1", db.UserRoleAdmin)
		require.NoError(t, err)
		require.Len(t, members.updated, 1)
		assert.Equal(t, db.UserRoleAdmin, response.Role)
		assert.Equal(t, "Ada", response.Name)
	})

	t.Run("update role refuses to change the owner", func(t *testing.T) {
		members := &fakeWorkspaceUsers{
			membership: &db.WorkspaceUser{Id: "m-1", UserId: "user-1", Role: db.UserRoleOwner},
		}

		_, err := newWorkspaceModule(workspace, members, &fakeSubscriptions{}).
			UpdateWorkspaceMemberRole("ws-1", "user-1", db.UserRoleAdmin)
		assert.ErrorIs(t, err, ErrWorkspaceOwnerProtected)
		assert.Empty(t, members.updated)
	})

	t.Run("delete refuses to remove the owner", func(t *testing.T) {
		members := &fakeWorkspaceUsers{membership: &db.WorkspaceUser{Id: "m-1", Role: db.UserRoleOwner}}

		err := newWorkspaceModule(workspace, members, &fakeSubscriptions{}).
			DeleteWorkspaceMember("ws-1", "user-1")
		assert.ErrorIs(t, err, ErrWorkspaceOwnerProtected)
		assert.Empty(t, members.deleted)
	})

	t.Run("delete returns not found for a missing member", func(t *testing.T) {
		err := newWorkspaceModule(workspace, &fakeWorkspaceUsers{}, &fakeSubscriptions{}).
			DeleteWorkspaceMember("ws-1", "user-1")
		assert.ErrorIs(t, err, ErrWorkspaceUserNotFound)
	})

	t.Run("delete removes the membership", func(t *testing.T) {
		members := &fakeWorkspaceUsers{membership: &db.WorkspaceUser{Id: "m-1"}}

		require.NoError(
			t,
			newWorkspaceModule(workspace, members, &fakeSubscriptions{}).
				DeleteWorkspaceMember("ws-1", "user-1"),
		)
		assert.Equal(t, []db.Id{"m-1"}, members.deleted)
	})
}
