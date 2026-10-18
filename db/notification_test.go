// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegrationNotificationRepository(t *testing.T) {
	database := openTestDB(t)
	workspace := createTestWorkspace(t, database)
	otherWorkspace := createTestWorkspace(t, database)
	user := createTestUser(t, database)
	other := createTestUser(t, database)
	repo := NewNotificationRepository(database)

	shared := &Notification{
		Audience:     NotificationAudienceWorkspace,
		WorkspaceId:  &workspace.Id,
		Type:         "access_key.expired",
		Severity:     NotificationSeverityWarning,
		Title:        "Access key expired",
		Message:      stringPtr("production-ci expired yesterday"),
		ResourceType: stringPtr("access_key"),
		ResourceId:   &user.Id,
		Data:         stringPtr(`{"name":"production-ci"}`),
		Checksum:     "critical",
	}
	personal := &Notification{
		Audience:    NotificationAudienceUser,
		WorkspaceId: &workspace.Id,
		UserId:      &user.Id,
		Type:        "invite.expiring",
		Title:       "Invite expires in 2 days",
	}
	account := &Notification{
		Audience: NotificationAudienceUser,
		UserId:   &user.Id,
		Type:     "password.changed",
		Title:    "Your password was changed",
	}
	system := &Notification{
		Audience: NotificationAudienceSystem,
		Type:     "health_check.missed",
		Severity: NotificationSeverityCritical,
		Title:    "Health checks missed their last run",
	}

	t.Run("create and get", func(t *testing.T) {
		require.NoError(t, repo.Create(shared))
		require.NoError(t, repo.Create(personal))
		require.NoError(t, repo.Create(account))
		require.NoError(t, repo.Create(system))

		got, err := repo.GetById(shared.Id)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, NotificationAudienceWorkspace, got.Audience)
		assert.Equal(t, "access_key.expired", got.Type)
		assert.Equal(t, NotificationSeverityWarning, got.Severity)
		assert.Nil(t, got.UserId)
		require.NotNil(t, got.ResourceType)
		assert.Equal(t, "access_key", *got.ResourceType)

		got, err = repo.GetById(personal.Id)
		require.NoError(t, err)
		assert.Equal(t, NotificationSeverityInfo, got.Severity)
	})

	t.Run("audience check", func(t *testing.T) {
		assert.Error(t, repo.Create(&Notification{Audience: NotificationAudienceUser, WorkspaceId: &workspace.Id, Type: "x", Title: "x"}))
		assert.Error(t, repo.Create(&Notification{Audience: NotificationAudienceWorkspace, Type: "x", Title: "x"}))
		assert.Error(t, repo.Create(&Notification{Audience: NotificationAudienceSystem, WorkspaceId: &workspace.Id, Type: "x", Title: "x"}))
		assert.Error(t, repo.Create(&Notification{WorkspaceId: &workspace.Id, Type: "x", Title: "x"}))
	})

	t.Run("latest matching", func(t *testing.T) {
		got, err := repo.GetLatestMatching(shared)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, shared.Id, got.Id)
		assert.Equal(t, "critical", got.Checksum)

		recovered := &Notification{
			Audience:     NotificationAudienceWorkspace,
			WorkspaceId:  &workspace.Id,
			Type:         shared.Type,
			Title:        "Access key rotated",
			ResourceType: shared.ResourceType,
			ResourceId:   shared.ResourceId,
			Checksum:     "passing",
		}
		require.NoError(t, repo.Create(recovered))

		got, err = repo.GetLatestMatching(shared)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, recovered.Id, got.Id)
		require.NoError(t, repo.Delete(recovered.Id))

		got, err = repo.GetLatestMatching(account)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, account.Id, got.Id)

		missing, err := repo.GetLatestMatching(&Notification{Audience: NotificationAudienceWorkspace, WorkspaceId: &workspace.Id, Type: "missing"})
		require.NoError(t, err)
		assert.Nil(t, missing)
	})

	t.Run("list for user", func(t *testing.T) {
		list, err := repo.ListForUser(workspace.Id, user.Id, false, 10, 0)
		require.NoError(t, err)
		require.Len(t, list, 3)
		assert.Nil(t, list[0].ReadAt)

		list, err = repo.ListForUser(workspace.Id, other.Id, false, 10, 0)
		require.NoError(t, err)
		require.Len(t, list, 1)
		assert.Equal(t, shared.Id, list[0].Id)
	})

	t.Run("system for admins", func(t *testing.T) {
		list, err := repo.ListForUser(workspace.Id, other.Id, true, 100, 0)
		require.NoError(t, err)

		ids := make([]Id, 0, len(list))
		for _, n := range list {
			ids = append(ids, n.Id)
		}
		assert.Contains(t, ids, system.Id)
		assert.Contains(t, ids, shared.Id)
		assert.NotContains(t, ids, personal.Id)
	})

	t.Run("mark read and dismiss only touch notifications the user can see", func(t *testing.T) {
		cases := []struct {
			name         string
			notification *Notification
			workspaceId  Id
			userId       Id
			admin        bool
			visible      bool
		}{
			{"own workspace", shared, workspace.Id, other.Id, false, true},
			{"other workspace", shared, otherWorkspace.Id, other.Id, false, false},
			{"own user", personal, workspace.Id, user.Id, false, true},
			{"own user in other workspace", personal, otherWorkspace.Id, user.Id, false, false},
			{"own user across workspaces", account, otherWorkspace.Id, user.Id, false, true},
			{"other user", personal, workspace.Id, other.Id, true, false},
			{"system for regular", system, workspace.Id, other.Id, false, false},
			{"system for admin", system, workspace.Id, other.Id, true, true},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				found, err := repo.MarkRead(tc.notification.Id, tc.workspaceId, tc.userId, tc.admin)
				require.NoError(t, err)
				assert.Equal(t, tc.visible, found)

				found, err = repo.Dismiss(tc.notification.Id, tc.workspaceId, tc.userId, tc.admin)
				require.NoError(t, err)
				assert.Equal(t, tc.visible, found)
			})
		}

		missing, err := NewId()
		require.NoError(t, err)
		found, err := repo.MarkRead(missing, workspace.Id, user.Id, true)
		require.NoError(t, err)
		assert.False(t, found)

		// Reset the receipts the cases wrote so the next subtests start clean.
		_, err = database.Exec(`DELETE FROM notification_reads WHERE user_id IN ($1, $2)`, user.Id.String(), other.Id.String())
		require.NoError(t, err)
	})

	t.Run("read state is per user", func(t *testing.T) {
		found, err := repo.MarkRead(shared.Id, workspace.Id, user.Id, false)
		require.NoError(t, err)
		assert.True(t, found)

		found, err = repo.MarkRead(shared.Id, workspace.Id, user.Id, false)
		require.NoError(t, err)
		assert.True(t, found)

		count, err := repo.CountUnreadForUser(workspace.Id, user.Id, false)
		require.NoError(t, err)
		assert.Equal(t, int64(2), count)

		count, err = repo.CountUnreadForUser(workspace.Id, other.Id, false)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)

		require.NoError(t, repo.MarkAllRead(workspace.Id, user.Id, false))

		count, err = repo.CountUnreadForUser(workspace.Id, user.Id, false)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)

		list, err := repo.ListForUser(workspace.Id, user.Id, false, 10, 0)
		require.NoError(t, err)
		for _, n := range list {
			assert.NotNil(t, n.ReadAt)
		}
	})

	t.Run("dismiss is per user", func(t *testing.T) {
		found, err := repo.Dismiss(shared.Id, workspace.Id, user.Id, false)
		require.NoError(t, err)
		assert.True(t, found)

		list, err := repo.ListForUser(workspace.Id, user.Id, false, 10, 0)
		require.NoError(t, err)
		ids := make([]Id, 0, len(list))
		for _, n := range list {
			ids = append(ids, n.Id)
		}
		assert.NotContains(t, ids, shared.Id)
		assert.Contains(t, ids, personal.Id)

		list, err = repo.ListForUser(workspace.Id, other.Id, false, 10, 0)
		require.NoError(t, err)
		require.Len(t, list, 1)
		assert.Equal(t, shared.Id, list[0].Id)

		found, err = repo.Dismiss(system.Id, workspace.Id, other.Id, true)
		require.NoError(t, err)
		assert.True(t, found)

		list, err = repo.ListForUser(workspace.Id, other.Id, true, 100, 0)
		require.NoError(t, err)
		for _, n := range list {
			assert.NotEqual(t, system.Id, n.Id)
		}

		count, err := repo.CountUnreadForUser(workspace.Id, other.Id, false)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)

		found, err = repo.MarkRead(shared.Id, workspace.Id, user.Id, false)
		require.NoError(t, err)
		assert.True(t, found)

		list, err = repo.ListForUser(workspace.Id, user.Id, false, 10, 0)
		require.NoError(t, err)
		for _, n := range list {
			assert.NotEqual(t, shared.Id, n.Id)
		}
	})

	t.Run("delete", func(t *testing.T) {
		for _, n := range []*Notification{shared, personal, account, system} {
			require.NoError(t, repo.Delete(n.Id))
		}

		missing, err := repo.GetById(shared.Id)
		require.NoError(t, err)
		assert.Nil(t, missing)
	})
}
