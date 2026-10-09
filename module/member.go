// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"time"

	"github.com/clivern/cognit/db"
)

// WorkspaceMemberResponse is a workspace member shaped for API responses.
type WorkspaceMemberResponse struct {
	Id          db.Id  `json:"id"`
	WorkspaceId db.Id  `json:"workspaceId"`
	UserId      db.Id  `json:"userId"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	Role        string `json:"role"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

// UpdateWorkspaceMemberRoleRequest is the body for updating a workspace member role.
type UpdateWorkspaceMemberRoleRequest struct {
	Role string `json:"role" validate:"required,oneof=admin regular readonly" label:"Role"`
}

// ListWorkspaceMembersResult is what you get back when listing workspace members.
type ListWorkspaceMembersResult struct {
	Members []*WorkspaceMemberResponse
	Total   int64
}

// ListWorkspaceMembers returns members in a workspace with pagination.
func (w *Workspace) ListWorkspaceMembers(workspaceId db.Id, limit, offset int) (*ListWorkspaceMembersResult, error) {
	workspace, err := w.WorkspaceRepository.GetById(workspaceId)
	if err != nil {
		return nil, err
	}
	if workspace == nil {
		return nil, ErrWorkspaceNotFound
	}

	list, err := w.WorkspaceUserRepository.ListByWorkspaceId(
		workspaceId,
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}

	total, err := w.WorkspaceUserRepository.CountByWorkspaceId(workspaceId)
	if err != nil {
		return nil, err
	}

	members := make([]*WorkspaceMemberResponse, 0, len(list))
	for _, member := range list {
		user, err := w.UserRepository.GetById(member.UserId)
		if err != nil {
			return nil, err
		}
		if user == nil {
			continue
		}

		members = append(members, &WorkspaceMemberResponse{
			Id:          member.Id,
			WorkspaceId: member.WorkspaceId,
			UserId:      member.UserId,
			Name:        user.Name,
			Email:       user.Email,
			Role:        member.Role,
			CreatedAt:   member.CreatedAt.UTC().Format(time.RFC3339),
			UpdatedAt:   member.UpdatedAt.UTC().Format(time.RFC3339),
		})
	}

	return &ListWorkspaceMembersResult{Members: members, Total: total}, nil
}

// GetWorkspaceMember returns the membership for a user in a workspace, or nil if not a member. TODO: remove if not used in future.
func (w *Workspace) GetWorkspaceMember(workspaceId, userId db.Id) (*db.WorkspaceUser, error) {
	return w.WorkspaceUserRepository.GetByWorkspaceAndUser(workspaceId, userId)
}

// UpdateWorkspaceMemberRole updates the role of a member in a workspace.
func (w *Workspace) UpdateWorkspaceMemberRole(workspaceId, userId db.Id, role string) (*WorkspaceMemberResponse, error) {
	workspace, err := w.WorkspaceRepository.GetById(workspaceId)
	if err != nil {
		return nil, err
	}
	if workspace == nil {
		return nil, ErrWorkspaceNotFound
	}

	m, err := w.WorkspaceUserRepository.GetByWorkspaceAndUser(workspaceId, userId)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, ErrWorkspaceUserNotFound
	}

	m.Role = role

	err = w.WorkspaceUserRepository.Update(m)
	if err != nil {
		return nil, err
	}

	user, err := w.UserRepository.GetById(m.UserId)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrWorkspaceUserNotFound
	}

	return &WorkspaceMemberResponse{
		Id:          m.Id,
		WorkspaceId: m.WorkspaceId,
		UserId:      m.UserId,
		Name:        user.Name,
		Email:       user.Email,
		Role:        m.Role,
		CreatedAt:   m.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:   m.UpdatedAt.UTC().Format(time.RFC3339),
	}, nil
}

// DeleteWorkspaceMember removes a member from a workspace.
func (w *Workspace) DeleteWorkspaceMember(workspaceId, userId db.Id) error {
	workspace, err := w.WorkspaceRepository.GetById(workspaceId)
	if err != nil {
		return err
	}
	if workspace == nil {
		return ErrWorkspaceNotFound
	}

	m, err := w.WorkspaceUserRepository.GetByWorkspaceAndUser(workspaceId, userId)
	if err != nil {
		return err
	}
	if m == nil {
		return ErrWorkspaceUserNotFound
	}

	return w.WorkspaceUserRepository.Delete(m.Id)
}
