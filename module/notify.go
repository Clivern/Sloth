// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/clivern/sloth/db"
	"github.com/clivern/sloth/pkg/util"

	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

var (
	ErrNotificationNotFound      = errors.New("notification not found")
	ErrFailedCreateNotification  = errors.New("failed create notification")
	ErrFailedListNotifications   = errors.New("failed list notifications")
	ErrFailedUpdateNotification  = errors.New("failed update notification")
	ErrInvalidNotificationTarget = errors.New("invalid notification target")
)

// Notify is the module for sending and reading notifications.
type Notify struct {
	NotificationRepository db.NotificationRepository
	UserRepository         db.UserRepository
}

// NewNotify creates a notify module with the given repositories.
func NewNotify(notifications db.NotificationRepository, users db.UserRepository) *Notify {
	return &Notify{
		NotificationRepository: notifications,
		UserRepository:         users,
	}
}

// NotifyRequest is what you pass when sending a notification.
type NotifyRequest struct {
	Type         string
	Severity     string
	Title        string
	Message      string
	ResourceType string
	ResourceId   db.Id
	Data         map[string]string
	Checksum     map[string]any
}

// NotificationResponse is a notification shaped for API responses.
type NotificationResponse struct {
	Id           db.Id           `json:"id"`
	Audience     string          `json:"audience"`
	WorkspaceId  *db.Id          `json:"workspaceId,omitempty"`
	UserId       *db.Id          `json:"userId,omitempty"`
	Type         string          `json:"type"`
	Severity     string          `json:"severity"`
	Title        string          `json:"title"`
	Message      *string         `json:"message,omitempty"`
	ResourceType *string         `json:"resourceType,omitempty"`
	ResourceId   *db.Id          `json:"resourceId,omitempty"`
	Data         json.RawMessage `json:"data,omitempty"`
	Read         bool            `json:"read"`
	ReadAt       *string         `json:"readAt,omitempty"`
	CreatedAt    string          `json:"createdAt"`
}

// ListNotificationsResponse is returned when listing notifications.
type ListNotificationsResponse struct {
	Notifications []*NotificationResponse
	Unread        int64
}

// NotifyWorkspace notifies every workspace member.
func (n *Notify) NotifyWorkspace(workspaceId db.Id, request NotifyRequest) (*db.Notification, error) {
	if lo.IsEmpty(workspaceId) {
		return nil, ErrInvalidNotificationTarget
	}

	notification := &db.Notification{
		Audience:     db.NotificationAudienceWorkspace,
		WorkspaceId:  &workspaceId,
		Type:         request.Type,
		Severity:     request.Severity,
		Title:        request.Title,
		Message:      lo.EmptyableToPtr(request.Message),
		ResourceType: lo.EmptyableToPtr(request.ResourceType),
		ResourceId:   lo.EmptyableToPtr(request.ResourceId),
	}
	if len(request.Data) > 0 {
		raw, err := json.Marshal(request.Data)
		if err != nil {
			log.Error().
				Err(err).
				Str("workspaceId", workspaceId.String()).
				Str("type", request.Type).
				Msg("Failed to encode notification data")
			return nil, fmt.Errorf("%w: %v", ErrFailedCreateNotification, err)
		}
		data := string(raw)
		notification.Data = &data
	}

	checksum, err := util.MapChecksum(request.Checksum)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("type", request.Type).
			Msg("Failed to compute notification checksum")
		return nil, fmt.Errorf("%w: %v", ErrFailedCreateNotification, err)
	}

	notification.Checksum = checksum

	err = n.NotificationRepository.Create(notification)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("type", request.Type).
			Msg("Failed to create notification")
		return nil, fmt.Errorf("%w: %v", ErrFailedCreateNotification, err)
	}

	log.Info().
		Str("workspaceId", workspaceId.String()).
		Str("notificationId", notification.Id.String()).
		Str("type", request.Type).
		Str("severity", request.Severity).
		Msg("Notification created")

	return notification, nil
}

// NotifyUser notifies one user, in all workspaces when workspaceId is nil.
func (n *Notify) NotifyUser(userId db.Id, workspaceId *db.Id, request NotifyRequest) (*db.Notification, error) {
	if lo.IsEmpty(userId) {
		return nil, ErrInvalidNotificationTarget
	}

	notification := &db.Notification{
		Audience:     db.NotificationAudienceUser,
		WorkspaceId:  workspaceId,
		UserId:       &userId,
		Type:         request.Type,
		Severity:     request.Severity,
		Title:        request.Title,
		Message:      lo.EmptyableToPtr(request.Message),
		ResourceType: lo.EmptyableToPtr(request.ResourceType),
		ResourceId:   lo.EmptyableToPtr(request.ResourceId),
	}
	if len(request.Data) > 0 {
		raw, err := json.Marshal(request.Data)
		if err != nil {
			log.Error().
				Err(err).
				Str("userId", userId.String()).
				Str("type", request.Type).
				Msg("Failed to encode notification data")
			return nil, fmt.Errorf("%w: %v", ErrFailedCreateNotification, err)
		}
		data := string(raw)
		notification.Data = &data
	}

	checksum, err := util.MapChecksum(request.Checksum)
	if err != nil {
		log.Error().
			Err(err).
			Str("userId", userId.String()).
			Str("type", request.Type).
			Msg("Failed to compute notification checksum")
		return nil, fmt.Errorf("%w: %v", ErrFailedCreateNotification, err)
	}

	notification.Checksum = checksum

	err = n.NotificationRepository.Create(notification)
	if err != nil {
		log.Error().
			Err(err).
			Str("userId", userId.String()).
			Str("type", request.Type).
			Msg("Failed to create notification")
		return nil, fmt.Errorf("%w: %v", ErrFailedCreateNotification, err)
	}

	log.Info().
		Str("userId", userId.String()).
		Str("notificationId", notification.Id.String()).
		Str("type", request.Type).
		Str("severity", request.Severity).
		Msg("Notification created")

	return notification, nil
}

// NotifySystem notifies platform admins.
func (n *Notify) NotifySystem(request NotifyRequest) (*db.Notification, error) {
	notification := &db.Notification{
		Audience:     db.NotificationAudienceSystem,
		Type:         request.Type,
		Severity:     request.Severity,
		Title:        request.Title,
		Message:      lo.EmptyableToPtr(request.Message),
		ResourceType: lo.EmptyableToPtr(request.ResourceType),
		ResourceId:   lo.EmptyableToPtr(request.ResourceId),
	}
	if len(request.Data) > 0 {
		raw, err := json.Marshal(request.Data)
		if err != nil {
			log.Error().
				Err(err).
				Str("type", request.Type).
				Msg("Failed to encode notification data")
			return nil, fmt.Errorf("%w: %v", ErrFailedCreateNotification, err)
		}
		data := string(raw)
		notification.Data = &data
	}

	checksum, err := util.MapChecksum(request.Checksum)
	if err != nil {
		log.Error().
			Err(err).
			Str("type", request.Type).
			Msg("Failed to compute notification checksum")
		return nil, fmt.Errorf("%w: %v", ErrFailedCreateNotification, err)
	}

	notification.Checksum = checksum

	err = n.NotificationRepository.Create(notification)
	if err != nil {
		log.Error().
			Err(err).
			Str("type", request.Type).
			Msg("Failed to create notification")
		return nil, fmt.Errorf("%w: %v", ErrFailedCreateNotification, err)
	}

	log.Info().
		Str("notificationId", notification.Id.String()).
		Str("type", request.Type).
		Str("severity", request.Severity).
		Msg("Notification created")

	return notification, nil
}

// ListNotifications lists a user's notifications in a workspace with the unread count.
func (n *Notify) ListNotifications(workspaceId, userId db.Id, limit, offset int) (*ListNotificationsResponse, error) {
	user, err := n.UserRepository.GetById(userId)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("userId", userId.String()).
			Msg("Failed to get user")
		return nil, fmt.Errorf("%w: %v", ErrFailedListNotifications, err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	admin := user.Role == db.UserRoleAdmin

	unread, err := n.NotificationRepository.CountUnreadForUser(workspaceId, userId, admin)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("userId", userId.String()).
			Msg("Failed to count unread notifications")
		return nil, fmt.Errorf("%w: %v", ErrFailedListNotifications, err)
	}

	notifications, err := n.NotificationRepository.ListForUser(workspaceId, userId, admin, limit, offset)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("userId", userId.String()).
			Msg("Failed to list notifications")
		return nil, fmt.Errorf("%w: %v", ErrFailedListNotifications, err)
	}

	list := make([]*NotificationResponse, 0, len(notifications))
	for _, item := range notifications {
		list = append(list, &NotificationResponse{
			Id:           item.Id,
			Audience:     item.Audience,
			WorkspaceId:  item.WorkspaceId,
			UserId:       item.UserId,
			Type:         item.Type,
			Severity:     item.Severity,
			Title:        item.Title,
			Message:      item.Message,
			ResourceType: item.ResourceType,
			ResourceId:   item.ResourceId,
			Data:         util.JSONRawFromString(item.Data),
			Read:         item.ReadAt != nil,
			ReadAt: lo.TernaryF(
				item.ReadAt != nil,
				func() *string { return new(item.ReadAt.UTC().Format(time.RFC3339)) },
				func() *string { return nil },
			),
			CreatedAt: item.CreatedAt.UTC().Format(time.RFC3339),
		})
	}

	return &ListNotificationsResponse{Notifications: list, Unread: unread}, nil
}

// CountUnread returns how many notifications a user has not read in a workspace.
func (n *Notify) CountUnread(workspaceId, userId db.Id) (int64, error) {
	user, err := n.UserRepository.GetById(userId)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("userId", userId.String()).
			Msg("Failed to get user")
		return 0, fmt.Errorf("%w: %v", ErrFailedListNotifications, err)
	}
	if user == nil {
		return 0, ErrUserNotFound
	}

	count, err := n.NotificationRepository.CountUnreadForUser(
		workspaceId,
		userId,
		user.Role == db.UserRoleAdmin,
	)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("userId", userId.String()).
			Msg("Failed to count unread notifications")
		return 0, fmt.Errorf("%w: %v", ErrFailedListNotifications, err)
	}

	return count, nil
}

// MarkRead marks a notification the user can see as read.
func (n *Notify) MarkRead(workspaceId, userId, notificationId db.Id) error {
	user, err := n.UserRepository.GetById(userId)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("userId", userId.String()).
			Msg("Failed to get user")
		return fmt.Errorf("%w: %v", ErrFailedUpdateNotification, err)
	}
	if user == nil {
		return ErrUserNotFound
	}

	found, err := n.NotificationRepository.MarkRead(
		notificationId,
		workspaceId,
		userId,
		user.Role == db.UserRoleAdmin,
	)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("userId", userId.String()).
			Str("notificationId", notificationId.String()).
			Msg("Failed to mark notification read")
		return fmt.Errorf("%w: %v", ErrFailedUpdateNotification, err)
	}
	if !found {
		return ErrNotificationNotFound
	}

	log.Info().
		Str("workspaceId", workspaceId.String()).
		Str("userId", userId.String()).
		Str("notificationId", notificationId.String()).
		Msg("Notification marked read")

	return nil
}

// Dismiss hides a notification the user can see.
func (n *Notify) Dismiss(workspaceId, userId, notificationId db.Id) error {
	user, err := n.UserRepository.GetById(userId)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("userId", userId.String()).
			Msg("Failed to get user")
		return fmt.Errorf("%w: %v", ErrFailedUpdateNotification, err)
	}
	if user == nil {
		return ErrUserNotFound
	}

	found, err := n.NotificationRepository.Dismiss(
		notificationId,
		workspaceId,
		userId,
		user.Role == db.UserRoleAdmin,
	)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("userId", userId.String()).
			Str("notificationId", notificationId.String()).
			Msg("Failed to dismiss notification")
		return fmt.Errorf("%w: %v", ErrFailedUpdateNotification, err)
	}
	if !found {
		return ErrNotificationNotFound
	}

	log.Info().
		Str("workspaceId", workspaceId.String()).
		Str("userId", userId.String()).
		Str("notificationId", notificationId.String()).
		Msg("Notification dismissed")

	return nil
}

// MarkAllRead marks every notification a user sees in a workspace as read.
func (n *Notify) MarkAllRead(workspaceId, userId db.Id) error {
	user, err := n.UserRepository.GetById(userId)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("userId", userId.String()).
			Msg("Failed to get user")
		return fmt.Errorf("%w: %v", ErrFailedUpdateNotification, err)
	}
	if user == nil {
		return ErrUserNotFound
	}

	err = n.NotificationRepository.MarkAllRead(
		workspaceId,
		userId,
		user.Role == db.UserRoleAdmin,
	)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("userId", userId.String()).
			Msg("Failed to mark all notifications read")
		return fmt.Errorf("%w: %v", ErrFailedUpdateNotification, err)
	}

	log.Info().
		Str("workspaceId", workspaceId.String()).
		Str("userId", userId.String()).
		Msg("All notifications marked read")

	return nil
}
