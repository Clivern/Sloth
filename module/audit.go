// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/clivern/cognit/db"
	"github.com/clivern/cognit/pkg/util"
)

var (
	ErrAuditEventNotFound    = errors.New("audit event not found")
	ErrFailedListAuditEvents = errors.New("failed list audit events")
	ErrFailedGetAuditEvent   = errors.New("failed get audit event")
)

// Audit is the module for workspace audit events.
type Audit struct {
	AuditRepository     db.AuditEventRepository
	WorkspaceRepository db.WorkspaceRepository
}

// NewAudit creates an audit module with the given repositories.
func NewAudit(audits db.AuditEventRepository, workspaces db.WorkspaceRepository) *Audit {
	return &Audit{
		AuditRepository:     audits,
		WorkspaceRepository: workspaces,
	}
}

// AuditEventResponse is an audit event shaped for API responses.
type AuditEventResponse struct {
	Id           db.Id           `json:"id"`
	WorkspaceId  db.Id           `json:"workspaceId"`
	UserId       *db.Id          `json:"userId,omitempty"`
	Action       string          `json:"action"`
	ResourceType *string         `json:"resourceType,omitempty"`
	ResourceId   *db.Id          `json:"resourceId,omitempty"`
	IPAddress    *string         `json:"ipAddress,omitempty"`
	UserAgent    *string         `json:"userAgent,omitempty"`
	Meta         json.RawMessage `json:"meta,omitempty"`
	CreatedAt    string          `json:"createdAt"`
}

// ListAuditEventsResponse is returned when listing audit events.
type ListAuditEventsResponse struct {
	Events []*AuditEventResponse
	Total  int64
}

// GetAuditEvent returns one audit event by id.
func (a *Audit) GetAuditEvent(workspaceId, auditId db.Id) (*AuditEventResponse, error) {
	workspace, err := a.WorkspaceRepository.GetById(workspaceId)
	if err != nil {
		return nil, err
	}
	if workspace == nil {
		return nil, ErrWorkspaceNotFound
	}

	event, err := a.AuditRepository.GetById(auditId)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedGetAuditEvent, err)
	}
	if event == nil || event.WorkspaceId != workspaceId {
		return nil, ErrAuditEventNotFound
	}

	return &AuditEventResponse{
		Id:           event.Id,
		WorkspaceId:  event.WorkspaceId,
		UserId:       event.UserId,
		Action:       event.Action,
		ResourceType: event.ResourceType,
		ResourceId:   event.ResourceId,
		IPAddress:    event.IPAddress,
		UserAgent:    event.UserAgent,
		Meta:         util.JSONRawFromString(event.Meta),
		CreatedAt:    event.CreatedAt.UTC().Format(time.RFC3339),
	}, nil
}

// ListAuditEvents returns paginated audit events for a workspace.
func (a *Audit) ListAuditEvents(workspaceId db.Id, limit, offset int) (*ListAuditEventsResponse, error) {
	workspace, err := a.WorkspaceRepository.GetById(workspaceId)
	if err != nil {
		return nil, err
	}
	if workspace == nil {
		return nil, ErrWorkspaceNotFound
	}

	total, err := a.AuditRepository.CountByWorkspaceId(workspaceId)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedListAuditEvents, err)
	}

	events, err := a.AuditRepository.ListByWorkspaceId(workspaceId, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedListAuditEvents, err)
	}

	list := make([]*AuditEventResponse, 0, len(events))
	for _, item := range events {
		list = append(list, &AuditEventResponse{
			Id:           item.Id,
			WorkspaceId:  item.WorkspaceId,
			UserId:       item.UserId,
			Action:       item.Action,
			ResourceType: item.ResourceType,
			ResourceId:   item.ResourceId,
			IPAddress:    item.IPAddress,
			UserAgent:    item.UserAgent,
			Meta:         util.JSONRawFromString(item.Meta),
			CreatedAt:    item.CreatedAt.UTC().Format(time.RFC3339),
		})
	}

	return &ListAuditEventsResponse{Events: list, Total: total}, nil
}
