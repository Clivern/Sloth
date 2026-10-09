// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"testing"

	"github.com/clivern/cognit/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitPermission(t *testing.T) {
	workspace := &db.Workspace{Id: "ws-1"}

	t.Run("can requires a workspace and a principal", func(t *testing.T) {
		perm := NewPerm(nil)
		perm.WorkspaceRepository = &fakeWorkspaces{}

		_, err := perm.WithUser(&db.User{Id: "user-1"}).Can(CanGetWorkspace)
		assert.ErrorIs(t, err, ErrPermissionWorkspaceMissing)

		_, err = perm.WithWorkspace(workspace).WithUser(nil).Can(CanGetWorkspace)
		assert.ErrorIs(t, err, ErrPermissionPrincipalMissing)
	})

	t.Run("a non member gets no permissions", func(t *testing.T) {
		allowed, err := newPerm(nil).WithUser(&db.User{Id: "user-1"}).WithWorkspace(workspace).Can(CanGetWorkspace)
		require.NoError(t, err)
		assert.False(t, allowed)
	})

	t.Run("a regular member can read but not change the workspace", func(t *testing.T) {
		perm := newPerm(&db.WorkspaceUser{Role: db.UserRoleRegular}).WithUser(&db.User{Id: "user-1"}).WithWorkspace(workspace)

		canRead, err := perm.Can(CanGetWorkspace)
		require.NoError(t, err)
		assert.True(t, canRead)

		canUpdate, err := perm.Can(CanUpdateWorkspace)
		require.NoError(t, err)
		assert.False(t, canUpdate)
	})

	t.Run("an admin or owner can manage agents", func(t *testing.T) {
		for _, role := range []string{db.UserRoleAdmin, db.UserRoleOwner} {
			allowed, err := newPerm(&db.WorkspaceUser{Role: role}).WithUser(&db.User{Id: "user-1"}).WithWorkspace(workspace).Can(CanDeleteAgent)
			require.NoError(t, err)
			assert.True(t, allowed, role)
		}
	})

	t.Run("unknown permissions are denied", func(t *testing.T) {
		allowed, err := newPerm(&db.WorkspaceUser{Role: db.UserRoleOwner}).WithUser(&db.User{Id: "user-1"}).WithWorkspace(workspace).Can("CAN_FLY")
		require.NoError(t, err)
		assert.False(t, allowed)
	})

	t.Run("an access key is limited to its own workspace", func(t *testing.T) {
		meta := `{"permissions":["` + CanRegisterInstance + `"]}`
		key := &db.AccessKey{WorkspaceId: "ws-2", Meta: &meta}

		allowed, err := newPerm(nil).WithAccessKey(key).WithWorkspace(workspace).Can(CanRegisterInstance)
		require.NoError(t, err)
		assert.False(t, allowed)
	})

	t.Run("an access key needs the permission in its metadata", func(t *testing.T) {
		withPermission := &db.AccessKey{WorkspaceId: "ws-1", Meta: new(`{"permissions":["` + CanRegisterInstance + `"]}`)}
		allowed, err := newPerm(nil).WithAccessKey(withPermission).WithWorkspace(workspace).Can(CanRegisterInstance)
		require.NoError(t, err)
		assert.True(t, allowed)

		withoutPermission := &db.AccessKey{WorkspaceId: "ws-1", Meta: new(`{"permissions":[]}`)}
		allowed, err = newPerm(nil).WithAccessKey(withoutPermission).WithWorkspace(workspace).Can(CanRegisterInstance)
		require.NoError(t, err)
		assert.False(t, allowed)
	})

	t.Run("an access key never gets permissions outside the allowed set", func(t *testing.T) {
		key := &db.AccessKey{WorkspaceId: "ws-1", Meta: new(`{"permissions":["CAN_FLY"]}`)}

		allowed, err := newPerm(nil).WithAccessKey(key).WithWorkspace(workspace).Can("CAN_FLY")
		require.NoError(t, err)
		assert.False(t, allowed)
	})

	t.Run("an access key with corrupt metadata returns an error", func(t *testing.T) {
		key := &db.AccessKey{WorkspaceId: "ws-1", Meta: new(`not json`)}

		_, err := newPerm(nil).WithAccessKey(key).WithWorkspace(workspace).Can(CanRegisterInstance)
		assert.Error(t, err)
	})

	t.Run("with user replaces an earlier access key and the reverse", func(t *testing.T) {
		perm := newPerm(&db.WorkspaceUser{Role: db.UserRoleOwner})

		perm.WithAccessKey(&db.AccessKey{Id: "access-1"}).WithUser(&db.User{Id: "user-1"})
		assert.Nil(t, perm.accessKey)

		perm.WithUser(&db.User{Id: "user-1"}).WithAccessKey(&db.AccessKey{Id: "access-1"})
		assert.Nil(t, perm.user)
	})

	t.Run("with workspace id loads the workspace from the repository", func(t *testing.T) {
		perm := newPerm(nil).WithWorkspaceId("ws-1")
		require.NotNil(t, perm.workspace)
		assert.Equal(t, db.Id("ws-1"), perm.workspace.Id)
	})

	t.Run("valid access key permissions accept the allowed set only", func(t *testing.T) {
		assert.True(t, ValidAccessKeyPermissions([]string{CanRegisterInstance}))
		assert.False(t, ValidAccessKeyPermissions([]string{"CAN_FLY"}))
		assert.False(t, ValidAccessKeyPermissions(nil))
	})
}
