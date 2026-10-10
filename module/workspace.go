// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"errors"
	"time"

	"github.com/clivern/cognit/db"
	"github.com/clivern/cognit/pkg/util"
)

var (
	ErrWorkspaceNotFound     = errors.New("workspace not found")
	ErrWorkspaceUserNotFound = errors.New("workspace user not found")
	ErrWorkspaceAccessDenied = errors.New("workspace access denied")
	ErrWorkspaceCheckAccess  = errors.New("failed check workspace access")
)

// Workspace is the module for workspace CRUD and members.
type Workspace struct {
	WorkspaceRepository     db.WorkspaceRepository
	WorkspaceUserRepository db.WorkspaceUserRepository
	SubscriptionRepository  db.SubscriptionRepository
	UserRepository          db.UserRepository
}

// NewWorkspace creates a workspace module with the given repositories.
func NewWorkspace(workspaces db.WorkspaceRepository, workspaceUsers db.WorkspaceUserRepository, subscriptions db.SubscriptionRepository, users db.UserRepository) *Workspace {
	return &Workspace{
		WorkspaceRepository:     workspaces,
		WorkspaceUserRepository: workspaceUsers,
		SubscriptionRepository:  subscriptions,
		UserRepository:          users,
	}
}

// CreateWorkspaceRequest is the body for creating a workspace.
type CreateWorkspaceRequest struct {
	Name string `json:"name" validate:"required,min=3,max=60" label:"Name"`
}

// UpdateWorkspaceRequest is the body for updating a workspace.
type UpdateWorkspaceRequest struct {
	Name string `json:"name" validate:"required,min=3,max=60" label:"Name"`
}

// WorkspaceResponse is a workspace shaped for API responses.
type WorkspaceResponse struct {
	Id           db.Id  `json:"id"`
	Name         string `json:"name"`
	Handle       string `json:"handle"`
	MembersCount int64  `json:"count"`
	Role         string `json:"role"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
}

// ListWorkspacesResponse is returned when listing workspace responses.
type ListWorkspacesResponse struct {
	Workspaces []*WorkspaceResponse
	Total      int64
}

// CreateWorkspace creates a workspace and adds the user as owner.
func (w *Workspace) CreateWorkspace(req *CreateWorkspaceRequest, user *db.User) (*WorkspaceResponse, error) {
	var handle string
	for {
		h, err := util.RandomHandle(4, 8)
		if err != nil {
			return nil, err
		}

		existing, err := w.WorkspaceRepository.GetByHandle(h)
		if err != nil {
			return nil, err
		}
		if existing == nil {
			handle = h
			break
		}
	}

	workspace := &db.Workspace{
		Name:   req.Name,
		Handle: handle,
	}

	err := w.WorkspaceRepository.Create(workspace)
	if err != nil {
		return nil, err
	}

	err = w.WorkspaceUserRepository.Create(&db.WorkspaceUser{
		WorkspaceId: workspace.Id,
		UserId:      user.Id,
		Role:        db.UserRoleOwner,
	})
	if err != nil {
		return nil, err
	}

	err = w.SubscriptionRepository.Create(&db.Subscription{
		WorkspaceId: workspace.Id,
	})
	if err != nil {
		return nil, err
	}

	count, err := w.WorkspaceUserRepository.CountByWorkspaceId(workspace.Id)
	if err != nil {
		return nil, err
	}

	return &WorkspaceResponse{
		Id:           workspace.Id,
		Name:         workspace.Name,
		Handle:       workspace.Handle,
		MembersCount: count,
		Role:         db.UserRoleOwner,
		CreatedAt:    workspace.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:    workspace.UpdatedAt.UTC().Format(time.RFC3339),
	}, nil
}

// ListWorkspaces retrieves workspace responses visible to the user.
func (w *Workspace) ListWorkspaces(user *db.User, limit, offset int) (*ListWorkspacesResponse, error) {
	workspaces, err := w.WorkspaceRepository.List(limit, offset, user.Id)
	if err != nil {
		return nil, err
	}

	total, err := w.WorkspaceRepository.Count(user.Id)
	if err != nil {
		return nil, err
	}

	resp := make([]*WorkspaceResponse, 0, len(workspaces))
	for _, workspace := range workspaces {
		role := ""
		membership, err := w.WorkspaceRepository.GetWorkspaceMembership(workspace.Id, user.Id)
		if err != nil {
			return nil, err
		}
		if membership != nil {
			role = membership.Role
		}

		count, err := w.WorkspaceUserRepository.CountByWorkspaceId(workspace.Id)
		if err != nil {
			return nil, err
		}

		resp = append(resp, &WorkspaceResponse{
			Id:           workspace.Id,
			Name:         workspace.Name,
			Handle:       workspace.Handle,
			MembersCount: count,
			Role:         role,
			CreatedAt:    workspace.CreatedAt.UTC().Format(time.RFC3339),
			UpdatedAt:    workspace.UpdatedAt.UTC().Format(time.RFC3339),
		})
	}

	return &ListWorkspacesResponse{Workspaces: resp, Total: total}, nil
}

// GetWorkspace retrieves a workspace after checking the user's access.
func (w *Workspace) GetWorkspace(workspaceId db.Id, user *db.User) (*WorkspaceResponse, error) {
	workspace, err := w.WorkspaceRepository.GetById(workspaceId)
	if err != nil {
		return nil, err
	}
	if workspace == nil {
		return nil, ErrWorkspaceNotFound
	}

	role := ""
	membership, err := w.WorkspaceRepository.GetWorkspaceMembership(workspace.Id, user.Id)
	if err != nil {
		return nil, err
	}
	if membership != nil {
		role = membership.Role
	}

	count, err := w.WorkspaceUserRepository.CountByWorkspaceId(workspace.Id)
	if err != nil {
		return nil, err
	}

	return &WorkspaceResponse{
		Id:           workspace.Id,
		Name:         workspace.Name,
		Handle:       workspace.Handle,
		MembersCount: count,
		Role:         role,
		CreatedAt:    workspace.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:    workspace.UpdatedAt.UTC().Format(time.RFC3339),
	}, nil
}

// UpdateWorkspace updates a workspace after checking the user's access.
func (w *Workspace) UpdateWorkspace(workspaceId db.Id, req *UpdateWorkspaceRequest, user *db.User) (*WorkspaceResponse, error) {
	workspace, err := w.WorkspaceRepository.GetById(workspaceId)
	if err != nil {
		return nil, err
	}
	if workspace == nil {
		return nil, ErrWorkspaceNotFound
	}

	workspace.Name = req.Name
	err = w.WorkspaceRepository.Update(workspace)
	if err != nil {
		return nil, err
	}

	role := ""
	membership, err := w.WorkspaceRepository.GetWorkspaceMembership(workspace.Id, user.Id)
	if err != nil {
		return nil, err
	}
	if membership != nil {
		role = membership.Role
	}

	count, err := w.WorkspaceUserRepository.CountByWorkspaceId(workspace.Id)
	if err != nil {
		return nil, err
	}

	return &WorkspaceResponse{
		Id:           workspace.Id,
		Name:         workspace.Name,
		Handle:       workspace.Handle,
		MembersCount: count,
		Role:         role,
		CreatedAt:    workspace.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:    workspace.UpdatedAt.UTC().Format(time.RFC3339),
	}, nil
}

// DeleteWorkspace deletes a workspace.
func (w *Workspace) DeleteWorkspace(workspaceId db.Id) error {
	workspace, err := w.WorkspaceRepository.GetById(workspaceId)
	if err != nil {
		return err
	}
	if workspace == nil {
		return ErrWorkspaceNotFound
	}

	return w.WorkspaceRepository.Delete(workspaceId)
}
