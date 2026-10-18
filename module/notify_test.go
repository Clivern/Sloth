// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/clivern/sloth/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitNotify(t *testing.T) {
	workspaceId := db.Id("ws-1")
	userId := db.Id("user-1")

	t.Run("workspace notification", func(t *testing.T) {
		notifications := &fakeNotifications{}
		notify := NewNotify(notifications, &fakeUsers{})

		notification, err := notify.NotifyWorkspace(workspaceId, NotifyRequest{
			Type:         "access_key.expired",
			Severity:     db.NotificationSeverityWarning,
			Title:        "Access key expired",
			Message:      "production-ci expired yesterday",
			ResourceType: "access_key",
			ResourceId:   "key-1",
			Data:         map[string]string{"name": "production-ci"},
		})
		require.NoError(t, err)
		require.NotNil(t, notification)
		require.Len(t, notifications.created, 1)
		assert.Equal(t, db.NotificationAudienceWorkspace, notification.Audience)
		assert.Equal(t, workspaceId, *notification.WorkspaceId)
		assert.Nil(t, notification.UserId)
		assert.Equal(t, "production-ci expired yesterday", *notification.Message)
		assert.Equal(t, "access_key", *notification.ResourceType)
		assert.JSONEq(t, `{"name":"production-ci"}`, *notification.Data)
	})

	t.Run("user notification", func(t *testing.T) {
		notifications := &fakeNotifications{}
		notify := NewNotify(notifications, &fakeUsers{})

		notification, err := notify.NotifyUser(userId, nil, NotifyRequest{Type: "password.changed", Title: "Password changed"})
		require.NoError(t, err)
		assert.Equal(t, db.NotificationAudienceUser, notification.Audience)
		assert.Equal(t, userId, *notification.UserId)
		assert.Nil(t, notification.WorkspaceId)
		assert.Nil(t, notification.Message)
		assert.Nil(t, notification.Data)
	})

	t.Run("system notification", func(t *testing.T) {
		notifications := &fakeNotifications{}
		notify := NewNotify(notifications, &fakeUsers{})

		notification, err := notify.NotifySystem(NotifyRequest{Type: "health_check.missed", Title: "Health checks missed"})
		require.NoError(t, err)
		assert.Equal(t, db.NotificationAudienceSystem, notification.Audience)
		assert.Nil(t, notification.WorkspaceId)
		assert.Nil(t, notification.UserId)
	})

	t.Run("missing target", func(t *testing.T) {
		notify := NewNotify(&fakeNotifications{}, &fakeUsers{})

		_, err := notify.NotifyWorkspace("", NotifyRequest{Type: "x", Title: "x"})
		assert.ErrorIs(t, err, ErrInvalidNotificationTarget)

		_, err = notify.NotifyUser("", nil, NotifyRequest{Type: "x", Title: "x"})
		assert.ErrorIs(t, err, ErrInvalidNotificationTarget)
	})

	t.Run("checksum is stored and every call notifies", func(t *testing.T) {
		notifications := &fakeNotifications{}
		notify := NewNotify(notifications, &fakeUsers{})

		for range 3 {
			_, err := notify.NotifyWorkspace(workspaceId, NotifyRequest{
				Type:     "agent.status",
				Title:    "support-bot is critical",
				Checksum: map[string]any{"status": "critical"},
			})
			require.NoError(t, err)
		}

		require.Len(t, notifications.created, 3)
		assert.NotEmpty(t, notifications.created[0].Checksum)
		assert.Equal(t, notifications.created[0].Checksum, notifications.created[2].Checksum)
	})

	t.Run("create wraps the repository error", func(t *testing.T) {
		notify := NewNotify(&fakeNotifications{err: errors.New("boom")}, &fakeUsers{})

		_, err := notify.NotifySystem(NotifyRequest{Type: "x", Title: "x"})
		assert.ErrorIs(t, err, ErrFailedCreateNotification)
	})

	t.Run("list shapes notifications and shows system ones to admins", func(t *testing.T) {
		createdAt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
		readAt := time.Date(2026, 1, 3, 3, 4, 5, 0, time.UTC)
		data := `{"name":"production-ci"}`
		notifications := &fakeNotifications{
			unread: 1,
			list: []*db.Notification{
				{Id: "n-1", Audience: db.NotificationAudienceWorkspace, Type: "access_key.expired", Data: &data, CreatedAt: createdAt},
				{Id: "n-2", Audience: db.NotificationAudienceSystem, Type: "health_check.missed", ReadAt: &readAt, CreatedAt: createdAt},
			},
		}
		notify := NewNotify(notifications, &fakeUsers{user: &db.User{Id: userId, Role: db.UserRoleAdmin}})

		response, err := notify.ListNotifications(workspaceId, userId, 10, 0)
		require.NoError(t, err)
		assert.True(t, notifications.admin)
		assert.Equal(t, int64(1), response.Unread)
		require.Len(t, response.Notifications, 2)
		assert.False(t, response.Notifications[0].Read)
		assert.Equal(t, "2026-01-02T03:04:05Z", response.Notifications[0].CreatedAt)
		assert.Equal(t, json.RawMessage(data), response.Notifications[0].Data)
		assert.True(t, response.Notifications[1].Read)
		assert.Equal(t, "2026-01-03T03:04:05Z", *response.Notifications[1].ReadAt)
	})

	t.Run("list hides system notifications from regular users", func(t *testing.T) {
		notifications := &fakeNotifications{}
		notify := NewNotify(notifications, &fakeUsers{user: &db.User{Id: userId, Role: db.UserRoleRegular}})

		_, err := notify.ListNotifications(workspaceId, userId, 10, 0)
		require.NoError(t, err)
		assert.False(t, notifications.admin)
	})

	t.Run("list returns not found for a missing user", func(t *testing.T) {
		notify := NewNotify(&fakeNotifications{}, &fakeUsers{})

		_, err := notify.ListNotifications(workspaceId, userId, 10, 0)
		assert.ErrorIs(t, err, ErrUserNotFound)

		_, err = notify.CountUnread(workspaceId, userId)
		assert.ErrorIs(t, err, ErrUserNotFound)
	})

	t.Run("count unread", func(t *testing.T) {
		notify := NewNotify(&fakeNotifications{unread: 3}, &fakeUsers{user: &db.User{Id: userId, Role: db.UserRoleRegular}})

		count, err := notify.CountUnread(workspaceId, userId)
		require.NoError(t, err)
		assert.Equal(t, int64(3), count)
	})

	t.Run("mark read", func(t *testing.T) {
		notifications := &fakeNotifications{}
		notify := NewNotify(notifications, &fakeUsers{user: &db.User{Id: userId, Role: db.UserRoleAdmin}})

		require.NoError(t, notify.MarkRead(workspaceId, userId, "n-1"))
		assert.Equal(t, []db.Id{"n-1"}, notifications.read)
		assert.True(t, notifications.admin)
	})

	t.Run("mark read returns not found when the user cannot see it", func(t *testing.T) {
		notify := NewNotify(&fakeNotifications{hidden: true}, &fakeUsers{user: &db.User{Id: userId, Role: db.UserRoleRegular}})

		err := notify.MarkRead(workspaceId, userId, "n-1")
		assert.ErrorIs(t, err, ErrNotificationNotFound)
	})

	t.Run("dismiss", func(t *testing.T) {
		notifications := &fakeNotifications{}
		notify := NewNotify(notifications, &fakeUsers{user: &db.User{Id: userId, Role: db.UserRoleRegular}})

		require.NoError(t, notify.Dismiss(workspaceId, userId, "n-1"))
		assert.Equal(t, []db.Id{"n-1"}, notifications.dismissed)
		assert.False(t, notifications.admin)
	})

	t.Run("dismiss returns not found when the user cannot see it", func(t *testing.T) {
		notify := NewNotify(&fakeNotifications{hidden: true}, &fakeUsers{user: &db.User{Id: userId}})

		err := notify.Dismiss(workspaceId, userId, "n-1")
		assert.ErrorIs(t, err, ErrNotificationNotFound)
	})

	t.Run("mark read and dismiss wrap the repository error", func(t *testing.T) {
		notify := NewNotify(&fakeNotifications{err: errors.New("boom")}, &fakeUsers{user: &db.User{Id: userId}})

		assert.ErrorIs(t, notify.MarkRead(workspaceId, userId, "n-1"), ErrFailedUpdateNotification)
		assert.ErrorIs(t, notify.Dismiss(workspaceId, userId, "n-1"), ErrFailedUpdateNotification)
	})

	t.Run("mark read and dismiss return not found for a missing user", func(t *testing.T) {
		notify := NewNotify(&fakeNotifications{}, &fakeUsers{})

		assert.ErrorIs(t, notify.MarkRead(workspaceId, userId, "n-1"), ErrUserNotFound)
		assert.ErrorIs(t, notify.Dismiss(workspaceId, userId, "n-1"), ErrUserNotFound)
	})

	t.Run("mark all read passes the admin flag", func(t *testing.T) {
		notifications := &fakeNotifications{}
		notify := NewNotify(notifications, &fakeUsers{user: &db.User{Id: userId, Role: db.UserRoleAdmin}})

		require.NoError(t, notify.MarkAllRead(workspaceId, userId))
		assert.True(t, notifications.admin)
	})
}
