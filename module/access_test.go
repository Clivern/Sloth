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

func TestUnitAccess(t *testing.T) {
	workspace := &fakeWorkspaces{
		workspace:  &db.Workspace{Id: "ws-1"},
		membership: &db.WorkspaceUser{Role: db.UserRoleAdmin},
	}
	admin := &Principal{User: &db.User{Id: "user-1"}}

	t.Run("create rejects unknown permissions", func(t *testing.T) {
		access := NewAccess(&fakeAccessKeys{}, workspace)

		_, err := access.CreateAccessKey(
			"ws-1",
			admin,
			&CreateAccessKeyRequest{Name: "agent", Permissions: []string{"workspace.delete"}},
		)
		assert.ErrorIs(t, err, ErrInvalidAccessKeyPermissions)
	})

	t.Run("create returns not found for a missing workspace", func(t *testing.T) {
		access := NewAccess(&fakeAccessKeys{}, &fakeWorkspaces{})

		_, err := access.CreateAccessKey(
			"ws-1",
			admin,
			&CreateAccessKeyRequest{Name: "agent", Permissions: []string{CanRegisterInstance}},
		)
		assert.ErrorIs(t, err, ErrWorkspaceNotFound)
	})

	t.Run("create rejects an invalid expiry", func(t *testing.T) {
		access := NewAccess(&fakeAccessKeys{}, workspace)

		_, err := access.CreateAccessKey("ws-1", admin, &CreateAccessKeyRequest{
			Name:        "agent",
			ExpiresAt:   "soon",
			Permissions: []string{CanRegisterInstance},
		})
		assert.ErrorIs(t, err, ErrInvalidExpiresAt)
	})

	t.Run("create denies permissions the user does not hold", func(t *testing.T) {
		member := &fakeWorkspaces{
			workspace:  &db.Workspace{Id: "ws-1"},
			membership: &db.WorkspaceUser{Role: db.UserRoleRegular},
		}
		access := NewAccess(&fakeAccessKeys{}, member)

		_, err := access.CreateAccessKey(
			"ws-1",
			admin,
			&CreateAccessKeyRequest{Name: "agent", Permissions: []string{CanRegisterInstance}},
		)
		assert.ErrorIs(t, err, ErrAccessKeyPermissionDenied)
	})

	t.Run("create denies permissions the access key does not hold", func(t *testing.T) {
		meta := `{"permissions":["` + CanCreateAccessKey + `"]}`
		key := &Principal{AccessKey: &db.AccessKey{Id: "access-1", WorkspaceId: "ws-1", Meta: &meta}}
		access := NewAccess(&fakeAccessKeys{}, workspace)

		_, err := access.CreateAccessKey(
			"ws-1",
			key,
			&CreateAccessKeyRequest{Name: "agent", Permissions: []string{CanUpdateWorkspaceBilling}},
		)
		assert.ErrorIs(t, err, ErrAccessKeyPermissionDenied)
	})

	t.Run("create allows permissions the access key holds", func(t *testing.T) {
		meta := `{"permissions":["` + CanCreateAccessKey + `","` + CanRegisterInstance + `"]}`
		key := &Principal{AccessKey: &db.AccessKey{Id: "access-1", WorkspaceId: "ws-1", Meta: &meta}}
		store := &fakeAccessKeys{}

		_, err := NewAccess(store, workspace).CreateAccessKey(
			"ws-1",
			key,
			&CreateAccessKeyRequest{Name: "agent", Permissions: []string{CanRegisterInstance}},
		)
		require.NoError(t, err)
		require.Len(t, store.created, 1)
	})

	t.Run("create stores the permissions and returns the secret once", func(t *testing.T) {
		store := &fakeAccessKeys{}
		access := NewAccess(store, workspace)

		response, err := access.CreateAccessKey(
			"ws-1",
			admin,
			&CreateAccessKeyRequest{Name: "agent", Permissions: []string{CanRegisterInstance}},
		)
		require.NoError(t, err)
		require.Len(t, store.created, 1)
		require.NotNil(t, store.created[0].Meta)
		assert.Contains(t, *store.created[0].Meta, CanRegisterInstance)
		assert.Equal(t, []string{CanRegisterInstance}, response.Permissions)
		assert.NotEmpty(t, response.Key)
		assert.Equal(t, response.Key, store.created[0].Key)
	})

	t.Run("create wraps the repository error", func(t *testing.T) {
		access := NewAccess(&fakeAccessKeys{err: errors.New("boom")}, workspace)

		_, err := access.CreateAccessKey(
			"ws-1",
			admin,
			&CreateAccessKeyRequest{Name: "agent", Permissions: []string{CanRegisterInstance}},
		)
		assert.ErrorIs(t, err, ErrFailedCreateAccessKey)
	})

	t.Run("list returns the keys with their permissions and the total", func(t *testing.T) {
		store := &fakeAccessKeys{
			list: []*db.AccessKey{
				{
					Id:   "access-1",
					Name: "agent",
					Meta: new(`{"permissions":["` + CanRegisterInstance + `"]}`),
				},
			},
			total: 4,
		}

		result, err := NewAccess(store, workspace).ListAccessKeys("ws-1", 20, 0)
		require.NoError(t, err)
		assert.Equal(t, int64(4), result.Total)
		assert.Equal(t, []string{CanRegisterInstance}, result.Keys[0].Permissions)
	})

	t.Run("get hides keys from another workspace", func(t *testing.T) {
		store := &fakeAccessKeys{
			item: &db.AccessKey{Id: "access-1", WorkspaceId: "ws-2", Meta: new(`{}`)},
		}

		_, err := NewAccess(store, workspace).GetAccessKey("ws-1", "access-1")
		assert.ErrorIs(t, err, ErrAccessKeyNotFound)
	})

	t.Run("get returns the key", func(t *testing.T) {
		store := &fakeAccessKeys{
			item: &db.AccessKey{
				Id:          "access-1",
				WorkspaceId: "ws-1",
				Name:        "agent",
				Meta:        new(`{}`),
			},
		}

		response, err := NewAccess(store, workspace).GetAccessKey("ws-1", "access-1")
		require.NoError(t, err)
		assert.Equal(t, "agent", response.Name)
		assert.Empty(t, response.Key)
	})

	t.Run("delete returns not found for a key in another workspace", func(t *testing.T) {
		store := &fakeAccessKeys{item: &db.AccessKey{Id: "access-1", WorkspaceId: "ws-2"}}

		err := NewAccess(store, workspace).DeleteAccessKey("ws-1", "access-1")
		assert.ErrorIs(t, err, ErrAccessKeyNotFound)
	})

	t.Run("delete removes the key", func(t *testing.T) {
		store := &fakeAccessKeys{item: &db.AccessKey{Id: "access-1", WorkspaceId: "ws-1"}}

		require.NoError(t, NewAccess(store, workspace).DeleteAccessKey("ws-1", "access-1"))
		assert.Equal(t, []db.Id{"access-1"}, store.deleted)
	})
}
