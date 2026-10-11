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

func TestUnitInvite(t *testing.T) {
	inviter := &db.User{Id: "user-1", Email: "owner@example.com"}

	t.Run("create returns not found for a missing workspace", func(t *testing.T) {
		invite := NewInvite(
			&fakeUserInvites{},
			&fakeUsers{},
			&fakeConfigs{},
			&fakeWorkspaces{},
			&fakeWorkspaceUsers{},
			&fakeMailer{},
		)

		_, err := invite.CreateInvite(
			"ws-1",
			&CreateInviteRequest{Email: "ada@example.com", Role: "regular"},
			inviter,
		)
		assert.ErrorIs(t, err, ErrWorkspaceNotFound)
	})

	t.Run("create rejects an email that is already a member", func(t *testing.T) {
		users := &fakeUsers{byEmail: &db.User{Id: "user-2", Email: "ada@example.com"}}
		members := &fakeWorkspaceUsers{membership: &db.WorkspaceUser{Id: "m-1"}}

		_, err := newInviteModule(&fakeUserInvites{}, users, members, &fakeMailer{}).
			CreateInvite("ws-1", &CreateInviteRequest{
				Email: "ada@example.com",
				Role:  "regular",
			}, inviter)
		assert.ErrorIs(t, err, ErrUserAlreadyInWorkspace)
	})

	t.Run("create rejects a second pending invite for the same email", func(t *testing.T) {
		invites := &fakeUserInvites{pendingCount: 1}

		_, err := newInviteModule(invites, &fakeUsers{}, &fakeWorkspaceUsers{}, &fakeMailer{}).
			CreateInvite("ws-1", &CreateInviteRequest{
				Email: "ada@example.com",
				Role:  "regular",
			}, inviter)
		assert.ErrorIs(t, err, ErrPendingInviteExists)
	})

	t.Run(
		"create stores a pending invite, lowercases the email and sends the mail",
		func(t *testing.T) {
			invites := &fakeUserInvites{}
			mailer := &fakeMailer{}

			response, err := newInviteModule(
				invites,
				&fakeUsers{},
				&fakeWorkspaceUsers{},
				mailer,
			).
				CreateInvite("ws-1", &CreateInviteRequest{
					Email: "  Ada@Example.com ",
					Role:  "regular",
				}, inviter)
			require.NoError(t, err)
			require.Len(t, invites.created, 1)
			assert.Equal(t, "ada@example.com", invites.created[0].Email)
			assert.Equal(t, InviteStatusPending, response.Status)
			assert.NotEmpty(t, response.Token)
			assert.Equal(t, []string{"ada@example.com"}, mailer.sent)
			assert.WithinDuration(
				t,
				time.Now().UTC().Add(7*24*time.Hour),
				invites.created[0].ExpiresAt,
				time.Minute*5,
			)
		},
	)

	t.Run("create still succeeds when the mail cannot be sent", func(t *testing.T) {
		invites := &fakeUserInvites{}

		_, err := newInviteModule(
			invites,
			&fakeUsers{},
			&fakeWorkspaceUsers{},
			&fakeMailer{err: errors.New("smtp down")},
		).
			CreateInvite("ws-1", &CreateInviteRequest{
				Email: "ada@example.com",
				Role:  "regular",
			}, inviter)
		require.NoError(t, err)
		assert.Len(t, invites.created, 1)
	})

	t.Run(
		"create accepts the invite for an existing user who is not a member yet",
		func(t *testing.T) {
			users := &fakeUsers{byEmail: &db.User{Id: "user-2", Email: "ada@example.com"}}
			members := &fakeWorkspaceUsers{}

			response, err := newInviteModule(
				&fakeUserInvites{},
				users,
				members,
				&fakeMailer{},
			).
				CreateInvite("ws-1", &CreateInviteRequest{
					Email: "ada@example.com",
					Role:  "regular",
				}, inviter)
			require.NoError(t, err)
			assert.Equal(t, InviteStatusAccepted, response.Status)
		},
	)

	t.Run("create wraps the invite repository error", func(t *testing.T) {
		_, err := newInviteModule(
			&fakeUserInvites{err: errors.New("boom")},
			&fakeUsers{},
			&fakeWorkspaceUsers{},
			&fakeMailer{},
		).
			CreateInvite("ws-1", &CreateInviteRequest{
				Email: "ada@example.com",
				Role:  "regular",
			}, inviter)
		assert.ErrorIs(t, err, ErrFailedCreateInvite)
	})

	t.Run("list returns the invites and the total", func(t *testing.T) {
		invites := &fakeUserInvites{
			list: []*db.UserInvite{
				{Id: "inv-1", Email: "ada@example.com", Status: InviteStatusPending},
			},
			total: 5,
		}

		result, err := newInviteModule(invites, &fakeUsers{}, &fakeWorkspaceUsers{}, &fakeMailer{}).
			ListInvites("ws-1", 20, 0)
		require.NoError(t, err)
		assert.Equal(t, int64(5), result.Total)
		require.Len(t, result.Invites, 1)
		assert.Nil(t, result.Invites[0].AcceptedAt)
	})

	t.Run("list returns not found for a missing workspace", func(t *testing.T) {
		invite := NewInvite(
			&fakeUserInvites{},
			&fakeUsers{},
			&fakeConfigs{},
			&fakeWorkspaces{},
			&fakeWorkspaceUsers{},
			&fakeMailer{},
		)

		_, err := invite.ListInvites("ws-1", 20, 0)
		assert.ErrorIs(t, err, ErrWorkspaceNotFound)
	})

	t.Run("get hides invites from another workspace", func(t *testing.T) {
		invites := &fakeUserInvites{item: &db.UserInvite{Id: "inv-1", WorkspaceId: "ws-2"}}

		_, err := newInviteModule(invites, &fakeUsers{}, &fakeWorkspaceUsers{}, &fakeMailer{}).
			GetInvite("ws-1", "inv-1")
		assert.ErrorIs(t, err, ErrInviteNotFound)
	})

	t.Run("get returns the accepted time when set", func(t *testing.T) {
		acceptedAt := time.Date(2026, 2, 3, 4, 5, 6, 0, time.UTC)
		invites := &fakeUserInvites{
			item: &db.UserInvite{Id: "inv-1", WorkspaceId: "ws-1", AcceptedAt: &acceptedAt},
		}

		response, err := newInviteModule(invites, &fakeUsers{}, &fakeWorkspaceUsers{}, &fakeMailer{}).
			GetInvite("ws-1", "inv-1")
		require.NoError(t, err)
		require.NotNil(t, response.AcceptedAt)
		assert.Equal(t, "2026-02-03T04:05:06Z", *response.AcceptedAt)
	})

	t.Run("delete returns not found for an invite from another workspace", func(t *testing.T) {
		invites := &fakeUserInvites{item: &db.UserInvite{Id: "inv-1", WorkspaceId: "ws-2"}}

		err := newInviteModule(invites, &fakeUsers{}, &fakeWorkspaceUsers{}, &fakeMailer{}).
			DeleteInvite("ws-1", "inv-1")
		assert.ErrorIs(t, err, ErrInviteNotFound)
	})

	t.Run("delete removes the invite", func(t *testing.T) {
		invites := &fakeUserInvites{item: &db.UserInvite{Id: "inv-1", WorkspaceId: "ws-1"}}

		require.NoError(
			t,
			newInviteModule(invites, &fakeUsers{}, &fakeWorkspaceUsers{}, &fakeMailer{}).
				DeleteInvite("ws-1", "inv-1"),
		)
		assert.Equal(t, []db.Id{"inv-1"}, invites.deleted)
	})

	t.Run(
		"attach pending adds the user to each workspace and accepts the invites",
		func(t *testing.T) {
			invites := &fakeUserInvites{pending: []*db.UserInvite{
				{Id: "inv-1", WorkspaceId: "ws-1", Role: "regular"},
				{Id: "inv-2", WorkspaceId: "ws-2", Role: "admin"},
			}}
			members := &fakeWorkspaceUsers{}
			invite := newInviteModule(invites, &fakeUsers{}, members, &fakeMailer{})

			require.NoError(t, invite.AttachPending(&db.User{Id: "user-2", Email: "ada@example.com"}))
			require.Len(t, members.created, 2)
			assert.Equal(t, "admin", members.created[1].Role)
			assert.Equal(t, []string{InviteStatusAccepted, InviteStatusAccepted}, invites.statusUpdates)
		},
	)

	t.Run("attach pending skips workspaces the user already belongs to", func(t *testing.T) {
		invites := &fakeUserInvites{
			pending: []*db.UserInvite{{Id: "inv-1", WorkspaceId: "ws-1", Role: "regular"}},
		}
		members := &fakeWorkspaceUsers{membership: &db.WorkspaceUser{Id: "m-1"}}

		require.NoError(
			t,
			newInviteModule(invites, &fakeUsers{}, members, &fakeMailer{}).
				AttachPending(&db.User{Id: "user-2", Email: "ada@example.com"}),
		)
		assert.Empty(t, members.created)
		assert.Equal(t, []string{InviteStatusAccepted}, invites.statusUpdates)
	})

	t.Run("attach pending wraps the repository error", func(t *testing.T) {
		invites := &fakeUserInvites{err: errors.New("boom")}

		err := newInviteModule(invites, &fakeUsers{}, &fakeWorkspaceUsers{}, &fakeMailer{}).
			AttachPending(&db.User{Id: "user-2"})
		assert.ErrorIs(t, err, ErrFailedAttachInvites)
	})

}
